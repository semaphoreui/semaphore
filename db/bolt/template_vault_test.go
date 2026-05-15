package bolt

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
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

// When a password vault references a deleted access key, GetTemplateVaults skips that row
// so API reads still work. UpdateTemplateVaults must still delete the underlying Bolt rows,
// otherwise replacements leave orphaned records.
func TestUpdateTemplateVaults_RemovesDanglingPasswordVault(t *testing.T) {
	store := CreateTestStore()

	proj, err := store.CreateProject(db.Project{
		Created: tz.Now(),
		Name:    "TestProject",
	})
	if err != nil {
		t.Fatal(err)
	}

	template, err := store.CreateTemplate(db.Template{
		ProjectID: proj.ID,
		Name:      "TestTemplate",
		Playbook:  "test.yml",
	})
	if err != nil {
		t.Fatal(err)
	}

	key, err := store.CreateAccessKey(db.AccessKey{
		ProjectID: &proj.ID,
		Type:      db.AccessKeyNone,
		Name:      "vault-key",
	})
	if err != nil {
		t.Fatal(err)
	}

	vkid := key.ID
	_, err = store.CreateTemplateVault(db.TemplateVault{
		ProjectID:  proj.ID,
		TemplateID: template.ID,
		Type:       "password",
		VaultKeyID: &vkid,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := store.DeleteAccessKey(proj.ID, key.ID); err != nil {
		t.Fatal(err)
	}

	vaults, err := store.GetTemplateVaults(proj.ID, template.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(vaults) != 0 {
		t.Fatalf("expected broken vault omitted from GetTemplateVaults, got %d", len(vaults))
	}

	raw, err := store.listRawTemplateVaults(proj.ID, template.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != 1 {
		t.Fatalf("expected 1 raw vault row before update, got %d", len(raw))
	}

	script := "echo hi"
	err = store.UpdateTemplateVaults(proj.ID, template.ID, []db.TemplateVault{
		{Type: "script", Script: &script},
	})
	if err != nil {
		t.Fatal(err)
	}

	raw, err = store.listRawTemplateVaults(proj.ID, template.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != 1 {
		t.Fatalf("expected 1 vault row after update, got %d", len(raw))
	}
	if raw[0].Type != "script" {
		t.Fatalf("expected script vault, got %s", raw[0].Type)
	}
}
