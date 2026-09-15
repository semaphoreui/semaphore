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

func TestHasEnvVar_POSIXCaseSensitive(t *testing.T) {
	env := []string{"PATH=/usr/bin", "HOME=/home/test"}

	assert.True(t, hasEnvVar(env, "PATH"))
	assert.False(t, hasEnvVar(env, "path"))
	assert.False(t, hasEnvVar(env, "Path"))
	assert.True(t, hasEnvVar(env, "HOME"))
	assert.False(t, hasEnvVar(env, "home"))
}

func TestAppendPlatformEnv_POSIXNoop(t *testing.T) {
	env := []string{"PATH=/usr/bin"}
	appendPlatformEnv(&env)
	assert.Len(t, env, 1)
	assert.Equal(t, "PATH=/usr/bin", env[0])
}
