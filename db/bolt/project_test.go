package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
	"testing"
	"time"
)

func TestGetProjects(t *testing.T) {
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

	_, err = store.CreateProjectUser(db.ProjectUser{
		ProjectID: proj1.ID,
		UserID:    usr.ID,
		Role:      db.ProjectOwner,
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	found, err := store.GetProjects(usr.ID)

	if err != nil {
		t.Fatal(err.Error())
	}

	if len(found) != 1 {
		t.Fatal(err.Error())
	}

}

func TestGetProject(t *testing.T) {
	store := CreateTestStore()

	proj, err := store.CreateProject(db.Project{
		Created: time.Now(),
		Name:    "Test1",
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	found, err := store.GetProject(proj.ID)

	if err != nil {
		t.Fatal(err.Error())
	}

	if found.Name != "Test1" {
		t.Fatal(err.Error())
	}

	err = store.DeleteProject(proj.ID)

	if err != nil {
		t.Fatal(err.Error())
	}
}
