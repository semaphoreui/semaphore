package sql

import (
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

func buildRolesQuery(rolesQuery db.RolesQuery) (sq.SelectBuilder, error) {
	query := sq.Select("*").From("`role`")
	var kinds db.RoleKind

	switch rolesQuery := rolesQuery.(type) {
	case db.GlobalRolesQuery:
		query = query.Where(sq.Eq{"project_id": nil})
		kinds = rolesQuery.Kinds
	case db.ProjectRolesQuery:
		query = query.Where(sq.Eq{"project_id": rolesQuery.ProjectID})
		kinds = db.RoleKindCustom
	case db.AvailableRolesQuery:
		query = query.Where(sq.Or{
			sq.Eq{"project_id": rolesQuery.ProjectID},
			sq.Eq{"project_id": nil},
		})
		kinds = rolesQuery.Kinds
	default:
		return query, fmt.Errorf("unsupported roles query: %T", rolesQuery)
	}

	switch kinds {
	case db.RoleKindBuiltin:
		return query.Where(sq.NotEq{"builtin_key": nil}), nil
	case db.RoleKindCustom:
		return query.Where(sq.Eq{"builtin_key": nil}), nil
	case db.RoleKindAll:
		return query, nil
	default:
		return query, fmt.Errorf("invalid role kind: %d", kinds)
	}
}

func (d *SqlDb) GetRoles(rolesQuery db.RolesQuery) ([]db.Role, error) {
	query, err := buildRolesQuery(rolesQuery)
	if err != nil {
		return nil, err
	}

	queryString, args, err := query.OrderBy("name").ToSql()
	if err != nil {
		return nil, err
	}

	var roles []db.Role
	_, err = d.selectAll(&roles, queryString, args...)
	return roles, err
}

func (d *SqlDb) GetRole(roleQuery db.RoleQuery) (db.Role, error) {
	query := sq.Select("*").From("`role`")

	switch roleQuery := roleQuery.(type) {
	case db.RoleByIDQuery:
		query = query.Where(sq.Eq{"id": roleQuery.ID})
	case db.BuiltinRoleQuery:
		query = query.Where(sq.Eq{"builtin_key": roleQuery.Key})
	default:
		return db.Role{}, fmt.Errorf("unsupported role query: %T", roleQuery)
	}

	queryString, args, err := query.ToSql()
	if err != nil {
		return db.Role{}, err
	}

	var role db.Role
	err = d.selectOne(&role, queryString, args...)
	return role, err
}

func sameProjectScope(left *int, right *int) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func (d *SqlDb) UpdateRole(role db.Role) error {
	existing, err := d.GetRole(db.RoleByIDQuery{ID: role.ID})
	if err != nil {
		return err
	}
	if existing.IsBuiltin() {
		return fmt.Errorf("built-in roles cannot be updated: %w", db.ErrInvalidOperation)
	}
	if err := db.ValidateCustomRole(role); err != nil {
		return err
	}
	if !sameProjectScope(existing.ProjectID, role.ProjectID) {
		return &common_errors.ValidationError{Message: "Role scope cannot be changed"}
	}

	_, err = d.exec(
		"update `role` set name=?, permissions=? where id=? and builtin_key is null",
		role.Name,
		role.Permissions,
		role.ID)
	return err
}

func (d *SqlDb) CreateRole(role db.Role) (db.Role, error) {
	if err := db.ValidateCustomRole(role); err != nil {
		return role, err
	}

	roleID, err := d.insert(
		"id",
		"insert into `role` (name, permissions, project_id, builtin_key) values (?, ?, ?, null)",
		role.Name,
		role.Permissions,
		role.ProjectID)
	if err != nil {
		return role, err
	}

	role.ID = roleID
	role.BuiltinKey = nil
	return role, nil
}

func (d *SqlDb) DeleteRole(roleID int) error {
	role, err := d.GetRole(db.RoleByIDQuery{ID: roleID})
	if err != nil {
		return err
	}
	if role.IsBuiltin() {
		return fmt.Errorf("built-in roles cannot be deleted: %w", db.ErrInvalidOperation)
	}

	res, err := d.exec("delete from `role` where id=? and builtin_key is null", roleID)
	err = validateMutationResult(res, err)
	if errors.Is(err, db.ErrInvalidOperation) {
		return fmt.Errorf("role cannot be deleted while it is assigned or granted: %w", err)
	}
	return err
}
