package tasks

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type outputsCaptureFake struct {
	document *string
	notes    []string
	err      error

	collected        bool
	terraformOutputs []byte
	closed           bool
}

func (c *outputsCaptureFake) Env() []string {
	return []string{"SEMAPHORE_OUTPUTS_FILE=/tmp/outputs.json"}
}

func (c *outputsCaptureFake) AnsibleCallbackDir() (string, error) {
	return "/tmp/callback_plugins", nil
}

func (c *outputsCaptureFake) Collect(terraformOutputs []byte) (*string, []string, error) {
	c.collected = true
	c.terraformOutputs = terraformOutputs
	return c.document, c.notes, c.err
}

func (c *outputsCaptureFake) Close() {
	c.closed = true
}

type outputsCollectorFake struct {
	capture *outputsCaptureFake
	err     error

	calls  int
	dir    string
	taskID int

	// validate overrides Validate; nil returns the document unchanged.
	validate  func(document string) (*string, error)
	validated []string
}

func (c *outputsCollectorFake) Begin(dir string, taskID int) (pro_interfaces.TaskOutputsCapture, error) {
	c.calls++
	c.dir = dir
	c.taskID = taskID
	if c.err != nil {
		return nil, c.err
	}
	return c.capture, nil
}

func (c *outputsCollectorFake) Validate(document string) (*string, error) {
	c.validated = append(c.validated, document)
	if c.validate != nil {
		return c.validate(document)
	}
	if c.err != nil {
		return nil, c.err
	}
	return &document, nil
}

// outputsApp is an app whose Run reports what the test tells it to.
type outputsApp struct {
	run func() error
}

func (a *outputsApp) SetLogger(logger task_logger.Logger) task_logger.Logger { return logger }

func (a *outputsApp) InstallRequirements(db_lib.LocalAppInstallingArgs) error { return nil }

func (a *outputsApp) Run(db_lib.LocalAppRunningArgs) error {
	if a.run == nil {
		return nil
	}
	return a.run()
}

func (a *outputsApp) Clear() {}

type recordingLogger struct {
	task_logger.NopLogger
	lines []string
}

func (l *recordingLogger) Log(msg string) {
	l.lines = append(l.lines, msg)
}

func setupOutputsConfig(t *testing.T) {
	t.Helper()
	prev := util.Config
	t.Cleanup(func() { util.Config = prev })
	util.Config = &util.ConfigType{TmpPath: "/tmp/semaphore"}
}

func TestLocalExecutor_beginOutputs(t *testing.T) {
	runID := 5

	tests := []struct {
		name         string
		hasCollector bool
		runID        *int
		app          db.TemplateApp
		wantCalls    int
		wantEnv      []string
	}{
		{"no collector", false, &runID, db.AppBash, 0, nil},
		{"task outside a workflow run", true, nil, db.AppBash, 0, nil},
		{"workflow shell task", true, &runID, db.AppBash, 1, []string{
			"SEMAPHORE_OUTPUTS_FILE=/tmp/outputs.json",
		}},
		{"workflow ansible task", true, &runID, db.AppAnsible, 1, []string{
			"SEMAPHORE_OUTPUTS_FILE=/tmp/outputs.json",
			"ANSIBLE_CALLBACK_PLUGINS=/custom/callbacks" + string(os.PathListSeparator) + "/tmp/callback_plugins",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupOutputsConfig(t)
			collector := &outputsCollectorFake{capture: &outputsCaptureFake{}}
			executor := &LocalExecutor{
				Task:     db.Task{ID: 9, WorkflowRunID: tt.runID},
				Template: db.Template{ProjectID: 3, App: tt.app},
			}
			if tt.hasCollector {
				executor.OutputsCollector = collector
			}

			env, err := executor.beginOutputs([]string{"ANSIBLE_CALLBACK_PLUGINS=/custom/callbacks"})

			require.NoError(t, err)
			assert.Equal(t, tt.wantEnv, env)
			assert.Equal(t, tt.wantCalls, collector.calls)
			if tt.wantCalls > 0 {
				assert.Equal(t, "/tmp/semaphore/project_3", collector.dir)
				assert.Equal(t, 9, collector.taskID)
			}
		})
	}
}

func TestLocalExecutor_beginOutputsError(t *testing.T) {
	setupOutputsConfig(t)
	runID := 5
	executor := &LocalExecutor{
		Task:             db.Task{WorkflowRunID: &runID},
		OutputsCollector: &outputsCollectorFake{err: errors.New("disk full")},
	}

	env, err := executor.beginOutputs(nil)

	assert.ErrorContains(t, err, "disk full")
	assert.Empty(t, env)
}

func TestLocalExecutor_ansibleCallbackPlugins(t *testing.T) {
	sep := string(os.PathListSeparator)
	const dir = "/tmp/callback_plugins"

	tests := []struct {
		name     string
		env      []string
		config   util.ConfigType
		osEnv    string
		expected string
	}{
		{
			name:     "keeps the value set for the task, the last one wins",
			env:      []string{"ANSIBLE_CALLBACK_PLUGINS=/first", "FOO=bar", "ANSIBLE_CALLBACK_PLUGINS=/second"},
			expected: "/second" + sep + dir,
		},
		{
			name:     "keeps the value from the config env vars",
			config:   util.ConfigType{EnvVars: map[string]string{"ANSIBLE_CALLBACK_PLUGINS": "/from/config"}},
			expected: "/from/config" + sep + dir,
		},
		{
			name:     "keeps the forwarded value of the server environment",
			config:   util.ConfigType{ForwardedEnvVars: []string{"ANSIBLE_CALLBACK_PLUGINS"}},
			osEnv:    "/from/os",
			expected: "/from/os" + sep + dir,
		},
		{
			name:     "ignores the server environment when it is not forwarded",
			osEnv:    "/from/os",
			expected: "~/.ansible/plugins/callback" + sep + "/usr/share/ansible/plugins/callback" + sep + dir,
		},
		{
			name:     "falls back to the default path of Ansible",
			expected: "~/.ansible/plugins/callback" + sep + "/usr/share/ansible/plugins/callback" + sep + dir,
		},
		{
			// expected is derived below from the helper the app uses for
			// ANSIBLE_HOME, so the two cannot drift apart.
			name:   "follows ANSIBLE_HOME of the template_dir home mode",
			config: util.ConfigType{HomeDirMode: util.HomeDirModeTemplateDir, TmpPath: "/tmp/semaphore"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prev := util.Config
			t.Cleanup(func() { util.Config = prev })
			config := tt.config
			util.Config = &config
			t.Setenv("ANSIBLE_CALLBACK_PLUGINS", tt.osEnv)

			executor := &LocalExecutor{
				Template:   db.Template{ID: 7, ProjectID: 3},
				Repository: db.Repository{ProjectID: 3},
			}
			if tt.config.HomeDirMode == util.HomeDirModeTemplateDir {
				tt.expected = filepath.Join(executor.Repository.GetHomePath(7), ".ansible/plugins/callback") +
					sep + "/usr/share/ansible/plugins/callback" + sep + dir
			}

			assert.Equal(t, tt.expected, executor.ansibleCallbackPlugins(tt.env, dir))
		})
	}
}

func newOutputsExecutor(capture *outputsCaptureFake, app *outputsApp) (*LocalExecutor, *recordingLogger) {
	logger := &recordingLogger{}
	executor := &LocalExecutor{
		App:            app,
		Logger:         logger,
		prepared:       true,
		outputsCapture: capture,
	}
	// Prepare is skipped, so make the transition it would have made.
	executor.SetStatus(task_logger.TaskRunningStatus)
	return executor, logger
}

func TestLocalExecutor_RunCollectsOutputs(t *testing.T) {
	setupExecutorConfig(t)
	document := `{"values":{"image_tag":"1.4.2"},"skipped":{"token":"sensitive"}}`
	capture := &outputsCaptureFake{
		document: &document,
		notes:    []string{`Output "token" was not captured: it is marked sensitive`},
	}
	executor, logger := newOutputsExecutor(capture, &outputsApp{})

	err := executor.Run("", nil, "")

	require.NoError(t, err)
	require.NotNil(t, executor.Outputs())
	assert.Equal(t, document, *executor.Outputs())
	assert.Nil(t, capture.terraformOutputs)
	assert.Equal(t, capture.notes, logger.lines)
	assert.True(t, capture.closed)
}

func TestLocalExecutor_RunFailsOnInvalidOutputs(t *testing.T) {
	setupExecutorConfig(t)
	capture := &outputsCaptureFake{err: errors.New("invalid outputs: the file must contain one JSON object")}
	executor, logger := newOutputsExecutor(capture, &outputsApp{})

	err := executor.Run("", nil, "")

	assert.ErrorContains(t, err, "task outputs: invalid outputs")
	assert.Nil(t, executor.Outputs())
	require.Len(t, logger.lines, 1)
	assert.Contains(t, logger.lines[0], "Invalid task outputs: ")
	assert.True(t, capture.closed)
}

func TestLocalExecutor_RunDoesNotCollectOutputsOfUnsuccessfulTask(t *testing.T) {
	document := `{"values":{"a":1}}`

	tests := []struct {
		name    string
		run     func(executor *LocalExecutor) error
		wantErr bool
	}{
		{
			name: "the app failed",
			run: func(*LocalExecutor) error {
				return errors.New("exit status 2")
			},
			wantErr: true,
		},
		{
			// A rejected Terraform plan: Run returns nil with a failed status.
			name: "the app ended without an error but not successfully",
			run: func(executor *LocalExecutor) error {
				executor.SetStatus(task_logger.TaskFailStatus)
				return nil
			},
		},
		{
			name: "the task was stopped",
			run: func(executor *LocalExecutor) error {
				executor.Kill()
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupExecutorConfig(t)
			capture := &outputsCaptureFake{document: &document}
			app := &outputsApp{}
			executor, _ := newOutputsExecutor(capture, app)
			app.run = func() error { return tt.run(executor) }

			err := executor.Run("", nil, "")

			assert.Equal(t, tt.wantErr, err != nil)
			assert.False(t, capture.collected)
			assert.Nil(t, executor.Outputs())
			assert.True(t, capture.closed)
		})
	}
}

func TestLocalExecutor_RunWithoutOutputsCapture(t *testing.T) {
	setupExecutorConfig(t)
	executor, _ := newOutputsExecutor(nil, &outputsApp{})
	executor.outputsCapture = nil

	err := executor.Run("", nil, "")

	require.NoError(t, err)
	assert.Nil(t, executor.Outputs())
}
