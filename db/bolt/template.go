package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
)

func (d *BoltDb) CreateTemplate(template db.Template) (newTemplate db.Template, err error) {
	newTpl, err := d.createObject(template.ProjectID, db.TemplateProps, template)
	if err != nil {
		return
	}
	newTemplate = newTpl.(db.Template)
	err = db.FillTemplate(d, &newTemplate)
	return
}

func (d *BoltDb) UpdateTemplate(template db.Template) error {
	return d.updateObject(template.ProjectID, db.TemplateProps, template)
}

func (d *BoltDb) GetTemplates(projectID int, params db.RetrieveQueryParams) (templates []db.Template, err error) {
	err = d.getObjects(projectID, db.TemplateProps, params, nil, &templates)

	if err != nil {
		return
	}

	err = db.FillTemplates(d, templates)

	return
}

func (d *BoltDb) GetTemplate(projectID int, templateID int) (template db.Template, err error) {
	err = d.getObject(projectID, db.TemplateProps, intObjectID(templateID), &template)
	if err != nil {
		return
	}
	err = db.FillTemplate(d, &template)
	return
}

func (d *BoltDb) DeleteTemplate(projectID int, templateID int) error {
	return d.deleteObject(projectID, db.TemplateProps, intObjectID(templateID))
}
