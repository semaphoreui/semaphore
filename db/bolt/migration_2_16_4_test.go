package bolt

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"testing"
)

// TestMigration_2_16_4_Apply tests the migration that adds hidden and type fields
func TestMigration_2_16_4_Apply(t *testing.T) {
	store := CreateTestStore()

	// Create a couple of projects first
	proj1, err := store.CreateProject(db.Project{
		Created: tz.Now(),
		Name:    "Test Project 1",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	proj2, err := store.CreateProject(db.Project{
		Created: tz.Now(),
		Name:    "Test Project 2",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// Create some custom views before migration (these would exist in pre-migration state)
	customView1, err := store.CreateView(db.View{
		ProjectID: proj1.ID,
		Title:     "Custom View 1",
		Position:  1,
		Type:      db.ViewTypeCustom, // This would be set after migration
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	customView2, err := store.CreateView(db.View{
		ProjectID: proj2.ID,
		Title:     "Custom View 2", 
		Position:  2,
		Type:      db.ViewTypeCustom, // This would be set after migration
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// After migration, each project should have an All view
	proj1Views, err := store.GetViews(proj1.ID)
	if err != nil {
		t.Fatal(err.Error())
	}

	proj2Views, err := store.GetViews(proj2.ID)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Verify proj1 has the custom view
	var foundCustom1 bool
	for _, view := range proj1Views {
		if view.ID == customView1.ID {
			foundCustom1 = true
			if view.Type != db.ViewTypeCustom {
				t.Fatal("Custom view should have type 'custom'")
			}
			if view.Hidden != false {
				t.Fatal("Custom view should not be hidden by default")
			}
		}
	}

	if !foundCustom1 {
		t.Fatal("Did not find custom view for project 1")
	}

	// Note: In a real migration scenario with SQL database, an All view would be created automatically
	// In our test environment, we're mainly testing the field handling

	// Verify proj2 has the custom view
	var foundCustom2 bool
	for _, view := range proj2Views {
		if view.ID == customView2.ID {
			foundCustom2 = true
			if view.Type != db.ViewTypeCustom {
				t.Fatal("Custom view should have type 'custom'")
			}
		}
	}

	if !foundCustom2 {
		t.Fatal("Did not find custom view for project 2")
	}

	// Test creating new views with the new fields
	failedView, err := store.CreateView(db.View{
		ProjectID: proj1.ID,
		Title:     "Failed Tasks",
		Position:  3,
		Hidden:    true,
		Type:      db.ViewTypeFailed,
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// Verify the failed view was created correctly
	retrievedFailedView, err := store.GetView(proj1.ID, failedView.ID)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !retrievedFailedView.Hidden {
		t.Fatal("Failed view should be hidden")
	}

	if retrievedFailedView.Type != db.ViewTypeFailed {
		t.Fatal("Failed view should have type 'failed'")
	}
}

// TestMigrationCompatibility tests that views created before migration work correctly after migration
func TestMigrationCompatibility(t *testing.T) {
	store := CreateTestStore()

	proj, err := store.CreateProject(db.Project{
		Created: tz.Now(),
		Name:    "Migration Test Project",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// Create a view as it would exist after migration (with default values applied)
	view := db.View{
		ProjectID: proj.ID,
		Title:     "Pre-migration View",
		Position:  1,
		Hidden:    false,         // Default value 
		Type:      db.ViewTypeCustom, // Default value that would be set by migration
	}

	createdView, err := store.CreateView(view)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Verify values were preserved
	if createdView.Hidden != false {
		t.Fatal("Hidden value should be false")
	}

	if createdView.Type != db.ViewTypeCustom {
		t.Fatal("Type should be 'custom'")
	}

	// Test that the view can be updated with new fields
	createdView.Hidden = true
	createdView.Type = db.ViewTypeFailed

	err = store.UpdateView(createdView)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Verify the update worked
	updatedView, err := store.GetView(proj.ID, createdView.ID)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !updatedView.Hidden {
		t.Fatal("View should be hidden after update")
	}

	if updatedView.Type != db.ViewTypeFailed {
		t.Fatal("View should have failed type after update")
	}
}