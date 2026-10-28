package sql

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

func (d *SqlDb) GetWorkflowRole(projectID int, workflowID int, id int) (perm db.WorkflowRolePerm, err error) {
	query, args, err := sq.Select("pwr.*", "r.slug role_slug").
		From("project__workflow_role pwr").
		Join("`role` r ON r.id = pwr.role_id").
		Where("pwr.project_id = ?", projectID).
		Where("pwr.workflow_id = ?", workflowID).
		Where("pwr.id = ?", id).
		ToSql()
	if err != nil {
		return
	}
	err = d.selectOne(&perm, query, args...)
	return
}

// GetWorkflowPermission returns the OR of the workflow grants of the member's roles.
func (d *SqlDb) GetWorkflowPermission(projectID int, workflowID int, userID int) (db.ProjectUserPermission, error) {
	return d.workflowGrants(`
		select
		  pwr.permissions
		from project__workflow_role pwr
		where pwr.project_id = ? and pwr.workflow_id = ?
		  and pwr.role_id in (
		    select pur.role_id from project__user_role pur where pur.project_id = ? and pur.user_id = ?
		    union
		    select m.role_id
		    from project__user_mapped_role mr
		    join group_role_mapping m on m.id = mr.mapping_id
		    where mr.project_id = ? and mr.user_id = ?
		  )
	`, projectID, workflowID, projectID, userID, projectID, userID)
}

// Per workflow of the project, the OR of the grants of the member's roles; no grant, no entry.
func (d *SqlDb) GetWorkflowGrantsForUser(projectID int, userID int) (map[int]db.ProjectUserPermission, error) {
	var rows []db.WorkflowRolePerm
	_, err := d.selectAll(&rows, `
		select
		  pwr.workflow_id, pwr.permissions
		from project__workflow_role pwr
		where pwr.project_id = ?
		  and pwr.role_id in (
		    select pur.role_id from project__user_role pur where pur.project_id = ? and pur.user_id = ?
		    union
		    select m.role_id
		    from project__user_mapped_role mr
		    join group_role_mapping m on m.id = mr.mapping_id
		    where mr.project_id = ? and mr.user_id = ?
		  )
	`, projectID, projectID, userID, projectID, userID)
	if err != nil {
		return nil, err
	}
	grants := map[int]db.ProjectUserPermission{}
	for _, row := range rows {
		grants[row.WorkflowID] |= row.Permissions
	}
	return grants, nil
}

func (d *SqlDb) workflowGrants(query string, args ...any) (db.ProjectUserPermission, error) {
	var rows []db.WorkflowRolePerm
	if _, err := d.selectAll(&rows, query, args...); err != nil {
		return 0, err
	}
	var permissions db.ProjectUserPermission
	for _, row := range rows {
		permissions |= row.Permissions
	}
	return permissions, nil
}

func (d *SqlDb) GetWorkflowRoles(projectID int, workflowID int) (roles []db.WorkflowRolePerm, err error) {
	query, args, err := sq.Select("pwr.*", "r.slug role_slug").
		From("project__workflow_role pwr").
		Join("`role` r ON r.id = pwr.role_id").
		Where("pwr.project_id = ?", projectID).
		Where("pwr.workflow_id = ?", workflowID).
		ToSql()
	if err != nil {
		return
	}
	_, err = d.selectAll(&roles, query, args...)
	return
}

func (d *SqlDb) CreateWorkflowRole(role db.WorkflowRolePerm) (newRole db.WorkflowRolePerm, err error) {
	available, err := d.Sql().SelectInt(d.PrepareQuery(
		"select count(*) from `role` where id=? and (project_id is null or project_id=?)"), role.RoleID, role.ProjectID)
	if err != nil {
		return
	}
	if available == 0 {
		err = &common_errors.ValidationError{Message: "Role is not available in this project"}
		return
	}
	granted, err := d.Sql().SelectInt(d.PrepareQuery(
		"select count(*) from project__workflow_role where workflow_id=? and role_id=?"), role.WorkflowID, role.RoleID)
	if err != nil {
		return
	}
	if granted > 0 {
		err = &common_errors.ValidationError{Message: "Role already has permissions on this workflow"}
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into project__workflow_role (project_id, workflow_id, role_id, permissions) values (?, ?, ?, ?)",
		role.ProjectID,
		role.WorkflowID,
		role.RoleID,
		role.Permissions)
	if err != nil {
		return
	}

	newRole = role
	newRole.ID = insertID
	return
}

func (d *SqlDb) DeleteWorkflowRole(projectID int, workflowID int, id int) error {
	res, err := d.exec("delete from project__workflow_role where project_id=? and workflow_id=? and id=?", projectID, workflowID, id)
	return requireDeletedRow(res, err)
}

func (d *SqlDb) UpdateWorkflowRole(role db.WorkflowRolePerm) error {
	_, err := d.exec(
		"update project__workflow_role set permissions=? where project_id=? and workflow_id=? and id=?",
		role.Permissions,
		role.ProjectID,
		role.WorkflowID,
		role.ID)
	return err
}
