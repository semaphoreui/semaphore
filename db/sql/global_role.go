package sql

import (
	stdsql "database/sql"
	"errors"
	"fmt"

	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

// globalAssignmentSelect reads assignments together with the role they grant.
const globalAssignmentSelect = "select a.id, a.user_id, a.role_id, r.slug as role_slug, r.name as role_name, r.global_permissions " +
	"from user__global_role as a join `role` as r on r.id = a.role_id"

// lockGlobalRoleState takes the row lock that serializes every transaction able to remove an administrator:
// changing global roles, their assignments or the admin flag. Without it two HA nodes could each see the
// other administrator and both remove one.
func (d *SqlDb) lockGlobalRoleState(tx *gorp.Transaction) error {
	res, err := d.execTx(tx, "update global_role_state set revision=revision+1 where id=1")
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return fmt.Errorf("global_role_state row is missing")
	}
	return nil
}

// administratorExclusion names what the caller is about to remove, so that it does not count.
type administratorExclusion struct {
	UserID       int // the user being deleted
	AdminOfUser  int // the user whose admin flag is being cleared
	AssignmentID int // the assignment being deleted
	RoleID       int // the role being deleted or losing CanManageGlobalRoles
}

// countAdministratorsTx counts the users who can still manage global roles after the exclusion: admins and
// holders of a global role with CanManageGlobalRoles.
func (d *SqlDb) countAdministratorsTx(tx gorp.SqlExecutor, exclude administratorExclusion) (int64, error) {
	return tx.SelectInt(d.PrepareQuery(`
		select count(distinct u.id)
		from `+"`user`"+` as u
		where u.id <> ?
		  and ((u.admin = true and u.id <> ?) or exists (
		    select 1
		    from user__global_role as a
		    join `+"`role`"+` as r on r.id = a.role_id
		    where a.user_id = u.id and a.id <> ? and a.role_id <> ? and (r.global_permissions & ?) = ?
		  ))
	`), exclude.UserID, exclude.AdminOfUser, exclude.AssignmentID, exclude.RoleID,
		db.CanManageGlobalRoles, db.CanManageGlobalRoles)
}

// requireAdministratorTx refuses the exclusion when it would remove the last user able to manage global roles.
// An instance that has no such user yet (first start, tests) is not protected: there is nothing to lose.
func (d *SqlDb) requireAdministratorTx(tx gorp.SqlExecutor, exclude administratorExclusion) error {
	after, err := d.countAdministratorsTx(tx, exclude)
	if err != nil || after > 0 {
		return err
	}
	before, err := d.countAdministratorsTx(tx, administratorExclusion{})
	if err != nil || before == 0 {
		return err
	}
	return db.ErrLastGlobalAdministrator
}

func (d *SqlDb) GetGlobalRoleAssignments(userID int) ([]db.GlobalRoleAssignment, error) {
	assignments := make([]db.GlobalRoleAssignment, 0)
	_, err := d.selectAll(&assignments, globalAssignmentSelect+" where a.user_id=? order by r.name, a.id", userID)
	return assignments, err
}

func (d *SqlDb) GetGlobalRoleAssignment(userID int, assignmentID int) (assignment db.GlobalRoleAssignment, err error) {
	err = d.selectOne(&assignment, globalAssignmentSelect+" where a.user_id=? and a.id=?", userID, assignmentID)
	return
}

// CreateGlobalRoleAssignment gives the user a global role. Only a custom global role can be assigned this way:
// a project role belongs to its project and a built-in role grants nothing outside a project.
func (d *SqlDb) CreateGlobalRoleAssignment(userID int, roleID int) (db.GlobalRoleAssignment, error) {
	var role db.Role
	err := d.selectOne(&role, "select * from `role` where id=? and project_id is null and builtin=false", roleID)
	if errors.Is(err, db.ErrNotFound) {
		return db.GlobalRoleAssignment{}, &common_errors.ValidationError{Message: "Global role not found"}
	}
	if err != nil {
		return db.GlobalRoleAssignment{}, err
	}
	if _, err = d.GetUser(userID); err != nil {
		return db.GlobalRoleAssignment{}, err
	}
	held, err := d.Sql().SelectInt(d.PrepareQuery(
		"select count(*) from user__global_role where user_id=? and role_id=?"), userID, roleID)
	if err != nil {
		return db.GlobalRoleAssignment{}, err
	}
	if held > 0 {
		return db.GlobalRoleAssignment{}, &common_errors.ValidationError{Message: "The user already holds this role"}
	}
	id, err := d.insert("id", "insert into user__global_role (user_id, role_id) values (?, ?)", userID, roleID)
	if err != nil {
		return db.GlobalRoleAssignment{}, err
	}
	return d.GetGlobalRoleAssignment(userID, id)
}

// DeleteGlobalRoleAssignment takes a global role away from the user, unless that leaves nobody who can
// manage global roles.
func (d *SqlDb) DeleteGlobalRoleAssignment(userID int, assignmentID int) error {
	tx, err := d.Sql().Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err = d.lockGlobalRoleState(tx); err != nil {
		return err
	}
	var assignment db.GlobalRoleAssignment
	err = tx.SelectOne(&assignment, d.PrepareQuery(globalAssignmentSelect+" where a.user_id=? and a.id=?"), userID, assignmentID)
	if err != nil {
		return d.wrapNotFound(err)
	}
	if assignment.GlobalPermissions.Can(db.CanManageGlobalRoles) {
		if err = d.requireAdministratorTx(tx, administratorExclusion{AssignmentID: assignmentID}); err != nil {
			return err
		}
	}
	res, err := d.execTx(tx, "delete from user__global_role where id=?", assignmentID)
	if err = requireDeletedRow(res, err); err != nil {
		return err
	}
	return tx.Commit()
}

// GetEffectiveGlobalPermissions is what the user may do instance-wide: everything for an admin, otherwise the
// union of the assigned global roles.
func (d *SqlDb) GetEffectiveGlobalPermissions(userID int) (db.GlobalPermission, error) {
	user, err := d.GetUser(userID)
	if err != nil {
		return 0, err
	}
	if user.Admin {
		return db.AllGlobalPermissions, nil
	}
	assignments, err := d.GetGlobalRoleAssignments(userID)
	if err != nil {
		return 0, err
	}
	return db.EffectiveGlobalPermissionMask(false, assignments), nil
}

// wrapNotFound turns the driver's empty result into db.ErrNotFound for reads done inside a transaction.
func (d *SqlDb) wrapNotFound(err error) error {
	if errors.Is(err, stdsql.ErrNoRows) {
		return db.ErrNotFound
	}
	return err
}
