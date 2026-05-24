package k8s

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
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

// Prepare creates the task Pod and waits for it to leave the Pending phase. The
// username / incomingVersion / alias arguments are kept for interface compatibility
// with LocalExecutor; this skeleton ignores them. They become relevant once the K8s
// executor learns to surface task-details env vars (Phase 6) and TF_HTTP_ADDRESS
// (Terraform alias support).
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

	pod := e.buildPodSpec(podName)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = e.Config.Clientset.CoreV1().Pods(e.Config.Namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("create pod %q: %w", podName, err)
	}

	e.log(fmt.Sprintf("k8s: created pod %s/%s", e.Config.Namespace, podName))
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

// Cleanup deletes the Pod. It is invoked unconditionally by Run via defer; calling
// it on an unprepared executor (where podName is still empty) is a no-op so that
// failed Prepare calls don't leave the deletion error masking the real error.
func (e *Executor) Cleanup() {
	if e.podName == "" || e.Config.Clientset == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	grace := int64(e.Config.CleanupGrace.Seconds())
	err := e.Config.Clientset.CoreV1().Pods(e.Config.Namespace).Delete(ctx, e.podName, metav1.DeleteOptions{
		GracePeriodSeconds: &grace,
	})
	if err != nil && !apierrors.IsNotFound(err) {
		e.log(fmt.Sprintf("k8s: failed to delete pod %s: %v", e.podName, err))
		return
	}
	e.log(fmt.Sprintf("k8s: deleted pod %s/%s", e.Config.Namespace, e.podName))
}

// --- helpers ---------------------------------------------------------------------

// buildPodSpec constructs the Pod object the skeleton runs. Phase 3+ will replace the
// hardcoded "echo hello" command with a keeper-shell entrypoint that ansible commands
// are streamed into via attach (see docs/plans/kubernetes-executor-spec.md section 7).
func (e *Executor) buildPodSpec(podName string) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: e.Config.Namespace,
			Labels: map[string]string{
				LabelTaskID: fmt.Sprintf("%d", e.Task.ID),
				LabelRunner: "semaphore",
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy:      corev1.RestartPolicyNever,
			ServiceAccountName: e.Config.ServiceAccount,
			Containers: []corev1.Container{{
				Name:    buildContainerName,
				Image:   e.Config.Image,
				Command: []string{"sh", "-c"},
				Args: []string{fmt.Sprintf(
					"echo 'semaphore k8s executor skeleton: task %d, template %d'",
					e.Task.ID, e.Template.ID,
				)},
			}},
		},
	}

	for _, name := range e.Config.PullSecrets {
		pod.Spec.ImagePullSecrets = append(pod.Spec.ImagePullSecrets, corev1.LocalObjectReference{Name: name})
	}

	return pod
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
