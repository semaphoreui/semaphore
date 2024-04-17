package api

import (
	"github.com/ansible-semaphore/semaphore/db"
	"net/http"
	"testing"
)

func TestIntegrationMatch(t *testing.T) {
	body := []byte("{\"hook_id\": 4856239453}")
	var header = make(http.Header)
	matched := Match(db.IntegrationMatcher{
		ID:            0,
		Name:          "Test",
		IntegrationID: 0,
		MatchType:     db.IntegrationMatchBody,
		Method:        db.IntegrationMatchMethodEquals,
		BodyDataType:  db.IntegrationBodyDataJSON,
		Key:           "hook_id",
		Value:         "4856239453",
	}, header, body)

	if !matched {
		t.Fatal()
	}
}
