package k8s

import (
	"context"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// init ensures util.Config is non-nil for K8s tests — production code reads
// util.Config.WebHost (Task.GetUrl) eagerly when building extra-vars JSON, and an
// uninitialized global would NPE the whole suite. A blank Config is enough: GetUrl
// returns nil when WebHost is empty, which is fine for the assertions we make.
func init() {
	if util.Config == nil {
		util.Config = &util.ConfigType{}
	}
}

// TestExecutorImplementsTasksExecutor is a compile-time guard: if Executor stops
// satisfying tasks.Executor the K8s executor can no longer be plugged into the runner
// job pool. Documents the contract introduced in Phase 2.
func TestExecutorImplementsTasksExecutor(t *testing.T) {
	var _ tasks.Executor = (*Executor)(nil)
}

func newTestConfig() Config {
	return Config{
		Clientset:      fake.NewSimpleClientset(),
		Namespace:      "semaphore-test",
		Image:          "alpine:latest",
		HelperImage:    "alpine/git:latest",
		ServiceAccount: "default",
		PollInterval:   10 * time.Millisecond,
		CleanupGrace:   5 * time.Second,
	}
}

func TestPrepare_CreatesPodWithExpectedSpec(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 42}, db.Template{ID: 7}, db.Inventory{}, db.Repository{}, db.Environment{})

	err := exec.Prepare("alice", nil, "")
	require.NoError(t, err)
	require.NotEmpty(t, exec.podName, "Prepare must generate a pod name")

	pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	require.NoError(t, err)

	assert.Equal(t, cfg.Namespace, pod.Namespace)
	assert.Contains(t, pod.Name, "semaphore-job-42-", "pod name follows semaphore-job-<task_id>-<rand> shape")
	assert.Equal(t, "42", pod.Labels[LabelTaskID])
	assert.Equal(t, "semaphore", pod.Labels[LabelRunner])

	require.Len(t, pod.Spec.Containers, 1, "skeleton pod has exactly one container")
	assert.Equal(t, buildContainerName, pod.Spec.Containers[0].Name)
	assert.Equal(t, cfg.Image, pod.Spec.Containers[0].Image)
	assert.Equal(t, cfg.ServiceAccount, pod.Spec.ServiceAccountName)
	assert.Equal(t, "Never", string(pod.Spec.RestartPolicy))
}

func TestPrepare_HonorsPullSecrets(t *testing.T) {
	cfg := newTestConfig()
	cfg.PullSecrets = []string{"private-registry", "second-registry"}

	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{}, db.Repository{}, db.Environment{})
	require.NoError(t, exec.Prepare("", nil, ""))

	pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	require.NoError(t, err)

	require.Len(t, pod.Spec.ImagePullSecrets, 2)
	assert.Equal(t, "private-registry", pod.Spec.ImagePullSecrets[0].Name)
	assert.Equal(t, "second-registry", pod.Spec.ImagePullSecrets[1].Name)
}

func TestPrepare_Idempotent(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 5}, db.Template{}, db.Inventory{}, db.Repository{}, db.Environment{})

	require.NoError(t, exec.Prepare("", nil, ""))
	firstName := exec.podName

	require.NoError(t, exec.Prepare("", nil, ""), "second Prepare must be a no-op, not an error")
	assert.Equal(t, firstName, exec.podName, "second Prepare must not change pod name")

	// Confirm only one Pod exists.
	pods, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).List(context.Background(), metav1.ListOptions{})
	require.NoError(t, err)
	assert.Len(t, pods.Items, 1)
}

func TestPrepare_RequiresClientset(t *testing.T) {
	exec := New(Config{}, db.Task{ID: 1}, db.Template{}, db.Inventory{}, db.Repository{}, db.Environment{})

	err := exec.Prepare("", nil, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "clientset")
}

func TestCleanup_DeletesPod(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 9}, db.Template{}, db.Inventory{}, db.Repository{}, db.Environment{})
	require.NoError(t, exec.Prepare("", nil, ""))

	podName := exec.podName
	exec.Cleanup()

	_, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), podName, metav1.GetOptions{})
	assert.True(t, apierrors.IsNotFound(err), "Cleanup must delete the Pod, got err=%v", err)
}

func TestCleanup_NoopWhenUnprepared(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 9}, db.Template{}, db.Inventory{}, db.Repository{}, db.Environment{})

	// Calling Cleanup before Prepare must not panic and must not touch the API server.
	assert.NotPanics(t, func() { exec.Cleanup() })

	pods, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).List(context.Background(), metav1.ListOptions{})
	require.NoError(t, err)
	assert.Empty(t, pods.Items)
}

func TestKill_SetsKilledAndCancelsRun(t *testing.T) {
	exec := New(newTestConfig(), db.Task{ID: 1}, db.Template{}, db.Inventory{}, db.Repository{}, db.Environment{})
	assert.False(t, exec.IsKilled())

	// Inject a cancel function the way Run would; verify Kill triggers it.
	called := false
	exec.mu.Lock()
	exec.cancelRun = func() { called = true }
	exec.mu.Unlock()

	exec.Kill()

	assert.True(t, exec.IsKilled(), "Kill must flip the killed flag")
	assert.True(t, called, "Kill must invoke the in-flight cancel function")
}

func TestGeneratePodName_Format(t *testing.T) {
	name, err := generatePodName(42)
	require.NoError(t, err)
	assert.Contains(t, name, "semaphore-job-42-")
	assert.Len(t, name, len("semaphore-job-42-")+6, "6 hex chars suffix")
}

// --- Phase 3: workspace volume + git-clone init container -----------------------

func TestPrepare_AttachesWorkspaceVolume(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{},
		db.Repository{GitURL: "https://example.com/repo.git", GitBranch: "main"},
		db.Environment{})

	require.NoError(t, exec.Prepare("", nil, ""))

	pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	require.NoError(t, err)

	require.Len(t, pod.Spec.Volumes, 1, "skeleton pod must declare exactly the workspace volume")
	assert.Equal(t, workspaceVolumeName, pod.Spec.Volumes[0].Name)
	require.NotNil(t, pod.Spec.Volumes[0].EmptyDir, "workspace is an emptyDir")

	require.Len(t, pod.Spec.Containers[0].VolumeMounts, 1)
	assert.Equal(t, workspaceVolumeName, pod.Spec.Containers[0].VolumeMounts[0].Name)
	assert.Equal(t, "/workspace", pod.Spec.Containers[0].VolumeMounts[0].MountPath)
	assert.Equal(t, "/workspace/repo", pod.Spec.Containers[0].WorkingDir,
		"build container starts inside the cloned repo")
}

func TestPrepare_AddsGitCloneInitContainerForRemoteRepo(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{},
		db.Repository{GitURL: "https://example.com/playbooks.git", GitBranch: "develop"},
		db.Environment{})

	require.NoError(t, exec.Prepare("", nil, ""))

	pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	require.NoError(t, err)

	require.Len(t, pod.Spec.InitContainers, 1, "remote HTTPS repo must trigger a git-clone init container")
	init := pod.Spec.InitContainers[0]
	assert.Equal(t, gitCloneInitContainerName, init.Name)
	assert.Equal(t, cfg.HelperImage, init.Image)
	require.Len(t, init.Args, 1)
	assert.Contains(t, init.Args[0], "https://example.com/playbooks.git",
		"clone script must reference the repository URL")
	assert.Contains(t, init.Args[0], "develop", "clone script must target the configured branch")
	assert.Contains(t, init.Args[0], "/workspace/repo", "clone script targets workspace subdirectory")

	require.Len(t, init.VolumeMounts, 1)
	assert.Equal(t, workspaceVolumeName, init.VolumeMounts[0].Name)
}

func TestPrepare_TaskBranchOverridesTemplateAndRepoBranch(t *testing.T) {
	cfg := newTestConfig()
	tplBranch := "release"
	taskBranch := "hotfix"
	exec := New(cfg, db.Task{ID: 1, GitBranch: &taskBranch},
		db.Template{GitBranch: &tplBranch},
		db.Inventory{},
		db.Repository{GitURL: "https://example.com/repo.git", GitBranch: "main"},
		db.Environment{})

	require.NoError(t, exec.Prepare("", nil, ""))

	pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	require.NoError(t, err)

	require.Len(t, pod.Spec.InitContainers, 1)
	script := pod.Spec.InitContainers[0].Args[0]
	assert.Contains(t, script, "hotfix", "task branch must win over template/repo branch")
	assert.NotContains(t, script, "release")
	assert.NotContains(t, script, " main ")
}

func TestPrepare_SkipsInitContainerForLocalRepo(t *testing.T) {
	cfg := newTestConfig()
	// A leading slash classifies the repository as RepositoryLocal — see db.Repository.GetType.
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{},
		db.Repository{GitURL: "/srv/repos/local"},
		db.Environment{})

	require.NoError(t, exec.Prepare("", nil, ""))

	pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	require.NoError(t, err)

	assert.Empty(t, pod.Spec.InitContainers, "local repos cannot be cloned into a Pod; init container must be omitted")
	assert.Equal(t, workspaceVolumeName, pod.Spec.Volumes[0].Name,
		"workspace volume is still attached so the build container has /workspace")
}

func TestBuildGitCloneScript_ProducesValidCommand(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		branch     string
		wantBranch bool
	}{
		{"branch supplied", "https://x/y.git", "main", true},
		{"empty branch falls back to HEAD", "https://x/y.git", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := buildGitCloneScript(tt.url, tt.branch, "/workspace/repo")
			assert.Contains(t, script, "set -e", "script must abort on first failure")
			assert.Contains(t, script, "git clone")
			assert.Contains(t, script, tt.url)
			assert.Contains(t, script, "/workspace/repo")
			if tt.wantBranch {
				assert.Contains(t, script, "--branch")
				assert.Contains(t, script, tt.branch)
			} else {
				assert.NotContains(t, script, "--branch")
			}
		})
	}
}
