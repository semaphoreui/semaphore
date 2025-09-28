package bolt

import (
	"github.com/semaphoreui/semaphore/db"
)

func (d *BoltDb) GetRole(roleID int) (role db.Role, err error) {
	err = d.getObject(0, db.RoleProps, intObjectID(roleID), &role)
	return
}

func (d *BoltDb) GetRoles() (roles []db.Role, err error) {
	err = d.getObjects(0, db.RoleProps, db.RetrieveQueryParams{}, nil, &roles)
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

func (d *BoltDb) DeleteRole(roleID int) error {
	return d.deleteObject(0, db.RoleProps, intObjectID(roleID), nil)
}
