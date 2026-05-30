package k8s

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/semaphoreui/semaphore/db"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// environmentEnvFromJSON parses Environment.ENV (a JSON object of string→string) into
// container env vars. Mirrors LocalExecutor.getEnvironmentENV: plain (non-secret)
// values from a Variable Group land in the process env of the task. Invalid JSON is
// treated as "no vars" rather than failing the task — matches the forgiving behaviour
// of the local path.
//
// Output is sorted by key so the resulting Pod spec is deterministic; without this
// Go's map iteration order would shuffle envs between Prepares and make tests flaky.
func environmentEnvFromJSON(envJSON *string) []corev1.EnvVar {
	if envJSON == nil || *envJSON == "" {
		return nil
	}
	parsed := map[string]string{}
	if err := json.Unmarshal([]byte(*envJSON), &parsed); err != nil {
		return nil
	}
	keys := make([]string, 0, len(parsed))
	for k := range parsed {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]corev1.EnvVar, 0, len(keys))
	for _, k := range keys {
		out = append(out, corev1.EnvVar{Name: k, Value: parsed[k]})
	}
	return out
}

// buildEnvSecret packs Environment.Secrets entries of type EnvironmentSecretEnv into
// a single K8s Secret and returns the matching envFrom reference. Returns (nil, nil)
// when no env-typed secrets exist so the caller can skip secret creation entirely.
//
// envFrom is preferred over inlining values as EnvVar literals because the secret
// material then never appears in the Pod spec (only the Secret reference does),
// which keeps the values out of `kubectl get pod -o yaml` and out of audit logs that
// capture Pod creation events.
//
// Entries with empty Name are skipped — an empty env var name would be rejected by
// the API server and is meaningless anyway.
func (e *Executor) buildEnvSecret() (*corev1.Secret, *corev1.EnvFromSource) {
	data := map[string][]byte{}
	for _, s := range e.Environment.Secrets {
		if s.Type != db.EnvironmentSecretEnv {
			continue
		}
		if s.Name == "" {
			continue
		}
		data[s.Name] = []byte(s.Secret)
	}
	if len(data) == 0 {
		return nil, nil
	}

	secretName := fmt.Sprintf("%s-env", e.podName)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: e.Config.Namespace,
			Labels: map[string]string{
				LabelRunner:  "semaphore",
				labelPodName: e.podName,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: data,
	}
	envFrom := &corev1.EnvFromSource{
		SecretRef: &corev1.SecretEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{Name: secretName},
		},
	}
	return secret, envFrom
}
