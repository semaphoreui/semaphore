//go:build !windows

package db_lib

// setEnvVar sets a key-value pair in the env map.
// On non-Windows systems, keys are case-sensitive so no deduplication is needed.
func setEnvVar(envMap map[string]string, k, v string) {
	envMap[k] = v
}
