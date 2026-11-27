package sql

import (
	"github.com/semaphoreui/semaphore/db"
)

// RoleMappingStoreImpl is the community stub: group role mappings are a Pro feature.
type RoleMappingStoreImpl struct {
}

func (d *RoleMappingStoreImpl) GetRoleMappings() ([]db.RoleMapping, error) {
	return []db.RoleMapping{}, nil
}

func (d *RoleMappingStoreImpl) CreateRoleMapping(mapping db.RoleMapping) (db.RoleMapping, error) {
	return mapping, db.ErrNotFound
}

func (d *RoleMappingStoreImpl) DeleteRoleMapping(mappingID int) (db.RoleMapping, []db.MappedRole, error) {
	return db.RoleMapping{}, nil, db.ErrNotFound
}

func (d *RoleMappingStoreImpl) SyncMappedRoles(userID int, provider string, groups []string) ([]db.MappedRole, []db.MappedRole, error) {
	return nil, nil, nil
}
