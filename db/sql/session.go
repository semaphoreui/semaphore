package sql

import (
	"database/sql"
	"github.com/ansible-semaphore/semaphore/db"
	"time"
)

func (d *SqlDb) CreateSession(session db.Session) (db.Session, error) {
	err := d.sql.Insert(&session)
	return session, err
}

func (d *SqlDb) CreateAPIToken(token db.APIToken) (db.APIToken, error) {
	token.Created = db.GetParsedTime(time.Now())
	err := d.sql.Insert(&token)
	return token, err
}

func (d *SqlDb) GetAPIToken(tokenID string) (token db.APIToken, err error) {
	err = d.selectOne(&token, d.PrepareQuery("select * from user__token where id=? and expired=false"), tokenID)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	return
}

func (d *SqlDb) ExpireAPIToken(userID int, tokenID string) (err error) {
	res, err := d.exec("update user__token set expired=true where id=? and user_id=?", tokenID, userID)

	return validateMutationResult(res, err)
}

func (d *SqlDb) GetSession(userID int, sessionID int) (session db.Session, err error) {
	err = d.selectOne(&session, "select * from session where id=? and user_id=? and expired=false", sessionID, userID)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	return
}

func (d *SqlDb) ExpireSession(userID int, sessionID int) error {
	res, err := d.exec("update session set expired=1 where id=? and user_id=?", sessionID, userID)

	return validateMutationResult(res, err)
}

func (d *SqlDb) TouchSession(userID int, sessionID int) error {
	_, err := d.exec("update session set last_active=? where id=? and user_id=?", time.Now(), sessionID, userID)

	return err
}

func (d *SqlDb) GetAPITokens(userID int) (tokens []db.APIToken, err error) {
	_, err = d.selectAll(&tokens, d.PrepareQuery("select * from user__token where user_id=?"), userID)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	return
}
