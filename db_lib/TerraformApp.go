package db_lib

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

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
	backendFilename  string          // backendFilename is the name of the backend file
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

	if app, ok := util.Config.Apps[t.Name]; ok {
		if app.AppPath != "" {
			command = app.AppPath
		}
		if app.AppArgs != nil {
			args = append(app.AppArgs, args...)
		}
	}

	if t.Name == string(db.AppTerragrunt) {
		hasTfPath := false
		for i := 0; i < len(args); i++ {
			a := args[i]
			if a == "--tf-path" || strings.HasPrefix(a, "--tf-path=") {
				hasTfPath = true
				break
			}
		}
		if !hasTfPath {
			args = append(args, "--tf-path=terraform")
		}
	}

	cmd := exec.Command(command, args...) //nolint: gas
	cmd.Dir = t.GetFullPath()

	cmd.Env = getEnvironmentVars()
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", getHomeDir(t.Repository, t.Template.ID)))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PWD=%s", cmd.Dir))

	if environmentVars != nil {
		cmd.Env = append(cmd.Env, environmentVars...)
	}

	cmd.SysProcAttr = util.Config.GetSysProcAttr()

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

func (t *TerraformApp) init(environmentVars []string, keyInstaller AccessKeyInstaller, params *db.TerraformTaskParams, extraArgs []string) error {

	keyInstallation, err := keyInstaller.Install(t.Inventory.SSHKey, db.AccessKeyRoleGit, t.Logger)
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

	// Add extra args specific to init stage
	if extraArgs != nil {
		args = append(args, extraArgs...)
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
	args := []string{"workspace", "list"}

	cmd := t.makeCmd(t.Name, args, environmentVars)
	err := cmd.Run()
	if err != nil {
		return false
	}

	return true
}

func (t *TerraformApp) selectWorkspace(workspace string, environmentVars []string) error {
	args := []string{"workspace", "select", "-or-create=true", workspace}
	if t.Name == string(db.AppTerragrunt) {
		args = append([]string{"run", "--"}, args...)
	}
	cmd := t.makeCmd(t.Name, args, environmentVars)
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

func (t *TerraformApp) Clear() {
	if t.backendFilename == "" {
		return
	}

	err := os.Remove(path.Join(t.GetFullPath(), t.backendFilename))
	if os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context": "terraform",
			"task_id": t.Template.ID,
		}).Warn("Unable to remove backend file")
	}
}

type TerraformInstallRequirementsArgs struct {
	LocalAppInstallingArgs
	InitArgs []string // Stage-specific args for init
}

func (t *TerraformApp) InstallRequirements(args LocalAppInstallingArgs) (err error) {
	return t.InstallRequirementsWithInitArgs(args, nil)
}

func (t *TerraformApp) InstallRequirementsWithInitArgs(args LocalAppInstallingArgs, initArgs []string) (err error) {

	tpl := args.TplParams.(*db.TerraformTemplateParams)
	p := args.Params.(*db.TerraformTaskParams)

	if tpl.OverrideBackend {
		t.backendFilename = "backend.tf"
		if tpl.BackendFilename != "" {
			t.backendFilename = tpl.BackendFilename
		}

		backendFile := path.Join(t.GetFullPath(), t.backendFilename)
		err = os.WriteFile(backendFile, []byte("terraform {\n  backend \"http\" {\n  }\n}\n"), 0644)
		if err != nil {
			return
		}
	}

	if err = t.init(args.EnvironmentVars, args.Installer, p, initArgs); err != nil {
		return
	}

	workspace := "default"

	if t.Inventory.Inventory != "" {
		workspace = t.Inventory.Inventory
	}

	if !t.isWorkspacesSupported(args.EnvironmentVars) {
		return
	}

	err = t.selectWorkspace(workspace, args.EnvironmentVars)
	return
}

func (t *TerraformApp) Plan(args []string, environmentVars []string, inputs map[string]string, cb func(*os.Process)) error {
	planArgs := []string{"plan", "-lock=false"}
	planArgs = append(planArgs, args...)
	cmd := t.makeCmd(t.Name, planArgs, environmentVars)
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
	applyArgs := []string{"apply", "-auto-approve", "-lock=false"}
	applyArgs = append(applyArgs, args...)
	cmd := t.makeCmd(t.Name, applyArgs, environmentVars)
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
	// Determine which args to use for plan and apply stages
	var planArgs []string
	var applyArgs []string

	// Use stage-specific args from map, with "default" fallback
	if pArgs, ok := args.CliArgs["plan"]; ok {
		planArgs = pArgs
	} else if aArgs, ok := args.CliArgs["apply"]; ok {
		applyArgs = aArgs
	} else if defaultArgs, ok := args.CliArgs["default"]; ok {
		planArgs = defaultArgs
	}

	if aArgs, ok := args.CliArgs["apply"]; ok {
		applyArgs = aArgs
	} else if defaultArgs, ok := args.CliArgs["default"]; ok {
		applyArgs = defaultArgs
	}

	err := t.Plan(planArgs, args.EnvironmentVars, args.Inputs, args.Callback)
	if err != nil {
		return err
	}

	params := args.TaskParams.(*db.TerraformTaskParams)
	tplParams := args.TemplateParams.(*db.TerraformTemplateParams)

	if t.PlanHasNoChanges || params.Plan {
		t.Logger.SetStatus(task_logger.TaskSuccessStatus)
		return nil
	}

	if tplParams.AutoApprove || tplParams.AllowAutoApprove && params.AutoApprove {
		t.Logger.SetStatus(task_logger.TaskRunningStatus)
		return t.Apply(applyArgs, args.EnvironmentVars, args.Inputs, args.Callback)
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
		return t.Apply(applyArgs, args.EnvironmentVars, args.Inputs, args.Callback)
	}

	return nil
}
