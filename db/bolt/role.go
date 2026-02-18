package bolt

import (
	"github.com/semaphoreui/semaphore/db"
)

func (d *BoltDb) GetGlobalRole(roleID int) (role db.Role, err error) {
	err = d.getObject(0, db.RoleProps, intObjectID(roleID), &role)
	return
}

func (d *BoltDb) GetGlobalRoleBySlug(slug string) (db.Role, error) {
	var roles []db.Role

	err := d.getObjects(0, db.RoleProps, db.RetrieveQueryParams{}, func(i any) bool {
		role := i.(db.Role)
		return role.Slug == slug && role.ProjectID == nil
	}, &roles)

	if err != nil {
		return db.Role{}, err
	}

	if len(roles) == 0 {
		return db.Role{}, db.ErrNotFound
	}

	return roles[0], nil
}

func (d *BoltDb) GetProjectRoles(projectID int) (roles []db.Role, err error) {
	err = d.getObjects(0, db.RoleProps, db.RetrieveQueryParams{}, func(i any) bool {
		role := i.(db.Role)
		return role.ProjectID != nil && *role.ProjectID == projectID
	}, &roles)
	return
}

func (d *BoltDb) GetGlobalRoles() (roles []db.Role, err error) {
	err = d.getObjects(0, db.RoleProps, db.RetrieveQueryParams{}, func(i any) bool {
		role := i.(db.Role)
		return role.ProjectID == nil
	}, &roles)
	return
}

func (d *BoltDb) UpdateRole(role db.Role) error {
	return d.updateObject(0, db.RoleProps, role)
}

func (d *BoltDb) CreateRole(role db.Role) (newRole db.Role, err error) {
	newRoleInterface, err := d.createObject(0, db.RoleProps, role)
	if err != nil {
		return
	}
	newRole = newRoleInterface.(db.Role)
	return
}

func (d *BoltDb) DeleteRole(slug string) error {
	return d.deleteObject(0, db.RoleProps, strObjectID(slug), nil)
}

func (d *BoltDb) GetProjectRole(projectID int, slug string) (db.Role, error) {
	var role db.Role
	err := d.getObject(0, db.RoleProps, strObjectID(slug), &role)
	if err != nil {
		return db.Role{}, err
	}

	// Verify the role belongs to the specified project
	if role.ProjectID == nil || *role.ProjectID != projectID {
		return db.Role{}, db.ErrNotFound
	}

	return role, nil
}

func (d *BoltDb) GetProjectOrGlobalRoleBySlug(projectID int, slug string) (db.Role, error) {
	var roles []db.Role

	err := d.getObjects(0, db.RoleProps, db.RetrieveQueryParams{}, func(i any) bool {
		role := i.(db.Role)
		return role.Slug == slug
	}, &roles)

	if err != nil {
		return db.Role{}, err
	}

	if len(roles) == 0 {
		return db.Role{}, db.ErrNotFound
	}

	return roles[0], nil
}
