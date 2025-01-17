package bolt

import (
	"testing"
	"time"

	"github.com/semaphoreui/semaphore/db"
)

func TestGetTemplateVaults(t *testing.T) {
	store := CreateTestStore()

	proj, err := store.CreateProject(db.Project{
		Created: time.Now(),
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

	vault, err := store.CreateTemplateVault(db.TemplateVault{
		ProjectID:  proj.ID,
		TemplateID: template.ID,
		Type:       "password",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	vaults, err := store.GetTemplateVaults(proj.ID, template.ID)
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(vaults) != 1 || vaults[0].ID != vault.ID {
		t.Fatalf("expected 1 vault, got %d", len(vaults))
	}
}

func TestCreateTemplateVault(t *testing.T) {
	store := CreateTestStore()

	proj, err := store.CreateProject(db.Project{
		Created: time.Now(),
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

	vault, err := store.CreateTemplateVault(db.TemplateVault{
		ProjectID:  proj.ID,
		TemplateID: template.ID,
		Type:       "password",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	foundVaults, err := store.GetTemplateVaults(proj.ID, template.ID)
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(foundVaults) != 1 || foundVaults[0].ID != vault.ID {
		t.Fatalf("expected 1 vault, got %d", len(foundVaults))
	}
}

func TestUpdateTemplateVaults(t *testing.T) {
	store := CreateTestStore()

	proj, err := store.CreateProject(db.Project{
		Created: time.Now(),
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

	_, err = store.CreateTemplateVault(db.TemplateVault{
		ProjectID:  proj.ID,
		TemplateID: template.ID,
		Type:       "password",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	vault2 := db.TemplateVault{
		ProjectID:  proj.ID,
		TemplateID: template.ID,
		Type:       "script",
	}

	err = store.UpdateTemplateVaults(proj.ID, template.ID, []db.TemplateVault{vault2})
	if err != nil {
		t.Fatal(err.Error())
	}

	vaults, err := store.GetTemplateVaults(proj.ID, template.ID)
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(vaults) != 1 || vaults[0].Type != "script" {
		t.Fatalf("expected 1 vault with type 'script', got %d", len(vaults))
	}
}
