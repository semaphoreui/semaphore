package sql

import (
	"database/sql"
	"errors"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"regexp"
)

func (d *SqlDb) SetSessionVerificationMethod(userID int, sessionID int, verificationMethod db.SessionVerificationMethod) error {
	return nil
}

func (d *SqlDb) VerifySession(userID int, sessionID int) error {
	_, err := d.exec("update session set verified = true where id=? and user_id=?", sessionID, userID)

	return err
}

func (d *SqlDb) CreateSession(session db.Session) (db.Session, error) {
	err := d.sql.Insert(&session)
	return session, err
}

func (d *SqlDb) CreateAPIToken(token db.APIToken) (db.APIToken, error) {
	token.Created = db.GetParsedTime(tz.Now())
	err := d.sql.Insert(&token)
	return token, err
}

func (d *SqlDb) GetAPIToken(tokenID string) (token db.APIToken, err error) {
	err = d.selectOne(&token, d.PrepareQuery("select * from user__token where id=? and expired=false"), tokenID)

	return
}

func (d *SqlDb) ExpireAPIToken(userID int, tokenID string) error {
	return validateMutationResult(d.exec("update user__token set expired=true where id=? and user_id=?", tokenID, userID))
}

func validateAPIToken(token string) error {
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9-_=]{8,}$`, token); !matched {
		return errors.New("invalid token format")
	}
	return nil
}

func (d *SqlDb) DeleteAPIToken(userID int, tokenPrefix string) (err error) {

	err = validateAPIToken(tokenPrefix)
	if err != nil {
		return
	}

	_, err = d.exec("DELETE FROM user__token WHERE id LIKE ? AND user_id=?", tokenPrefix+"%", userID)

	return
}

func (d *SqlDb) GetSession(userID int, sessionID int) (session db.Session, err error) {
	err = d.selectOne(&session, "select * from session where id=? and user_id=? and expired=false", sessionID, userID)

	return
}

func (d *SqlDb) ExpireSession(userID int, sessionID int) error {
	res, err := d.exec("update session set expired=true where id=? and user_id=?", sessionID, userID)

	return validateMutationResult(res, err)
}

func (d *SqlDb) TouchSession(userID int, sessionID int) error {
	_, err := d.exec("update session set last_active=? where id=? and user_id=?", tz.Now(), sessionID, userID)

	return err
}

func (d *SqlDb) GetAPITokens(userID int) (tokens []db.APIToken, err error) {
	_, err = d.selectAll(&tokens, d.PrepareQuery("select * from user__token where user_id=? order by created desc"), userID)

	if errors.Is(err, sql.ErrNoRows) {
		err = db.ErrNotFound
	}

	return
}
