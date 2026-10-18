package sql

import (
	"github.com/Masterminds/squirrel"
	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

func (d *SqlDb) GetSharedRoleBySlug(slug string) (role db.Role, err error) {
	err = d.selectOne(&role, "select * from `role` where lower(slug)=lower(?) and project_id is null", slug)
	return
}

func (d *SqlDb) GetProjectRole(projectID int, slug string) (role db.Role, err error) {
	err = d.selectOne(&role, "select * from `role` where lower(slug)=lower(?) and project_id=?", slug, projectID)
	return
}

// GetProjectOrSharedRoleBySlug prefers the project role when a global role has the same slug.
func (d *SqlDb) GetProjectOrSharedRoleBySlug(projectID int, slug string) (role db.Role, err error) {
	err = d.selectOne(&role, `
		select
		  *
		from role
		where lower(slug)=lower(?)
		  and (project_id=? or project_id is null)
		order by project_id is null
		limit 1
	`, slug, projectID)
	return
}

func (d *SqlDb) GetProjectRoles(projectID int) (roles []db.Role, err error) {
	_, err = d.selectAll(&roles, "select * from `role` where project_id=? order by name", projectID)
	return
}

func (d *SqlDb) GetSharedRoles() (roles []db.Role, err error) {
	_, err = d.selectAll(&roles, "select * from `role` where project_id is null order by name")
	return
}

// CreateRole inserts a custom role. A global slug clashes with every role, a project slug with global roles and its own project.
func (d *SqlDb) CreateRole(role db.Role) (db.Role, error) {
	scope := 0
	if role.ProjectID != nil {
		scope = *role.ProjectID
	}
	taken, err := d.Sql().SelectInt(d.PrepareQuery(
		"select count(*) from `role` where lower(slug)=lower(?) and (?=0 or scope in (0, ?))"), role.Slug, scope, scope)
	if err != nil {
		return role, err
	}
	if taken > 0 {
		return role, &common_errors.ValidationError{Message: "Role slug is taken"}
	}

	role.ID, err = d.insert("id", `
		insert into role (slug, name, permissions, project_id, scope, builtin)
		values (?, ?, ?, ?, ?, false)
	`, role.Slug, role.Name, role.Permissions, role.ProjectID, scope)
	role.Scope, role.Builtin = scope, false
	return role, err
}

// UpdateRole changes only the name and the permissions. Built-in roles never match.
func (d *SqlDb) UpdateRole(role db.Role) error {
	_, err := d.exec("update `role` set name=?, permissions=? where id=? and builtin=false",
		role.Name, role.Permissions, role.ID)
	return err
}

// DeleteRole removes a custom role. The foreign keys refuse a role that is still assigned or granted.
func (d *SqlDb) DeleteRole(roleID int) error {
	res, err := d.exec("delete from `role` where id=? and builtin=false", roleID)
	return requireDeletedRow(res, err)
}

// GetRoleAssignments lists the members who hold a role, with all their roles, the template grants of the role,
// and the group mappings of the role with their assignments.
func (d *SqlDb) GetRoleAssignments(roleID int) (db.RoleAssignments, error) {
	return d.roleAssignments(d.Sql(), roleID)
}

// roleAssignments takes the executor so that a transaction reads with its own connection, SQLite has only one.
func (d *SqlDb) roleAssignments(q gorp.SqlExecutor, roleID int) (assigned db.RoleAssignments, err error) {
	// A mapping of this role goes with it, so only mappings of other roles keep the membership.
	query, args, err := squirrel.Select("pur.project_id", "pur.user_id", "u.username", "r.id as role_id", "r.slug").
		Column(squirrel.Alias(squirrel.Expr(`exists (
			select 1 from project__user_mapped_role mr
			join group_role_mapping gm on gm.id = mr.mapping_id
			where mr.project_id = pur.project_id and mr.user_id = pur.user_id and gm.role_id <> ?
		)`, roleID), "mapped")).
		From("project__user_role as pur").
		Join("`user` as u on u.id = pur.user_id").
		Join("`role` as r on r.id = pur.role_id").
		Where(`exists (
			select 1 from project__user_role m
			where m.project_id = pur.project_id and m.user_id = pur.user_id and m.role_id = ?
		)`, roleID).
		OrderBy("pur.project_id", "pur.user_id", "r.id").
		ToSql()
	if err != nil {
		return
	}
	var rows []struct {
		ProjectID int    `db:"project_id"`
		UserID    int    `db:"user_id"`
		Username  string `db:"username"`
		RoleID    int    `db:"role_id"`
		Slug      string `db:"slug"`
		Mapped    bool   `db:"mapped"`
	}
	if _, err = q.Select(&rows, d.PrepareQuery(query), args...); err != nil {
		return
	}
	// The rows come ordered by member, so a member's roles are next to each other.
	for _, row := range rows {
		last := len(assigned.Members) - 1
		if last < 0 || assigned.Members[last].ProjectID != row.ProjectID || assigned.Members[last].UserID != row.UserID {
			assigned.Members = append(assigned.Members, db.RoleMember{ProjectID: row.ProjectID, UserID: row.UserID, Username: row.Username, Mapped: row.Mapped})
			last++
		}
		assigned.Members[last].Roles = append(assigned.Members[last].Roles, db.Role{ID: row.RoleID, Slug: row.Slug})
	}
	_, err = q.Select(&assigned.Grants, d.PrepareQuery("select id, template_id, project_id from project__template_role where role_id=?"), roleID)
	if err != nil {
		return
	}
	_, err = q.Select(&assigned.Mappings, d.PrepareQuery(`
		select
		  m.*, p.name as project_name, r.slug as role_slug, r.name as role_name
		from group_role_mapping m
		join project p on p.id = m.project_id
		join role r on r.id = m.role_id
		where m.role_id = ?
		order by m.id
	`), roleID)
	if err != nil {
		return
	}
	query, args, err = MappedRolesQuery().
		Where("m.role_id = ?", roleID).
		OrderBy("r.name", "m.provider", "m.group_name").
		ToSql()
	if err != nil {
		return
	}
	_, err = q.Select(&assigned.MappedRoles, d.PrepareQuery(query), args...)
	return
}

// DeleteRoleWithAssignments removes a custom role with its member assignments and template grants and returns
// what it removed. A member whose only role it was stops being a member. Without withMappings the foreign key
// refuses a role that a group mapping uses, with db.ErrInvalidOperation, and nothing is removed.
func (d *SqlDb) DeleteRoleWithAssignments(roleID int, withMappings bool) (db.RoleAssignments, error) {
	tx, err := d.Sql().Begin()
	if err != nil {
		return db.RoleAssignments{}, err
	}
	assigned, err := d.roleAssignments(tx, roleID)
	if err != nil {
		_ = tx.Rollback()
		return db.RoleAssignments{}, err
	}
	if withMappings {
		// Their mapped assignments go with them by the cascade.
		if _, err = d.execTx(tx, "delete from group_role_mapping where role_id=?", roleID); err != nil {
			_ = tx.Rollback()
			return db.RoleAssignments{}, err
		}
	}
	if _, err = d.execTx(tx, "delete from project__user_role where role_id=?", roleID); err != nil {
		_ = tx.Rollback()
		return db.RoleAssignments{}, err
	}
	if _, err = d.execTx(tx, "delete from project__template_role where role_id=?", roleID); err != nil {
		_ = tx.Rollback()
		return db.RoleAssignments{}, err
	}
	res, err := d.execTx(tx, "delete from `role` where id=? and builtin=false", roleID)
	if err = requireDeletedRow(res, err); err != nil {
		_ = tx.Rollback()
		return db.RoleAssignments{}, err
	}
	return assigned, tx.Commit()
}
