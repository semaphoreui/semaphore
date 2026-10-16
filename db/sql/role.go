package sql

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

func (d *SqlDb) GetGlobalRoleBySlug(slug string) (role db.Role, err error) {
	err = d.selectOne(&role, "select * from `role` where lower(slug)=lower(?) and project_id is null", slug)
	return
}

func (d *SqlDb) GetProjectRole(projectID int, slug string) (role db.Role, err error) {
	err = d.selectOne(&role, "select * from `role` where lower(slug)=lower(?) and project_id=?", slug, projectID)
	return
}

// GetProjectOrGlobalRoleBySlug prefers the project role when a global role has the same slug.
func (d *SqlDb) GetProjectOrGlobalRoleBySlug(projectID int, slug string) (role db.Role, err error) {
	err = d.selectOne(&role, `
		select
		  *
		from role
		where lower(slug)=lower(?)
		  and (project_id=? or project_id is null)
		order by project_id is null
		limit 1
	`, slug, projectID)
	return
}

func (d *SqlDb) GetProjectRoles(projectID int) (roles []db.Role, err error) {
	_, err = d.selectAll(&roles, "select * from `role` where project_id=? order by name", projectID)
	return
}

func (d *SqlDb) GetGlobalRoles() (roles []db.Role, err error) {
	_, err = d.selectAll(&roles, "select * from `role` where project_id is null order by name")
	return
}

// CreateRole inserts a custom role. A global slug clashes with every role, a project slug with global roles and its own project.
func (d *SqlDb) CreateRole(role db.Role) (db.Role, error) {
	scope := 0
	if role.ProjectID != nil {
		scope = *role.ProjectID
	}
	taken, err := d.Sql().SelectInt(d.PrepareQuery(
		"select count(*) from `role` where lower(slug)=lower(?) and (?=0 or scope in (0, ?))"), role.Slug, scope, scope)
	if err != nil {
		return role, err
	}
	if taken > 0 {
		return role, &common_errors.ValidationError{Message: "Role slug is taken"}
	}

	role.ID, err = d.insert("id", `
		insert into role (slug, name, permissions, project_id, scope, builtin)
		values (?, ?, ?, ?, ?, false)
	`, role.Slug, role.Name, role.Permissions, role.ProjectID, scope)
	role.Scope, role.Builtin = scope, false
	return role, err
}

// UpdateRole changes only the name and the permissions. Built-in roles never match.
func (d *SqlDb) UpdateRole(role db.Role) error {
	_, err := d.exec("update `role` set name=?, permissions=? where id=? and builtin=false",
		role.Name, role.Permissions, role.ID)
	return err
}

// DeleteRole removes a custom role. The foreign keys refuse a role that is still assigned or granted.
func (d *SqlDb) DeleteRole(roleID int) error {
	res, err := d.exec("delete from `role` where id=? and builtin=false", roleID)
	return requireDeletedRow(res, err)
}
