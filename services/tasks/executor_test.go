package tasks

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveGitBranch(t *testing.T) {
	tests := []struct {
		name       string
		repoBranch string
		template   db.Template
		task       db.Task
		expected   string
	}{
		{
			name:       "no overrides falls back to repository branch",
			repoBranch: "main",
			template:   db.Template{},
			task:       db.Task{},
			expected:   "main",
		},
		{
			name:       "template branch overrides repository branch",
			repoBranch: "main",
			template:   db.Template{GitBranch: new("release")},
			task:       db.Task{},
			expected:   "release",
		},
		{
			name:       "task branch ignored when override not allowed",
			repoBranch: "main",
			template:   db.Template{AllowOverrideBranchInTask: false},
			task:       db.Task{GitBranch: new("attacker")},
			expected:   "main",
		},
		{
			name:       "task branch ignored over pinned template branch when not allowed",
			repoBranch: "main",
			template:   db.Template{GitBranch: new("release"), AllowOverrideBranchInTask: false},
			task:       db.Task{GitBranch: new("attacker")},
			expected:   "release",
		},
		{
			name:       "task branch applied when override allowed",
			repoBranch: "main",
			template:   db.Template{AllowOverrideBranchInTask: true},
			task:       db.Task{GitBranch: new("feature")},
			expected:   "feature",
		},
		{
			name:       "empty task branch ignored even when override allowed",
			repoBranch: "main",
			template:   db.Template{GitBranch: new("release"), AllowOverrideBranchInTask: true},
			task:       db.Task{GitBranch: new("")},
			expected:   "release",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, resolveGitBranch(tt.repoBranch, tt.template, tt.task))
		})
	}
}

func TestWithEffectiveBranch(t *testing.T) {
	repository := db.Repository{GitBranch: "main"}
	template := db.Template{AllowOverrideBranchInTask: true}
	task := db.Task{GitBranch: new("release")}

	effective := withEffectiveBranch(repository, template, task)

	assert.Equal(t, "release", effective.GitBranch)
	assert.Equal(t, "main", repository.GitBranch)
}

// A monorepo is shared on the repository's own branch, so neither a branch the
// template still carries nor a task-supplied branch moves it.
func TestWithEffectiveBranch_MonorepoIgnoresOverrides(t *testing.T) {
	repository := db.Repository{GitBranch: "main", Monorepo: true}
	template := db.Template{GitBranch: new("release"), AllowOverrideBranchInTask: true}
	task := db.Task{GitBranch: new("feature")}

	assert.Equal(t, "main", withEffectiveBranch(repository, template, task).GitBranch)

	repository.Monorepo = false
	assert.Equal(t, "feature", withEffectiveBranch(repository, template, task).GitBranch)
}

// Each task of a Subversion monorepo runs in its own directory; a git monorepo
// and a Subversion repository checked out per template keep running in the
// checkout.
func TestWithEffectiveBranch_SvnMonorepoRunsInItsOwnDir(t *testing.T) {
	previous := util.Config
	util.Config = &util.ConfigType{TmpPath: t.TempDir()}
	t.Cleanup(func() { util.Config = previous })

	svn := db.Repository{ID: 7, ProjectID: 1, GitURL: "svn://svn.example.com/repo", GitBranch: "trunk", Monorepo: true}

	one := withEffectiveBranch(svn, db.Template{}, db.Task{ID: 1})
	two := withEffectiveBranch(svn, db.Template{}, db.Task{ID: 2})
	assert.Equal(t, svn.GetTaskRunPath(1), one.RunDir)
	assert.NotEqual(t, one.RunDir, two.RunDir)
	assert.Equal(t, one.GetCheckoutPath(5), two.GetCheckoutPath(6), "they still share the checkout")

	perTemplate := svn
	perTemplate.Monorepo = false
	assert.Empty(t, withEffectiveBranch(perTemplate, db.Template{}, db.Task{ID: 1}).RunDir)

	git := svn
	git.GitURL = "https://example.com/x.git"
	assert.Empty(t, withEffectiveBranch(git, db.Template{}, db.Task{ID: 1}).RunDir)
}

// The task's copy is made from the shared working copy and removed when the
// task ends; the working copy stays for the next task.
func TestLocalExecutor_TaskRunDirLifecycle(t *testing.T) {
	for _, bin := range []string{"svn", "svnadmin"} {
		if _, err := exec.LookPath(bin); err != nil {
			t.Skip(bin + " is not installed")
		}
	}

	previous := util.Config
	util.Config = &util.ConfigType{TmpPath: t.TempDir()}
	t.Cleanup(func() { util.Config = previous })

	root := filepath.Join(t.TempDir(), "repo")
	runCmd(t, "", "svnadmin", "create", root)
	url := "file://" + filepath.ToSlash(root)
	runCmd(t, "", "svn", "mkdir", "--non-interactive", "-m", "layout", url+"/trunk")
	runCmd(t, "", "svn", "import", "--non-interactive", "-m", "site", writeTemp(t, "site.yml", "v1"), url+"/trunk/site.yml")

	// A file:// URL is not routed to the svn client, so the run directory is
	// set by hand here; withEffectiveBranch is covered above.
	repository := db.Repository{ID: 7, ProjectID: 1, GitURL: url, GitBranch: "trunk", Monorepo: true}
	repository.RunDir = repository.GetTaskRunPath(42)
	checkout := repository.GetCheckoutPath(3)
	require.NoError(t, os.MkdirAll(filepath.Dir(checkout), 0755))
	runCmd(t, "", "svn", "checkout", "--non-interactive", url+"/trunk", checkout)

	executor := &LocalExecutor{
		Template:     db.Template{ID: 3},
		Repository:   repository,
		Logger:       task_logger.NopLogger{},
		KeyInstaller: &KeyInstallerMock{},
	}

	require.NoError(t, executor.copyRepositoryToRunDir())
	content, err := os.ReadFile(filepath.Join(repository.RunDir, "site.yml"))
	require.NoError(t, err)
	assert.Equal(t, "v1", string(content))
	assert.NoDirExists(t, filepath.Join(repository.RunDir, ".svn"))

	executor.Cleanup()

	assert.NoDirExists(t, repository.RunDir)
	assert.DirExists(t, filepath.Join(checkout, ".svn"))
}

func runCmd(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(file, []byte(content), 0644))
	return file
}

// TestLocalExecutorImplementsExecutor is a compile-time guard: if LocalExecutor stops
// satisfying the Executor interface, this file fails to build. It documents the
// contract introduced in Phase 1 of the Kubernetes runner refactor.
func TestLocalExecutorImplementsExecutor(t *testing.T) {
	var _ Executor = (*LocalExecutor)(nil)
	var _ Job = (*LocalExecutor)(nil)
}

func TestLocalExecutorCleanup_ZeroValueSafe(t *testing.T) {
	// Cleanup runs in a defer after a potentially-failed Prepare, so it must tolerate
	// being called on a barely-initialized executor — in particular App may be nil
	// when Prepare didn't even get that far.
	exec := &LocalExecutor{}

	assert.NotPanics(t, func() {
		exec.Cleanup()
	})
}

func TestLocalExecutorIsKilled(t *testing.T) {
	exec := &LocalExecutor{}
	assert.False(t, exec.IsKilled(), "fresh executor should not have a kill request")

	exec.Kill() // Process is nil so this only flips the flag
	assert.True(t, exec.IsKilled(), "Kill() must record the kill request")
}
