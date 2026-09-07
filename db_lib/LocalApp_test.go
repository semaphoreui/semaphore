package db_lib

import (
	"os"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
)

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.HasPrefix(s, item) {
			return true
		}
	}
	return false
}

func TestGetEnvironmentVars(t *testing.T) {
	t.Setenv("SEMAPHORE_TEST", "test123")
	t.Setenv("SEMAPHORE_TEST2", "test222")
	t.Setenv("PASSWORD", "test222")
	t.Setenv("HTTP_PROXY", "http://proxy.corp:8080")
	t.Setenv("NO_PROXY", "localhost,.domain.com")
	t.Setenv("SSL_CERT_FILE", "/etc/ssl/certs/custom-ca.pem")
	t.Setenv("GIT_SSL_NO_VERIFY", "true")

	util.Config = &util.ConfigType{
		ForwardedEnvVars: []string{"SEMAPHORE_TEST"},
		EnvVars: map[string]string{
			"ANSIBLE_FORCE_COLOR": "False",
			"HTTP_PROXY":          "http://override.proxy:9090",
		},
	}

	res := getEnvironmentVars()

	expectedSubstrings := []string{
		"SEMAPHORE_TEST=test123",
		"ANSIBLE_FORCE_COLOR=False",
		"PATH=",
		"HTTP_PROXY=http://override.proxy:9090",
		"NO_PROXY=localhost,.domain.com",
		"SSL_CERT_FILE=/etc/ssl/certs/custom-ca.pem",
	}

	for _, e := range expectedSubstrings {
		if !contains(res, e) {
			t.Errorf("Expected result to contain %v, but got %v", e, res)
		}
	}

	// Ambient value must be overridden, not duplicated
	if contains(res, "HTTP_PROXY=http://proxy.corp:8080") {
		t.Errorf("HTTP_PROXY should have been overridden by Config.EnvVars, but found ambient value in %v", res)
	}

	// Verify no duplicate keys exist
	seenKeys := make(map[string]bool)
	for _, envStr := range res {
		parts := strings.SplitN(envStr, "=", 2)
		key := parts[0]
		if seenKeys[key] {
			t.Errorf("Duplicate environment variable key found: %s", key)
		}
		seenKeys[key] = true
	}

	// Unforwarded variables must not be leaked
	if contains(res, "PASSWORD=") {
		t.Errorf("PASSWORD should not be in environment variables, got %v", res)
	}
	if contains(res, "SEMAPHORE_TEST2=") {
		t.Errorf("SEMAPHORE_TEST2 should not be in environment variables without being in ForwardedEnvVars, got %v", res)
	}
	// Security-sensitive GIT_SSL_NO_VERIFY must not be forwarded by default
	if contains(res, "GIT_SSL_NO_VERIFY=") {
		t.Errorf("GIT_SSL_NO_VERIFY should not be forwarded automatically by default, got %v", res)
	}

	// Explicit forwarding of GIT_SSL_NO_VERIFY when configured
	util.Config.ForwardedEnvVars = []string{"SEMAPHORE_TEST", "GIT_SSL_NO_VERIFY"}
	resExplicit := getEnvironmentVars()
	if !contains(resExplicit, "GIT_SSL_NO_VERIFY=true") {
		t.Errorf("GIT_SSL_NO_VERIFY should be included when explicitly listed in ForwardedEnvVars, got %v", resExplicit)
	}
}

func TestGetEnvironmentVars_WindowsMixedCaseAndOSParity(t *testing.T) {
	// 1. Test mixed-case proxy and system variables commonly found on Windows/POSIX
	t.Setenv("HTTP_PROXY", "http://upper.proxy:8080")
	t.Setenv("http_proxy", "http://lower.proxy:8080")
	t.Setenv("SYSTEMROOT", `C:\Windows`)
	t.Setenv("SystemRoot", `C:\Windows`)
	t.Setenv("WINDIR", `C:\Windows`)
	t.Setenv("APPDATA", `C:\Users\test\AppData\Roaming`)
	t.Setenv("LOCALAPPDATA", `C:\Users\test\AppData\Local`)
	t.Setenv("COMSPEC", `C:\Windows\system32\cmd.exe`)
	t.Setenv("PATHEXT", `.COM;.EXE;.BAT;.CMD`)
	t.Setenv("USERNAME", "TestRunner")
	t.Setenv("TMP", `C:\Temp`)
	t.Setenv("TEMP", `C:\Temp`)

	util.Config = &util.ConfigType{
		ForwardedEnvVars: []string{},
		EnvVars: map[string]string{
			"HTTP_PROXY": "http://override.proxy:9090",
		},
	}

	res := getEnvironmentVars()

	// Config.EnvVars must override HTTP_PROXY
	if !contains(res, "HTTP_PROXY=http://override.proxy:9090") {
		t.Errorf("Expected HTTP_PROXY override, got %v", res)
	}

	// Windows system variables must be present
	expectedVars := []string{
		`SYSTEMROOT=C:\Windows`,
		`WINDIR=C:\Windows`,
		`APPDATA=C:\Users\test\AppData\Roaming`,
		`COMSPEC=C:\Windows\system32\cmd.exe`,
		`PATHEXT=.COM;.EXE;.BAT;.CMD`,
		`USERNAME=TestRunner`,
		`TMP=C:\Temp`,
		`TEMP=C:\Temp`,
	}

	for _, ev := range expectedVars {
		if !contains(res, ev) {
			t.Errorf("Expected result to contain %s, got %v", ev, res)
		}
	}
}

func TestGetHomeDir(t *testing.T) {
	repo := db.Repository{
		ProjectID: 42,
	}
	templateID := 114

	// Set a known HOME value for testing
	originalHome := os.Getenv("HOME")
	testHome := "/home/testuser"
	os.Setenv("HOME", testHome) //nolint:errcheck
	defer os.Setenv("HOME", originalHome) //nolint:errcheck

	// Save original config and restore after all tests
	originalConfig := util.Config
	defer func() { util.Config = originalConfig }()

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
			expectedHome: testHome,
			description:  "Should return real user HOME",
		},
		{
			name:         "UserHome mode",
			homeDirMode:  util.HomeDirModeUserHome,
			tmpPath:      "/tmp/semaphore",
			expectedHome: testHome,
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
			// Setup config for this test case
			util.Config = &util.ConfigType{
				HomeDirMode: tt.homeDirMode,
				TmpPath:     tt.tmpPath,
			}

			// Call getHomeDir
			result := getHomeDir(repo, templateID)

			// Verify the result
			if result != tt.expectedHome {
				t.Errorf("%s: expected HOME=%s, got HOME=%s",
					tt.description, tt.expectedHome, result)
			}
		})
	}
}
