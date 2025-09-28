package sql

import "github.com/semaphoreui/semaphore/db"

func (d *SqlDb) GetRole(roleID int) (db.Role, error) {
	var role db.Role
	err := d.selectOne(&role, "select * from `role` where id=?", roleID)
	return role, err
}

func (d *SqlDb) GetRoles() ([]db.Role, error) {
	var roles []db.Role
	_, err := d.selectAll(&roles, "select * from `role` order by name")
	return roles, err
}

func (d *SqlDb) UpdateRole(role db.Role) error {
	_, err := d.exec(
		"update `role` set slug=?, name=?, permissions=? where id=?",
		role.Slug,
		role.Name,
		role.Permissions,
		role.ID)
	return err
}

func (d *SqlDb) CreateRole(role db.Role) (db.Role, error) {
	insertID, err := d.insert(
		"id",
		"insert into `role` (slug, name, permissions) values (?, ?, ?)",
		role.Slug,
		role.Name,
		role.Permissions)

	if err != nil {
		return role, err
	}

	role.ID = insertID
	return role, nil
}

func (d *SqlDb) DeleteRole(roleID int) error {
	res, err := d.exec("delete from `role` where id=?", roleID)
	return validateMutationResult(res, err)
}

func (d *SqlDb) GetRoleBySlug(slug string) (db.Role, error) {
	var role db.Role
	err := d.selectOne(&role, "select * from `role` where slug=?", slug)
	return role, err
}
