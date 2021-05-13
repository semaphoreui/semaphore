package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
)

func (d *BoltDb) CreateTemplate(template db.Template) (db.Template, error) {
	newTemplate, err := d.createObject(template.ProjectID, db.TemplateProps, template)
	return newTemplate.(db.Template), err
}

func (d *BoltDb) UpdateTemplate(template db.Template) error {
	return d.updateObject(template.ProjectID, db.TemplateProps, template)
}

func (d *BoltDb) GetTemplates(projectID int, params db.RetrieveQueryParams) (templates []db.Template, err error) {
	err = d.getObjects(projectID, db.TemplateProps, params, nil, &templates)
	return
}

func (d *BoltDb) GetTemplate(projectID int, templateID int) (template db.Template, err error) {
	err = d.getObject(projectID, db.TemplateProps, intObjectID(templateID), &template)
	return
}

func (d *BoltDb) DeleteTemplate(projectID int, templateID int) error {
	return d.deleteObject(projectID, db.TemplateProps, intObjectID(templateID))
}
