package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
	"testing"
	"time"
)

func TestGetViews(t *testing.T) {
	store := createStore()
	err := store.Connect()

	if err != nil {
		t.Fatal(err.Error())
	}

	proj1, err := store.CreateProject(db.Project{
		Created: time.Now(),
		Name: "Test1",
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = store.CreateView(db.View{
		ProjectID: proj1.ID,
		Title: "Test",
		Position: 1,
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	found, err := store.GetViews(proj1.ID)

	if err != nil {
		t.Fatal(err.Error())
	}

	if len(found) != 1 {
		t.Fatal()
	}

	view, err := store.GetView(proj1.ID, found[0].ID)

	if err != nil {
		t.Fatal(err.Error())
	}

	if view.ID != found[0].ID || view.Title != found[0].Title || view.Position != found[0].Position {
		t.Fatal()
	}
}