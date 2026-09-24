//go:build windows

package db_lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetEnvVar_WindowsCaseInsensitiveDeduplication(t *testing.T) {
	envMap := make(map[string]string)

	// Seed initial mixed-case entry
	setEnvVar(envMap, "Path", `C:\Windows`)
	assert.Equal(t, `C:\Windows`, envMap["Path"])

	// Apply uppercase PATH override; must deduplicate and keep only the latest key
	setEnvVar(envMap, "PATH", `C:\Windows;C:\Program Files`)
	assert.Len(t, envMap, 1, "expected exactly one key/value in envMap")
	assert.Equal(t, `C:\Windows;C:\Program Files`, envMap["PATH"])
	assert.NotContains(t, envMap, "Path")

	// Apply lowercase path override
	setEnvVar(envMap, "path", `C:\Custom`)
	assert.Len(t, envMap, 1, "expected exactly one key/value in envMap")
	assert.Equal(t, `C:\Custom`, envMap["path"])
	assert.NotContains(t, envMap, "PATH")

	// Verify case-insensitive override on proxy variables as well
	setEnvVar(envMap, "http_proxy", "http://ambient:8080")
	setEnvVar(envMap, "HTTP_PROXY", "http://override:9090")
	assert.Len(t, envMap, 2, "expected path and HTTP_PROXY")
	assert.Equal(t, "http://override:9090", envMap["HTTP_PROXY"])
	assert.NotContains(t, envMap, "http_proxy")
}

func TestHasNonEmptyEnvVar_WindowsCaseInsensitive(t *testing.T) {
	env := []string{`Path=C:\Windows`, `USERPROFILE=C:\Users\test`}

	assert.True(t, hasNonEmptyEnvVar(env, "PATH"))
	assert.True(t, hasNonEmptyEnvVar(env, "path"))
	assert.True(t, hasNonEmptyEnvVar(env, "Path"))
	assert.True(t, hasNonEmptyEnvVar(env, "userprofile"))
	assert.True(t, hasNonEmptyEnvVar(env, "USERPROFILE"))
	assert.False(t, hasNonEmptyEnvVar(env, "NONEXISTENT"))
}

func TestHasNonEmptyEnvVar_WindowsEmptyValueCountsAsAbsent(t *testing.T) {
	assert.False(t, hasNonEmptyEnvVar([]string{"userprofile="}, "USERPROFILE"))
	assert.True(t, hasNonEmptyEnvVar([]string{`userprofile=C:\Users\test`}, "USERPROFILE"))
}

func TestAppendPlatformEnv_WindowsUserProfile(t *testing.T) {
	t.Setenv("USERPROFILE", `C:\Users\AmbientUser`)

	// 1. When USERPROFILE is absent, appendPlatformEnv adds it
	var env []string
	appendPlatformEnv(&env)
	assert.Contains(t, env, `USERPROFILE=C:\Users\AmbientUser`)

	// 2. When userprofile is already present (even in lowercase), ambient fallback is not appended
	envWithProfile := []string{`userprofile=C:\Users\ExplicitUser`}
	appendPlatformEnv(&envWithProfile)
	assert.Len(t, envWithProfile, 1)
	assert.Equal(t, `userprofile=C:\Users\ExplicitUser`, envWithProfile[0])
}
