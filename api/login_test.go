package api

import (
	"bytes"
	"encoding/json"
	"github.com/ansible-semaphore/semaphore/util"
	//_ "github.com/snikch/goodman/hooks"
	//_ "github.com/snikch/goodman/transaction"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthRegister(t *testing.T) {
	util.Config = &util.ConfigType{}

	body, err := json.Marshal(map[string]string{
		"name":  "Toby",
		"email": "Toby@example.com",
	})

	if err != nil {
		t.Fail()
	}

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	r := Route()

	r.ServeHTTP(rr, req)

	if rr.Code != 400 {
		t.Errorf("Response code should be 400 but got %d", rr.Code)
	}
}
