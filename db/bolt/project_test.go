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

	proj1, err := store.CreateProject(db.Project{
		Created: time.Now(),
		Name: "Test1",
	})

	if err != nil {
		t.Fatal()
	}

	_, err = store.CreateProjectUser(db.ProjectUser{
		ProjectID: proj1.ID,
		UserID: usr.ID,
		Admin: true,
	})

	if err != nil {
		t.Fatal()
	}

	found, err := store.GetProjects(usr.ID)

	if err != nil {
		t.Fatal()
	}

	if len(found) != 1 {
		t.Fatal()
	}

}

func TestGetProject(t *testing.T) {
	store := createStore()
	err := store.Connect()

	if err != nil {
		t.Fatal()
	}

	proj, err := store.CreateProject(db.Project{
		Created: time.Now(),
		Name: "Test1",
	})


	if err != nil {
		t.Fatal()
	}

	found, err := store.GetProject(proj.ID)

	if err != nil {
		t.Fatal()
	}

	if found.Name != "Test1" {
		t.Fatal()
	}

	err = store.DeleteProject(proj.ID)

	if err != nil {
		t.Fatal()
	}
}
