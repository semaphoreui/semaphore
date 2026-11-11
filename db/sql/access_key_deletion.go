package sql

import (
	"errors"
	"fmt"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/pkg/tz"
	log "github.com/sirupsen/logrus"
)

// ClaimAccessKeyDeletion and the link writes in environment_secret.go are the two halves of one protocol.
// The deletion commits its claim and then reads the links; a link is committed and then the claim is read.
// Whatever the interleaving, of a deletion and a link racing each other at least one sees the other and
// steps back, so the secret is never removed from its storage under a link that stays. It needs no lock
// and no transaction around the call to the storage, and it holds across HA nodes: the claim is a column.
func (d *SqlDb) ClaimAccessKeyDeletion(projectID int, accessKeyID int, exceptSecretID int) (claim int64, err error) {
	now := tz.Now()
	claim = now.Add(db.AccessKeyDeleteClaimTTL).UnixMilli()

	// One statement decides who deletes: a live claim of somebody else matches no row.
	res, err := d.exec(
		"update access_key set delete_claim=? "+
			"where project_id=? and id=? and (delete_claim is null or delete_claim <= ?)",
		claim, projectID, accessKeyID, now.UnixMilli())
	if err = validateMutationResult(res, err); err != nil {
		return 0, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	if affected == 0 {
		if _, err = d.GetAccessKey(projectID, accessKeyID); err != nil {
			return 0, err
		}
		return 0, fmt.Errorf("%w: the key is being deleted", db.ErrInvalidOperation)
	}

	links, err := d.Sql().SelectInt(
		d.PrepareQuery("select count(*) from project__environment_secret where project_id=? and key_id=? and id<>?"),
		projectID, accessKeyID, exceptSecretID)
	if err == nil && links > 0 {
		err = fmt.Errorf("%w: the key is used by a variable group", db.ErrInvalidOperation)
	}

	if err != nil {
		if releaseErr := d.ReleaseAccessKeyDeletion(projectID, accessKeyID, claim); releaseErr != nil {
			log.WithError(releaseErr).WithFields(log.Fields{
				"project_id": projectID, "key_id": accessKeyID,
			}).Error("failed to release the deletion claim of a key; it expires on its own")
		}
		return 0, err
	}

	return claim, nil
}

func (d *SqlDb) ReleaseAccessKeyDeletion(projectID int, accessKeyID int, claim int64) error {
	_, err := d.exec(
		"update access_key set delete_claim=null where project_id=? and id=? and delete_claim=?",
		projectID, accessKeyID, claim)
	return err
}

func (d *SqlDb) DeleteAccessKeyWithSecret(projectID int, accessKeyID int, environmentID int, secretID int) error {
	tx, err := d.Sql().Begin()
	if err != nil {
		return err
	}

	// The link first: the key is still referenced by it.
	err = requireDeletedRow(d.execTx(tx,
		"delete from project__environment_secret where project_id=? and environment_id=? and id=? and key_id=?",
		projectID, environmentID, secretID, accessKeyID))
	if err == nil {
		err = requireDeletedRow(d.execTx(tx,
			"delete from access_key where project_id=? and id=?", projectID, accessKeyID))
	}

	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// refuseKeyBeingDeleted is called after a link to the key has been committed: see ClaimAccessKeyDeletion.
// A key of another project is refused here as well.
func (d *SqlDb) refuseKeyBeingDeleted(projectID int, accessKeyID int) error {
	key, err := d.GetAccessKey(projectID, accessKeyID)
	if errors.Is(err, db.ErrNotFound) {
		return common_errors.NewValidationError(fmt.Sprintf("key %d does not exist in this project", accessKeyID))
	}
	if err != nil {
		return err
	}

	if key.DeletionClaimed(tz.Now()) {
		return fmt.Errorf("%w: key %s is being deleted", db.ErrInvalidOperation, key.Name)
	}

	return nil
}
