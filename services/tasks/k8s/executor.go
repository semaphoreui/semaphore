package k8s

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// LabelRunner tags every K8s resource created by this executor so orphan cleanup
	// on runner restart can find them. Value is the runner's name / token-prefix —
	// kept as a label key constant so Phase 5 (orphan cleanup) can list-by-selector.
	LabelRunner = "semaphoreui.com/runner"

	// LabelTaskID identifies the Semaphore task ID that owns the resource.
	LabelTaskID = "semaphoreui.com/task-id"

	// buildContainerName is the name of the Pod's primary container. Stays constant
	// across tasks so attach/log calls don't need to discover it.
	buildContainerName = "build"

	// gitCloneInitContainerName is the init container that materializes the task
	// repository into the shared workspace volume. Lives only for the lifetime of
	// the clone and exits — the build container starts once it succeeds.
	gitCloneInitContainerName = "git-clone"

	// workspaceVolumeName is the emptyDir shared between the init container (writes
	// the repository into it) and the build container (consumes it).
	workspaceVolumeName = "workspace"

	// workspaceMountPath is where the shared workspace appears inside containers.
	// The repository is cloned into workspaceMountPath/repo so future phases can
	// also drop helper files (inventory, extra-vars JSON, .exit marker) alongside
	// it without colliding with repo contents.
	workspaceMountPath = "/workspace"

	// workspaceRepoSubpath is the subdirectory under workspaceMountPath where the
	// task repository ends up.
	workspaceRepoSubpath = "repo"
)

// Executor runs one Semaphore task inside an ephemeral Kubernetes Pod. The skeleton
// implementation (Phase 2) creates a Pod that prints a single hello message, waits for
// it to terminate, streams its logs back into the task logger, and deletes the Pod.
// Phases 3+ wire in git clone, secret mounting, and real ansible-playbook execution.
//
// Executor satisfies services/tasks.Executor (the interface that the runner job pool
// dispatches into). It is constructed by the runner factory per task and never reused.
type Executor struct {
	Task        db.Task
	Template    db.Template
	Inventory   db.Inventory
	Repository  db.Repository
	Environment db.Environment

	Logger task_logger.Logger
	Config Config

	mu     sync.Mutex
	killed bool

	// podName is generated in Prepare and reused by Run / Cleanup. Empty until prepared.
	podName string

	// secretNames are the K8s Secrets this executor created for the task (SSH keys,
	// vault passwords, extra-vars JSON). Tracked so Cleanup can delete them —
	// ownerReferences would also work but explicit tracking means a half-failed
	// Prepare doesn't strand resources.
	secretNames []string

	// configmapNames are the K8s ConfigMaps this executor created (today: only
	// the static inventory ConfigMap, when applicable). Same lifecycle story as
	// secretNames.
	configmapNames []string

	// cancelRun fires when the user requests a stop. The log-stream / status-poll
	// goroutines watch this context and bail out.
	cancelRun context.CancelFunc
}

// New builds an Executor from the runtime Config and the per-task data the runner
// pulled off the server. The Logger is wired separately via SetLogger before Run is
// invoked, mirroring how LocalExecutor's logger is attached by the job pool.
func New(cfg Config, task db.Task, template db.Template, inventory db.Inventory, repository db.Repository, environment db.Environment) *Executor {
	return &Executor{
		Task:        task,
		Template:    template,
		Inventory:   inventory,
		Repository:  repository,
		Environment: environment,
		Config:      cfg,
	}
}

// --- task_logger / Job interface plumbing -----------------------------------------

// SetLogger wires the per-task log sink into the executor. Called by the job pool
// after construction but before Run.
func (e *Executor) SetLogger(logger task_logger.Logger) {
	e.Logger = logger
}

// SetStatus forwards to the bound Logger. Safe to call before SetLogger (no-op).
func (e *Executor) SetStatus(status task_logger.TaskStatus) {
	if e.Logger != nil {
		e.Logger.SetStatus(status)
	}
}

func (e *Executor) log(msg string) {
	if e.Logger != nil {
		e.Logger.Log(msg)
	}
}

// IsKilled reports whether Kill has been invoked. Used by the job pool to decide
// whether a returned error means "stopped by user" vs "failed".
func (e *Executor) IsKilled() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.killed
}

// Kill is invoked by the job pool when the user asks to stop the task. It flips the
// killed flag and signals the Run loop to bail out; the actual Pod deletion happens
// in Cleanup (called via defer in Run).
func (e *Executor) Kill() {
	e.mu.Lock()
	e.killed = true
	cancel := e.cancelRun
	e.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

// --- Executor interface methods --------------------------------------------------

// Prepare creates the per-task K8s resources: one Secret per installed SSH key, then
// the build Pod that references them. The order matters — Pod creation references
// the Secrets by name, so they must exist first (otherwise the kubelet would block
// the Pod in ContainerCreating until they appear). On failure, any Secrets already
// created are torn down via Cleanup (caller invokes it via defer).
//
// The username / incomingVersion / alias arguments are kept for interface compatibility
// with LocalExecutor. They become relevant once the K8s executor learns to surface
// task-details env vars (Phase 6) and TF_HTTP_ADDRESS (Terraform alias support).
func (e *Executor) Prepare(username string, incomingVersion *string, alias string) error {
	if e.Config.Clientset == nil {
		return errors.New("k8s executor: clientset is not configured")
	}

	if e.podName != "" {
		return nil // idempotent — Run may have called Prepare already
	}

	podName, err := generatePodName(e.Task.ID)
	if err != nil {
		return fmt.Errorf("generate pod name: %w", err)
	}
	e.podName = podName

	sshInstalls := e.collectSSHKeys()

	// Ansible-specific resources (vault password Secrets, inventory ConfigMap,
	// extra-vars JSON Secret). Nil for non-Ansible templates — those keep the
	// skeleton "ls /workspace" behaviour until later phases.
	ansible, err := e.buildAnsiblePrep(username, incomingVersion)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err = e.createSSHSecrets(ctx, sshInstalls); err != nil {
		return err
	}
	if err = e.createAnsibleResources(ctx, ansible); err != nil {
		return err
	}

	pod := e.buildPodSpec(podName, sshInstalls, ansible)
	if _, err = e.Config.Clientset.CoreV1().Pods(e.Config.Namespace).Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("create pod %q: %w", podName, err)
	}

	e.log(fmt.Sprintf("k8s: created pod %s/%s (%d ssh keys installed)", e.Config.Namespace, podName, len(sshInstalls)))
	return nil
}

// createAnsibleResources writes vault Secrets, the extra-vars Secret, and the inventory
// ConfigMap (when applicable) into the cluster. Names are recorded on the executor so
// Cleanup catches them even when this method returns partway through.
func (e *Executor) createAnsibleResources(ctx context.Context, prep *ansiblePrep) error {
	if prep == nil {
		return nil
	}

	if prep.configMap != nil {
		_, err := e.Config.Clientset.CoreV1().ConfigMaps(e.Config.Namespace).Create(ctx, prep.configMap, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("create configmap %q: %w", prep.configMap.Name, err)
		}
		e.configmapNames = append(e.configmapNames, prep.configMap.Name)
	}

	for _, secret := range prep.secrets {
		_, err := e.Config.Clientset.CoreV1().Secrets(e.Config.Namespace).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("create secret %q: %w", secret.Name, err)
		}
		e.secretNames = append(e.secretNames, secret.Name)
	}
	return nil
}

// createSSHSecrets writes one K8s Secret per installed SSH key. Each name created is
// recorded on the executor so Cleanup can delete it even when this method bails out
// partway through.
func (e *Executor) createSSHSecrets(ctx context.Context, installs []sshKeyInstallation) error {
	for _, inst := range installs {
		secret := buildSSHSecret(e.podName, e.Config.Namespace, inst)
		if _, err := e.Config.Clientset.CoreV1().Secrets(e.Config.Namespace).Create(ctx, secret, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("create ssh secret for key %d: %w", inst.key.ID, err)
		}
		e.secretNames = append(e.secretNames, secret.Name)
	}
	return nil
}

// Run is the entry point the runner job pool calls. The skeleton flow is:
//  1. Prepare creates the Pod.
//  2. Wait for the Pod to reach a terminal phase (Succeeded / Failed) or the runner
//     to request a stop (Kill).
//  3. Stream the Pod's logs into the task logger.
//  4. Cleanup deletes the Pod (via defer).
//
// The killed flag is checked after Prepare so a stop request that arrived during
// creation still propagates correctly.
func (e *Executor) Run(username string, incomingVersion *string, alias string) (err error) {
	defer e.Cleanup()

	if err = e.Prepare(username, incomingVersion, alias); err != nil {
		return err
	}

	if e.IsKilled() {
		e.SetStatus(task_logger.TaskStoppedStatus)
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	e.mu.Lock()
	e.cancelRun = cancel
	e.mu.Unlock()
	defer cancel()

	finalPhase, err := e.waitForPodCompletion(ctx)
	if err != nil {
		if e.IsKilled() {
			// Cancellation is expected; surface as a clean stop.
			return nil
		}
		return fmt.Errorf("wait for pod completion: %w", err)
	}

	if streamErr := e.streamPodLogs(ctx); streamErr != nil {
		// Log streaming failure should not mask the pod result; just record it.
		e.log(fmt.Sprintf("k8s: failed to stream logs: %v", streamErr))
	}

	if finalPhase == corev1.PodFailed {
		return fmt.Errorf("pod %s finished with status Failed", e.podName)
	}
	return nil
}

// Cleanup deletes the Pod and every Secret this executor created (one per installed
// SSH key today). Invoked unconditionally by Run via defer; calling it on an
// unprepared executor is a no-op so that failed Prepare calls don't leave the
// deletion error masking the real error.
//
// Deletion order is Pod-first so the kubelet is the one to unmount the Secret
// volumes; deleting Secrets while a running Pod still mounts them is allowed by
// K8s but produces confusing kubelet logs. NotFound errors are ignored so the
// method is safe to call repeatedly and on partially-prepared executors.
func (e *Executor) Cleanup() {
	if e.Config.Clientset == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if e.podName != "" {
		grace := int64(e.Config.CleanupGrace.Seconds())
		err := e.Config.Clientset.CoreV1().Pods(e.Config.Namespace).Delete(ctx, e.podName, metav1.DeleteOptions{
			GracePeriodSeconds: &grace,
		})
		if err != nil && !apierrors.IsNotFound(err) {
			e.log(fmt.Sprintf("k8s: failed to delete pod %s: %v", e.podName, err))
		} else {
			e.log(fmt.Sprintf("k8s: deleted pod %s/%s", e.Config.Namespace, e.podName))
		}
	}

	for _, name := range e.secretNames {
		err := e.Config.Clientset.CoreV1().Secrets(e.Config.Namespace).Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			e.log(fmt.Sprintf("k8s: failed to delete secret %s: %v", name, err))
		}
	}

	for _, name := range e.configmapNames {
		err := e.Config.Clientset.CoreV1().ConfigMaps(e.Config.Namespace).Delete(ctx, name, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			e.log(fmt.Sprintf("k8s: failed to delete configmap %s: %v", name, err))
		}
	}
}

// --- helpers ---------------------------------------------------------------------

// buildPodSpec constructs the Pod object the executor runs. The Pod always contains a
// shared workspace emptyDir; for any non-local repository it also gets a git-clone init
// container that materializes the repo into /workspace/repo before the build container
// starts. Each installed SSH key gets a Secret-backed volume mounted at
// /secrets/ssh/<id>/, mode 0400. Local repos (file paths on the runner host) are not
// meaningful inside a Pod — the init container is skipped and the build container
// will see an empty workspace.
//
// Phases 5+ replace the build container's command with a keeper-shell entrypoint that
// ansible commands are streamed into via attach (see spec section 7).
func (e *Executor) buildPodSpec(podName string, sshInstalls []sshKeyInstallation, ansible *ansiblePrep) *corev1.Pod {
	workspaceMount := corev1.VolumeMount{
		Name:      workspaceVolumeName,
		MountPath: workspaceMountPath,
	}

	repoPath := workspaceMountPath + "/" + workspaceRepoSubpath

	volumes := []corev1.Volume{{
		Name: workspaceVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}}
	volumes = append(volumes, sshSecretVolumes(podName, sshInstalls)...)

	buildMounts := []corev1.VolumeMount{workspaceMount}
	buildMounts = append(buildMounts, sshSecretVolumeMounts(sshInstalls)...)

	if ansible != nil {
		volumes = append(volumes, ansible.volumes...)
		buildMounts = append(buildMounts, ansible.volumeMounts...)
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: e.Config.Namespace,
			Labels: map[string]string{
				LabelTaskID:  fmt.Sprintf("%d", e.Task.ID),
				LabelRunner:  "semaphore",
				labelPodName: podName,
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy:      corev1.RestartPolicyNever,
			ServiceAccountName: e.Config.ServiceAccount,
			Volumes:            volumes,
			Containers: []corev1.Container{{
				Name:         buildContainerName,
				Image:        e.Config.Image,
				WorkingDir:   repoPath,
				Command:      []string{"sh", "-c"},
				Args:         []string{e.buildContainerScript(repoPath, sshInstalls, ansible)},
				VolumeMounts: buildMounts,
				Env:          buildContainerEnv(sshInstalls),
			}},
		},
	}

	if initContainer, ok := e.gitCloneInitContainer(repoPath, workspaceMount); ok {
		pod.Spec.InitContainers = []corev1.Container{initContainer}
	}

	for _, name := range e.Config.PullSecrets {
		pod.Spec.ImagePullSecrets = append(pod.Spec.ImagePullSecrets, corev1.LocalObjectReference{Name: name})
	}

	return pod
}

// buildContainerEnv returns the environment variables that should always be set on
// the build container.
//
// HOME and ANSIBLE_LOCAL_TEMP point into the writable workspace volume because some
// clusters (notably OpenShift's restricted-v2 SCC) assign each Pod a random UID that
// has no /etc/passwd entry. Without HOME, getpwuid() fails and Ansible falls back to
// writing to "/" (or an empty string), which it cannot do — surfacing as
// "Permission denied: '/.ansible'". The workspace emptyDir is writable by any UID.
//
// ANSIBLE_HOST_KEY_CHECKING=False is added only when SSH keys are present: without it
// Ansible refuses to connect to never-before-seen hosts because the Pod has no
// persistent known_hosts file. Phases 6+ will add SEMAPHORE_TASK_* env vars sourced
// from db.Task.
func buildContainerEnv(installs []sshKeyInstallation) []corev1.EnvVar {
	env := []corev1.EnvVar{
		{Name: "HOME", Value: workspaceMountPath},
		{Name: "ANSIBLE_LOCAL_TEMP", Value: workspaceMountPath + "/.ansible/tmp"},
	}
	if len(installs) > 0 {
		env = append(env, corev1.EnvVar{Name: "ANSIBLE_HOST_KEY_CHECKING", Value: "False"})
	}
	return env
}

// buildContainerScript returns the inline shell program the build container runs.
//
// Layout:
//  1. Header (echoes which task/template runs).
//  2. ssh-agent guard + ssh-add for each installed SSH key (Phase 4). The guard
//     keeps slim images without openssh-client from failing the whole task — they
//     simply skip the SSH setup with a clear warning.
//  3. App-specific body. For Ansible: cd into repo, exec ansible-playbook with the
//     args computed by buildAnsiblePrep. For other apps: skeleton ls /workspace
//     output, the same as Phase 2-4. `set -e` aborts on the first command failure.
//  4. Trailing "exit code" echo so the runner-side log shows the final status even
//     when ansible-playbook prints no terminator of its own.
func (e *Executor) buildContainerScript(repoPath string, installs []sshKeyInstallation, ansible *ansiblePrep) string {
	var b strings.Builder
	b.WriteString("set -e\n")
	fmt.Fprintf(&b, "echo 'semaphore k8s executor: task %d, template %d'\n", e.Task.ID, e.Template.ID)

	if agent := sshAgentSetupScript(installs); agent != "" {
		// Without `command -v` the script would abort here (set -e + missing ssh-agent),
		// taking down the whole task. We'd rather complete it without SSH if the user
		// happens to be running a no-auth playbook in a stripped-down image.
		b.WriteString("if command -v ssh-agent >/dev/null 2>&1 && command -v ssh-add >/dev/null 2>&1; then\n")
		b.WriteString(indent(agent, "  "))
		b.WriteString("else\n")
		b.WriteString("  echo 'k8s: ssh-agent / ssh-add not found in image; skipping SSH key install'\n")
		b.WriteString("fi\n")
	}

	if ansible != nil {
		fmt.Fprintf(&b, "cd %s\n", repoPath)
		fmt.Fprintf(&b, "echo 'k8s: running ansible-playbook %s'\n", shellJoin(ansible.playbookArgs))
		b.WriteString("set +e\n")
		fmt.Fprintf(&b, "ansible-playbook %s\n", shellJoin(ansible.playbookArgs))
		b.WriteString("ANSIBLE_EXIT=$?\n")
		b.WriteString("set -e\n")
		b.WriteString("echo \"k8s: ansible-playbook exited with status ${ANSIBLE_EXIT}\"\n")
		b.WriteString("exit $ANSIBLE_EXIT\n")
		return b.String()
	}

	// Skeleton body (non-Ansible templates aren't fully supported yet).
	b.WriteString("echo 'workspace contents:'\n")
	fmt.Fprintf(&b, "ls -la %s 2>/dev/null || echo '(workspace is empty)'\n", workspaceMountPath)
	fmt.Fprintf(&b, "if [ -d %s ]; then echo 'repo cloned at %s:'; ls -la %s; fi\n", repoPath, repoPath, repoPath)

	return b.String()
}

// indent prefixes every non-empty line of s with prefix. Used so the ssh-agent setup
// snippet nests cleanly inside the `if command -v` guard.
func indent(s, prefix string) string {
	if s == "" {
		return ""
	}
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

// shellJoin renders argv as a single shell-quoted command. Uses single quotes so the
// embedded JSON / paths don't get re-expanded; ' inside an arg is escaped via the
// classic "close, escape, reopen" trick.
func shellJoin(argv []string) string {
	parts := make([]string, len(argv))
	for i, a := range argv {
		parts[i] = "'" + strings.ReplaceAll(a, "'", `'"'"'`) + "'"
	}
	return strings.Join(parts, " ")
}

// gitCloneInitContainer builds the init container that fetches the task repository
// into the workspace volume. Returns ok=false for repositories that have no remote
// URL (db.RepositoryLocal): cloning a host filesystem path from inside a Pod makes
// no sense, and the skeleton lets the build container handle the resulting empty
// workspace gracefully (the future Phase 5 will surface a clear error to the user).
//
// Authentication is intentionally absent here. Phase 4 wires SSH and HTTP credentials
// in via mounted K8s Secrets; until then, only public HTTPS clones will actually
// succeed at runtime. The Pod spec itself is still produced correctly for SSH/private
// repos so factory / test coverage works against any repository shape.
func (e *Executor) gitCloneInitContainer(repoPath string, workspaceMount corev1.VolumeMount) (corev1.Container, bool) {
	if e.Repository.GetType() == db.RepositoryLocal {
		return corev1.Container{}, false
	}

	url := e.Repository.GetGitURL(true) // secure=true: do not embed credentials yet
	if url == "" {
		return corev1.Container{}, false
	}

	branch := e.Repository.GitBranch
	if e.Template.GitBranch != nil && *e.Template.GitBranch != "" {
		branch = *e.Template.GitBranch
	}
	if e.Task.GitBranch != nil && *e.Task.GitBranch != "" {
		branch = *e.Task.GitBranch
	}

	script := buildGitCloneScript(url, branch, repoPath)

	return corev1.Container{
		Name:         gitCloneInitContainerName,
		Image:        e.Config.HelperImage,
		Command:      []string{"sh", "-c"},
		Args:         []string{script},
		VolumeMounts: []corev1.VolumeMount{workspaceMount},
	}, true
}

// buildGitCloneScript renders the shell script the init container runs. Kept as a
// function (not a constant) so future phases can layer in commit-hash checkout,
// authentication, and submodule init without rewiring the container plumbing.
func buildGitCloneScript(url, branch, repoPath string) string {
	if branch == "" {
		// `git clone --branch` is mandatory only when the branch differs from HEAD;
		// omitting it lets the remote's default branch take over.
		return fmt.Sprintf("set -e\ngit clone --depth 1 %q %q", url, repoPath)
	}
	return fmt.Sprintf("set -e\ngit clone --depth 1 --branch %q %q %q", branch, url, repoPath)
}

// waitForPodCompletion polls Pod status until it enters a terminal phase or the
// context is cancelled. Polling (not watch) is fine for the skeleton; a future phase
// will switch to a SharedInformer for efficiency on busy runners.
func (e *Executor) waitForPodCompletion(ctx context.Context) (corev1.PodPhase, error) {
	ticker := time.NewTicker(e.Config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
			pod, err := e.Config.Clientset.CoreV1().Pods(e.Config.Namespace).Get(ctx, e.podName, metav1.GetOptions{})
			if err != nil {
				return "", err
			}
			switch pod.Status.Phase {
			case corev1.PodSucceeded, corev1.PodFailed:
				return pod.Status.Phase, nil
			}
		}
	}
}

// streamPodLogs pulls the build container's log stream and forwards each line into
// the task logger. The Pod has already terminated by the time we get here (the
// skeleton serializes log read after completion to keep the flow simple); future
// phases will stream logs concurrently with execution via pods/attach.
func (e *Executor) streamPodLogs(ctx context.Context) error {
	req := e.Config.Clientset.CoreV1().Pods(e.Config.Namespace).GetLogs(e.podName, &corev1.PodLogOptions{
		Container: buildContainerName,
	})
	stream, err := req.Stream(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = stream.Close() }()

	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		e.log(scanner.Text())
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		return err
	}
	return nil
}

// generatePodName produces a DNS-1123-compatible Pod name. Format mirrors the spec:
// semaphore-job-<task_id>-<rand>. Six random hex chars give 16 million combinations,
// enough to avoid collisions for the lifetime of a runner.
func generatePodName(taskID int) (string, error) {
	buf := make([]byte, 3)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return fmt.Sprintf("semaphore-job-%d-%s", taskID, hex.EncodeToString(buf)), nil
}

// Compile-time guarantee that Executor matches the shape the runner expects. The
// import of kubernetes.Interface here ensures the fake clientset (used in tests)
// stays drop-in compatible with the production type.
var _ kubernetes.Interface = (kubernetes.Interface)(nil)
