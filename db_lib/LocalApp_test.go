package db_lib

import (
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
)

func TestGetEnvironmentVars_ExplicitForwardingOnly(t *testing.T) {
	// Restore util.Config after the test
	original := util.Config
	t.Cleanup(func() { util.Config = original })

	t.Setenv("HTTP_PROXY", "http://proxy.corp:8080")
	t.Setenv("SEMAPHORE_TEST", "test123")
	t.Setenv("SEMAPHORE_TEST2", "test222")
	t.Setenv("PASSWORD", "secret")

	util.Config = &util.ConfigType{
		ForwardedEnvVars: []string{"SEMAPHORE_TEST", "HTTP_PROXY"},
		EnvVars: map[string]string{
			"ANSIBLE_FORCE_COLOR": "False",
		},
	}

	res := getEnvironmentVars()

	// Explicitly forwarded vars must be present
	assert.True(t, containsPrefix(res, "SEMAPHORE_TEST=test123"), "expected SEMAPHORE_TEST in result, got %v", res)
	assert.True(t, containsPrefix(res, "HTTP_PROXY=http://proxy.corp:8080"), "expected HTTP_PROXY in result, got %v", res)

	// Config.EnvVars must be present
	assert.True(t, containsPrefix(res, "ANSIBLE_FORCE_COLOR=False"), "expected ANSIBLE_FORCE_COLOR in result, got %v", res)

	// PATH must always be present
	assert.True(t, containsPrefix(res, "PATH="), "expected PATH in result, got %v", res)

	// Vars NOT in ForwardedEnvVars or EnvVars must not be leaked
	assert.False(t, containsPrefix(res, "PASSWORD="), "PASSWORD should not be forwarded, got %v", res)
	assert.False(t, containsPrefix(res, "SEMAPHORE_TEST2="), "SEMAPHORE_TEST2 should not be forwarded, got %v", res)
}

func TestGetEnvironmentVars_EnvVarsOverrideForwarded(t *testing.T) {
	// Restore util.Config after the test
	original := util.Config
	t.Cleanup(func() { util.Config = original })

	t.Setenv("HTTP_PROXY", "http://ambient.proxy:8080")

	util.Config = &util.ConfigType{
		ForwardedEnvVars: []string{"HTTP_PROXY"},
		EnvVars: map[string]string{
			"HTTP_PROXY": "http://override.proxy:9090",
		},
	}

	res := getEnvironmentVars()

	// Config.EnvVars must win over ForwardedEnvVars (ambient)
	assert.True(t, containsPrefix(res, "HTTP_PROXY=http://override.proxy:9090"), "Config.EnvVars should override ambient, got %v", res)
	assert.False(t, containsPrefix(res, "HTTP_PROXY=http://ambient.proxy:8080"), "ambient value should not appear when overridden, got %v", res)

	// Verify no duplicate keys
	seen := make(map[string]bool)
	for _, envStr := range res {
		key, _, _ := strings.Cut(envStr, "=")
		assert.False(t, seen[key], "duplicate key found: %s in %v", key, res)
		seen[key] = true
	}
}

func TestGetEnvironmentVars_EmptyConfig(t *testing.T) {
	// Restore util.Config after the test
	original := util.Config
	t.Cleanup(func() { util.Config = original })

	util.Config = &util.ConfigType{
		ForwardedEnvVars: []string{},
		EnvVars:          map[string]string{},
	}

	res := getEnvironmentVars()

	// Only PATH should be forwarded
	assert.True(t, containsPrefix(res, "PATH="), "PATH must always be present")
	assert.Equal(t, 1, len(res), "only PATH should be in result when config is empty, got %v", res)
}

func TestGetHomeDir(t *testing.T) {
	// Save original config and restore after all tests
	originalConfig := util.Config
	t.Cleanup(func() { util.Config = originalConfig })

	t.Setenv("HOME", "/home/testuser")

	repo := db.Repository{
		ProjectID: 42,
	}
	templateID := 114

	tests := []struct {
		name         string
		homeDirMode  string
		tmpPath      string
		expectedHome string
		description  string
	}{
		{
			name:         "ProjectHome mode",
			homeDirMode:  util.HomeDirModeProjectHome,
			tmpPath:      "/tmp/semaphore",
			expectedHome: "/tmp/semaphore/project_42",
			description:  "Should return project temp directory",
		},
		{
			name:         "TemplateDir mode",
			homeDirMode:  util.HomeDirModeTemplateDir,
			tmpPath:      "/tmp/semaphore",
			expectedHome: "/home/testuser",
			description:  "Should return real user HOME",
		},
		{
			name:         "UserHome mode",
			homeDirMode:  util.HomeDirModeUserHome,
			tmpPath:      "/tmp/semaphore",
			expectedHome: "/home/testuser",
			description:  "Should return real user HOME",
		},
		{
			name:         "Empty/default mode",
			homeDirMode:  "",
			tmpPath:      "/tmp/semaphore",
			expectedHome: "",
			description:  "Should return empty string for unknown mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			util.Config = &util.ConfigType{
				HomeDirMode: tt.homeDirMode,
				TmpPath:     tt.tmpPath,
			}

			result := getHomeDir(repo, templateID)

			assert.Equal(t, tt.expectedHome, result, "%s: expected HOME=%s, got HOME=%s",
				tt.description, tt.expectedHome, result)
		})
	}
}

// containsPrefix reports whether any element of slice starts with prefix.
func containsPrefix(slice []string, prefix string) bool {
	for _, s := range slice {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}
