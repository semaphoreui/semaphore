package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
)

func (d *BoltDb) CreateTemplate(template db.Template) (newTemplate db.Template, err error) {
	err = template.Validate()

	if err != nil {
		return
	}

	newTpl, err := d.createObject(template.ProjectID, db.TemplateProps, template)
	if err != nil {
		return
	}
	newTemplate = newTpl.(db.Template)
	err = db.FillTemplate(d, &newTemplate)
	return
}

func (d *BoltDb) UpdateTemplate(template db.Template) error {
	err := template.Validate()

	if err != nil {
		return err
	}

	return d.updateObject(template.ProjectID, db.TemplateProps, template)
}

func (d *BoltDb) getTemplates(projectID int, viewID *int, params db.RetrieveQueryParams) (templates []db.Template, err error) {
	var filter func(interface{}) bool
	if viewID != nil {
		filter = func (tpl interface{}) bool {
			template := tpl.(db.Template)
			return template.ViewID != nil && *template.ViewID == *viewID
		}
	}

	err = d.getObjects(projectID, db.TemplateProps, params, filter, &templates)

	if err != nil {
		return
	}

	err = db.FillTemplates(d, templates)

	return
}

func (d *BoltDb) GetTemplates(projectID int, params db.RetrieveQueryParams) ( []db.Template,  error) {
	return d.getTemplates(projectID, nil, params)
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
