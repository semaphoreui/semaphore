package api

import (
	"bytes"
	"encoding/json"
	"github.com/ansible-semaphore/semaphore/db/bolt"
	"github.com/ansible-semaphore/semaphore/util"
	"math/rand"
	"strconv"
	"time"

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

func createBoltDb() bolt.BoltDb {
	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	fn := "/tmp/test_semaphore_db_" + strconv.Itoa(r.Int())
	return bolt.BoltDb{
		Filename: fn,
	}
}
