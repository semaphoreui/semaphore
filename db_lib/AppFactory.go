package db_lib

import (
	"github.com/ansible-semaphore/semaphore/db"
)

func CreateApp(template db.Template, repository db.Repository) LocalApp {
	switch template.App {
	case db.TemplateAnsible:
		return &AnsibleApp{
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
