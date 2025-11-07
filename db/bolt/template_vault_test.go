package bolt

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"testing"
)

func TestGetTemplateVaults(t *testing.T) {
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

func TestUpdateTemplatePreservesVaults(t *testing.T) {
	store := CreateTestStore()

	proj, err := store.CreateProject(db.Project{
		Created: tz.Now(),
		Name:    "TestProject",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// Create a template with a vault
	template, err := store.CreateTemplate(db.Template{
		ProjectID: proj.ID,
		Name:      "TestTemplate",
		Playbook:  "test.yml",
		Vaults: []db.TemplateVault{
			{
				Type: "password",
			},
		},
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// Verify vault was created
	vaultsBefore, err := store.GetTemplateVaults(proj.ID, template.ID)
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(vaultsBefore) != 1 {
		t.Fatalf("expected 1 vault before update, got %d", len(vaultsBefore))
	}

	// Update the template without passing vaults (simulating an API call that only updates name)
	template.Name = "UpdatedName"
	template.Vaults = nil // This simulates what happens when the API receives a request without vaults field
	err = store.UpdateTemplate(template)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Verify vaults were NOT deleted
	vaultsAfter, err := store.GetTemplateVaults(proj.ID, template.ID)
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(vaultsAfter) != 1 {
		t.Fatalf("expected 1 vault after update, got %d (vaults should be preserved)", len(vaultsAfter))
	}
}

func TestUpdateTemplateWithVaults(t *testing.T) {
	store := CreateTestStore()

	proj, err := store.CreateProject(db.Project{
		Created: tz.Now(),
		Name:    "TestProject",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// Create a template with a vault
	template, err := store.CreateTemplate(db.Template{
		ProjectID: proj.ID,
		Name:      "TestTemplate",
		Playbook:  "test.yml",
		Vaults: []db.TemplateVault{
			{
				Type: "password",
			},
		},
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// Verify vault was created
	vaultsBefore, err := store.GetTemplateVaults(proj.ID, template.ID)
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(vaultsBefore) != 1 || vaultsBefore[0].Type != "password" {
		t.Fatalf("expected 1 password vault before update, got %d", len(vaultsBefore))
	}

	// Update the template WITH vaults (explicitly changing vaults)
	template.Name = "UpdatedName"
	template.Vaults = []db.TemplateVault{
		{
			Type: "script",
		},
	}
	err = store.UpdateTemplate(template)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Verify vaults were updated to the new vaults
	vaultsAfter, err := store.GetTemplateVaults(proj.ID, template.ID)
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(vaultsAfter) != 1 {
		t.Fatalf("expected 1 vault after update, got %d", len(vaultsAfter))
	}
	if vaultsAfter[0].Type != "script" {
		t.Fatalf("expected script vault after update, got %s", vaultsAfter[0].Type)
	}
}
