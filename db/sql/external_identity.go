package sql

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
)

func (d *SqlDb) CreateExternalIdentity(identity db.UserExternalIdentity) (db.UserExternalIdentity, error) {
	identity.Created = db.GetParsedTime(tz.Now())
	if identity.Type == "" {
		identity.Type = db.IdentityTypeOidc
	}
	err := d.Sql().Insert(&identity)
	return identity, err
}

func (d *SqlDb) GetExternalIdentity(idType string, provider string, externalUID string) (identity db.UserExternalIdentity, err error) {
	err = d.selectOne(
		&identity,
		"select * from user__external_identity where type=? and provider=? and external_uid=?",
		idType, provider, externalUID)
	return
}

func (d *SqlDb) GetUserExternalIdentities(userID int) (identities []db.UserExternalIdentity, err error) {
	_, err = d.selectAll(
		&identities,
		d.PrepareQuery("select * from user__external_identity where user_id=? order by created desc"),
		userID)
	return
}

func (d *SqlDb) DeleteExternalIdentity(userID int, idType string, provider string) error {
	_, err := d.exec(
		"delete from user__external_identity where user_id=? and type=? and provider=?",
		userID, idType, provider)
	return err
}
