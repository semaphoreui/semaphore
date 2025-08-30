package bolt

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"testing"
)

func TestViewTypesAndHidden(t *testing.T) {
	store := CreateTestStore()

	proj1, err := store.CreateProject(db.Project{
		Created: tz.Now(),
		Name:    "Test1",
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	// Test creating views with different types
	customView, err := store.CreateView(db.View{
		ProjectID: proj1.ID,
		Title:     "Custom View",
		Position:  1,
		Hidden:    false,
		Type:      db.ViewTypeCustom,
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	allView, err := store.CreateView(db.View{
		ProjectID: proj1.ID,
		Title:     "All",
		Position:  -1,
		Hidden:    false,
		Type:      db.ViewTypeAll,
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	failedView, err := store.CreateView(db.View{
		ProjectID: proj1.ID,
		Title:     "Failed Tasks",
		Position:  2,
		Hidden:    true,
		Type:      db.ViewTypeFailed,
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	// Test view type helper methods
	if !customView.IsCustomView() {
		t.Fatal("Custom view should return true for IsCustomView()")
	}

	if !allView.IsAllView() {
		t.Fatal("All view should return true for IsAllView()")
	}

	if !failedView.IsFailedView() {
		t.Fatal("Failed view should return true for IsFailedView()")
	}

	// Test retrieving views
	views, err := store.GetViews(proj1.ID)
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(views) != 3 {
		t.Fatalf("Expected 3 views, got %d", len(views))
	}

	// Test that hidden and type fields are properly stored and retrieved
	var foundCustom, foundAll, foundFailed bool
	for _, view := range views {
		switch view.Type {
		case db.ViewTypeCustom:
			foundCustom = true
			if view.Hidden || view.Title != "Custom View" {
				t.Fatal("Custom view properties not preserved")
			}
		case db.ViewTypeAll:
			foundAll = true
			if view.Hidden || view.Title != "All" || view.Position != -1 {
				t.Fatal("All view properties not preserved")
			}
		case db.ViewTypeFailed:
			foundFailed = true
			if !view.Hidden || view.Title != "Failed Tasks" {
				t.Fatal("Failed view properties not preserved")
			}
		}
	}

	if !foundCustom || !foundAll || !foundFailed {
		t.Fatal("Not all view types were found")
	}
}

func TestViewValidation(t *testing.T) {
	// Test valid view types
	validView := db.View{
		Title: "Test",
		Type:  db.ViewTypeCustom,
	}
	if err := validView.Validate(); err != nil {
		t.Fatal("Valid view should pass validation")
	}

	// Test invalid view type
	invalidView := db.View{
		Title: "Test",
		Type:  "invalid_type",
	}
	if err := invalidView.Validate(); err == nil {
		t.Fatal("Invalid view type should fail validation")
	}

	// Test empty title
	emptyTitleView := db.View{
		Title: "",
		Type:  db.ViewTypeCustom,
	}
	if err := emptyTitleView.Validate(); err == nil {
		t.Fatal("Empty title should fail validation")
	}
}

func TestShouldAllTabBeAtEnd(t *testing.T) {
	// Test case 1: All view at beginning (position -1)
	views1 := []db.View{
		{Title: "All", Position: -1, Type: db.ViewTypeAll, Hidden: false},
		{Title: "Custom 1", Position: 1, Type: db.ViewTypeCustom, Hidden: false},
		{Title: "Custom 2", Position: 2, Type: db.ViewTypeCustom, Hidden: false},
	}
	if db.ShouldAllTabBeAtEnd(views1) {
		t.Fatal("All tab should be at beginning when position is -1")
	}

	// Test case 2: All view at end (position greater than others)
	views2 := []db.View{
		{Title: "Custom 1", Position: 1, Type: db.ViewTypeCustom, Hidden: false},
		{Title: "Custom 2", Position: 2, Type: db.ViewTypeCustom, Hidden: false},
		{Title: "All", Position: 3, Type: db.ViewTypeAll, Hidden: false},
	}
	if !db.ShouldAllTabBeAtEnd(views2) {
		t.Fatal("All tab should be at end when position is greater than others")
	}

	// Test case 3: Hidden All view should return false
	views3 := []db.View{
		{Title: "Custom 1", Position: 1, Type: db.ViewTypeCustom, Hidden: false},
		{Title: "All", Position: 3, Type: db.ViewTypeAll, Hidden: true},
	}
	if db.ShouldAllTabBeAtEnd(views3) {
		t.Fatal("Hidden All view should result in false")
	}

	// Test case 4: No All view should return false
	views4 := []db.View{
		{Title: "Custom 1", Position: 1, Type: db.ViewTypeCustom, Hidden: false},
		{Title: "Custom 2", Position: 2, Type: db.ViewTypeCustom, Hidden: false},
	}
	if db.ShouldAllTabBeAtEnd(views4) {
		t.Fatal("No All view should result in false")
	}
}

func TestUpdateViewWithNewFields(t *testing.T) {
	store := CreateTestStore()

	proj1, err := store.CreateProject(db.Project{
		Created: tz.Now(),
		Name:    "Test1",
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	// Create a view
	view, err := store.CreateView(db.View{
		ProjectID: proj1.ID,
		Title:     "Test View",
		Position:  1,
		Hidden:    false,
		Type:      db.ViewTypeCustom,
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	// Update the view with new field values
	view.Hidden = true
	view.Type = db.ViewTypeFailed
	view.Title = "Updated Failed View"

	err = store.UpdateView(view)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Retrieve and verify the updated view
	updatedView, err := store.GetView(proj1.ID, view.ID)
	if err != nil {
		t.Fatal(err.Error())
	}

	if updatedView.Hidden != true {
		t.Fatal("Hidden field should be updated to true")
	}

	if updatedView.Type != db.ViewTypeFailed {
		t.Fatal("Type field should be updated to failed")
	}

	if updatedView.Title != "Updated Failed View" {
		t.Fatal("Title field should be updated")
	}
}

func TestViewConstants(t *testing.T) {
	// Test that view type constants are correctly defined
	if db.ViewTypeCustom != "custom" {
		t.Fatal("ViewTypeCustom should be 'custom'")
	}

	if db.ViewTypeAll != "all" {
		t.Fatal("ViewTypeAll should be 'all'")
	}

	if db.ViewTypeFailed != "failed" {
		t.Fatal("ViewTypeFailed should be 'failed'")
	}
}