package models

import (
	"encoding/json"
)

type Session struct {
	ID string `json:"-"`

	UserID *int `json:"user_id"`
}

func (session *Session) Encode() []byte {
	js, err := json.Marshal(session)
	if err != nil {
		panic(err)
	}

	return js
}

func DecodeSession(ID string, sess string) (Session, error) {
	var session Session
	err := json.Unmarshal([]byte(sess), &session)

	session.ID = ID

	return session, err
}
