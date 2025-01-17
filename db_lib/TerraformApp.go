package db_lib

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

type TerraformApp struct {
	Logger           task_logger.Logger
	Template         db.Template
	Repository       db.Repository
	Inventory        db.Inventory
	reader           terraformReader // reader
	Name             string          // Name is the name of the terraform binary
	PlanHasNoChanges bool            // PlanHasNoChanges is true if terraform plan has no changes
}

type terraformReader struct {
	EOF    bool
	status task_logger.TaskStatus
	logger task_logger.Logger
}

func (r *terraformReader) Read(p []byte) (n int, err error) {
	if r.EOF {
		return 0, io.EOF
	}

	if r.status != task_logger.TaskWaitingConfirmation {
		time.Sleep(time.Second * 3)
		return 0, nil
	}

	for {
		time.Sleep(time.Second * 3)
		if r.status.IsFinished() ||
			r.status == task_logger.TaskConfirmed ||
			r.status == task_logger.TaskRejected {
			break
		}
	}

	r.EOF = true

	switch r.status {
	case task_logger.TaskConfirmed:
		copy(p, "yes\n")
		r.logger.SetStatus(task_logger.TaskRunningStatus)
		return 4, nil
	case task_logger.TaskRejected:
		copy(p, "no\n")
		r.logger.SetStatus(task_logger.TaskRunningStatus)
		return 3, nil
	default:
		copy(p, "\n")
		return 1, nil
	}
}

func (t *TerraformApp) makeCmd(command string, args []string, environmentVars []string) *exec.Cmd {
	cmd := exec.Command(command, args...) //nolint: gas
	cmd.Dir = t.GetFullPath()

	cmd.Env = getEnvironmentVars()
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", util.Config.TmpPath))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PWD=%s", cmd.Dir))

	if environmentVars != nil {
		cmd.Env = append(cmd.Env, environmentVars...)
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
	logger.AddStatusListener(func(status task_logger.TaskStatus) {
		t.reader.status = status
	})

	t.reader.logger = logger
	t.Logger = logger
	return logger
}

func (t *TerraformApp) init(environmentVars []string, params *db.TerraformTaskParams) error {

	keyInstallation, err := t.Inventory.SSHKey.Install(db.AccessKeyRoleGit, t.Logger)
	if err != nil {
		return err
	}
	defer keyInstallation.Destroy() //nolint: errcheck

	args := []string{"init", "-lock=false"}

	if params.Upgrade {
		args = append(args, "-upgrade")
	}

	if params.Reconfigure {
		args = append(args, "-reconfigure")
	} else {
		args = append(args, "-migrate-state")
	}

	cmd := t.makeCmd(t.Name, args, environmentVars)
	cmd.Env = append(cmd.Env, keyInstallation.GetGitEnv()...)
	t.Logger.LogCmd(cmd)

	t.Logger.AddLogListener(func(new time.Time, msg string) {
		s := strings.TrimSpace(msg)
		if strings.Contains(s, "Do you want to copy ") {
			t.Logger.SetStatus(task_logger.TaskWaitingConfirmation)
		} else if strings.Contains(msg, "has been successfully initialized!") ||
			strings.Contains(msg, "Error:") {
			t.reader.EOF = true
		}
	})

	cmd.Stdin = &t.reader
	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	t.Logger.WaitLog()
	return nil
}

func (t *TerraformApp) isWorkspacesSupported(environmentVars []string) bool {
	cmd := t.makeCmd(t.Name, []string{"workspace", "list"}, environmentVars)
	err := cmd.Run()
	if err != nil {
		return false
	}

	return true
}

func (t *TerraformApp) selectWorkspace(workspace string, environmentVars []string) error {
	cmd := t.makeCmd(t.Name, []string{"workspace", "select", "-or-create=true", workspace}, environmentVars)
	t.Logger.LogCmd(cmd)

	err := cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	t.Logger.WaitLog()
	return nil
}

func (t *TerraformApp) InstallRequirements(environmentVars []string, params interface{}) (err error) {

	p := params.(*db.TerraformTaskParams)

	err = t.init(environmentVars, p)
	if err != nil {
		return
	}

	workspace := "default"

	if t.Inventory.Inventory != "" {
		workspace = t.Inventory.Inventory
	}

	if !t.isWorkspacesSupported(environmentVars) {
		return
	}

	err = t.selectWorkspace(workspace, environmentVars)
	return
}

func (t *TerraformApp) Plan(args []string, environmentVars []string, inputs map[string]string, cb func(*os.Process)) error {
	args = append([]string{"plan", "-lock=false"}, args...)
	cmd := t.makeCmd(t.Name, args, environmentVars)
	t.Logger.LogCmd(cmd)

	t.reader.logger.AddLogListener(func(new time.Time, msg string) {
		if strings.Contains(msg, "No changes.") {
			t.PlanHasNoChanges = true
		}
	})

	cmd.Stdin = strings.NewReader("")
	err := cmd.Start()
	if err != nil {
		return err
	}

	cb(cmd.Process)

	err = cmd.Wait()
	if err != nil {
		return err
	}

	t.Logger.WaitLog()
	return nil
}

func (t *TerraformApp) Apply(args []string, environmentVars []string, inputs map[string]string, cb func(*os.Process)) error {
	args = append([]string{"apply", "-auto-approve", "-lock=false"}, args...)
	cmd := t.makeCmd(t.Name, args, environmentVars)
	t.Logger.LogCmd(cmd)
	cmd.Stdin = strings.NewReader("")
	err := cmd.Start()
	if err != nil {
		return err
	}
	cb(cmd.Process)

	err = cmd.Wait()
	if err != nil {
		return err
	}

	t.Logger.WaitLog()
	return nil
}

func (t *TerraformApp) Run(args LocalAppRunningArgs) error {
	err := t.Plan(args.CliArgs, args.EnvironmentVars, args.Inputs, args.Callback)
	if err != nil {
		return err
	}

	params := args.TaskParams.(*db.TerraformTaskParams)

	if t.PlanHasNoChanges || params.Plan {
		t.Logger.SetStatus(task_logger.TaskSuccessStatus)
		return nil
	}

	if params.AutoApprove {
		t.Logger.SetStatus(task_logger.TaskRunningStatus)
		return t.Apply(args.CliArgs, args.EnvironmentVars, args.Inputs, args.Callback)
	}

	t.Logger.SetStatus(task_logger.TaskWaitingConfirmation)

	for {
		time.Sleep(time.Second * 3)
		if t.reader.status.IsFinished() ||
			t.reader.status == task_logger.TaskConfirmed ||
			t.reader.status == task_logger.TaskRejected {
			break
		}
	}

	switch t.reader.status {
	case task_logger.TaskRejected:
		t.Logger.SetStatus(task_logger.TaskFailStatus)
	case task_logger.TaskConfirmed:
		t.Logger.SetStatus(task_logger.TaskRunningStatus)
		return t.Apply(args.CliArgs, args.EnvironmentVars, args.Inputs, args.Callback)
	}

	return nil
}
