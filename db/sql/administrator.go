package sql

import (
	"fmt"

	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/db"
)

// LockAdministratorState takes the row lock that serializes every transaction able to remove an
// administrator: clearing the admin flag, deleting an admin and, in the Pro edition, changing global roles
// or their assignments. Without it two HA nodes could each see the other administrator and both remove one.
func (d *SqlDb) LockAdministratorState(tx *gorp.Transaction) error {
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

// requireAnotherAdmin refuses to remove the admin flag or the account of the last admin. An instance that
// has no admin yet (first start, tests) is not protected: there is nothing to lose.
func (d *SqlDb) requireAnotherAdmin(tx gorp.SqlExecutor, exceptUserID int) error {
	others, err := tx.SelectInt(d.PrepareQuery("select count(*) from `user` where admin = true and id <> ?"), exceptUserID)
	if err != nil || others > 0 {
		return err
	}
	all, err := tx.SelectInt(d.PrepareQuery("select count(*) from `user` where admin = true"))
	if err != nil || all == 0 {
		return err
	}
	return db.ErrLastGlobalAdministrator
}
