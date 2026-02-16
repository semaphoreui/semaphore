package db_lib

import (
	"fmt"
	"os"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

// getHomeDir returns the HOME directory value for a task based on the configured
// HomeDirMode. For "project_home" it returns the project tmp directory.
// For "template_dir" and "user_home" it returns the real user HOME (no override).
func getHomeDir(repo db.Repository, templateID int) string {
	switch util.Config.HomeDirMode {
	case util.HomeDirModeProjectHome:
		return util.Config.GetProjectTmpDir(repo.ProjectID)
	case util.HomeDirModeTemplateDir, util.HomeDirModeUserHome:
		return os.Getenv("HOME")
	default:
		return ""
	}
}

func getEnvironmentVars() []string {
	res := []string{
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
	}

	for _, e := range util.Config.ForwardedEnvVars {
		v := os.Getenv(e)
		if v != "" {
			res = append(res, fmt.Sprintf("%s=%s", e, v))
		}
	}

	for k, v := range util.Config.EnvVars {
		res = append(res, fmt.Sprintf("%s=%s", k, v))
	}

	return res
}

type LocalAppRunningArgs struct {
	CliArgs         map[string][]string // Stage-specific args (e.g., "init", "apply", "default")
	EnvironmentVars []string
	Inputs          map[string]string
	TaskParams      any
	TemplateParams  any
	Callback        func(*os.Process)
}

type LocalAppInstallingArgs struct {
	EnvironmentVars []string
	TplParams       any
	Params          any
	Installer       AccessKeyInstaller
}

type LocalApp interface {
	SetLogger(logger task_logger.Logger) task_logger.Logger
	InstallRequirements(args LocalAppInstallingArgs) error
	Run(args LocalAppRunningArgs) error
	Clear()
}
