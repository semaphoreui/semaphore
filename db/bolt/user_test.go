package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
	"testing"
	"time"
)

func TestBoltDb_UpdateProjectUser(t *testing.T) {
	store := CreateTestStore()

	usr, err := store.CreateUser(db.UserWithPwd{
		Pwd: "123456",
		User: db.User{
			Email:    "denguk@example.com",
			Name:     "Denis Gukov",
			Username: "fiftin",
		},
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	proj1, err := store.CreateProject(db.Project{
		Created: time.Now(),
		Name:    "Test1",
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	projUser, err := store.CreateProjectUser(db.ProjectUser{
		ProjectID: proj1.ID,
		UserID:    usr.ID,
		Role:      db.ProjectOwner,
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	projUser.Role = db.ProjectOwner
	err = store.UpdateProjectUser(projUser)

	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestGetUsers(t *testing.T) {
	store := CreateTestStore()

	_, err := store.CreateUser(db.UserWithPwd{
		Pwd: "123456",
		User: db.User{
			Email:    "denguk@example.com",
			Name:     "Denis Gukov",
			Username: "fiftin",
		},
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	found, err := store.GetUsers(db.RetrieveQueryParams{})

	if err != nil {
		t.Fatal(err.Error())
	}

	if len(found) != 1 {
		t.Fatal(err.Error())
	}

}

func TestGetUser(t *testing.T) {
	store := CreateTestStore()

	usr, err := store.CreateUser(db.UserWithPwd{
		Pwd: "123456",
		User: db.User{
			Email:    "denguk@example.com",
			Name:     "Denis Gukov",
			Username: "fiftin",
		},
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	found, err := store.GetUser(usr.ID)

	if err != nil {
		t.Fatal(err.Error())
	}

	if found.Username != "fiftin" {
		t.Fatal(err.Error())
	}

	err = store.DeleteUser(usr.ID)

	if err != nil {
		t.Fatal(err.Error())
	}
}
