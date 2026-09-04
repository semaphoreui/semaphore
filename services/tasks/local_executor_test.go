package tasks

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupExecutorConfig(t *testing.T) {
	t.Helper()
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })
	if util.Config == nil {
		util.Config = &util.ConfigType{}
	}
}

// TestGetShellArgs_PassesSurveySecretVar verifies that Survey variables of type
// "Secret" (delivered via LocalExecutor.Secret) are passed to Bash/Shell tasks,
// alongside plain survey vars that arrive in Environment.JSON.
func TestGetShellArgs_PassesSurveySecretVar(t *testing.T) {
	setupExecutorConfig(t)

	exec := &LocalExecutor{
		Template: db.Template{
			Type:     db.TemplateTask,
			Playbook: "date.sh",
		},
		Environment: db.Environment{
			JSON: `{"PLAIN_VAR":"hello"}`,
		},
		Secret: `{"MY_VAR":"s3cr3t"}`,
	}

	args, err := exec.getShellArgs("admin", nil)
	require.NoError(t, err)

	assert.Contains(t, args, "PLAIN_VAR=hello", "plain survey var must be passed")
	assert.Contains(t, args, "MY_VAR=s3cr3t", "secret survey var must be passed")
}

// TestGetEnvironmentExtraVars_MergesSecret checks the shared helper merges the
// Secret field into the extra vars map used by Shell and Terraform tasks.
func TestGetEnvironmentExtraVars_MergesSecret(t *testing.T) {
	setupExecutorConfig(t)

	exec := &LocalExecutor{
		Environment: db.Environment{JSON: `{"PLAIN_VAR":"hello"}`},
		Secret:      `{"MY_VAR":"s3cr3t"}`,
	}

	extraVars, err := exec.getEnvironmentExtraVars("admin", nil)
	require.NoError(t, err)

	assert.Equal(t, "hello", extraVars["PLAIN_VAR"])
	assert.Equal(t, "s3cr3t", extraVars["MY_VAR"])
}

// TestGetEnvironmentExtraVars_EmptySecret pins that "" and "{}" are equivalent
// "no survey secrets" values: "" comes from DB-loaded tasks and API clients
// omitting the field, "{}" from the UI and the AddTask sanitization.
func TestGetEnvironmentExtraVars_EmptySecret(t *testing.T) {
	setupExecutorConfig(t)

	for _, secret := range []string{"", "{}"} {
		t.Run("secret "+secret, func(t *testing.T) {
			exec := &LocalExecutor{
				Environment: db.Environment{JSON: `{"PLAIN_VAR":"hello"}`},
				Secret:      secret,
			}

			extraVars, err := exec.getEnvironmentExtraVars("admin", nil)
			require.NoError(t, err)

			assert.Equal(t, "hello", extraVars["PLAIN_VAR"])
			assert.Len(t, extraVars, 2) // PLAIN_VAR + semaphore_vars only
		})
	}
}

// TestGetEnvironmentExtraVars_SkipsEnvTargetVars verifies that survey vars with
// Target "env" are excluded from the extra-vars map (they must not reach
// --extra-vars / -var / shell CLI args).
func TestGetEnvironmentExtraVars_SkipsEnvTargetVars(t *testing.T) {
	setupExecutorConfig(t)

	exec := &LocalExecutor{
		Template: db.Template{
			SurveyVars: []db.SurveyVar{
				{Name: "ENV_VAR", Target: db.SurveyVarTargetEnv},
				{Name: "CLI_VAR"},
			},
		},
		Environment: db.Environment{JSON: `{"ENV_VAR":"via-env","CLI_VAR":"via-cli"}`},
		Secret:      `{"SECRET_ENV_VAR":"s3cr3t"}`,
	}
	exec.Template.SurveyVars = append(exec.Template.SurveyVars,
		db.SurveyVar{Name: "SECRET_ENV_VAR", Type: "secret", Target: db.SurveyVarTargetEnv})

	extraVars, err := exec.getEnvironmentExtraVars("admin", nil)
	require.NoError(t, err)

	assert.NotContains(t, extraVars, "ENV_VAR")
	assert.NotContains(t, extraVars, "SECRET_ENV_VAR")
	assert.Equal(t, "via-cli", extraVars["CLI_VAR"])
}

// TestGetSurveyEnvVars verifies env-target survey vars are collected as
// NAME=value pairs from both Environment.JSON and the Secret field, and that
// CLI-target vars are ignored.
func TestGetSurveyEnvVars(t *testing.T) {
	setupExecutorConfig(t)

	exec := &LocalExecutor{
		Template: db.Template{
			SurveyVars: []db.SurveyVar{
				{Name: "ENV_VAR", Target: db.SurveyVarTargetEnv},
				{Name: "SECRET_ENV_VAR", Type: "secret", Target: db.SurveyVarTargetEnv},
				{Name: "CLI_VAR"},
				{Name: "MISSING_VAR", Target: db.SurveyVarTargetEnv}, // no value provided
			},
		},
		Environment: db.Environment{JSON: `{"ENV_VAR":"via-env","CLI_VAR":"via-cli"}`},
		Secret:      `{"SECRET_ENV_VAR":"s3cr3t"}`,
	}

	envVars, err := exec.getSurveyEnvVars()
	require.NoError(t, err)

	assert.Contains(t, envVars, "ENV_VAR=via-env")
	assert.Contains(t, envVars, "SECRET_ENV_VAR=s3cr3t")
	assert.Len(t, envVars, 2, "CLI-target and valueless vars must not be included")
}

// TestFormatVarValue verifies values from multi-select survey vars (arrays)
// are JSON-encoded instead of degrading to Go's fmt representation.
func TestFormatVarValue(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"string", "hello", "hello"},
		{"array", []any{"1", "2"}, `["1","2"]`},
		{"empty array", []any{}, `[]`},
		{"object", map[string]any{"k": "v"}, `{"k":"v"}`},
		{"number", float64(42), "42"},
		{"bool", true, "true"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatVarValue(tt.input))
		})
	}
}

// TestGetSurveyEnvVars_MultiSelect verifies env-target multi-select survey
// vars are delivered as JSON arrays, not Go-formatted slices like "[1 2]".
func TestGetSurveyEnvVars_MultiSelect(t *testing.T) {
	setupExecutorConfig(t)

	exec := &LocalExecutor{
		Template: db.Template{
			SurveyVars: []db.SurveyVar{
				{Name: "MULTI_VAR", Type: db.SurveyVarSelect, Target: db.SurveyVarTargetEnv},
			},
		},
		Environment: db.Environment{JSON: `{"MULTI_VAR":["1","2"]}`},
	}

	envVars, err := exec.getSurveyEnvVars()
	require.NoError(t, err)

	assert.Contains(t, envVars, `MULTI_VAR=["1","2"]`)
}

// TestGetShellArgs_MultiSelect verifies multi-select survey vars reach shell
// tasks as JSON arrays in KEY=value CLI args, not Go's "[1 2]" formatting.
func TestGetShellArgs_MultiSelect(t *testing.T) {
	setupExecutorConfig(t)

	exec := &LocalExecutor{
		Template: db.Template{
			Type:     db.TemplateTask,
			Playbook: "run.sh",
		},
		Environment: db.Environment{JSON: `{"multi_var":["1","2"]}`},
	}

	args, err := exec.getShellArgs("admin", nil)
	require.NoError(t, err)

	assert.Contains(t, args, `multi_var=["1","2"]`)
	assert.NotContains(t, args, "multi_var=[1 2]")
}

// TestGetTerraformArgs_MultiSelect verifies multi-select survey vars reach
// terraform as -var name=<JSON list>, which terraform parses as a list value.
func TestGetTerraformArgs_MultiSelect(t *testing.T) {
	setupExecutorConfig(t)

	exec := &LocalExecutor{
		Template: db.Template{
			Type: db.TemplateTask,
			App:  db.AppTerraform,
		},
		Environment: db.Environment{JSON: `{"multi_var":["1","2"]}`},
	}

	argsMap, err := exec.getTerraformArgs("admin", nil)
	require.NoError(t, err)

	defaultArgs := argsMap["default"]
	found := false
	for i, arg := range defaultArgs {
		if arg == "-var" && i+1 < len(defaultArgs) && defaultArgs[i+1] == `multi_var=["1","2"]` {
			found = true
		}
	}
	assert.True(t, found, "expected -var multi_var=[\"1\",\"2\"] in %v", defaultArgs)
}

// TestGetArgs_AnsibleForks verifies passing -f / --forks in template and task arguments,
// including invalid JSON error handling and AllowOverrideArgsInTask behavior.
func TestGetArgs_AnsibleForks(t *testing.T) {
	setupExecutorConfig(t)

	tests := []struct {
		name                  string
		templateArgs          *string
		allowOverride         bool
		taskArgs              *string
		expectedEffectiveFork string
		expectedForksSubArgs  []string
		mustNotContain        []string
		expectError           bool
		expectedErrorMsg      string
	}{
		{
			name:                  "Template arguments with --forks 10",
			templateArgs:          strPtr(`["--forks", "10"]`),
			allowOverride:         false,
			taskArgs:              nil,
			expectedEffectiveFork: "10",
			expectedForksSubArgs:  []string{"--forks", "10"},
		},
		{
			name:                  "Template arguments with -f 10",
			templateArgs:          strPtr(`["-f", "10"]`),
			allowOverride:         false,
			taskArgs:              nil,
			expectedEffectiveFork: "10",
			expectedForksSubArgs:  []string{"-f", "10"},
		},
		{
			name:                  "Task level overrides template forks when AllowOverrideArgsInTask is true",
			templateArgs:          strPtr(`["--forks", "5"]`),
			allowOverride:         true,
			taskArgs:              strPtr(`["--forks", "10"]`),
			expectedEffectiveFork: "10",
			expectedForksSubArgs:  []string{"--forks", "5", "--forks", "10"},
		},
		{
			name:                  "Task level ignored when AllowOverrideArgsInTask is false",
			templateArgs:          strPtr(`["--forks", "5"]`),
			allowOverride:         false,
			taskArgs:              strPtr(`["--forks", "10"]`),
			expectedEffectiveFork: "5",
			expectedForksSubArgs:  []string{"--forks", "5"},
			mustNotContain:        []string{"10"},
		},
		{
			name:             "Invalid JSON in template arguments returns descriptive error",
			templateArgs:     strPtr(`--forks 10`),
			allowOverride:    false,
			taskArgs:         nil,
			expectError:      true,
			expectedErrorMsg: "invalid format of the template extra arguments, must be valid JSON",
		},
		{
			name:             "Invalid JSON in task arguments returns descriptive error when AllowOverrideArgsInTask is true",
			templateArgs:     strPtr(`["--forks", "5"]`),
			allowOverride:    true,
			taskArgs:         strPtr(`invalid-json`),
			expectError:      true,
			expectedErrorMsg: "invalid format of the TaskRunner extra arguments, must be valid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exec := &LocalExecutor{
				Logger: task_logger.NopLogger{},
				Template: db.Template{
					Type:                    db.TemplateTask,
					Playbook:                "site.yml",
					Arguments:               tt.templateArgs,
					AllowOverrideArgsInTask: tt.allowOverride,
				},
				Task: db.Task{
					Arguments: tt.taskArgs,
				},
				Inventory: db.Inventory{
					Type: db.InventoryStatic,
				},
			}

			args, _, err := exec.getPlaybookArgs("admin", nil)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				return
			}

			require.NoError(t, err)

			// Check that expected forks sub-arguments are present in sequence
			if len(tt.expectedForksSubArgs) > 0 {
				found := false
				for i := 0; i <= len(args)-len(tt.expectedForksSubArgs); i++ {
					match := true
					for j := 0; j < len(tt.expectedForksSubArgs); j++ {
						if args[i+j] != tt.expectedForksSubArgs[j] {
							match = false
							break
						}
					}
					if match {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected sequence %v in generated args: %v", tt.expectedForksSubArgs, args)
			}

			// Verify the effective (last) fork argument value matches the expected setting
			if tt.expectedEffectiveFork != "" {
				var lastForkVal string
				for i := 0; i < len(args)-1; i++ {
					if args[i] == "--forks" || args[i] == "-f" {
						lastForkVal = args[i+1]
					}
				}
				assert.Equal(t, tt.expectedEffectiveFork, lastForkVal, "Expected final effective fork value to be %s", tt.expectedEffectiveFork)
			}

			// Verify disqualified values are absent if specified
			for _, prohibited := range tt.mustNotContain {
				assert.NotContains(t, args, prohibited, "Generated args must not contain overridden task arg %q: %v", prohibited, args)
			}

			// Verify playbook is the trailing argument
			assert.Equal(t, "site.yml", args[len(args)-1], "Playbook must be the last argument")
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func TestResolveTaskCopyPath(t *testing.T) {
	setupExecutorConfig(t)
	util.Config = &util.ConfigType{TmpPath: "/tmp"}

	parallelTemplate := db.Template{ID: 5, AllowParallelTasks: true}
	task := db.Task{ID: 9}

	tests := []struct {
		name         string
		repository   db.Repository
		template     db.Template
		expectedPath string
	}{
		{
			name:         "parallel template gets its own task copy",
			repository:   db.Repository{ID: 1, GitURL: "https://example.com/x.git"},
			template:     parallelTemplate,
			expectedPath: "/tmp/project_0/repository_1_template_5_task_9",
		},
		{
			name:         "non-parallel template keeps the shared repository",
			repository:   db.Repository{ID: 1, GitURL: "https://example.com/x.git"},
			template:     db.Template{ID: 5, AllowParallelTasks: false},
			expectedPath: "",
		},
		{
			name:         "local repository type keeps the shared repository",
			repository:   db.Repository{ID: 1, GitURL: "/local/path"},
			template:     parallelTemplate,
			expectedPath: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedPath, resolveTaskCopyPath(tt.repository, tt.template, task))
		})
	}
}

func gitCmd(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	return string(out)
}

func gitCommitFile(t *testing.T, dir, name string) string {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), nil, 0644))
	gitCmd(t, dir, "add", name)
	gitCmd(t, dir, "commit", "-qm", "add "+name)
	return strings.TrimSpace(gitCmd(t, dir, "rev-parse", "HEAD"))
}

// startTask builds the executor for one task the way the task pool does, and
// resolves its branch the way prepareRun does just before the repository update.
func startTask(t *testing.T, provider *LocalExecutorProvider, template db.Template,
	repository db.Repository, taskID int, branch string) *LocalExecutor {

	t.Helper()

	task := db.Task{ID: taskID, TemplateID: template.ID}
	if branch != "" {
		task.GitBranch = &branch
	}

	executor, err := provider.NewExecutor(task, template, db.Inventory{}, repository, db.Environment{}, "")
	require.NoError(t, err)

	localExecutor := executor.(*LocalExecutor)
	localExecutor.SetLogger(task_logger.NopLogger{})
	localExecutor.Repository.GitBranch = resolveGitBranch(repository.GitBranch, template, task)
	return localExecutor
}

// Two tasks of one template start at the same time on different branches: one on
// the repository's default branch, the other overriding it. Neither pins a commit,
// so each takes the tip of its own branch. Each task must end up with the files of
// the branch it asked for, and must report the commit it ran.
func TestUpdateAndCheckoutRepository_EachTaskRunsItsOwnBranch(t *testing.T) {
	for _, gitClientId := range []string{util.CmdGitClientId, util.GoGitClientId} {
		t.Run(gitClientId, func(t *testing.T) {
			setupExecutorConfig(t)
			util.Config = &util.ConfigType{
				TmpPath:     t.TempDir(),
				GitClientId: gitClientId,
				Process:     &util.ConfigProcess{},
			}

			// The branches fork from a common commit: each one carries a file the
			// other does not have, which is what the assertions below read.
			upstream := t.TempDir()
			gitCmd(t, upstream, "init", "-q", "-b", "main")
			gitCmd(t, upstream, "config", "user.email", "t@t")
			gitCmd(t, upstream, "config", "user.name", "t")
			gitCommitFile(t, upstream, "base.txt")

			gitCmd(t, upstream, "checkout", "-q", "-b", "feature")
			featureCommit := gitCommitFile(t, upstream, "feature.txt")

			gitCmd(t, upstream, "checkout", "-q", "main")
			mainCommit := gitCommitFile(t, upstream, "main.txt")

			repository := db.Repository{
				ID:        1,
				ProjectID: 1,
				GitURL:    "file://" + upstream,
				GitBranch: "main",
				SSHKey:    db.AccessKey{Type: db.AccessKeyNone},
			}
			template := db.Template{ID: 1, AllowParallelTasks: true, AllowOverrideBranchInTask: true}

			// Build the executors the way the task pool does, so the wiring is covered too.
			provider := NewLocalExecutorProvider(&KeyInstallerMock{})
			mainTask := startTask(t, provider, template, repository, 201, "")
			featureTask := startTask(t, provider, template, repository, 202, "feature")

			var wg sync.WaitGroup
			var mainErr, featureErr error
			wg.Add(2)
			go func() { defer wg.Done(); mainErr = mainTask.updateAndCheckoutRepository() }()
			go func() { defer wg.Done(); featureErr = featureTask.updateAndCheckoutRepository() }()
			wg.Wait()

			require.NoError(t, mainErr)
			require.NoError(t, featureErr)

			mainWorkdir := mainTask.Repository.WorkingCopyPath
			assert.FileExists(t, filepath.Join(mainWorkdir, "main.txt"))
			assert.NoFileExists(t, filepath.Join(mainWorkdir, "feature.txt"),
				"the task must not see the files of the other branch")
			require.NotNil(t, mainTask.Task.CommitHash)
			assert.Equal(t, mainCommit, *mainTask.Task.CommitHash,
				"the task must record the tip of the branch it ran")

			featureWorkdir := featureTask.Repository.WorkingCopyPath
			assert.FileExists(t, filepath.Join(featureWorkdir, "feature.txt"))
			assert.NoFileExists(t, filepath.Join(featureWorkdir, "main.txt"),
				"the task must not see the files of the other branch")
			require.NotNil(t, featureTask.Task.CommitHash)
			assert.Equal(t, featureCommit, *featureTask.Task.CommitHash,
				"the task must record the tip of the branch it ran")
		})
	}
}
