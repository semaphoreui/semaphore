package models

import (
	"encoding/json"
)

type SessionLogin struct {
	UserID             string  `json:"user_id"`
	AccessLevelID      string  `json:"access_level_id"`
	UserName           string  `json:"user_name"`
	UserEmail          string  `json:"user_email"`
	Check              int     `json:"check"`
	MealID             *string `json:"meal_id"`
	GroupLoginUsername *string `json:"group_login_username"`
	GroupLogin         *bool   `json:"group_login"`
	FakeLogin          *int    `json:"fake_login"`
	FakeLoginUserID    *int    `json:"fake_login_user_id"`
	FakeLoginRealName  *string `json:"fake_login_real_name"`
}

type Session struct {
	ID string `json:"-"`

	ClientThemeMainBGColourBefore     *string `json:"client_theme_main_bg_colour_before,omitempty"`
	ClientThemeMainBGColour           *string `json:"client_theme_main_bg_colour,omitempty"`
	ClientThemeMainBGColourDarker     *string `json:"client_theme_main_bg_colour_darker,omitempty"`
	ClientThemeMainBGColourEvenDarker *string `json:"client_theme_main_bg_colour_even_darker,omitempty"`

	Login *SessionLogin `json:"login"`
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
