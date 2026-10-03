//go:build !windows

package db_lib

import "strings"

// setEnvVar sets a key-value pair in the env map.
// On non-Windows systems, keys are case-sensitive so no deduplication is needed.
func setEnvVar(envMap map[string]string, k, v string) {
	envMap[k] = v
}

// hasNonEmptyEnvVar reports whether an env slice "KEY=VAL" holds the given key
// with a value (case-sensitive on POSIX). An empty value counts as absent: it
// is only ever used for HOME and USERPROFILE, where "HOME=" hides the user's
// git configuration just as effectively as not setting it at all.
func hasNonEmptyEnvVar(env []string, key string) bool {
	for _, e := range env {
		k, v, _ := strings.Cut(e, "=")
		if k == key && v != "" {
			return true
		}
	}
	return false
}

// appendPlatformEnv appends platform-specific ambient environment variables if not already present.
// On POSIX systems, no extra ambient variables are required.
func appendPlatformEnv(env *[]string) {
}
