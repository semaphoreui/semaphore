//go:build windows

package db_lib

import "strings"

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
