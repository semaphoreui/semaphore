package db_lib

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

type AppVersionConfig struct {
	Path string
	Args []string
}

func ResolveAppVersion(template db.Template) AppVersionConfig {
	if template.AppVersionID != nil {
		if v, ok := util.Config.AppVersions[*template.AppVersionID]; ok {
			return AppVersionConfig{
				Path: v.AppPath,
				Args: v.AppArgs,
			}
		}
	}

	if app, ok := util.Config.Apps[string(template.App)]; ok {
		return AppVersionConfig{
			Path: app.AppPath,
			Args: app.AppArgs,
		}
	}

	return AppVersionConfig{}
}

func CreateApp(template db.Template, repository db.Repository, inventory db.Inventory, logger task_logger.Logger) LocalApp {
	versionCfg := ResolveAppVersion(template)

	switch template.App {
	case db.AppAnsible:
		return &AnsibleApp{
			Template:   template,
			Repository: repository,
			Logger:     logger,
			Playbook: &AnsiblePlaybook{
				TemplateID: template.ID,
				Repository: repository,
				Logger:     logger,
				VersionCfg: versionCfg,
			},
		}
	case db.AppTerraform, db.AppTofu, db.AppTerragrunt:
		return &TerraformApp{
			Template:   template,
			Repository: repository,
			Logger:     logger,
			Name:       string(template.App),
			Inventory:  inventory,
			VersionCfg: versionCfg,
		}
	default:
		return &ShellApp{
			Template:   template,
			Repository: repository,
			Logger:     logger,
			App:        template.App,
			VersionCfg: versionCfg,
		}
	}
}
