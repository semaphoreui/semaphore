package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
)

func (d *BoltDb) CreateTemplate(template db.Template) (newTemplate db.Template, err error) {
	err = template.Validate()

	if err != nil {
		return
	}

	template.SurveyVarsJSON = db.ObjectToJSON(template.SurveyVars)
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

	template.SurveyVarsJSON = db.ObjectToJSON(template.SurveyVars)
	return d.updateObject(template.ProjectID, db.TemplateProps, template)
}

func (d *BoltDb) GetTemplates(projectID int, filter db.TemplateFilter, params db.RetrieveQueryParams) (templates []db.Template, err error) {
	var ftr = func(tpl interface{}) bool {
		template := tpl.(db.Template)
		var res = true
		if filter.ViewID != nil {
			res = res && template.ViewID != nil && *template.ViewID == *filter.ViewID
		}
		if filter.BuildTemplateID != nil {
			res = res && template.BuildTemplateID != nil && *template.BuildTemplateID == *filter.BuildTemplateID
			if filter.AutorunOnly {
				res = res && template.Autorun
			}
		}
		return res
	}

	err = d.getObjects(projectID, db.TemplateProps, params, ftr, &templates)

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
