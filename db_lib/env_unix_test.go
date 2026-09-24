//go:build !windows

package db_lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetEnvVar_POSIXCaseSensitive(t *testing.T) {
	envMap := make(map[string]string)
	setEnvVar(envMap, "Path", "/custom/bin")
	setEnvVar(envMap, "PATH", "/usr/bin")

	// On POSIX, environment variable keys are case-sensitive
	assert.Len(t, envMap, 2)
	assert.Equal(t, "/custom/bin", envMap["Path"])
	assert.Equal(t, "/usr/bin", envMap["PATH"])
}

func TestHasNonEmptyEnvVar_POSIXCaseSensitive(t *testing.T) {
	env := []string{"PATH=/usr/bin", "HOME=/home/test"}

	assert.True(t, hasNonEmptyEnvVar(env, "PATH"))
	assert.False(t, hasNonEmptyEnvVar(env, "path"))
	assert.False(t, hasNonEmptyEnvVar(env, "Path"))
	assert.True(t, hasNonEmptyEnvVar(env, "HOME"))
	assert.False(t, hasNonEmptyEnvVar(env, "home"))
}

func TestHasNonEmptyEnvVar_EmptyValueCountsAsAbsent(t *testing.T) {
	assert.False(t, hasNonEmptyEnvVar([]string{"HOME="}, "HOME"))
	assert.False(t, hasNonEmptyEnvVar([]string{"HOME"}, "HOME"))
	assert.True(t, hasNonEmptyEnvVar([]string{"HOME=", "HOME=/home/test"}, "HOME"))
}

func TestAppendPlatformEnv_POSIXNoop(t *testing.T) {
	env := []string{"PATH=/usr/bin"}
	appendPlatformEnv(&env)
	assert.Len(t, env, 1)
	assert.Equal(t, "PATH=/usr/bin", env[0])
}
