package db_lib

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/pkg/task_logger"
	"github.com/ansible-semaphore/semaphore/util"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

type TerraformApp struct {
	Logger     task_logger.Logger
	Template   db.Template
	Repository db.Repository
	Inventory  db.Inventory
	reader     terraformReader
	Name       string
	noChanges  bool
}

type terraformReaderResult int

const (
	terraformReaderConfirmed terraformReaderResult = iota
	terraformReaderFailed
)

type terraformReader struct {
	result *terraformReaderResult
}

func (t *TerraformApp) makeCmd(command string, args []string, environmentVars *[]string) *exec.Cmd {
	cmd := exec.Command(command, args...) //nolint: gas
	cmd.Dir = t.GetFullPath()

	cmd.Env = removeSensitiveEnvs(os.Environ())
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", util.Config.TmpPath))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PWD=%s", cmd.Dir))

	if environmentVars != nil {
		cmd.Env = append(cmd.Env, *environmentVars...)
	}

	return cmd
}

func (t *TerraformApp) runCmd(command string, args []string) error {
	cmd := t.makeCmd(command, args, nil)
	t.Logger.LogCmd(cmd)
	return cmd.Run()
}

func (t *TerraformApp) GetFullPath() string {
	return path.Join(t.Repository.GetFullPath(t.Template.ID), strings.TrimPrefix(t.Template.Playbook, "/"))
}

func (t *TerraformApp) SetLogger(logger task_logger.Logger) task_logger.Logger {
	t.Logger = logger

	t.Logger.AddLogListener(func(new time.Time, msg string) {
		if strings.Contains(msg, "No changes.") {
			t.noChanges = true
		}
	})

	t.Logger.AddStatusListener(func(status task_logger.TaskStatus) {
		var result terraformReaderResult

		switch status {
		case task_logger.TaskConfirmed:
			result = terraformReaderConfirmed
			t.reader.result = &result
		case task_logger.TaskFailStatus, task_logger.TaskStoppedStatus:
			result = terraformReaderFailed
			t.reader.result = &result
		}
	})

	return logger
}

func (t *TerraformApp) init() error {
	cmd := t.makeCmd(t.Name, []string{"init"}, nil)
	t.Logger.LogCmd(cmd)
	err := cmd.Start()
	if err != nil {
		return err
	}

	return cmd.Wait()
}

func (t *TerraformApp) selectWorkspace(workspace string) error {
	cmd := t.makeCmd(string(t.Name), []string{"workspace", "select", "-or-create=true", workspace}, nil)
	t.Logger.LogCmd(cmd)
	err := cmd.Start()
	if err != nil {
		return err
	}

	return cmd.Wait()
}

func (t *TerraformApp) InstallRequirements() (err error) {
	err = t.init()
	if err != nil {
		return
	}

	workspace := "default"

	if t.Inventory.Inventory != "" {
		workspace = t.Inventory.Inventory
	}

	err = t.selectWorkspace(workspace)
	return
}

func (t *TerraformApp) Plan(args []string, environmentVars *[]string, inputs map[string]string, cb func(*os.Process)) error {
	args = append([]string{"plan"}, args...)
	cmd := t.makeCmd(t.Name, args, environmentVars)
	t.Logger.LogCmd(cmd)
	cmd.Stdin = strings.NewReader("")
	err := cmd.Start()
	if err != nil {
		return err
	}
	cb(cmd.Process)
	return cmd.Wait()
}

func (t *TerraformApp) Apply(args []string, environmentVars *[]string, inputs map[string]string, cb func(*os.Process)) error {
	args = append([]string{"apply", "-auto-approve"}, args...)
	cmd := t.makeCmd(t.Name, args, environmentVars)
	t.Logger.LogCmd(cmd)
	cmd.Stdin = strings.NewReader("")
	err := cmd.Start()
	if err != nil {
		return err
	}
	cb(cmd.Process)
	return cmd.Wait()
}

func (t *TerraformApp) Run(args []string, environmentVars *[]string, inputs map[string]string, cb func(*os.Process)) error {
	err := t.Plan(args, environmentVars, inputs, cb)
	if err != nil {
		return err
	}

	if t.noChanges {
		t.Logger.SetStatus(task_logger.TaskSuccessStatus)
		return nil
	}

	t.Logger.SetStatus(task_logger.TaskWaitingConfirmation)

	for {
		time.Sleep(time.Second * 3)
		if t.reader.result != nil {
			break
		}
	}

	switch *t.reader.result {
	case terraformReaderFailed:
		return nil
	case terraformReaderConfirmed:
		t.Logger.SetStatus(task_logger.TaskRunningStatus)
		return t.Apply(args, environmentVars, inputs, cb)
	default:
		return fmt.Errorf("unknown plan result")
	}
}
