package sql

import (
	"encoding/json"

	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) CreateTemplate(template db.Template) (newTemplate db.Template, err error) {
	err = template.Validate()

	if err != nil {
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into project__template ("+
			"project_id, inventory_id, repository_id, environment_id, name, "+
			"playbook, arguments, allow_override_args_in_task, description, `type`, "+
			"start_version, build_template_id, view_id, autorun, survey_vars, "+
			"suppress_success_alerts, app, git_branch, runner_tag, task_params, "+
			"allow_override_branch_in_task, allow_parallel_tasks)"+
			"values ("+
			"?, ?, ?, ?, ?, "+
			"?, ?, ?, ?, ?, "+
			"?, ?, ?, ?, ?, "+
			"?, ?, ?, ?, ?,"+
			"?, ?)",
		template.ProjectID,
		template.InventoryID,
		template.RepositoryID,
		template.EnvironmentID,
		template.Name,

		template.Playbook,
		template.Arguments,
		template.AllowOverrideArgsInTask,
		template.Description,
		template.Type,

		template.StartVersion,
		template.BuildTemplateID,
		template.ViewID,
		template.Autorun,
		db.ObjectToJSON(template.SurveyVars),

		template.SuppressSuccessAlerts,
		template.App,
		template.GitBranch,
		template.RunnerTag,
		template.TaskParams,

		template.AllowOverrideBranchInTask,
		template.AllowParallelTasks,
	)

	if err != nil {
		return
	}

	err = d.UpdateTemplateVaults(template.ProjectID, insertID, template.Vaults)
	if err != nil {
		return
	}

	err = db.FillTemplate(d, &newTemplate)

	if err != nil {
		return
	}

	newTemplate = template
	newTemplate.ID = insertID

	return
}

func (d *SqlDb) UpdateTemplate(template db.Template) error {
	err := template.Validate()

	if err != nil {
		return err
	}

	_, err = d.exec("update project__template set "+
		"inventory_id=?, "+
		"repository_id=?, "+
		"environment_id=?, "+
		"name=?, "+
		"playbook=?, "+
		"arguments=?, "+
		"allow_override_args_in_task=?, "+
		"description=?, "+
		"`type`=?, "+
		"start_version=?,"+
		"build_template_id=?, "+
		"view_id=?, "+
		"autorun=?, "+
		"survey_vars=?, "+
		"suppress_success_alerts=?, "+
		"app=?, "+
		"`git_branch`=?, "+
		"task_params=?, "+
		"runner_tag=?, "+
		"allow_override_branch_in_task=?, "+
		"allow_parallel_tasks=? "+
		"where id=? and project_id=?",
		template.InventoryID,
		template.RepositoryID,
		template.EnvironmentID,
		template.Name,
		template.Playbook,
		template.Arguments,
		template.AllowOverrideArgsInTask,
		template.Description,
		template.Type,
		template.StartVersion,
		template.BuildTemplateID,
		template.ViewID,
		template.Autorun,
		db.ObjectToJSON(template.SurveyVars),
		template.SuppressSuccessAlerts,
		template.App,
		template.GitBranch,
		template.TaskParams,
		template.RunnerTag,
		template.AllowOverrideBranchInTask,
		template.AllowParallelTasks,

		template.ID,
		template.ProjectID,
	)
	if err != nil {
		return err
	}

	err = d.UpdateTemplateVaults(template.ProjectID, template.ID, template.Vaults)

	return err
}
func (d *SqlDb) SetTemplateDescription(projectID int, templateID int, description string) (err error) {

	_, err = d.exec("update project__template set "+
		"description=? "+
		"where id=? and project_id=?",
		description,
		templateID,
		projectID,
	)

	return
}

func (d *SqlDb) getTemplates(projectID int, userID *int, filter db.TemplateFilter, params db.RetrieveQueryParams) (templates []db.TemplateWithPerms, err error) {

	pp, err := params.Validate(db.TemplateProps)
	if err != nil {
		return
	}

	templates = []db.TemplateWithPerms{}

	type templateWithLastTask struct {
		db.TemplateWithPerms
		LastTaskID *int `db:"last_task_id"`
	}

	var view db.View

	if filter.ViewID != nil {
		view, err = d.GetView(projectID, *filter.ViewID)
		if err != nil {
			return
		}
	}

	q := squirrel.Select("pt.id",
		"pt.project_id",
		"pt.inventory_id",
		"pt.repository_id",
		"pt.environment_id",
		"pt.name",
		"pt.description",
		"pt.playbook",
		"pt.arguments",
		"pt.allow_override_args_in_task",
		"pt.build_template_id",
		"pt.start_version",
		"pt.view_id",
		"pt.`app`",
		"pt.`git_branch`",
		"pt.survey_vars",
		"pt.`type`",
		"pt.`tasks`",
		"pt.runner_tag",
		"pt.task_params",
		"pt.allow_override_branch_in_task",
		"pt.allow_parallel_tasks",
		"(SELECT `id` FROM `task` WHERE template_id = pt.id ORDER BY `id` DESC LIMIT 1) last_task_id").
		From("project__template pt")

	if filter.App != nil {
		q = q.Where("pt.app=?", *filter.App)
	}

	if filter.ViewID != nil {
		switch view.Type {
		case db.ViewTypeCustom:
			q = q.Where("pt.view_id=?", *filter.ViewID)
		case db.ViewTypeAll:
			if view.Filter != nil {
				// TODO: implement filter
			}
		}
	}

	if filter.BuildTemplateID != nil {
		q = q.Where("pt.build_template_id=?", *filter.BuildTemplateID)
		if filter.AutorunOnly {
			q = q.Where("pt.autorun=true")
		}
	}

	order := "ASC"
	var sortBy string

	if pp.SortBy != "" { // order by query param has priority
		sortBy = pp.SortBy
		if pp.SortInverted {
			order = "DESC"
		}
	} else if filter.ViewID != nil && view.SortColumn != nil {
		sortBy = *view.SortColumn
		if view.SortReverse {
			order = "DESC"
		}
	}

	switch sortBy {
	case "name", "playbook":
		q = q.Where("pt.project_id=?", projectID).
			OrderBy("pt." + sortBy + " " + order)
	case "inventory":
		q = q.LeftJoin("project__inventory pi ON (pt.inventory_id = pi.id)").
			Where("pt.project_id=?", projectID).
			OrderBy("pi.name " + order)
	case "environment":
		q = q.LeftJoin("project__environment pe ON (pt.environment_id = pe.id)").
			Where("pt.project_id=?", projectID).
			OrderBy("pe.name " + order)
	case "repository":
		q = q.LeftJoin("project__repository pr ON (pt.repository_id = pr.id)").
			Where("pt.project_id=?", projectID).
			OrderBy("pr.name " + order)
	default:
		q = q.Where("pt.project_id=?", projectID).
			OrderBy("pt.name " + order)
	}

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	var tpls []templateWithLastTask

	_, err = d.selectAll(&tpls, query, args...)

	if err != nil {
		return
	}

	taskIDs := make([]int, 0)

	for _, tpl := range tpls {
		if tpl.LastTaskID != nil {
			taskIDs = append(taskIDs, *tpl.LastTaskID)
		}
	}

	var tasks []db.TaskWithTpl
	err = d.getTasks(projectID, nil, taskIDs, db.RetrieveQueryParams{}, &tasks)

	if err != nil {
		return
	}

	for _, tpl := range tpls {
		template := tpl.TemplateWithPerms

		if tpl.LastTaskID != nil {
			for _, tsk := range tasks {
				if tsk.ID == *tpl.LastTaskID {
					err = tsk.Fill(d)
					if err != nil {
						return
					}
					template.LastTask = &tsk
					break
				}
			}
		}

		if tpl.SurveyVarsJSON != nil {
			err = json.Unmarshal([]byte(*tpl.SurveyVarsJSON), &template.SurveyVars)
		}

		if err != nil {
			return
		}

		template.Vaults, err = d.GetTemplateVaults(projectID, template.ID)
		if err != nil {
			return
		}

		templates = append(templates, template)
	}

	return
}

func (d *SqlDb) GetTemplatesWithPermissions(projectID int, userID int, filter db.TemplateFilter, params db.RetrieveQueryParams) (templates []db.TemplateWithPerms, err error) {
	return d.getTemplates(projectID, &userID, filter, params)
}

func (d *SqlDb) GetTemplates(projectID int, filter db.TemplateFilter, params db.RetrieveQueryParams) (templates []db.Template, err error) {
	res, err := d.getTemplates(projectID, nil, filter, params)
	if err != nil {
		return
	}

	for _, tpl := range res {
		templates = append(templates, tpl.Template)
	}

	return
}

func (d *SqlDb) GetTemplate(projectID int, templateID int) (template db.Template, err error) {
	err = d.selectOne(
		&template,
		"select * from project__template where project_id=? and id=?",
		projectID,
		templateID)

	if err != nil {
		return
	}

	err = db.FillTemplate(d, &template)
	return
}

func (d *SqlDb) DeleteTemplate(projectID int, templateID int) error {
	_, err := d.exec("delete from project__template where project_id=? and id=?", projectID, templateID)
	return err
}

func (d *SqlDb) GetTemplateRefs(projectID int, templateID int) (db.ObjectReferrers, error) {
	return d.getObjectRefs(projectID, db.TemplateProps, templateID)
}

func (d *SqlDb) GetTemplatePermission(projectID int, templateID int, userID int) (perm db.ProjectUserPermission, err error) {
	var projectUser db.ProjectUser
	projectUser, err = d.GetProjectUser(projectID, userID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			err = nil // user not in project, no permissions
		}
		return
	}

	perm = projectUser.Role.GetPermissions()

	var roleIDs []int
	query, args, err := squirrel.Select("role_id").
		From("project__user_role").
		Where("project_id = ?", projectID).
		Where("user_id = ?", userID).
		ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(&roleIDs, query, args...)
	if err != nil {
		return
	}

	if len(roleIDs) == 0 {
		return
	}

	var templateRoles []db.TemplatePerm
	query, args, err = squirrel.Select("*").
		From("template__role").
		Where("project_id = ?", projectID).
		Where("template_id = ?", templateID).
		Where(squirrel.Eq{"role_id": roleIDs}).
		ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(&templateRoles, query, args...)
	if err != nil {
		return
	}

	for _, tr := range templateRoles {
		perm |= tr.Permissions
	}

	return
}

func (d *SqlDb) GetTemplateRoles(projectID int, templateID int) (roles []db.TemplatePerm, err error) {
	query, args, err := squirrel.Select("*").
		From("template__role").
		Where("project_id = ?", projectID).
		Where("template_id = ?", templateID).
		ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(&roles, query, args...)
	return
}
func (d *SqlDb) CreateTemplateRole(role db.TemplatePerm) (newRole db.TemplatePerm, err error) {
	insertID, err := d.insert(
		"id",
		"insert into template__role (project_id, template_id, role_id, permissions) values (?, ?, ?, ?)",
		role.ProjectID,
		role.TemplateID,
		role.RoleID,
		role.Permissions)

	if err != nil {
		return
	}

	newRole = role
	newRole.ID = insertID
	return
}
func (d *SqlDb) DeleteTemplateRole(projectID int, templateID int, roleID int) error {
	_, err := d.exec("delete from template__role where project_id=? and template_id=? and role_id=?", projectID, templateID, roleID)
	return err
}
func (d *SqlDb) UpdateTemplateRole(role db.TemplatePerm) error {
	_, err := d.exec(
		"update template__role set permissions=? where id=?",
		role.Permissions,
		role.ID)

	return err
}
