package db_lib

import (
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/pkg/task_logger"
)

func CreateApp(template db.Template, repository db.Repository, inventory db.Inventory, logger task_logger.Logger) LocalApp {
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
			},
		}
	case db.AppTerraform, db.AppTofu:
		return &TerraformApp{
			Template:   template,
			Repository: repository,
			Logger:     logger,
			Name:       string(template.App),
			Inventory:  inventory,
		}
	default:
		return &ShellApp{
			Template:   template,
			Repository: repository,
			Logger:     logger,
			App:        template.App,
		}
	}
}
