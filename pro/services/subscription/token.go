package subscription

import (
	"errors"
	"github.com/semaphoreui/semaphore/db"
	"time"
)

type Token struct {
	Company   string    `json:"company,omitempty"`
	State     string    `json:"state"`
	Key       string    `json:"key"`
	Plan      string    `json:"plan"`
	Users     int       `json:"users"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// ParseToken Function to verify a JWT
func ParseToken(tokenString string) (res Token, err error) {
	err = errors.New("invalid token")
	return
}

func (t *Token) Validate() error {
	return nil
}

func CanAddProUser(store db.Store) (ok bool, err error) {
	return
}

func GetToken(store db.Store) (res Token, err error) {
	err = errors.New("can not get token")
	return
}

func HasActiveSubscription(store db.Store) bool {
	return false
}
