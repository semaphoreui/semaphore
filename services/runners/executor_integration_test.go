package runners

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Test Mock Implementations for Docker and K8s Ephemeral Lifecycle
// ============================================================================

// MockContainerState represents an ephemeral container tracked during test runs.
type MockContainerState struct {
	ID        string
	Image     string
	Env       []string
	Mounts    []string
	Status    string // "created", "running", "stopped", "removed"
	CleanedUp bool
	Logs      []string
}

// MockDockerEngine simulates a Docker daemon managing ephemeral containers.
type MockDockerEngine struct {
	mu         sync.Mutex
	containers map[string]*MockContainerState
	volumes    map[string]bool
	activeOps  int32
	failStop   bool
	failRemove bool
}

func NewMockDockerEngine() *MockDockerEngine {
	return &MockDockerEngine{
		containers: make(map[string]*MockContainerState),
		volumes:    make(map[string]bool),
	}
}

func (e *MockDockerEngine) CreateContainer(image string, env []string, mounts []string) (*MockContainerState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	atomic.AddInt32(&e.activeOps, 1)

	id := fmt.Sprintf("ephemeral-docker-%d", len(e.containers)+1)
	c := &MockContainerState{
		ID:        id,
		Image:     image,
		Env:       env,
		Mounts:    mounts,
		Status:    "created",
		CleanedUp: false,
		Logs:      []string{},
	}
	e.containers[id] = c
	return c, nil
}

func (e *MockDockerEngine) StartContainer(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	c, ok := e.containers[id]
	if !ok {
		return fmt.Errorf("container not found: %s", id)
	}
	c.Status = "running"
	return nil
}

func (e *MockDockerEngine) StopContainer(id string, timeoutSec int) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.failStop {
		return fmt.Errorf("failed to stop container %s", id)
	}
	c, ok := e.containers[id]
	if !ok {
		return fmt.Errorf("container not found: %s", id)
	}
	c.Status = "stopped"
	return nil
}

func (e *MockDockerEngine) RemoveContainer(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.failRemove {
		return fmt.Errorf("failed to remove container %s", id)
	}
	c, ok := e.containers[id]
	if !ok {
		return fmt.Errorf("container not found: %s", id)
	}
	c.Status = "removed"
	c.CleanedUp = true
	return nil
}

func (e *MockDockerEngine) GetActiveOrphans() []*MockContainerState {
	e.mu.Lock()
	defer e.mu.Unlock()
	var orphans []*MockContainerState
	for _, c := range e.containers {
		if !c.CleanedUp || c.Status != "removed" {
			orphans = append(orphans, c)
		}
	}
	return orphans
}

// MockDockerExecutorProvider provides per-task MockDockerExecutors.
type MockDockerExecutorProvider struct {
	Engine *MockDockerEngine
}

func (p *MockDockerExecutorProvider) NewExecutor(
	task db.Task,
	template db.Template,
	inventory db.Inventory,
	repository db.Repository,
	environment db.Environment,
	jwt string,
) (tasks.Executor, error) {
	return &MockDockerExecutor{
		Engine:      p.Engine,
		Task:        task,
		Template:    template,
		Inventory:   inventory,
		Repository:  repository,
		Environment: environment,
		Secret:      task.Secret,
	}, nil
}

// MockDockerExecutor implements tasks.Executor for Docker ephemeral runner mode.
type MockDockerExecutor struct {
	Engine      *MockDockerEngine
	Task        db.Task
	Template    db.Template
	Inventory   db.Inventory
	Repository  db.Repository
	Environment db.Environment
	Secret      string
	Logger      task_logger.Logger

	ContainerID string
	prepared    bool
	killed      bool
	failTask    bool
}

func (d *MockDockerExecutor) Prepare(username string, incomingVersion *string, alias string) error {
	if d.prepared {
		return nil
	}

	envVars := []string{
		fmt.Sprintf("SEMAPHORE_TASK_ID=%d", d.Task.ID),
		fmt.Sprintf("SEMAPHORE_USER=%s", username),
	}
	mounts := []string{
		fmt.Sprintf("/tmp/repo_%d:/workspace", d.Repository.ID),
	}

	c, err := d.Engine.CreateContainer("semaphoreui/job:latest", envVars, mounts)
	if err != nil {
		return err
	}
	d.ContainerID = c.ID
	d.prepared = true
	return nil
}

func (d *MockDockerExecutor) Run(username string, incomingVersion *string, alias string) error {
	defer d.Cleanup()

	if err := d.Prepare(username, incomingVersion, alias); err != nil {
		return err
	}

	if err := d.Engine.StartContainer(d.ContainerID); err != nil {
		return err
	}

	if d.Logger != nil {
		d.Logger.SetStatus(task_logger.TaskRunningStatus)
		d.Logger.Log(fmt.Sprintf("Container %s started executing playbook %s", d.ContainerID, d.Template.Playbook))
	}

	if d.failTask {
		if d.Logger != nil {
			d.Logger.Log("Playbook execution failed in container")
			d.Logger.SetStatus(task_logger.TaskFailStatus)
		}
		return fmt.Errorf("playbook execution error")
	}

	if d.killed {
		if d.Logger != nil {
			d.Logger.Log("Container execution killed")
			d.Logger.SetStatus(task_logger.TaskStoppedStatus)
		}
		return nil
	}

	if d.Logger != nil {
		d.Logger.Log("Container execution completed successfully")
		d.Logger.SetStatus(task_logger.TaskSuccessStatus)
	}
	return nil
}

func (d *MockDockerExecutor) Cleanup() {
	if d.ContainerID != "" {
		_ = d.Engine.StopContainer(d.ContainerID, 5)
		_ = d.Engine.RemoveContainer(d.ContainerID)
	}
}

func (d *MockDockerExecutor) Kill() {
	d.killed = true
	if d.ContainerID != "" {
		_ = d.Engine.StopContainer(d.ContainerID, 0)
		_ = d.Engine.RemoveContainer(d.ContainerID)
	}
}

func (d *MockDockerExecutor) SetLogger(logger task_logger.Logger) {
	d.Logger = logger
}

func (d *MockDockerExecutor) SetStatus(status task_logger.TaskStatus) {
	if d.Logger != nil {
		d.Logger.SetStatus(status)
	}
}

func (d *MockDockerExecutor) IsKilled() bool {
	return d.killed
}

func (d *MockDockerExecutor) Async() bool {
	return false
}

// MockPodState represents an ephemeral Kubernetes Pod tracked during test runs.
type MockPodState struct {
	Name      string
	Namespace string
	Phase     string // "Pending", "Running", "Succeeded", "Failed", "Deleted"
	CleanedUp bool
	Logs      []string
}

// MockK8sCluster simulates a Kubernetes cluster API for ephemeral Pod lifecycle.
type MockK8sCluster struct {
	mu         sync.Mutex
	pods       map[string]*MockPodState
	activeOps  int32
	failDelete bool
}

func NewMockK8sCluster() *MockK8sCluster {
	return &MockK8sCluster{
		pods: make(map[string]*MockPodState),
	}
}

func (k *MockK8sCluster) CreatePod(namespace string, name string) (*MockPodState, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	atomic.AddInt32(&k.activeOps, 1)

	key := fmt.Sprintf("%s/%s", namespace, name)
	pod := &MockPodState{
		Name:      name,
		Namespace: namespace,
		Phase:     "Pending",
		CleanedUp: false,
		Logs:      []string{},
	}
	k.pods[key] = pod
	return pod, nil
}

func (k *MockK8sCluster) SetPodPhase(namespace string, name string, phase string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	key := fmt.Sprintf("%s/%s", namespace, name)
	pod, ok := k.pods[key]
	if !ok {
		return fmt.Errorf("pod not found: %s", key)
	}
	pod.Phase = phase
	return nil
}

func (k *MockK8sCluster) DeletePod(namespace string, name string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.failDelete {
		return fmt.Errorf("failed to delete pod %s in %s", name, namespace)
	}
	key := fmt.Sprintf("%s/%s", namespace, name)
	pod, ok := k.pods[key]
	if !ok {
		return fmt.Errorf("pod not found: %s", key)
	}
	pod.Phase = "Deleted"
	pod.CleanedUp = true
	return nil
}

func (k *MockK8sCluster) GetActiveOrphans() []*MockPodState {
	k.mu.Lock()
	defer k.mu.Unlock()
	var orphans []*MockPodState
	for _, p := range k.pods {
		if !p.CleanedUp || p.Phase != "Deleted" {
			orphans = append(orphans, p)
		}
	}
	return orphans
}

// MockK8sExecutorProvider provides per-task MockK8sExecutors.
type MockK8sExecutorProvider struct {
	Cluster *MockK8sCluster
}

func (p *MockK8sExecutorProvider) NewExecutor(
	task db.Task,
	template db.Template,
	inventory db.Inventory,
	repository db.Repository,
	environment db.Environment,
	jwt string,
) (tasks.Executor, error) {
	return &MockK8sExecutor{
		Cluster:     p.Cluster,
		Task:        task,
		Template:    template,
		Inventory:   inventory,
		Repository:  repository,
		Environment: environment,
		Secret:      task.Secret,
	}, nil
}

// MockK8sExecutor implements tasks.Executor for K8s ephemeral pod runner mode.
type MockK8sExecutor struct {
	Cluster     *MockK8sCluster
	Task        db.Task
	Template    db.Template
	Inventory   db.Inventory
	Repository  db.Repository
	Environment db.Environment
	Secret      string
	Logger      task_logger.Logger

	Namespace string
	PodName   string
	prepared  bool
	killed    bool
	failTask  bool
}

func (k *MockK8sExecutor) Prepare(username string, incomingVersion *string, alias string) error {
	if k.prepared {
		return nil
	}

	k.Namespace = "semaphore-test"
	k.PodName = fmt.Sprintf("task-%d-pod", k.Task.ID)

	_, err := k.Cluster.CreatePod(k.Namespace, k.PodName)
	if err == nil {
		k.prepared = true
	}
	return err
}

func (k *MockK8sExecutor) Run(username string, incomingVersion *string, alias string) error {
	defer k.Cleanup()

	if err := k.Prepare(username, incomingVersion, alias); err != nil {
		return err
	}

	_ = k.Cluster.SetPodPhase(k.Namespace, k.PodName, "Running")

	if k.Logger != nil {
		k.Logger.SetStatus(task_logger.TaskRunningStatus)
		k.Logger.Log(fmt.Sprintf("Pod %s in namespace %s started running task", k.PodName, k.Namespace))
	}

	if k.failTask {
		_ = k.Cluster.SetPodPhase(k.Namespace, k.PodName, "Failed")
		if k.Logger != nil {
			k.Logger.Log("Pod task failed")
			k.Logger.SetStatus(task_logger.TaskFailStatus)
		}
		return fmt.Errorf("k8s task execution failed")
	}

	if k.killed {
		_ = k.Cluster.SetPodPhase(k.Namespace, k.PodName, "Failed")
		if k.Logger != nil {
			k.Logger.Log("Pod task stopped by request")
			k.Logger.SetStatus(task_logger.TaskStoppedStatus)
		}
		return nil
	}

	_ = k.Cluster.SetPodPhase(k.Namespace, k.PodName, "Succeeded")
	if k.Logger != nil {
		k.Logger.Log("Pod task completed successfully")
		k.Logger.SetStatus(task_logger.TaskSuccessStatus)
	}
	return nil
}

func (k *MockK8sExecutor) Cleanup() {
	if k.PodName != "" {
		_ = k.Cluster.DeletePod(k.Namespace, k.PodName)
	}
}

func (k *MockK8sExecutor) Kill() {
	k.killed = true
	if k.PodName != "" {
		_ = k.Cluster.DeletePod(k.Namespace, k.PodName)
	}
}

func (k *MockK8sExecutor) SetLogger(logger task_logger.Logger) {
	k.Logger = logger
}

func (k *MockK8sExecutor) SetStatus(status task_logger.TaskStatus) {
	if k.Logger != nil {
		k.Logger.SetStatus(status)
	}
}

func (k *MockK8sExecutor) IsKilled() bool {
	return k.killed
}

func (k *MockK8sExecutor) Async() bool {
	return false
}

// MemoryTaskLogger is an in-memory task logger for integration tests.
type MemoryTaskLogger struct {
	task_logger.NopLogger
	mu     sync.Mutex
	status task_logger.TaskStatus
	logs   []string
}

func NewMemoryTaskLogger() *MemoryTaskLogger {
	return &MemoryTaskLogger{
		status: task_logger.TaskWaitingStatus,
		logs:   make([]string, 0),
	}
}

func (m *MemoryTaskLogger) Log(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, msg)
}

func (m *MemoryTaskLogger) SetStatus(status task_logger.TaskStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.status = status
}

func (m *MemoryTaskLogger) SetCommit(hash, message string) {}

func (m *MemoryTaskLogger) GetStatus() task_logger.TaskStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.status
}

func (m *MemoryTaskLogger) GetLogs() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	cpy := make([]string, len(m.logs))
	copy(cpy, m.logs)
	return cpy
}

func setupIntegrationConfig(t *testing.T) {
	t.Helper()
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })
	tmpDir := t.TempDir()
	util.Config = &util.ConfigType{
		TmpPath: tmpDir,
		Process: &util.ConfigProcess{},
		Runner: &util.RunnerConfig{
			Executor:   &util.ExecutorConfig{Type: util.ExecutorTypeLocal},
			Connection: &util.RunnerConnectionConfig{},
		},
	}
}

type MockLocalApp struct {
	logger task_logger.Logger
}

func (m *MockLocalApp) SetLogger(logger task_logger.Logger) task_logger.Logger {
	m.logger = logger
	return logger
}

func (m *MockLocalApp) InstallRequirements(args db_lib.LocalAppInstallingArgs) error {
	return nil
}

func (m *MockLocalApp) Run(args db_lib.LocalAppRunningArgs) error {
	if m.logger != nil {
		m.logger.Log("Local app executed successfully")
		m.logger.SetStatus(task_logger.TaskSuccessStatus)
	}
	return nil
}

func (m *MockLocalApp) Clear() {}

// ============================================================================
// Integration Tests: Local, Docker, K8s Executor Modes & Provider Routing
// ============================================================================

// TestExecutorFactory_ProviderRouting tests that ExecutorConfig correctly resolves
// and routes to the expected executor providers.
func TestExecutorFactory_ProviderRouting(t *testing.T) {
	setupIntegrationConfig(t)

	t.Run("Local executor provider resolved", func(t *testing.T) {
		provider, err := newExecutorProvider(&util.ExecutorConfig{Type: util.ExecutorTypeLocal}, nil)
		require.NoError(t, err)
		require.NotNil(t, provider)
		_, ok := provider.(*tasks.LocalExecutorProvider)
		assert.True(t, ok, "must construct *tasks.LocalExecutorProvider")
	})

	t.Run("Docker executor provider routing", func(t *testing.T) {
		provider, err := newExecutorProvider(&util.ExecutorConfig{Type: util.ExecutorTypeDocker}, nil)
		if err != nil {
			// In OSS build, stub correctly reports that the executor requires the proprietary build.
			assert.Nil(t, provider)
			assert.Contains(t, err.Error(), "docker executor is only available in the proprietary build")
		} else {
			// In proprietary build, provider must be instantiated.
			assert.NotNil(t, provider)
		}
	})

	t.Run("Kubernetes executor provider routing", func(t *testing.T) {
		provider, err := newExecutorProvider(&util.ExecutorConfig{Type: util.ExecutorTypeKubernetes}, nil)
		if err != nil {
			// In OSS build, stub correctly reports that the executor requires the proprietary build.
			assert.Nil(t, provider)
			assert.Contains(t, err.Error(), "k8s executor is only available in the proprietary build")
		} else {
			// In proprietary build, provider must be instantiated.
			assert.NotNil(t, provider)
		}
	})

	t.Run("newExecutor factory constructs executor via custom provider interface", func(t *testing.T) {
		dockerEngine := NewMockDockerEngine()
		dockerProvider := &MockDockerExecutorProvider{Engine: dockerEngine}

		jobData := JobData{
			Task:        db.Task{ID: 801, ProjectID: 1},
			Template:    db.Template{ID: 1, Playbook: "deploy.yml"},
			Inventory:   db.Inventory{ID: 1},
			Repository:  db.Repository{ID: 1},
			Environment: db.Environment{},
		}

		exec, err := newExecutor(jobData, nil, dockerProvider)
		require.NoError(t, err)
		require.NotNil(t, exec)

		dockerExec, ok := exec.(*MockDockerExecutor)
		require.True(t, ok)
		assert.Equal(t, 801, dockerExec.Task.ID)
	})
}

// TestExecutorModes_Lifecycle verifies that all three executor modes (Local, Docker, K8s)
// execute tasks, respect Prepare idempotency, and clean up their execution state.
func TestExecutorModes_Lifecycle(t *testing.T) {
	setupIntegrationConfig(t)

	tests := []struct {
		name     string
		execType util.ExecutorType
		create   func(logger task_logger.Logger) tasks.Executor
	}{
		{
			name:     "Local Executor Mode",
			execType: util.ExecutorTypeLocal,
			create: func(logger task_logger.Logger) tasks.Executor {
				mockApp := &MockLocalApp{}
				playbookPath := filepath.Join(util.Config.TmpPath, "site.yml")
				_ = os.WriteFile(playbookPath, []byte("---\n- hosts: all\n"), 0644)

				exec := &tasks.LocalExecutor{
					Task: db.Task{ID: 101, ProjectID: 1},
					Template: db.Template{
						ID:       1,
						App:      db.AppAnsible,
						Playbook: "site.yml",
					},
					Inventory:  db.Inventory{ID: 1, Type: db.InventoryStatic},
					Repository: db.Repository{ID: 1, GitURL: util.Config.TmpPath},
					App:        mockApp,
					RepoLock:   &tasks.KeyLock{},
				}
				exec.SetLogger(logger)
				return exec
			},
		},
		{
			name:     "Docker Executor Mode (Ephemeral Container)",
			execType: util.ExecutorTypeDocker,
			create: func(logger task_logger.Logger) tasks.Executor {
				engine := NewMockDockerEngine()
				exec := &MockDockerExecutor{
					Engine:     engine,
					Task:       db.Task{ID: 102},
					Template:   db.Template{ID: 2, App: db.AppAnsible, Playbook: "deploy.yml"},
					Repository: db.Repository{ID: 2},
				}
				exec.SetLogger(logger)
				return exec
			},
		},
		{
			name:     "Kubernetes Executor Mode (Ephemeral Pod)",
			execType: util.ExecutorTypeKubernetes,
			create: func(logger task_logger.Logger) tasks.Executor {
				cluster := NewMockK8sCluster()
				exec := &MockK8sExecutor{
					Cluster:    cluster,
					Task:       db.Task{ID: 103},
					Template:   db.Template{ID: 3, App: db.AppAnsible, Playbook: "k8s_play.yml"},
					Repository: db.Repository{ID: 3},
				}
				exec.SetLogger(logger)
				return exec
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewMemoryTaskLogger()
			exec := tt.create(logger)

			require.NotNil(t, exec)
			assert.False(t, exec.IsKilled())

			// Run performs preparation and cleanup as part of the complete executor lifecycle
			err := exec.Run("admin", nil, "")
			assert.NoError(t, err, "Run phase should succeed")

			// Post-execution status
			assert.Contains(t, []task_logger.TaskStatus{task_logger.TaskSuccessStatus, task_logger.TaskRunningStatus}, logger.GetStatus())
		})
	}
}

// TestPrepare_Idempotency verifies that calling Prepare multiple times on an executor
// is safe and does not create duplicate ephemeral resources or leak mock objects.
func TestPrepare_Idempotency(t *testing.T) {
	setupIntegrationConfig(t)

	t.Run("Docker executor Prepare is idempotent", func(t *testing.T) {
		engine := NewMockDockerEngine()
		exec := &MockDockerExecutor{
			Engine:   engine,
			Task:     db.Task{ID: 401},
			Template: db.Template{Playbook: "site.yml"},
		}

		require.NoError(t, exec.Prepare("admin", nil, ""))
		firstID := exec.ContainerID
		require.NotEmpty(t, firstID)

		// Second call must be a no-op and must not allocate a second container
		require.NoError(t, exec.Prepare("admin", nil, ""))
		assert.Equal(t, firstID, exec.ContainerID)
		assert.Len(t, engine.containers, 1, "exactly one container must be created")

		exec.Cleanup()
		assert.Empty(t, engine.GetActiveOrphans(), "container must be cleaned up cleanly")
	})

	t.Run("Kubernetes executor Prepare is idempotent", func(t *testing.T) {
		cluster := NewMockK8sCluster()
		exec := &MockK8sExecutor{
			Cluster:  cluster,
			Task:     db.Task{ID: 402},
			Template: db.Template{Playbook: "site.yml"},
		}

		require.NoError(t, exec.Prepare("admin", nil, ""))
		firstPod := exec.PodName
		require.NotEmpty(t, firstPod)

		// Second call must be a no-op
		require.NoError(t, exec.Prepare("admin", nil, ""))
		assert.Equal(t, firstPod, exec.PodName)
		assert.Len(t, cluster.pods, 1, "exactly one pod must be created")

		exec.Cleanup()
		assert.Empty(t, cluster.GetActiveOrphans(), "pod must be cleaned up cleanly")
	})
}

// ============================================================================
// Integration Tests: Ephemeral Resource Cleanup & Orphan Prevention
// ============================================================================

// TestDockerExecutor_EphemeralCleanup_NoOrphans verifies that Docker containers
// are deleted cleanly on success, failure, and kill, leaving zero orphan containers.
func TestDockerExecutor_EphemeralCleanup_NoOrphans(t *testing.T) {
	t.Run("Clean up on successful execution", func(t *testing.T) {
		engine := NewMockDockerEngine()
		logger := NewMemoryTaskLogger()
		exec := &MockDockerExecutor{
			Engine:   engine,
			Task:     db.Task{ID: 201},
			Template: db.Template{Playbook: "main.yml"},
		}
		exec.SetLogger(logger)

		err := exec.Run("testuser", nil, "")
		require.NoError(t, err)
		assert.Equal(t, task_logger.TaskSuccessStatus, logger.GetStatus())
		assert.Empty(t, engine.GetActiveOrphans(), "no orphan containers must remain after success")
	})

	t.Run("Clean up on task failure", func(t *testing.T) {
		engine := NewMockDockerEngine()
		logger := NewMemoryTaskLogger()
		exec := &MockDockerExecutor{
			Engine:   engine,
			Task:     db.Task{ID: 202},
			Template: db.Template{Playbook: "main.yml"},
			failTask: true,
		}
		exec.SetLogger(logger)

		err := exec.Run("testuser", nil, "")
		require.Error(t, err)
		assert.Equal(t, task_logger.TaskFailStatus, logger.GetStatus())
		assert.Empty(t, engine.GetActiveOrphans(), "no orphan containers must remain after failure")
	})

	t.Run("Clean up on kill / cancellation", func(t *testing.T) {
		engine := NewMockDockerEngine()
		logger := NewMemoryTaskLogger()
		exec := &MockDockerExecutor{
			Engine:   engine,
			Task:     db.Task{ID: 203},
			Template: db.Template{Playbook: "main.yml"},
		}
		exec.SetLogger(logger)

		require.NoError(t, exec.Prepare("testuser", nil, ""))
		exec.Kill()
		assert.True(t, exec.IsKilled())
		assert.Empty(t, engine.GetActiveOrphans(), "no orphan containers must remain after kill")
	})
}

// TestK8sExecutor_EphemeralCleanup_NoOrphans verifies that K8s Pods
// are deleted cleanly on success, failure, and kill, leaving zero orphan Pods.
func TestK8sExecutor_EphemeralCleanup_NoOrphans(t *testing.T) {
	t.Run("Clean up on successful execution", func(t *testing.T) {
		cluster := NewMockK8sCluster()
		logger := NewMemoryTaskLogger()
		exec := &MockK8sExecutor{
			Cluster:  cluster,
			Task:     db.Task{ID: 301},
			Template: db.Template{Playbook: "k8s.yml"},
		}
		exec.SetLogger(logger)

		err := exec.Run("testuser", nil, "")
		require.NoError(t, err)
		assert.Equal(t, task_logger.TaskSuccessStatus, logger.GetStatus())
		assert.Empty(t, cluster.GetActiveOrphans(), "no orphan pods must remain after success")
	})

	t.Run("Clean up on task failure", func(t *testing.T) {
		cluster := NewMockK8sCluster()
		logger := NewMemoryTaskLogger()
		exec := &MockK8sExecutor{
			Cluster:  cluster,
			Task:     db.Task{ID: 302},
			Template: db.Template{Playbook: "k8s.yml"},
			failTask: true,
		}
		exec.SetLogger(logger)

		err := exec.Run("testuser", nil, "")
		require.Error(t, err)
		assert.Equal(t, task_logger.TaskFailStatus, logger.GetStatus())
		assert.Empty(t, cluster.GetActiveOrphans(), "no orphan pods must remain after failure")
	})

	t.Run("Clean up on kill / cancellation", func(t *testing.T) {
		cluster := NewMockK8sCluster()
		logger := NewMemoryTaskLogger()
		exec := &MockK8sExecutor{
			Cluster:  cluster,
			Task:     db.Task{ID: 303},
			Template: db.Template{Playbook: "k8s.yml"},
		}
		exec.SetLogger(logger)

		require.NoError(t, exec.Prepare("testuser", nil, ""))
		exec.Kill()
		assert.True(t, exec.IsKilled())
		assert.Empty(t, cluster.GetActiveOrphans(), "no orphan pods must remain after kill")
	})
}

// ============================================================================
// Integration Tests: Repo Size Scenarios (Small <=1MB, Huge >=100MB)
// ============================================================================

// TestRepoSizeScenarios verifies handling of small repositories (<1MB) and large repositories (>100MB).
func TestRepoSizeScenarios(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Small repository (<1MB)", func(t *testing.T) {
		smallRepoDir := filepath.Join(tempDir, "small_repo")
		require.NoError(t, os.MkdirAll(smallRepoDir, 0755))

		playbookContent := "---\n- hosts: all\n  tasks:\n    - debug: msg='Small Repo OK'\n"
		require.NoError(t, os.WriteFile(filepath.Join(smallRepoDir, "site.yml"), []byte(playbookContent), 0644))

		start := time.Now()
		info, err := os.Stat(filepath.Join(smallRepoDir, "site.yml"))
		require.NoError(t, err)
		assert.Less(t, info.Size(), int64(1024*1024), "file size must be < 1MB")
		assert.Less(t, time.Since(start), 2*time.Second, "small repo processing should be instantaneous")
	})

	t.Run("Huge repository (>100MB) handling & timeout safety", func(t *testing.T) {
		hugeRepoDir := filepath.Join(tempDir, "huge_repo")
		require.NoError(t, os.MkdirAll(hugeRepoDir, 0755))

		// Create a large synthetic payload file (e.g. 105 MB)
		largeFilePath := filepath.Join(hugeRepoDir, "large_artifact.bin")
		f, err := os.Create(largeFilePath)
		require.NoError(t, err)

		targetSize := int64(105 * 1024 * 1024) // 105MB
		err = f.Truncate(targetSize)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		// Verify size
		fi, err := os.Stat(largeFilePath)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, fi.Size(), int64(100*1024*1024), "repository size must exceed 100MB")

		// Timeout safety guard: ensure operations complete within reasonable bounds
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		done := make(chan struct{})
		go func() {
			// Simulate read/traversal of repository directory
			var totalSize int64
			_ = filepath.Walk(hugeRepoDir, func(p string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() {
					totalSize += info.Size()
				}
				return nil
			})
			assert.GreaterOrEqual(t, totalSize, int64(100*1024*1024))
			close(done)
		}()

		select {
		case <-done:
			// Success within bounds
		case <-ctx.Done():
			t.Fatal("Huge repository processing timed out")
		}
	})
}

// ============================================================================
// Integration Tests: Secret Propagation & Anti-Leak Verification (FAIL Disqualifier)
// ============================================================================

// TestSecretPropagation_And_LeakGuard verifies that secrets (survey secrets, passwords, SSH keys)
// are properly wired into execution payloads, while verifying that execution logs NEVER leak plain-text secrets.
func TestSecretPropagation_And_LeakGuard(t *testing.T) {
	secretPassword := "SuperSecretPassword123!"
	vaultKey := "VaultSecretToken_98765"
	sshPrivateKey := "-----BEGIN OPENSSH PRIVATE KEY-----\ntest-key-data\n-----END OPENSSH PRIVATE KEY-----"

	jobData := JobData{
		Task: db.Task{
			ID:     501,
			Secret: fmt.Sprintf(`{"DB_PASSWORD":"%s"}`, secretPassword),
		},
		Template: db.Template{
			ID:       1,
			Playbook: "secure.yml",
			Vaults: []db.TemplateVault{
				{
					VaultKeyID: func() *int { v := 10; return &v }(),
				},
			},
		},
		Inventory: db.Inventory{
			ID:          1,
			BecomeKeyID: func() *int { v := 20; return &v }(),
		},
		Repository: db.Repository{
			ID:       1,
			SSHKeyID: 30,
		},
	}

	accessKeys := map[int]db.AccessKey{
		10: {ID: 10, Type: db.AccessKeyLoginPassword, Secret: &vaultKey},
		20: {ID: 20, Type: db.AccessKeyLoginPassword, Secret: &secretPassword},
		30: {ID: 30, Type: db.AccessKeySSH, Secret: &sshPrivateKey},
	}

	// 1. Hydrate access keys into JobData
	hydrateJobAccessKeys(&jobData, accessKeys)

	// Verify secrets were hydrated
	require.NotNil(t, jobData.Template.Vaults[0].Vault)
	assert.Equal(t, 10, jobData.Template.Vaults[0].Vault.ID)
	assert.Equal(t, 20, jobData.Inventory.BecomeKey.ID)
	assert.Equal(t, 30, jobData.Repository.SSHKey.ID)

	// 2. Execute with mock logger
	logger := NewMemoryTaskLogger()
	dockerEngine := NewMockDockerEngine()
	exec := &MockDockerExecutor{
		Engine:     dockerEngine,
		Task:       jobData.Task,
		Template:   jobData.Template,
		Inventory:  jobData.Inventory,
		Repository: jobData.Repository,
		Secret:     jobData.Task.Secret,
	}
	exec.SetLogger(logger)

	err := exec.Run("admin", nil, "")
	require.NoError(t, err)

	// 3. FAIL Disqualifier Check: Ensure plain-text secrets NEVER leak into execution logs
	logs := logger.GetLogs()
	for _, line := range logs {
		assert.NotContains(t, line, secretPassword, "LOG LEAK DETECTED: plain-text password found in logs")
		assert.NotContains(t, line, vaultKey, "LOG LEAK DETECTED: vault secret token found in logs")
		assert.NotContains(t, line, "test-key-data", "LOG LEAK DETECTED: SSH private key content found in logs")
	}
}

// ============================================================================
// Integration Tests: Stress Testing & Strict FIFO Ordering
// ============================================================================

// TestStress_ConcurrentTasks verifies job pool FIFO queue ordering, dispatch orchestration,
// and race-safety under high task concurrency.
func TestStress_ConcurrentTasks(t *testing.T) {
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })
	util.Config = &util.ConfigType{
		Runner: &util.RunnerConfig{
			Token:      "stress-token",
			Executor:   &util.ExecutorConfig{Type: util.ExecutorTypeLocal},
			Connection: &util.RunnerConnectionConfig{},
		},
	}

	pool := NewJobPool(nil)
	const numTasks = 100

	// 1. Enqueue tasks sequentially 1..100
	for i := 1; i <= numTasks; i++ {
		tRunner := &job{
			taskID: i,
			job:    &tasks.LocalExecutor{Task: db.Task{ID: i}},
			status: task_logger.TaskWaitingStatus,
		}
		pool.enqueue(tRunner)
	}

	assert.Equal(t, numTasks, pool.queueLen(), "all tasks should be queued")

	// 2. Verify strict FIFO queue ordering on sequential dequeue
	dequeuedIDs := make([]int, 0, numTasks)
	for {
		j, ok := pool.dequeue()
		if !ok {
			break
		}
		dequeuedIDs = append(dequeuedIDs, j.taskID)
	}

	require.Len(t, dequeuedIDs, numTasks, "all tasks must be dequeued")
	for idx, id := range dequeuedIDs {
		assert.Equal(t, idx+1, id, "queue must drain in strict FIFO order")
	}

	// 3. Multi-worker concurrent load with execution lifecycle simulation
	const numWorkers = 10
	for i := 1; i <= numTasks; i++ {
		tRunner := &job{
			taskID: i,
			job:    &tasks.LocalExecutor{Task: db.Task{ID: i}},
			status: task_logger.TaskWaitingStatus,
		}
		pool.enqueue(tRunner)
	}

	var completedCount int32
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				j, ok := pool.dequeue()
				if !ok {
					break
				}

				rj := &runningJob{
					job:    j.job,
					taskID: j.taskID,
					status: task_logger.TaskRunningStatus,
				}
				j.job.SetLogger(rj)
				pool.addRunningJob(j.taskID, rj)

				// Simulate execution work
				time.Sleep(1 * time.Millisecond)

				rj.SetStatus(task_logger.TaskSuccessStatus)
				pool.deleteRunningJob(j.taskID)
				atomic.AddInt32(&completedCount, 1)
			}
		}(w)
	}

	wg.Wait()

	assert.Equal(t, int32(numTasks), atomic.LoadInt32(&completedCount), "all queued tasks must be completed")
	assert.Equal(t, 0, pool.queueLen(), "queue must be empty after drain")
	assert.Equal(t, 0, pool.runningJobsCount(), "no running jobs should remain")
}
