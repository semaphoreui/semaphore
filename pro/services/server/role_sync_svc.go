package server

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
)

// RoleSyncServiceImpl is the community stub: group role mappings are a Pro feature.
type RoleSyncServiceImpl struct{}

func NewRoleSyncService(mappings db.RoleMappingStore) pro_interfaces.RoleSyncService {
	return &RoleSyncServiceImpl{}
}

func (s *RoleSyncServiceImpl) SyncLoginRoles(userID int, provider string, claims map[string]any) ([]db.MappedRole, []db.MappedRole, error) {
	return nil, nil, nil
}
