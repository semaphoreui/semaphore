package db_lib

import (
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/pkg/task_logger"
)

func CreateApp(template db.Template, repository db.Repository, logger task_logger.Logger) LocalApp {
	switch template.App {
	case db.TemplateAnsible:
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
	case db.TemplateTerraform, db.TemplateTofu:
		return &TerraformApp{
			Template:   template,
			Repository: repository,
			Logger:     logger,
			Name:       TerraformAppName(template.App),
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
