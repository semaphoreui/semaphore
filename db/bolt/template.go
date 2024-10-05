package bolt

import (
	"errors"
	"github.com/ansible-semaphore/semaphore/db"
	"go.etcd.io/bbolt"
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
	err = d.UpdateTemplateVaults(template.ProjectID, newTemplate.ID, template.Vaults)
	if err != nil {
		return
	}
	err = db.FillTemplate(d, &newTemplate)
	return
}

func (d *BoltDb) UpdateTemplate(template db.Template) error {
	err := template.Validate()

	if err != nil {
		return err
	}

	template.SurveyVarsJSON = db.ObjectToJSON(template.SurveyVars)
	err = d.updateObject(template.ProjectID, db.TemplateProps, template)
	if err != nil {
		return err
	}
	return d.UpdateTemplateVaults(template.ProjectID, template.ID, template.Vaults)
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

	templatesMap := make(map[int]*db.Template)

	for i := 0; i < len(templates); i++ {
		templates[i].Vaults, err = d.GetTemplateVaults(projectID, templates[i].ID)
		templatesMap[templates[i].ID] = &templates[i]
	}

	unfilledTemplateCount := len(templates)

	var errEndOfTemplates = errors.New("no more templates to filling")

	err = d.apply(projectID, db.TaskProps, db.RetrieveQueryParams{}, func(i interface{}) error {
		task := i.(db.Task)

		tpl, ok := templatesMap[task.TemplateID]
		if !ok {
			return nil
		}

		if tpl.LastTask != nil {
			return nil
		}

		tpl.LastTask = &db.TaskWithTpl{
			Task:             task,
			TemplatePlaybook: tpl.Playbook,
			TemplateAlias:    tpl.Name,
			TemplateType:     tpl.Type,
			TemplateApp:      tpl.App,
		}

		unfilledTemplateCount--

		if unfilledTemplateCount <= 0 {
			return errEndOfTemplates
		}

		return nil
	})

	if errors.Is(err, errEndOfTemplates) {
		err = nil
	}

	return
}

func (d *BoltDb) getRawTemplate(projectID int, templateID int) (template db.Template, err error) {
	err = d.getObject(projectID, db.TemplateProps, intObjectID(templateID), &template)
	return
}

func (d *BoltDb) GetTemplate(projectID int, templateID int) (template db.Template, err error) {
	template, err = d.getRawTemplate(projectID, templateID)
	if err != nil {
		return
	}
	err = db.FillTemplate(d, &template)
	return
}

func (d *BoltDb) deleteTemplate(projectID int, templateID int, tx *bbolt.Tx) (err error) {
	inUse, err := d.isObjectInUse(projectID, db.TemplateProps, intObjectID(templateID), db.TemplateProps)

	if err != nil {
		return err
	}

	if inUse {
		return db.ErrInvalidOperation
	}

	tasks, err := d.GetTemplateTasks(projectID, templateID, db.RetrieveQueryParams{})
	if err != nil {
		return
	}
	for _, task := range tasks {
		err = d.deleteTaskWithOutputs(projectID, task.ID, true, tx)
		if err != nil {
			return
		}
	}

	schedules, err := d.GetTemplateSchedules(projectID, templateID)
	if err != nil {
		return
	}
	for _, sch := range schedules {
		err = d.deleteSchedule(projectID, sch.ID, tx)
		if err != nil {
			return
		}
	}

	return d.deleteObject(projectID, db.TemplateProps, intObjectID(templateID), tx)
}

func (d *BoltDb) DeleteTemplate(projectID int, templateID int) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		return d.deleteTemplate(projectID, templateID, tx)
	})
}

func (d *BoltDb) GetTemplateRefs(projectID int, templateID int) (db.ObjectReferrers, error) {
	return d.getObjectRefs(projectID, db.TemplateProps, templateID)
}
