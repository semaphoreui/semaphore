// Package k8s implements a Kubernetes-backed task executor for Semaphore runners.
//
// The model is borrowed from GitLab runner's Kubernetes executor: one runner process
// drives many short-lived Pods, one Pod per task. Phase 2 (this skeleton) only proves
// the wiring — it creates a Pod that prints a hello message and tears it down. Later
// phases will add git clone init containers, secret mounting, SSH/Vault key handling,
// ansible-playbook execution via attach, and cancellation. See
// docs/plans/kubernetes-executor-spec.md for the full plan.
package k8s

import (
	"fmt"
	"strings"
	"time"

	"github.com/semaphoreui/semaphore/util"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultNamespace      = "semaphore"
	defaultImage          = "alpine:latest"
	defaultHelperImage    = "alpine/git:latest"
	defaultServiceAccount = "default"
	defaultPollInterval   = 3 * time.Second
	defaultCleanupGrace   = 30 * time.Second
)

// Config holds everything the K8s executor needs at runtime. Built once at runner
// startup from util.Config.Runner.K8s, then handed to each per-task Executor.
type Config struct {
	// Clientset is the typed K8s client. Production code builds it from KubeconfigPath
	// (or in-cluster); tests inject a fake clientset.
	Clientset kubernetes.Interface

	Namespace      string
	Image          string
	HelperImage    string
	ServiceAccount string
	PullSecrets    []string
	PollInterval   time.Duration
	CleanupGrace   time.Duration
}

// ConfigFromRunnerConfig materializes runtime config from the on-disk runner config,
// applying defaults for fields the user left blank. The clientset is built from
// KubeconfigPath when set, otherwise from in-cluster service-account credentials.
func ConfigFromRunnerConfig(rc util.RunnerK8sConfig) (Config, error) {
	restCfg, err := buildRestConfig(rc.KubeconfigPath)
	if err != nil {
		return Config{}, fmt.Errorf("build kube REST config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return Config{}, fmt.Errorf("build kube clientset: %w", err)
	}

	return Config{
		Clientset:      clientset,
		Namespace:      firstNonEmpty(rc.Namespace, defaultNamespace),
		Image:          firstNonEmpty(rc.Image, defaultImage),
		HelperImage:    firstNonEmpty(rc.HelperImage, defaultHelperImage),
		ServiceAccount: firstNonEmpty(rc.ServiceAccount, defaultServiceAccount),
		PullSecrets:    parsePullSecrets(rc.PullSecrets),
		PollInterval:   secondsOrDefault(rc.PollIntervalSeconds, defaultPollInterval),
		CleanupGrace:   secondsOrDefault(rc.CleanupGraceSeconds, defaultCleanupGrace),
	}, nil
}

// buildRestConfig picks between kubeconfig-file mode and in-cluster mode. Out-of-cluster
// mode wins when KubeconfigPath is non-empty so operators can override during dev.
func buildRestConfig(kubeconfigPath string) (*rest.Config, error) {
	if kubeconfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}
	return rest.InClusterConfig()
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func secondsOrDefault(seconds int, fallback time.Duration) time.Duration {
	if seconds <= 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}

func parsePullSecrets(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
