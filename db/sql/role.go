package sql

import (
	"database/sql"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

func (d *SqlDb) GetGlobalRoleBySlug(slug string) (role db.Role, err error) {
	err = d.selectOne(&role, "select * from `role` where lower(slug)=lower(?) and project_id is null", slug)
	return
}

func (d *SqlDb) GetProjectRole(projectID int, slug string) (role db.Role, err error) {
	err = d.selectOne(&role, "select * from `role` where lower(slug)=lower(?) and project_id=?", slug, projectID)
	return
}

// GetProjectOrGlobalRoleBySlug prefers the project role when a global role has the same slug.
func (d *SqlDb) GetProjectOrGlobalRoleBySlug(projectID int, slug string) (role db.Role, err error) {
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

func (d *SqlDb) GetGlobalRoles() (roles []db.Role, err error) {
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
		insert into role (slug, name, permissions, global_permissions, project_id, scope, builtin)
		values (?, ?, ?, ?, ?, ?, false)
	`, role.Slug, role.Name, role.Permissions, role.GlobalPermissions, role.ProjectID, scope)
	role.Scope, role.Builtin = scope, false
	return role, err
}

// UpdateRole changes only the name and the permissions. Built-in roles never match. Taking CanManageGlobalRoles
// away from a role is refused when its holders are the only administrators left.
func (d *SqlDb) UpdateRole(role db.Role) error {
	tx, err := d.Sql().Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err = d.LockGlobalRoleState(tx); err != nil {
		return err
	}
	var current db.Role
	err = tx.SelectOne(&current, d.PrepareQuery("select * from `role` where id=? and builtin=false"), role.ID)
	if errors.Is(err, sql.ErrNoRows) {
		// A built-in or unknown role: the update below changes nothing, as before.
		return nil
	}
	if err != nil {
		return err
	}
	if current.GlobalPermissions.Can(db.CanManageGlobalRoles) && !role.GlobalPermissions.Can(db.CanManageGlobalRoles) {
		if err = d.RequireAdministrator(tx, db.AdministratorExclusion{RoleID: role.ID}); err != nil {
			return err
		}
	}
	_, err = d.execTx(tx, "update `role` set name=?, permissions=?, global_permissions=? where id=? and builtin=false",
		role.Name, role.Permissions, role.GlobalPermissions, role.ID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// DeleteRole removes a custom role. The foreign keys refuse a role that is still assigned or granted.
func (d *SqlDb) DeleteRole(roleID int) error {
	res, err := d.exec("delete from `role` where id=? and builtin=false", roleID)
	return requireDeletedRow(res, err)
}

// GetRoleAssignments lists the members who hold a role, with all their roles, and the template grants of the role.
func (d *SqlDb) GetRoleAssignments(roleID int) (db.RoleAssignments, error) {
	return d.roleAssignments(d.Sql(), roleID)
}

// roleAssignments takes the executor so that a transaction reads with its own connection, SQLite has only one.
func (d *SqlDb) roleAssignments(q gorp.SqlExecutor, roleID int) (assigned db.RoleAssignments, err error) {
	query, args, err := squirrel.Select("pur.project_id", "pur.user_id", "u.username", "r.id as role_id", "r.slug").
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
	}
	if _, err = q.Select(&rows, d.PrepareQuery(query), args...); err != nil {
		return
	}
	// The rows come ordered by member, so a member's roles are next to each other.
	for _, row := range rows {
		last := len(assigned.Members) - 1
		if last < 0 || assigned.Members[last].ProjectID != row.ProjectID || assigned.Members[last].UserID != row.UserID {
			assigned.Members = append(assigned.Members, db.RoleMember{ProjectID: row.ProjectID, UserID: row.UserID, Username: row.Username})
			last++
		}
		assigned.Members[last].Roles = append(assigned.Members[last].Roles, db.Role{ID: row.RoleID, Slug: row.Slug})
	}
	_, err = q.Select(&assigned.Grants, d.PrepareQuery("select id, template_id, project_id from project__template_role where role_id=?"), roleID)
	if err != nil {
		return
	}
	assigned.GlobalMembers = make([]db.GlobalRoleAssignment, 0)
	_, err = q.Select(&assigned.GlobalMembers, d.PrepareQuery(GlobalRoleAssignmentSelect+" where a.role_id=? order by a.id"), roleID)
	return
}

// DeleteRoleWithAssignments removes a custom role with its member assignments, global assignments and template
// grants and returns what it removed. A member whose only role it was stops being a member. The role is kept
// when its global holders are the only administrators left.
func (d *SqlDb) DeleteRoleWithAssignments(roleID int) (db.RoleAssignments, error) {
	tx, err := d.Sql().Begin()
	if err != nil {
		return db.RoleAssignments{}, err
	}
	if err = d.LockGlobalRoleState(tx); err != nil {
		_ = tx.Rollback()
		return db.RoleAssignments{}, err
	}
	assigned, err := d.roleAssignments(tx, roleID)
	if err != nil {
		_ = tx.Rollback()
		return db.RoleAssignments{}, err
	}
	if len(assigned.GlobalMembers) > 0 && assigned.GlobalMembers[0].GlobalPermissions.Can(db.CanManageGlobalRoles) {
		if err = d.RequireAdministrator(tx, db.AdministratorExclusion{RoleID: roleID}); err != nil {
			_ = tx.Rollback()
			return db.RoleAssignments{}, err
		}
	}
	if _, err = d.execTx(tx, "delete from user__global_role where role_id=?", roleID); err != nil {
		_ = tx.Rollback()
		return db.RoleAssignments{}, err
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
