package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
	"reflect"
	"time"
)

type globalToken struct {
	ID     string `db:"id" json:"id"`
	UserID int    `db:"user_id" json:"user_id"`
}

var globalTokenObject = db.ObjectProps{
	TableName:         "token",
	PrimaryColumnName: "id",
	Type:              reflect.TypeOf(globalToken{}),
	IsGlobal:          true,
}

func (d *BoltDb) CreateSession(session db.Session) (db.Session, error) {
	newSession, err := d.createObject(session.UserID, db.SessionProps, session)
	if err != nil {
		return db.Session{}, err
	}
	return newSession.(db.Session), nil
}

func (d *BoltDb) CreateAPIToken(token db.APIToken) (db.APIToken, error) {
	// create token in bucket "token_<user id>"
	newToken, err := d.createObject(token.UserID, db.TokenProps, token)
	if err != nil {
		return db.APIToken{}, err
	}

	// create token in bucket "token"
	_, err = d.createObject(0, globalTokenObject, globalToken{ID: token.ID, UserID: token.UserID})
	if err != nil {
		return db.APIToken{}, err
	}

	return newToken.(db.APIToken), nil
}

func (d *BoltDb) GetAPIToken(tokenID string) (token db.APIToken, err error) {
	var t globalToken
	err = d.getObject(0, globalTokenObject, strObjectID(tokenID), &t)
	if err != nil {
		return
	}
	err = d.getObject(t.UserID, db.TokenProps, strObjectID(tokenID), &token)
	return
}

func (d *BoltDb) ExpireAPIToken(userID int, tokenID string) (err error) {
	var token db.APIToken
	err = d.getObject(userID, db.TokenProps, strObjectID(tokenID), &token)
	if err != nil {
		return
	}
	token.Expired = true
	err = d.updateObject(userID, db.TokenProps, token)
	return
}

func (d *BoltDb) DeleteAPIToken(userID int, tokenID string) (err error) {
	err = d.ExpireAPIToken(userID, tokenID)
	if err != nil {
		return
	}

	err = d.deleteObject(userID, db.TokenProps, strObjectID(tokenID), nil)
	return
}

func (d *BoltDb) GetSession(userID int, sessionID int) (session db.Session, err error) {
	err = d.getObject(userID, db.SessionProps, intObjectID(sessionID), &session)
	return
}

func (d *BoltDb) ExpireSession(userID int, sessionID int) (err error) {
	var session db.Session
	err = d.getObject(userID, db.SessionProps, intObjectID(sessionID), &session)
	if err != nil {
		return
	}
	session.Expired = true
	err = d.updateObject(userID, db.SessionProps, session)
	return
}

func (d *BoltDb) TouchSession(userID int, sessionID int) (err error) {
	var session db.Session
	err = d.getObject(userID, db.SessionProps, intObjectID(sessionID), &session)
	if err != nil {
		return
	}
	session.LastActive = time.Now()
	err = d.updateObject(userID, db.SessionProps, session)
	return
}

func (d *BoltDb) GetAPITokens(userID int) (tokens []db.APIToken, err error) {
	err = d.getObjects(userID, db.TokenProps, db.RetrieveQueryParams{}, nil, &tokens)
	return
}
