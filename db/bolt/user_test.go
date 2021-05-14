package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
	"testing"
)

func TestGetUsers(t *testing.T) {
	store := createStore()
	err := store.Connect()

	if err != nil {
		t.Failed()
	}

	_, err = store.CreateUser(db.UserWithPwd{
		Pwd: "123456",
		User: db.User{
			Email: "denguk@example.com",
			Name: "Denis Gukov",
			Username: "fiftin",
		},
	})

	if err != nil {
		t.Failed()
	}

	found, err := store.GetUsers(db.RetrieveQueryParams{})

	if err != nil {
		t.Failed()
	}

	if len(found) != 1 {
		t.Failed()
	}

}

func TestGetUser(t *testing.T) {
	store := createStore()
	err := store.Connect()

	if err != nil {
		t.Failed()
	}

	usr, err := store.CreateUser(db.UserWithPwd{
		Pwd: "123456",
		User: db.User{
			Email: "denguk@example.com",
			Name: "Denis Gukov",
			Username: "fiftin",
		},
	})

	if err != nil {
		t.Failed()
	}

	found, err := store.GetUser(usr.ID)

	if err != nil {
		t.Failed()
	}

	if found.Username != "fiftin" {
		t.Failed()
	}

	err = store.DeleteUser(usr.ID)

	if err != nil {
		t.Failed()
	}
}
