package db_lib

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

type ShellApp struct {
	Logger     task_logger.Logger
	Template   db.Template
	Repository db.Repository
	App        db.TemplateApp
	reader     bashReader
}

type bashReader struct {
	input  *string
	logger task_logger.Logger
}

func (r *bashReader) Read(p []byte) (n int, err error) {

	r.logger.SetStatus(task_logger.TaskWaitingConfirmation)

	for {
		time.Sleep(time.Second * 3)
		if r.input != nil {
			break
		}
	}

	copy(p, *r.input+"\n")
	r.logger.SetStatus(task_logger.TaskRunningStatus)
	return len(*r.input) + 1, nil
}

func (t *ShellApp) makeCmd(command string, args []string, environmentVars *[]string) *exec.Cmd {
	cmd := exec.Command(command, args...) //nolint: gas
	cmd.Dir = t.GetFullPath()

	cmd.Env = getEnvironmentVars()
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", util.Config.TmpPath))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PWD=%s", cmd.Dir))

	if environmentVars != nil {
		cmd.Env = append(cmd.Env, *environmentVars...)
	}

	return cmd
}

func (t *ShellApp) runCmd(command string, args []string) error {
	cmd := t.makeCmd(command, args, nil)
	t.Logger.LogCmd(cmd)
	return cmd.Run()
}

func (t *ShellApp) GetFullPath() (path string) {
	path = t.Repository.GetFullPath(t.Template.ID)
	return
}

func (t *ShellApp) SetLogger(logger task_logger.Logger) task_logger.Logger {
	t.Logger = logger
	t.Logger.AddStatusListener(func(status task_logger.TaskStatus) {

	})
	t.reader.logger = logger
	return logger
}

func (t *ShellApp) InstallRequirements(environmentVars *[]string) error {
	return nil
}

func (t *ShellApp) makeShellCmd(args []string, environmentVars *[]string) *exec.Cmd {
	var command string
	var appArgs []string
	switch t.App {
	case db.AppBash:
		command = "bash"
	case db.AppPython:
		command = "python3"
	case db.AppPowerShell:
		command = "powershell"
		appArgs = []string{"-File"}
	default:
		command = string(t.App)
	}

	if app, ok := util.Config.Apps[string(t.App)]; ok {
		if app.AppPath != "" {
			command = app.AppPath
		}
		if app.AppArgs != nil {
			appArgs = app.AppArgs
		}
	}

	return t.makeCmd(command, append(appArgs, args...), environmentVars)
}

func (t *ShellApp) Run(args []string, environmentVars *[]string, inputs map[string]string, cb func(*os.Process)) error {
	cmd := t.makeShellCmd(args, environmentVars)
	t.Logger.LogCmd(cmd)
	//cmd.Stdin = &t.reader
	cmd.Stdin = strings.NewReader("")
	err := cmd.Start()
	if err != nil {
		return err
	}
	cb(cmd.Process)
	return cmd.Wait()
}
