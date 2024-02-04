package db_lib

import (
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/lib"
)

func CreateApp(template db.Template, repository db.Repository, logger lib.Logger) LocalApp {
	switch template.App {
	case db.TemplateAnsible:
		return &AnsibleApp{
			Template:   template,
			Repository: repository,
			Logger:     logger,
			Playbook: &AnsiblePlaybook{
				TemplateID: template.ID,
				Repository: repository,
			},
		}
	case db.TemplateTerraform:
		return &TerraformApp{
			Template:   template,
			Repository: repository,
			Playbook: &AnsiblePlaybook{
				TemplateID: template.ID,
				Repository: repository,
			},
		}
	default:
		panic("unknown app")
	}
}
