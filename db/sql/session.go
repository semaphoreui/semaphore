package sql

import (
	"database/sql"
	"github.com/ansible-semaphore/semaphore/db"
	"time"
)

func (d *SqlDb) CreateAPIToken(token db.APIToken) (db.APIToken, error) {
	token.Created = db.GetParsedTime(time.Now())
	err := d.sql.Insert(&token)
	return token, err
}

func (d *SqlDb) GetAPIToken(tokenID string) (token db.APIToken, err error) {
	err = d.sql.SelectOne(&token, "select * from user__token where id=? and expired=0", tokenID)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	return
}

func (d *SqlDb) ExpireAPIToken(userID int, tokenID string) (err error) {
	res, err := d.sql.Exec("update user__token set expired=1 where id=? and user_id=?", tokenID, userID)

	return validateMutationResult(res, err)
}

func (d *SqlDb) GetSession(userID int, sessionID int) (session db.Session, err error) {
	err = d.sql.SelectOne(&session, "select * from session where id=? and user_id=? and expired=0", sessionID, userID)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	return
}

func (d *SqlDb) ExpireSession(userID int, sessionID int) error {
	res, err := d.sql.Exec("update session set expired=1 where id=? and user_id=?", sessionID, userID)

	return validateMutationResult(res, err)
}

func (d *SqlDb) TouchSession(userID int, sessionID int) error {
	_, err := d.sql.Exec("update session set last_active=? where id=? and user_id=?", time.Now(), sessionID, userID)

	return err
}

func (d *SqlDb) GetAPITokens(userID int) (tokens []db.APIToken, err error) {
	_, err = d.sql.Select(&tokens, "select * from user__token where user_id=?", userID)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	return
}

