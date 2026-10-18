package db

// RoleMappingStore keeps the IdP group mappings and the roles members hold through them.
type RoleMappingStore interface {
	GetRoleMappings() ([]RoleMapping, error)
	CreateRoleMapping(mapping RoleMapping) (RoleMapping, error)
	// DeleteRoleMapping removes a mapping and returns it with the assignments that went with it.
	DeleteRoleMapping(mappingID int) (RoleMapping, []MappedRole, error)
	// SyncMappedRoles makes the user's mapped roles equal to the provider's mappings for these groups.
	SyncMappedRoles(userID int, provider string, groups []string) (added []MappedRole, removed []MappedRole, err error)
}
