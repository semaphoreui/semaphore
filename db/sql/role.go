package sql

import "github.com/semaphoreui/semaphore/db"

func (d *SqlDb) GetGlobalRoleBySlug(slug string) (db.Role, error) {
	var role db.Role
	err := d.selectOne(&role, "select * from `role` where slug=? and project_id is null", slug)
	return role, err
}

func (d *SqlDb) GetProjectRoles(projectID int) ([]db.Role, error) {
	var roles []db.Role
	_, err := d.selectAll(&roles, "select * from `role` where project_id=? order by name", projectID)
	return roles, err
}

func (d *SqlDb) GetGlobalRoles() ([]db.Role, error) {
	var roles []db.Role
	_, err := d.selectAll(&roles, "select * from `role` where project_id is null order by name")
	return roles, err
}

// roleScope limits a role query to the global roles or to one project.
func roleScope(slug string, projectID *int) (string, []any) {
	if projectID == nil {
		return " where slug=? and project_id is null", []any{slug}
	}
	return " where slug=? and project_id=?", []any{slug, *projectID}
}

func (d *SqlDb) UpdateRole(role db.Role) error {
	where, args := roleScope(role.Slug, role.ProjectID)
	res, err := d.exec(
		"update `role` set name=?, permissions=?"+where,
		append([]any{role.Name, role.Permissions}, args...)...)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil || affected > 0 {
		return err
	}

	// MySQL reports no affected rows for an update that changes nothing.
	var existing db.Role
	return d.selectOne(&existing, "select * from `role`"+where, args...)
}

func (d *SqlDb) CreateRole(role db.Role) (db.Role, error) {
	_, err := d.insert(
		"",
		"insert into `role` (slug, name, permissions, project_id) values (?, ?, ?, ?)",
		role.Slug,
		role.Name,
		role.Permissions,
		role.ProjectID)

	if err != nil {
		return role, err
	}

	return role, nil
}

func (d *SqlDb) DeleteRole(slug string, projectID *int) error {
	where, args := roleScope(slug, projectID)
	res, err := d.exec("delete from `role`"+where, args...)
	if err = validateMutationResult(res, err); err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return db.ErrNotFound
	}
	return nil
}

func (d *SqlDb) GetProjectRole(projectID int, slug string) (db.Role, error) {
	var role db.Role
	err := d.selectOne(&role, "select * from `role` where slug=? and project_id=?", slug, projectID)
	return role, err
}

func (d *SqlDb) GetProjectOrGlobalRoleBySlug(projectID int, slug string) (db.Role, error) {
	var role db.Role
	err := d.selectOne(
		&role,
		"select * from `role` where slug=? and (project_id=? or project_id is null)",
		slug,
		projectID)
	return role, err
}
