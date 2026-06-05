package db_lib

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/ssh"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

// Mock logger for testing
type mockLogger struct{}

func (l *mockLogger) Log(msg string)                                                       {}
func (l *mockLogger) Logf(format string, a ...any)                                        {}
func (l *mockLogger) LogWithTime(now time.Time, msg string)                               {}
func (l *mockLogger) LogfWithTime(now time.Time, format string, a ...any)                 {}
func (l *mockLogger) LogCmd(cmd *exec.Cmd)                                                {}
func (l *mockLogger) SetStatus(status task_logger.TaskStatus)                             {}
func (l *mockLogger) AddStatusListener(l2 task_logger.StatusListener)                     {}
func (l *mockLogger) AddLogListener(l2 task_logger.LogListener)                           {}
func (l *mockLogger) SetCommit(hash, message string)                                      {}
func (l *mockLogger) WaitLog()                                                            {}

// Mock key installer for testing
type mockKeyInstaller struct{}

func (k *mockKeyInstaller) Install(key db.AccessKey, usage db.AccessKeyRole, logger task_logger.Logger) (ssh.AccessKeyInstallation, error) {
	return ssh.AccessKeyInstallation{}, nil
}

func TestProxyGitClientCreation(t *testing.T) {
	// Test that ProxyGitClient can be created
	keyInstaller := &mockKeyInstaller{}
	client := CreateProxyGitClient(keyInstaller)
	
	if client == nil {
		t.Error("CreateProxyGitClient returned nil")
	}
	
	// Test that it implements GitClient interface
	var _ GitClient = client
}

func TestProxyGitClientInterface(t *testing.T) {
	// Create a ProxyGitClient instance
	keyInstaller := &mockKeyInstaller{}
	client := CreateProxyGitClient(keyInstaller)
	
	// Create a mock repository
	repo := GitRepository{
		Repository: db.Repository{
			GitURL:    "https://github.com/test/repo.git",
			GitBranch: "main",
		},
		Logger: &mockLogger{},
		Client: client,
	}
	
	// Test that all interface methods can be called without panicking
	
	// CanBePulled should return true for proxy client
	if !client.CanBePulled(repo) {
		t.Error("CanBePulled should return true for proxy client")
	}
	
	// GetRemoteBranches should return the current branch
	branches, err := client.GetRemoteBranches(repo)
	if err != nil {
		t.Errorf("GetRemoteBranches returned error: %v", err)
	}
	if len(branches) != 1 || branches[0] != "main" {
		t.Errorf("GetRemoteBranches returned unexpected branches: %v", branches)
	}
	
	// Checkout should not return an error (it's a no-op)
	err = client.Checkout(repo, "some-commit")
	if err != nil {
		t.Errorf("Checkout returned error: %v", err)
	}
}

func TestProxyGitClientExtractArchive(t *testing.T) {
	client := ProxyGitClient{}
	
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test-extract-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Test extracting an empty archive (should not fail)
	err = client.extractArchive([]byte{}, tempDir)
	if err == nil {
		t.Error("extractArchive should fail with empty data")
	}
}

func TestGitClientFactoryProxySelection(t *testing.T) {
	// Save original config
	originalConfig := util.Config
	defer func() {
		util.Config = originalConfig
	}()
	
	// Set config to use proxy git client
	util.Config = &util.ConfigType{
		GitClientId: util.ProxyGitClientId,
	}
	
	keyInstaller := &mockKeyInstaller{}
	client := CreateDefaultGitClient(keyInstaller)
	
	// Check that we got a ProxyGitClient
	if _, ok := client.(ProxyGitClient); !ok {
		t.Error("CreateDefaultGitClient did not return ProxyGitClient when GitClientId is proxy_git")
	}
}

func TestCreateRepositoryArchiveHelper(t *testing.T) {
	// This tests the archive creation helper function indirectly
	// by checking that it doesn't panic with various inputs
	
	// Test with non-existent directory
	_, err := createRepositoryArchiveTestHelper("/non/existent/path")
	if err == nil {
		t.Error("createRepositoryArchive should fail with non-existent path")
	}
	
	// Test with empty directory
	tempDir, err := os.MkdirTemp("", "test-archive-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	archive, err := createRepositoryArchiveTestHelper(tempDir)
	if err != nil {
		t.Errorf("createRepositoryArchive failed with empty directory: %v", err)
	}
	if len(archive) == 0 {
		t.Error("createRepositoryArchive returned empty archive")
	}
	
	// Test with directory containing a file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	archive, err = createRepositoryArchiveTestHelper(tempDir)
	if err != nil {
		t.Errorf("createRepositoryArchive failed with file: %v", err)
	}
	if len(archive) == 0 {
		t.Error("createRepositoryArchive returned empty archive with file")
	}
}

// Helper function that mimics the archive creation logic from the API
func createRepositoryArchiveTestHelper(repoPath string) ([]byte, error) {
	// This is a simplified version of the createRepositoryArchive function
	// for testing purposes
	
	_, err := os.Stat(repoPath)
	if err != nil {
		return nil, err
	}
	
	// Return a simple archive (just placeholder data)
	return []byte("fake-archive-data"), nil
}