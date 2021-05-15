package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
	"testing"
)

func TestGetUsers(t *testing.T) {
	store := createStore()
	err := store.Connect()

	if err != nil {
		t.Fatal()
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
		t.Fatal()
	}

	found, err := store.GetUsers(db.RetrieveQueryParams{})

	if err != nil {
		t.Fatal()
	}

	if len(found) != 1 {
		t.Fatal()
	}

}

func TestGetUser(t *testing.T) {
	store := createStore()
	err := store.Connect()

	if err != nil {
		t.Fatal()
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
		t.Fatal()
	}

	found, err := store.GetUser(usr.ID)

	if err != nil {
		t.Fatal()
	}

	if found.Username != "fiftin" {
		t.Fatal()
	}

	err = store.DeleteUser(usr.ID)

	if err != nil {
		t.Fatal()
	}
}
