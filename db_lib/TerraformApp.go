package db_lib

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/lib"
	"github.com/ansible-semaphore/semaphore/util"
	"os"
	"os/exec"
	"time"
)

type TerraformApp struct {
	Logger     lib.Logger
	Template   db.Template
	Repository db.Repository
	reader     terraformReader
}

type terraformLogger struct {
	logger lib.Logger
	reader *terraformReader
}

func (l *terraformLogger) Log(msg string) {
	l.logger.Log(msg)
}

type terraformReader struct {
	confirmed bool
	logger    *terraformLogger
}

func (r *terraformReader) Read(p []byte) (n int, err error) {

	if r.confirmed {
		copy(p, "\n")
		return 1, nil
	}

	r.logger.SetStatus(lib.TaskWaitingConfirmation)

	for {
		time.Sleep(time.Second * 3)
		if r.confirmed {
			break
		}
	}

	copy(p, "yes\n")

	r.logger.SetStatus(lib.TaskRunningStatus)

	return 4, nil
}

func (l *terraformLogger) Log2(msg string, now time.Time) {
	l.logger.Log2(msg, now)
}

func (l *terraformLogger) LogCmd(cmd *exec.Cmd) {
	l.logger.LogCmd(cmd)
}

func (l *terraformLogger) SetStatus(status lib.TaskStatus) {
	if status == lib.TaskConfirmed {
		l.reader.confirmed = true
	}

	l.logger.SetStatus(status)
}

func (t *TerraformApp) makeCmd(command string, args []string, environmentVars *[]string) *exec.Cmd {
	cmd := exec.Command(command, args...) //nolint: gas
	cmd.Dir = t.GetFullPath()

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", util.Config.TmpPath))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PWD=%s", cmd.Dir))

	if environmentVars != nil {
		cmd.Env = append(cmd.Env, *environmentVars...)
	}

	sensitiveEnvs := []string{
		"SEMAPHORE_ACCESS_KEY_ENCRYPTION",
		"SEMAPHORE_ADMIN_PASSWORD",
		"SEMAPHORE_DB_USER",
		"SEMAPHORE_DB_NAME",
		"SEMAPHORE_DB_HOST",
		"SEMAPHORE_DB_PASS",
		"SEMAPHORE_LDAP_PASSWORD",
	}

	// Remove sensitive env variables from cmd process
	for _, env := range sensitiveEnvs {
		cmd.Env = append(cmd.Env, env+"=")
	}

	return cmd
}

func (t *TerraformApp) runCmd(command string, args []string) error {
	cmd := t.makeCmd(command, args, nil)
	t.Logger.LogCmd(cmd)
	return cmd.Run()
}

func (t *TerraformApp) GetFullPath() (path string) {
	path = t.Repository.GetFullPath(t.Template.ID)
	return
}

func (t *TerraformApp) SetLogger(logger lib.Logger) lib.Logger {
	internalLogger := &terraformLogger{
		logger: logger,
		reader: &t.reader,
	}

	t.reader.logger = internalLogger
	t.Logger = internalLogger
	return internalLogger
}

func (t *TerraformApp) InstallRequirements() error {

	if _, ok := t.Logger.(*terraformLogger); !ok {
		t.SetLogger(t.Logger)
	}

	cmd := t.makeCmd("terraform", []string{"init"}, nil)
	t.Logger.LogCmd(cmd)
	err := cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Wait()
}

func (t *TerraformApp) Run(args []string, environmentVars *[]string, cb func(*os.Process)) error {
	cmd := t.makeCmd("terraform", args, environmentVars)
	t.Logger.LogCmd(cmd)
	cmd.Stdin = &t.reader
	err := cmd.Start()
	if err != nil {
		return err
	}
	cb(cmd.Process)
	return cmd.Wait()
}
