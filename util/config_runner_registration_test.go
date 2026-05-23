package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigInit_LoadsRunnerRegistrationTokenFromFile(t *testing.T) {
	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "registration.txt")
	if err := os.WriteFile(tokenPath, []byte("reg-token-from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("SEMAPHORE_RUNNER_REGISTRATION_TOKEN_FILE", tokenPath)
	t.Setenv("SEMAPHORE_RUNNER_REGISTRATION_TOKEN", "")

	ConfigInit("", true)

	if Config.RunnerRegistrationToken != "reg-token-from-file" {
		t.Fatalf("RunnerRegistrationToken: got %q want %q", Config.RunnerRegistrationToken, "reg-token-from-file")
	}
	if Config.Runner == nil || Config.Runner.RegistrationToken != "reg-token-from-file" {
		got := ""
		if Config.Runner != nil {
			got = Config.Runner.RegistrationToken
		}
		t.Fatalf("Runner.RegistrationToken: got %q want %q", got, "reg-token-from-file")
	}
}

func TestConfigInit_InlineRunnerRegistrationTokenNotOverriddenByFile(t *testing.T) {
	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "registration.txt")
	if err := os.WriteFile(tokenPath, []byte("from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("SEMAPHORE_RUNNER_REGISTRATION_TOKEN_FILE", tokenPath)
	t.Setenv("SEMAPHORE_RUNNER_REGISTRATION_TOKEN", "from-env")

	ConfigInit("", true)

	if Config.RunnerRegistrationToken != "from-env" {
		t.Fatalf("RunnerRegistrationToken: got %q want from env", Config.RunnerRegistrationToken)
	}
	if Config.Runner == nil || Config.Runner.RegistrationToken != "from-env" {
		got := ""
		if Config.Runner != nil {
			got = Config.Runner.RegistrationToken
		}
		t.Fatalf("Runner.RegistrationToken: got %q want from env", got)
	}
}
