//go:build windows

package db_lib

import (
	"fmt"
	"os"
	"strings"
)

// setEnvVar sets a key-value pair in the env map with case-insensitive deduplication.
// On Windows, environment variable names are case-insensitive, so any existing entry
// with the same name (regardless of case) is removed before inserting the new value.
func setEnvVar(envMap map[string]string, k, v string) {
	for existing := range envMap {
		if strings.EqualFold(existing, k) {
			delete(envMap, existing)
		}
	}
	envMap[k] = v
}

// hasEnvVar checks if an env slice "KEY=VAL" contains the specified key (case-insensitive on Windows).
func hasEnvVar(env []string, key string) bool {
	for _, e := range env {
		k, _, _ := strings.Cut(e, "=")
		if strings.EqualFold(k, key) {
			return true
		}
	}
	return false
}

// appendPlatformEnv appends Windows-specific ambient environment variables if not already present.
// If USERPROFILE is not already in the environment, it is added from the ambient environment.
func appendPlatformEnv(env *[]string) {
	if !hasEnvVar(*env, "USERPROFILE") {
		if up := os.Getenv("USERPROFILE"); up != "" {
			*env = append(*env, fmt.Sprintf("USERPROFILE=%s", up))
		}
	}
}
