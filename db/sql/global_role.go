package sql

import (
	"fmt"

	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/db"
)

// GlobalRoleAssignmentSelect reads global role assignments together with the global role they grant. The Pro
// store that manages global roles and this store's user deletion share it.
const GlobalRoleAssignmentSelect = "select a.id, a.user_id, a.role_id, u.username, r.slug as role_slug, r.name as role_name, r.permissions as global_permissions " +
	"from user__global_role as a join `user` as u on u.id = a.user_id join global_role as r on r.id = a.role_id"

// LockGlobalRoleState takes the row lock that serializes every transaction able to remove an administrator:
// changing global roles, their assignments or the admin flag. Without it two HA nodes could each see the
// other administrator and both remove one.
func (d *SqlDb) LockGlobalRoleState(tx *gorp.Transaction) error {
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

// CountAdministrators counts the users who can still manage global roles after the exclusion: admins and
// holders of a global role with CanManageGlobalRoles.
func (d *SqlDb) CountAdministrators(tx gorp.SqlExecutor, exclude db.AdministratorExclusion) (int64, error) {
	return tx.SelectInt(d.PrepareQuery(`
		select count(distinct u.id)
		from `+"`user`"+` as u
		where u.id <> ?
		  and ((u.admin = true and u.id <> ?) or exists (
		    select 1
		    from user__global_role as a
		    join global_role as r on r.id = a.role_id
		    where a.user_id = u.id and a.id <> ? and a.role_id <> ? and (r.permissions & ?) = ?
		  ))
	`), exclude.UserID, exclude.AdminOfUser, exclude.AssignmentID, exclude.RoleID,
		db.CanManageGlobalRoles, db.CanManageGlobalRoles)
}

// RequireAdministrator refuses the exclusion when it would remove the last user able to manage global roles.
// An instance that has no such user yet (first start, tests) is not protected: there is nothing to lose.
func (d *SqlDb) RequireAdministrator(tx gorp.SqlExecutor, exclude db.AdministratorExclusion) error {
	after, err := d.CountAdministrators(tx, exclude)
	if err != nil || after > 0 {
		return err
	}
	before, err := d.CountAdministrators(tx, db.AdministratorExclusion{})
	if err != nil || before == 0 {
		return err
	}
	return db.ErrLastGlobalAdministrator
}
