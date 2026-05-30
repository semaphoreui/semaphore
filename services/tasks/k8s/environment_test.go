package k8s

import (
	"context"
	"fmt"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func envPtr(s string) *string { return &s }

func TestPrepare_PlainEnvironmentENVReachesContainer(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{}, db.Repository{},
		db.Environment{
			ENV: envPtr(`{"FOO": "bar", "TARGET_HOST": "10.0.0.5"}`),
		})
	require.NoError(t, exec.Prepare("", nil, ""))

	pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	require.NoError(t, err)

	envByName := map[string]string{}
	for _, env := range pod.Spec.Containers[0].Env {
		envByName[env.Name] = env.Value
	}
	assert.Equal(t, "bar", envByName["FOO"])
	assert.Equal(t, "10.0.0.5", envByName["TARGET_HOST"])
	// Base env vars must still be there.
	assert.Equal(t, workspaceMountPath, envByName["HOME"])
}

func TestPrepare_EnvironmentENVUserValueOverridesBase(t *testing.T) {
	// If a user puts HOME in their Variable Group it must win — they explicitly asked
	// for a specific value. K8s uses last-wins semantics on duplicate EnvVar names.
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{}, db.Repository{},
		db.Environment{ENV: envPtr(`{"HOME": "/custom"}`)})
	require.NoError(t, exec.Prepare("", nil, ""))

	pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	require.NoError(t, err)

	var homeValues []string
	for _, env := range pod.Spec.Containers[0].Env {
		if env.Name == "HOME" {
			homeValues = append(homeValues, env.Value)
		}
	}
	require.Len(t, homeValues, 2, "base HOME + user HOME both present so last-wins applies")
	assert.Equal(t, workspaceMountPath, homeValues[0])
	assert.Equal(t, "/custom", homeValues[1])
}

func TestPrepare_InvalidEnvironmentENVJSONDoesNotFail(t *testing.T) {
	// LocalExecutor swallows env-parse errors and runs anyway; matching that here keeps
	// behaviour parity. The task shouldn't die on a typo in the Variable Group ENV box.
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{}, db.Repository{},
		db.Environment{ENV: envPtr(`{not valid json`)})
	require.NoError(t, exec.Prepare("", nil, ""))

	pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	require.NoError(t, err)
	// Base envs still present, no junk vars from the broken JSON.
	assert.NotEmpty(t, pod.Spec.Containers[0].Env)
}

func TestPrepare_EnvironmentENVOutputIsDeterministic(t *testing.T) {
	// Map iteration is randomised in Go; ensure two consecutive Prepares (across two
	// Executors with the same input) produce identical EnvVar order so the Pod spec
	// is stable for diff-based reviews and golden-file tests.
	cfg := newTestConfig()
	envJSON := envPtr(`{"A": "1", "B": "2", "C": "3", "D": "4", "E": "5", "F": "6"}`)

	var orders [][]string
	for i := 0; i < 2; i++ {
		exec := New(cfg, db.Task{ID: i + 1}, db.Template{}, db.Inventory{}, db.Repository{},
			db.Environment{ENV: envJSON})
		require.NoError(t, exec.Prepare("", nil, ""))

		pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
		require.NoError(t, err)

		var names []string
		for _, env := range pod.Spec.Containers[0].Env {
			if len(env.Name) == 1 { // pick out our single-letter user vars only
				names = append(names, env.Name)
			}
		}
		orders = append(orders, names)
	}
	assert.Equal(t, orders[0], orders[1], "env order must be deterministic across Prepares")
	assert.Equal(t, []string{"A", "B", "C", "D", "E", "F"}, orders[0])
}

func TestPrepare_EnvSecretsCreateSecretAndEnvFrom(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{}, db.Repository{},
		db.Environment{
			Secrets: []db.EnvironmentSecret{
				{Type: db.EnvironmentSecretEnv, Name: "API_TOKEN", Secret: "s3cret"},
				{Type: db.EnvironmentSecretEnv, Name: "DB_PASSWORD", Secret: "hunter2"},
				{Type: db.EnvironmentSecretVar, Name: "ignored_var", Secret: "extravar"},
			},
		})
	require.NoError(t, exec.Prepare("", nil, ""))

	envSecretName := fmt.Sprintf("%s-env", exec.podName)
	secret, err := cfg.Clientset.CoreV1().Secrets(cfg.Namespace).Get(context.Background(), envSecretName, metav1.GetOptions{})
	require.NoError(t, err, "env-typed secrets must produce a K8s Secret")

	assert.Equal(t, []byte("s3cret"), secret.Data["API_TOKEN"])
	assert.Equal(t, []byte("hunter2"), secret.Data["DB_PASSWORD"])
	_, hasVar := secret.Data["ignored_var"]
	assert.False(t, hasVar, "var-typed secrets must NOT end up in the env Secret (they go to --extra-vars)")

	pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	require.NoError(t, err)
	require.Len(t, pod.Spec.Containers[0].EnvFrom, 1, "envFrom must reference the env Secret")
	assert.Equal(t, envSecretName, pod.Spec.Containers[0].EnvFrom[0].SecretRef.Name)
}

func TestPrepare_NoEnvSecrets_NoEnvSecretCreated(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{}, db.Repository{}, db.Environment{})
	require.NoError(t, exec.Prepare("", nil, ""))

	envSecretName := fmt.Sprintf("%s-env", exec.podName)
	_, err := cfg.Clientset.CoreV1().Secrets(cfg.Namespace).Get(context.Background(), envSecretName, metav1.GetOptions{})
	assert.True(t, apierrors.IsNotFound(err), "no env-typed secrets means no env Secret")

	pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	require.NoError(t, err)
	assert.Empty(t, pod.Spec.Containers[0].EnvFrom)
}

func TestPrepare_EmptyNameEnvSecretSkipped(t *testing.T) {
	// An EnvironmentSecret with empty Name would be rejected by the API server; drop
	// it client-side with a clear contract instead of failing Prepare with a confusing
	// "invalid value for field 'data'" error from K8s.
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{}, db.Repository{},
		db.Environment{
			Secrets: []db.EnvironmentSecret{
				{Type: db.EnvironmentSecretEnv, Name: "", Secret: "ignored"},
				{Type: db.EnvironmentSecretEnv, Name: "REAL", Secret: "kept"},
			},
		})
	require.NoError(t, exec.Prepare("", nil, ""))

	envSecretName := fmt.Sprintf("%s-env", exec.podName)
	secret, err := cfg.Clientset.CoreV1().Secrets(cfg.Namespace).Get(context.Background(), envSecretName, metav1.GetOptions{})
	require.NoError(t, err)
	assert.Len(t, secret.Data, 1)
	assert.Equal(t, []byte("kept"), secret.Data["REAL"])
}

func TestCleanup_DeletesEnvSecret(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{}, db.Repository{},
		db.Environment{
			Secrets: []db.EnvironmentSecret{
				{Type: db.EnvironmentSecretEnv, Name: "TOKEN", Secret: "x"},
			},
		})
	require.NoError(t, exec.Prepare("", nil, ""))

	envSecretName := fmt.Sprintf("%s-env", exec.podName)
	exec.Cleanup()

	_, err := cfg.Clientset.CoreV1().Secrets(cfg.Namespace).Get(context.Background(), envSecretName, metav1.GetOptions{})
	assert.True(t, apierrors.IsNotFound(err), "env Secret must be cleaned up with the Pod")
}
