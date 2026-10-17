package sql

import (
	"github.com/semaphoreui/semaphore/db"
)

// GlobalRoleStoreImpl is the community stub: global roles grant nothing, only the admin flag counts.
type GlobalRoleStoreImpl struct {
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
