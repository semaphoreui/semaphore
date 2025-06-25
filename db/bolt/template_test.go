package bolt

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"testing"
)

func Test_SetTemplateDescription(t *testing.T) {
	store := CreateTestStore()

	proj, err := store.CreateProject(db.Project{
		Created: tz.Now(),
		Name:    "TestProject",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	template, err := store.CreateTemplate(db.Template{
		ProjectID: proj.ID,
		Name:      "TestTemplate",
		Playbook:  "test.yml",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	err = store.SetTemplateDescription(proj.ID, template.ID, "New description")
	if err != nil {
		t.Fatal(err.Error())
	}

	tpl, err := store.GetTemplate(proj.ID, template.ID)
	if err != nil {
		t.Fatal(err.Error())
	}

	if *tpl.Description != "New description" {
		t.Fatalf("expected description to be 'New description', got '%s'", *tpl.Description)
	}
}
