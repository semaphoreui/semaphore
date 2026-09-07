package db_lib

import (
	"fmt"
	"os"
	"runtime"
	"strings"

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

var defaultForwardedEnvVars = []string{
	// Proxy
	"HTTP_PROXY", "http_proxy",
	"HTTPS_PROXY", "https_proxy",
	"NO_PROXY", "no_proxy",
	"ALL_PROXY", "all_proxy",
	"FTP_PROXY", "ftp_proxy",
	// SSL / TLS certificates
	"SSL_CERT_FILE", "SSL_CERT_DIR",
	"CURL_CA_BUNDLE", "REQUESTS_CA_BUNDLE",
	"GIT_SSL_CAINFO", "GIT_SSL_CAPATH",
	// System / User / Locale
	"USER", "LOGNAME", "USERNAME",
	"LANG", "LC_ALL", "LC_CTYPE",
	"TZ",
	"TMPDIR", "TEMP", "TMP",
	// Windows system vars
	"SYSTEMROOT", "SystemRoot",
	"WINDIR", "windir",
	"APPDATA", "LOCALAPPDATA",
	"COMSPEC", "ComSpec",
	"PATHEXT",
}

func setEnvVar(envMap map[string]string, k, v string) {
	if runtime.GOOS == "windows" {
		for existing := range envMap {
			if strings.EqualFold(existing, k) {
				delete(envMap, existing)
			}
		}
	}
	envMap[k] = v
}

func getEnvironmentVars() []string {
	envMap := make(map[string]string)

	setEnvVar(envMap, "PATH", os.Getenv("PATH"))

	for _, e := range defaultForwardedEnvVars {
		if v := os.Getenv(e); v != "" {
			setEnvVar(envMap, e, v)
		}
	}

	for _, e := range util.Config.ForwardedEnvVars {
		if v := os.Getenv(e); v != "" {
			setEnvVar(envMap, e, v)
		}
	}

	for k, v := range util.Config.EnvVars {
		setEnvVar(envMap, k, v)
	}

	res := make([]string, 0, len(envMap))
	for k, v := range envMap {
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
