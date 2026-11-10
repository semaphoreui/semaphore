package tasks

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hostsApp is an app whose Run gets the environment of the task.
type hostsApp struct {
	outputsApp
	run func(env []string) error
}

func (a *hostsApp) Run(args db_lib.LocalAppRunningArgs) error {
	if a.run == nil {
		return nil
	}
	return a.run(args.EnvironmentVars)
}

func newHostsExecutor(t *testing.T, app db.TemplateApp, run func(env []string) error) *LocalExecutor {
	t.Helper()
	setupExecutorConfig(t)
	return &LocalExecutor{
		Task:       db.Task{ID: 9, ProjectID: 3},
		Template:   db.Template{ID: 7, ProjectID: 3, App: app, Playbook: "site.yml"},
		Repository: db.Repository{ProjectID: 3, GitURL: t.TempDir()},
		Inventory:  db.Inventory{ID: 2, ProjectID: 3, Type: db.InventoryStatic, Inventory: "localhost"},
		App:        &hostsApp{run: run},
		Logger:     &recordingLogger{},
	}
}

func envValues(env []string, key string) (values []string) {
	for _, e := range env {
		if v, ok := strings.CutPrefix(e, key+"="); ok {
			values = append(values, v)
		}
	}
	return
}

func TestLocalExecutor_PrepareAddsHostsPlugin(t *testing.T) {
	executor := newHostsExecutor(t, db.AppAnsible, nil)
	t.Cleanup(executor.Cleanup)

	err := executor.Prepare("", nil, "")

	require.NoError(t, err)
	files := envValues(executor.preparedEnv, hostsFileEnv)
	require.Len(t, files, 1)
	assert.True(t, strings.HasPrefix(files[0], util.Config.GetProjectTmpDir(3)+string(os.PathSeparator)))
	plugins := envValues(executor.preparedEnv, ansibleCallbackPluginsEnv)
	require.NotEmpty(t, plugins)
	paths := filepath.SplitList(plugins[len(plugins)-1])
	assert.FileExists(t, filepath.Join(paths[len(paths)-1], "semaphore_hosts.py"))

	t.Run("not for terraform", func(t *testing.T) {
		executor := newHostsExecutor(t, db.AppTerraform, nil)
		t.Cleanup(executor.Cleanup)

		err := executor.Prepare("", nil, "")

		require.NoError(t, err)
		assert.Empty(t, envValues(executor.preparedEnv, hostsFileEnv))
	})
}

func TestLocalExecutor_RunLoadsOutputsAndHostsPlugins(t *testing.T) {
	var plugins []string
	executor := newHostsExecutor(t, db.AppAnsible, func(env []string) error {
		plugins = envValues(env, ansibleCallbackPluginsEnv)
		return nil
	})
	home := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(home, ".ansible.cfg"), []byte("[defaults]\n"), 0o644))
	t.Setenv("HOME", home)
	runID := 5
	executor.Task.WorkflowRunID = &runID
	executor.OutputsCollector = &outputsCollectorFake{capture: &outputsCaptureFake{}}

	err := executor.Run("", nil, "")

	require.NoError(t, err)
	require.NotEmpty(t, plugins)
	paths := filepath.SplitList(plugins[len(plugins)-1])
	require.Len(t, paths, 4)
	assert.Equal(t, []string{
		"~/.ansible/plugins/callback",
		"/usr/share/ansible/plugins/callback",
		"/tmp/callback_plugins",
	}, paths[:3])
	assert.True(t, strings.HasPrefix(paths[3], util.Config.GetProjectTmpDir(3)+string(os.PathSeparator)))
}

func TestLocalExecutor_beginHostsError(t *testing.T) {
	executor := newHostsExecutor(t, db.AppAnsible, nil)
	// The project tmp dir cannot be created under a regular file.
	util.Config.TmpPath = filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(util.Config.TmpPath, nil, 0o600))

	env := executor.beginHosts(nil)

	assert.Empty(t, env)
	assert.Nil(t, executor.hostsCapture)
	lines := executor.Logger.(*recordingLogger).lines
	require.Len(t, lines, 1)
	assert.Contains(t, lines[0], "Host stats will not be captured: ")
}

func TestLocalExecutor_RunCollectsHostsOfFailedTask(t *testing.T) {
	document := `{"version":1,"hosts":{"a":{"failed":1}}}`
	executor := newHostsExecutor(t, db.AppAnsible, func(env []string) error {
		files := envValues(env, hostsFileEnv)
		require.Len(t, files, 1)
		if err := os.WriteFile(files[0], []byte(document), 0o600); err != nil {
			return err
		}
		return errors.New("exit status 2")
	})

	err := executor.Run("", nil, "")

	assert.ErrorContains(t, err, "exit status 2")
	require.NotNil(t, executor.Hosts())
	assert.Equal(t, document, *executor.Hosts())
}
