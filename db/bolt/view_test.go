package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
	"sort"
	"testing"
	"time"
)

func TestGetViews(t *testing.T) {
	store := CreateTestStore()

	proj1, err := store.CreateProject(db.Project{
		Created: time.Now(),
		Name:    "Test1",
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = store.CreateView(db.View{
		ProjectID: proj1.ID,
		Title:     "Test",
		Position:  1,
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

func TestSetViewPositions(t *testing.T) {
	store := CreateTestStore()

	proj1, err := store.CreateProject(db.Project{
		Created: time.Now(),
		Name:    "Test1",
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	v1, err := store.CreateView(db.View{
		ProjectID: proj1.ID,
		Title:     "Test",
		Position:  4,
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	v2, err := store.CreateView(db.View{
		ProjectID: proj1.ID,
		Title:     "Test",
		Position:  2,
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	found, err := store.GetViews(proj1.ID)

	if err != nil {
		t.Fatal(err.Error())
	}

	if len(found) != 2 {
		t.Fatal()
	}

	sort.Slice(found, func(i, j int) bool {
		return found[i].Position < found[j].Position
	})

	if found[0].Position != v2.Position || found[1].Position != v1.Position {
		t.Fatal()
	}

	err = store.SetViewPositions(proj1.ID, map[int]int{
		v1.ID: 3,
		v2.ID: 6,
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	found, err = store.GetViews(proj1.ID)

	if err != nil {
		t.Fatal(err.Error())
	}

	if len(found) != 2 {
		t.Fatal()
	}

	sort.Slice(found, func(i, j int) bool {
		return found[i].Position < found[j].Position
	})

	if found[0].Position != 3 || found[1].Position != 6 {
		t.Fatal()
	}
}
