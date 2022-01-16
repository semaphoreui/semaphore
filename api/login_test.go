package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/db/bolt"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
	"github.com/gorilla/securecookie"
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

func createStore() db.Store {
	store := createBoltDb()
	err := store.Connect()
	if err != nil {
		panic(err)
	}
	return &store
}

func initCookies() {
	var encryption []byte
	hash, _ := base64.StdEncoding.DecodeString("ikt5uEXMZa7qinEqRa7tO1y9QpBAMG8goTxsJ57h6O8=")
	encryption, _ = base64.StdEncoding.DecodeString("bEjxOq4fhKdiYO50mEF99aR1LJPnevvViVvXfhZV5QY=")
	util.Cookie = securecookie.New(hash, encryption)
}

func TestAuthRegister2(t *testing.T) {
	initCookies()

	util.Config = &util.ConfigType{
		RegisterFirstUser: true,
	}

	body, err := json.Marshal(map[string]string{
		"name":     "Toby",
		"email":    "Toby@example.com",
		"username": "toby",
		"password": "Test123",
	})

	if err != nil {
		t.Fail()
	}

	store := createStore()

	err = store.CreatePlaceholderUser()

	if err != nil {
		t.Fail()
	}

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	r := Route()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			context.Set(r, "store", store)
			next.ServeHTTP(w, r)
		})
	})

	r.ServeHTTP(rr, req)

	if rr.Code != 204 {
		t.Errorf("Response code should be 204 but got %d", rr.Code)
	}
}
