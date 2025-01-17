package bolt

import (
	"sort"
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
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
func TestGetView(t *testing.T) {
	store := CreateTestStore()

	proj1, err := store.CreateProject(db.Project{
		Created: time.Now(),
		Name:    "Test1",
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	view, err := store.CreateView(db.View{
		ProjectID: proj1.ID,
		Title:     "Test",
		Position:  1,
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	found, err := store.GetView(proj1.ID, view.ID)

	if err != nil {
		t.Fatal(err.Error())
	}

	if found.ID != view.ID || found.Title != view.Title || found.Position != view.Position {
		t.Fatal()
	}
}

func TestUpdateView(t *testing.T) {
	store := CreateTestStore()

	proj1, err := store.CreateProject(db.Project{
		Created: time.Now(),
		Name:    "Test1",
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	view, err := store.CreateView(db.View{
		ProjectID: proj1.ID,
		Title:     "Test",
		Position:  1,
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	view.Title = "Updated Test"
	err = store.UpdateView(view)

	if err != nil {
		t.Fatal(err.Error())
	}

	updatedView, err := store.GetView(proj1.ID, view.ID)

	if err != nil {
		t.Fatal(err.Error())
	}

	if updatedView.Title != "Updated Test" {
		t.Fatal()
	}
}

func TestCreateView(t *testing.T) {
	store := CreateTestStore()

	proj1, err := store.CreateProject(db.Project{
		Created: time.Now(),
		Name:    "Test1",
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	view, err := store.CreateView(db.View{
		ProjectID: proj1.ID,
		Title:     "Test",
		Position:  1,
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	found, err := store.GetView(proj1.ID, view.ID)

	if err != nil {
		t.Fatal(err.Error())
	}

	if found.ID != view.ID || found.Title != view.Title || found.Position != view.Position {
		t.Fatal()
	}
}

func TestDeleteView(t *testing.T) {
	store := CreateTestStore()

	proj1, err := store.CreateProject(db.Project{
		Created: time.Now(),
		Name:    "Test1",
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	view, err := store.CreateView(db.View{
		ProjectID: proj1.ID,
		Title:     "Test",
		Position:  1,
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	err = store.DeleteView(proj1.ID, view.ID)

	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = store.GetView(proj1.ID, view.ID)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}
