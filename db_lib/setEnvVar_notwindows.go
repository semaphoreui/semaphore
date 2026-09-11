//go:build !windows

package db_lib

import "strings"

// setEnvVar sets a key-value pair in the env map.
// On non-Windows systems, keys are case-sensitive so no deduplication is needed.
func setEnvVar(envMap map[string]string, k, v string) {
	envMap[k] = v
}

// hasEnvVar checks if an env slice "KEY=VAL" contains the specified key (case-sensitive on POSIX).
func hasEnvVar(env []string, key string) bool {
	for _, e := range env {
		k, _, _ := strings.Cut(e, "=")
		if k == key {
			return true
		}
	}
	return false
}

// appendPlatformEnv appends platform-specific ambient environment variables if not already present.
// On POSIX systems, no extra ambient variables are required.
func appendPlatformEnv(env *[]string) {
}
