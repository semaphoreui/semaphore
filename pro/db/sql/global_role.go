package sql

import (
	"github.com/semaphoreui/semaphore/db"
)

// GlobalRoleStoreImpl is the community stub: there are no global roles, only the admin flag counts.
type GlobalRoleStoreImpl struct {
}

func (d *GlobalRoleStoreImpl) GetGlobalRoles() ([]db.GlobalRole, error) {
	return []db.GlobalRole{}, nil
}

func (d *GlobalRoleStoreImpl) GetGlobalRole(roleID int) (db.GlobalRole, error) {
	return db.GlobalRole{}, db.ErrNotFound
}

func (d *GlobalRoleStoreImpl) CreateGlobalRole(role db.GlobalRole) (db.GlobalRole, error) {
	return db.GlobalRole{}, db.ErrNotFound
}

func (d *GlobalRoleStoreImpl) UpdateGlobalRole(role db.GlobalRole) error {
	return db.ErrNotFound
}

func (d *GlobalRoleStoreImpl) DeleteGlobalRole(roleID int) error {
	return db.ErrNotFound
}

func (d *GlobalRoleStoreImpl) DeleteGlobalRoleWithAssignments(roleID int) ([]db.GlobalRoleAssignment, error) {
	return nil, db.ErrNotFound
}

func (d *GlobalRoleStoreImpl) GetGlobalRoleHolders(roleID int) ([]db.GlobalRoleAssignment, error) {
	return []db.GlobalRoleAssignment{}, nil
}

func (d *GlobalRoleStoreImpl) GetGlobalRoleAssignments(userID int) ([]db.GlobalRoleAssignment, error) {
	return []db.GlobalRoleAssignment{}, nil
}

func (d *GlobalRoleStoreImpl) GetGlobalRoleAssignment(userID int, assignmentID int) (db.GlobalRoleAssignment, error) {
	return db.GlobalRoleAssignment{}, db.ErrNotFound
}

func (d *GlobalRoleStoreImpl) CreateGlobalRoleAssignment(userID int, roleID int) (db.GlobalRoleAssignment, error) {
	return db.GlobalRoleAssignment{}, db.ErrNotFound
}

func (d *GlobalRoleStoreImpl) DeleteGlobalRoleAssignment(userID int, assignmentID int) error {
	return db.ErrNotFound
}

func (d *GlobalRoleStoreImpl) GetEffectiveGlobalPermissions(userID int) (db.GlobalPermission, error) {
	return 0, nil
}
