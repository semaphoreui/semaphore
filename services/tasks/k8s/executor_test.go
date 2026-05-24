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
	corev1 "k8s.io/api/core/v1"
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

// --- Appendix A.1: private SSH clone --------------------------------------------

func TestPrepare_InitContainerMountsRepositorySSHKey(t *testing.T) {
	cfg := newTestConfig()
	key := sshAccessKey(77, "")
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{},
		db.Repository{GitURL: "git@example.com:org/playbooks.git", GitBranch: "main", SSHKey: key},
		db.Environment{})
	require.NoError(t, exec.Prepare("", nil, ""))

	pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	require.NoError(t, err)

	require.Len(t, pod.Spec.InitContainers, 1)
	init := pod.Spec.InitContainers[0]

	// Init container must mount workspace + the repository's SSH key (and only that key).
	mountByName := map[string]string{}
	for _, m := range init.VolumeMounts {
		mountByName[m.Name] = m.MountPath
	}
	assert.Equal(t, "/workspace", mountByName[workspaceVolumeName])
	assert.Equal(t, sshMountPath(key.ID), mountByName[sshVolumeName(key.ID)],
		"init container must mount the repository SSH key at /secrets/ssh/<key_id>/")
}

func TestPrepare_InitContainerScriptStartsSSHAgentForPrivateRepo(t *testing.T) {
	cfg := newTestConfig()
	key := sshAccessKey(7, "")
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{},
		db.Repository{GitURL: "git@example.com:org/repo.git", GitBranch: "main", SSHKey: key},
		db.Environment{})
	require.NoError(t, exec.Prepare("", nil, ""))

	pod, _ := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	script := pod.Spec.InitContainers[0].Args[0]

	assert.Contains(t, script, "command -v ssh-agent", "ssh-agent guard mirrors the build container's defensive pattern")
	assert.Contains(t, script, "ssh-add "+sshMountPath(key.ID)+"/"+sshSecretFilePrivateKey,
		"private key must be added to the agent before clone")
	assert.Contains(t, script, "GIT_SSH_COMMAND", "GIT_SSH_COMMAND must override ssh options")
	assert.Contains(t, script, "StrictHostKeyChecking=accept-new",
		"unknown-host prompt must be suppressed — no stdin in a Pod")
	assert.Contains(t, script, "git clone")
}

func TestPrepare_InitContainerNoSSHForPublicHTTPSRepo(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{},
		db.Repository{GitURL: "https://example.com/public.git", GitBranch: "main"}, // no SSH key
		db.Environment{})
	require.NoError(t, exec.Prepare("", nil, ""))

	pod, _ := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	init := pod.Spec.InitContainers[0]

	assert.NotContains(t, init.Args[0], "ssh-agent",
		"public repos must not boot ssh-agent — keeps the helper image's footprint clear of unused setup")
	assert.NotContains(t, init.Args[0], "GIT_SSH_COMMAND")

	// Only the workspace mount is present; no SSH key Secret was installed so no
	// extra mounts get attached either.
	require.Len(t, init.VolumeMounts, 1)
	assert.Equal(t, workspaceVolumeName, init.VolumeMounts[0].Name)
}

func TestPrepare_InitContainerOnlyRepoKey_NotInventoryKey(t *testing.T) {
	cfg := newTestConfig()
	repoKey := sshAccessKey(11, "")
	invKey := sshAccessKey(22, "")
	exec := New(cfg, db.Task{ID: 1}, db.Template{},
		db.Inventory{SSHKey: invKey},
		db.Repository{GitURL: "git@example.com:org/r.git", GitBranch: "main", SSHKey: repoKey},
		db.Environment{})
	require.NoError(t, exec.Prepare("", nil, ""))

	pod, _ := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	init := pod.Spec.InitContainers[0]

	mountNames := map[string]bool{}
	for _, m := range init.VolumeMounts {
		mountNames[m.Name] = true
	}
	assert.True(t, mountNames[sshVolumeName(repoKey.ID)], "repo key mounted in init container")
	assert.False(t, mountNames[sshVolumeName(invKey.ID)],
		"inventory key must NOT leak into the init container (least-privilege per spec A.1)")

	// And the ssh-add line references only the repo key, not the inventory key.
	script := init.Args[0]
	assert.Contains(t, script, "ssh-add "+sshMountPath(repoKey.ID)+"/"+sshSecretFilePrivateKey)
	assert.NotContains(t, script, sshMountPath(invKey.ID))
}

func TestPrepare_InitContainerPassphraseProtectedRepoKeyUsesAskPass(t *testing.T) {
	cfg := newTestConfig()
	key := sshAccessKey(9, "s3cret")
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{},
		db.Repository{GitURL: "git@example.com:org/r.git", GitBranch: "main", SSHKey: key},
		db.Environment{})
	require.NoError(t, exec.Prepare("", nil, ""))

	pod, _ := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	script := pod.Spec.InitContainers[0].Args[0]

	assert.Contains(t, script, "SSH_ASKPASS",
		"passphrase-protected repo keys reuse the same askpass flow as the build container")
	assert.Contains(t, script, sshMountPath(key.ID)+"/"+sshSecretFilePassphrase)
}

// terminate marks the Pod's containers as terminated and pushes a final phase via
// the fake clientset's UpdateStatus path. Used by tests to simulate the real
// kubelet/scheduler updating container state — the fake clientset doesn't do any
// of that automatically.
func terminate(t *testing.T, cfg Config, podName string, phase corev1.PodPhase, initNames, buildNames []string) {
	t.Helper()
	pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), podName, metav1.GetOptions{})
	require.NoError(t, err)

	mk := func(name string) corev1.ContainerStatus {
		return corev1.ContainerStatus{
			Name: name,
			State: corev1.ContainerState{
				Terminated: &corev1.ContainerStateTerminated{ExitCode: 0},
			},
		}
	}
	for _, n := range initNames {
		pod.Status.InitContainerStatuses = append(pod.Status.InitContainerStatuses, mk(n))
	}
	for _, n := range buildNames {
		pod.Status.ContainerStatuses = append(pod.Status.ContainerStatuses, mk(n))
	}
	pod.Status.Phase = phase

	_, err = cfg.Clientset.CoreV1().Pods(cfg.Namespace).UpdateStatus(context.Background(), pod, metav1.UpdateOptions{})
	require.NoError(t, err)
}

func TestStreamPodLogsLive_FetchesInitContainerAndBuildContainer(t *testing.T) {
	cfg := newTestConfig()
	cfg.PollInterval = 5 * time.Millisecond
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{},
		db.Repository{GitURL: "https://example.com/r.git", GitBranch: "main"},
		db.Environment{})
	require.NoError(t, exec.Prepare("", nil, ""))

	// Simulate the kubelet reporting both containers as terminated and the Pod as
	// Succeeded. Without this the streamer loop never sees a containerHasStarted
	// transition and would block forever.
	terminate(t, cfg, exec.podName, corev1.PodSucceeded,
		[]string{gitCloneInitContainerName},
		[]string{buildContainerName})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	phase, err := exec.streamPodLogsLive(ctx)
	require.NoError(t, err)
	assert.Equal(t, corev1.PodSucceeded, phase)

	// What we assert here is that the executor requested logs for both containers
	// — init first, then build. Without this, a failed `git clone` would land users
	// with a Pod-failed status and zero diagnostic output. We also verify Follow=true
	// so users see logs *live* rather than as a post-mortem dump.
	want := map[string]bool{gitCloneInitContainerName: false, buildContainerName: false}
	for _, action := range cfg.Clientset.(*fake.Clientset).Actions() {
		if action.GetVerb() != "get" || action.GetSubresource() != "log" {
			continue
		}
		if v, ok := action.(interface{ GetValue() interface{} }); ok {
			if opts, ok := v.GetValue().(*corev1.PodLogOptions); ok {
				want[opts.Container] = true
				assert.True(t, opts.Follow, "log stream for %s must use Follow=true for live tailing", opts.Container)
			}
		}
	}
	assert.True(t, want[gitCloneInitContainerName], "init container log stream must be opened")
	assert.True(t, want[buildContainerName], "build container log stream must be opened")
}

func TestStreamPodLogsLive_NoInitContainersStreamOnlyBuild(t *testing.T) {
	cfg := newTestConfig()
	cfg.PollInterval = 5 * time.Millisecond
	// Local repo → no init container is added; only the build container exists.
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{},
		db.Repository{GitURL: "/srv/local/repo"},
		db.Environment{})
	require.NoError(t, exec.Prepare("", nil, ""))

	terminate(t, cfg, exec.podName, corev1.PodSucceeded, nil, []string{buildContainerName})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	phase, err := exec.streamPodLogsLive(ctx)
	require.NoError(t, err)
	assert.Equal(t, corev1.PodSucceeded, phase)

	sawInitFetch := false
	for _, action := range cfg.Clientset.(*fake.Clientset).Actions() {
		if action.GetVerb() != "get" || action.GetSubresource() != "log" {
			continue
		}
		if v, ok := action.(interface{ GetValue() interface{} }); ok {
			if opts, ok := v.GetValue().(*corev1.PodLogOptions); ok {
				if opts.Container == gitCloneInitContainerName {
					sawInitFetch = true
				}
			}
		}
	}
	assert.False(t, sawInitFetch, "init container fetch must be skipped when there are no init containers")
}

func TestStreamPodLogsLive_PropagatesFailedPhase(t *testing.T) {
	cfg := newTestConfig()
	cfg.PollInterval = 5 * time.Millisecond
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{}, db.Repository{}, db.Environment{})
	require.NoError(t, exec.Prepare("", nil, ""))

	terminate(t, cfg, exec.podName, corev1.PodFailed, nil, []string{buildContainerName})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	phase, err := exec.streamPodLogsLive(ctx)
	require.NoError(t, err)
	assert.Equal(t, corev1.PodFailed, phase, "PodFailed must propagate so Run can return an error")
}

func TestContainerHasStarted(t *testing.T) {
	assert.False(t, containerHasStarted(corev1.ContainerState{
		Waiting: &corev1.ContainerStateWaiting{Reason: "PodInitializing"},
	}), "waiting containers have no logs to stream yet")

	assert.True(t, containerHasStarted(corev1.ContainerState{
		Running: &corev1.ContainerStateRunning{},
	}), "running containers are streamable")

	assert.True(t, containerHasStarted(corev1.ContainerState{
		Terminated: &corev1.ContainerStateTerminated{ExitCode: 1},
	}), "fast-exiting containers (Waiting→Terminated without observed Running) must still be streamed")
}

func TestRepositorySSHInstall_FiltersByOrigin(t *testing.T) {
	repo := sshKeyInstallation{key: sshAccessKey(1, ""), origin: "repository"}
	inv := sshKeyInstallation{key: sshAccessKey(2, ""), origin: "inventory"}

	got, ok := repositorySSHInstall([]sshKeyInstallation{inv, repo})
	require.True(t, ok)
	assert.Equal(t, 1, got.key.ID, "must return the entry with origin=repository regardless of order")

	_, ok = repositorySSHInstall([]sshKeyInstallation{inv})
	assert.False(t, ok, "no repo entry → ok=false")
}

// --- Appendix A.3: OpenShift random-UID passwd fixup ----------------------------

func TestPrepare_BuildScriptIncludesPasswdFixup(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{},
		db.Repository{}, db.Environment{})
	require.NoError(t, exec.Prepare("", nil, ""))

	pod, _ := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	script := pod.Spec.Containers[0].Args[0]

	// The fixup must run before anything that might invoke ssh/git, since both shell
	// out to getpwuid which trips "No user exists for uid" under OpenShift random UIDs.
	assert.Contains(t, script, "whoami >/dev/null 2>&1",
		"passwd fixup uses whoami as a getpwuid proxy")
	assert.Contains(t, script, ">> /etc/passwd",
		"fixup appends a synthetic passwd entry rather than rewriting")
	assert.Contains(t, script, "$(id -u)")
}

func TestPrepare_InitScriptIncludesPasswdFixup(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{},
		db.Repository{GitURL: "https://example.com/r.git", GitBranch: "main"},
		db.Environment{})
	require.NoError(t, exec.Prepare("", nil, ""))

	pod, _ := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	script := pod.Spec.InitContainers[0].Args[0]

	// Init container needs the fixup just as badly — git over ssh fails first.
	assert.Contains(t, script, ">> /etc/passwd")
	assert.Contains(t, script, "whoami >/dev/null")
}

func TestPrepare_InitContainerSetsHomeEnv(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{},
		db.Repository{GitURL: "https://example.com/r.git"},
		db.Environment{})
	require.NoError(t, exec.Prepare("", nil, ""))

	pod, _ := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	init := pod.Spec.InitContainers[0]

	var sawHome bool
	for _, env := range init.Env {
		if env.Name == "HOME" {
			assert.Equal(t, "/workspace", env.Value,
				"HOME must point at the writable workspace volume so the passwd fixup synthesizes a usable entry")
			sawHome = true
		}
	}
	assert.True(t, sawHome, "init container must set HOME for the passwd fixup to land a usable home dir")
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
			script := buildGitCloneScript(tt.url, tt.branch, "/workspace/repo", sshKeyInstallation{}, false)
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
