package sql

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

func buildRoleQuery(roleQuery db.RoleQuery) (sq.SelectBuilder, error) {
	query := sq.Select("*").From("`role`")
	var kinds db.RoleKind

	switch roleQuery := roleQuery.(type) {
	case db.GlobalRoleQuery:
		query = query.Where(sq.Eq{"project_id": nil})
		kinds = roleQuery.Kinds
	case db.ProjectRoleQuery:
		query = query.Where(sq.Eq{"project_id": roleQuery.ProjectID})
		kinds = db.RoleKindCustom
	case db.AvailableRoleQuery:
		query = query.Where(sq.Or{
			sq.Eq{"project_id": roleQuery.ProjectID},
			sq.Eq{"project_id": nil},
		})
		kinds = roleQuery.Kinds
	default:
		return query, fmt.Errorf("unsupported role query: %T", roleQuery)
	}

	switch kinds {
	case db.RoleKindBuiltin:
		return query.Where(sq.Eq{"is_builtin": true}), nil
	case db.RoleKindCustom:
		return query.Where(sq.Eq{"is_builtin": false}), nil
	case db.RoleKindAll:
		return query, nil
	default:
		return query, fmt.Errorf("invalid role kind: %d", kinds)
	}
}

func (d *SqlDb) GetRoles(roleQuery db.RoleQuery) ([]db.Role, error) {
	query, err := buildRoleQuery(roleQuery)
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

func (d *SqlDb) GetRoleBySlug(slug string, roleQuery db.RoleQuery) (db.Role, error) {
	query, err := buildRoleQuery(roleQuery)
	if err != nil {
		return db.Role{}, err
	}

	queryString, args, err := query.Where(sq.Eq{"slug": slug}).ToSql()
	if err != nil {
		return db.Role{}, err
	}

	var role db.Role
	err = d.selectOne(&role, queryString, args...)
	return role, err
}

func (d *SqlDb) UpdateRole(role db.Role) error {
	_, err := d.exec(
		"update `role` set name=?, permissions=? where slug=? and is_builtin=false",
		role.Name,
		role.Permissions,
		role.Slug)
	return err
}

func (d *SqlDb) CreateRole(role db.Role) (db.Role, error) {
	if err := db.ValidateRole(role); err != nil {
		return role, err
	}

	_, err := d.insert(
		"",
		"insert into `role` (slug, name, permissions, project_id, is_builtin) values (?, ?, ?, ?, ?)",
		role.Slug,
		role.Name,
		role.Permissions,
		role.ProjectID,
		role.IsBuiltin)

	if err != nil {
		return role, err
	}

	return role, nil
}

func (d *SqlDb) DeleteRole(slug string) error {
	res, err := d.exec("delete from `role` where slug=? and is_builtin=false", slug)
	return validateMutationResult(res, err)
}
