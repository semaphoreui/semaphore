package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
)

func (d *BoltDb) CreateTemplate(template db.Template) (db.Template, error) {
	newTemplate, err := d.createObject(template.ProjectID, db.TemplateObject, template)
	return newTemplate.(db.Template), err
}

func (d *BoltDb) UpdateTemplate(template db.Template) error {
	return d.updateObject(template.ProjectID, db.TemplateObject, template)
}

func (d *BoltDb) GetTemplates(projectID int, params db.RetrieveQueryParams) (templates []db.Template, err error) {
	err = d.getObjects(projectID, db.TemplateObject, params, &templates)
	return
}

func (d *BoltDb) GetTemplate(projectID int, templateID int) (template db.Template, err error) {
	err = d.getObject(projectID, db.TemplateObject, templateID, &template)
	return
}

func (d *BoltDb) DeleteTemplate(projectID int, templateID int) error {
	return d.deleteObject(projectID, db.TemplateObject, templateID)
}
