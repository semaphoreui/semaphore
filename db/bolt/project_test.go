package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
	"testing"
	"time"
)

func TestGetProjects(t *testing.T) {
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

	proj1, err := store.CreateProject(db.Project{
		Created: time.Now(),
		Name: "Test1",
	})

	if err != nil {
		t.Failed()
	}

	_, err = store.CreateProjectUser(db.ProjectUser{
		ProjectID: proj1.ID,
		UserID: usr.ID,
		Admin: true,
	})

	if err != nil {
		t.Failed()
	}

	found, err := store.GetProjects(usr.ID)

	if err != nil {
		t.Failed()
	}

	if len(found) != 1 {
		t.Failed()
	}

}

func TestGetProject(t *testing.T) {
	store := createStore()
	err := store.Connect()

	if err != nil {
		t.Failed()
	}

	proj, err := store.CreateProject(db.Project{
		Created: time.Now(),
		Name: "Test1",
	})


	if err != nil {
		t.Failed()
	}

	found, err := store.GetProject(proj.ID)

	if err != nil {
		t.Failed()
	}

	if found.Name != "Test1" {
		t.Failed()
	}

}
