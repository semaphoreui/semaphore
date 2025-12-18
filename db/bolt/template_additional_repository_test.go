package bolt

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/tz"
	"testing"
)

func TestGetTemplateAdditionalRepositories(t *testing.T) {
	store := CreateTestStore()

	proj, err := store.CreateProject(db.Project{
		Created: tz.Now(),
		Name:    "TestProject",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// Create access key first
	key, err := store.CreateAccessKey(db.AccessKey{
		Name:      "TestKey",
		Type:      "none",
		ProjectID: &proj.ID,
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID,
		Name:      "TestRepo",
		GitURL:    "https://github.com/test/repo",
		GitBranch: "main",
		SSHKeyID:  key.ID,
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

	addRepos := []db.TemplateAdditionalRepository{
		{
			TemplateID:   template.ID,
			RepositoryID: repo.ID,
			Path:         "libs/mylib",
			Position:     0,
		},
	}

	err = store.UpdateTemplateAdditionalRepositories(proj.ID, template.ID, addRepos)
	if err != nil {
		t.Fatalf("UpdateTemplateAdditionalRepositories failed: %v", err)
	}

	repos, err := store.GetTemplateAdditionalRepositories(proj.ID, template.ID)
	if err != nil {
		t.Fatalf("GetTemplateAdditionalRepositories failed: %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("expected 1 additional repository, got %d", len(repos))
	}

	if repos[0].Path != "libs/mylib" {
		t.Fatalf("expected path 'libs/mylib', got '%s'", repos[0].Path)
	}
}

func TestUpdateTemplateAdditionalRepositories(t *testing.T) {
	store := CreateTestStore()

	proj, err := store.CreateProject(db.Project{
		Created: tz.Now(),
		Name:    "TestProject",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// Create access key first
	key, err := store.CreateAccessKey(db.AccessKey{
		Name:      "TestKey",
		Type:      "none",
		ProjectID: &proj.ID,
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	repo1, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID,
		Name:      "TestRepo1",
		GitURL:    "https://github.com/test/repo1",
		GitBranch: "main",
		SSHKeyID:  key.ID,
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	repo2, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID,
		Name:      "TestRepo2",
		GitURL:    "https://github.com/test/repo2",
		GitBranch: "main",
		SSHKeyID:  key.ID,
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

	// Create initial additional repositories
	addRepos := []db.TemplateAdditionalRepository{
		{
			TemplateID:   template.ID,
			RepositoryID: repo1.ID,
			Path:         "libs/lib1",
			Position:     0,
		},
	}

	err = store.UpdateTemplateAdditionalRepositories(proj.ID, template.ID, addRepos)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Update with different repositories
	updatedRepos := []db.TemplateAdditionalRepository{
		{
			TemplateID:   template.ID,
			RepositoryID: repo2.ID,
			Path:         "libs/lib2",
			Position:     0,
		},
		{
			TemplateID:   template.ID,
			RepositoryID: repo1.ID,
			Path:         "libs/lib1-updated",
			GitBranch:    strPtr("develop"),
			Position:     1,
		},
	}

	err = store.UpdateTemplateAdditionalRepositories(proj.ID, template.ID, updatedRepos)
	if err != nil {
		t.Fatal(err.Error())
	}

	repos, err := store.GetTemplateAdditionalRepositories(proj.ID, template.ID)
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(repos) != 2 {
		t.Fatalf("expected 2 additional repositories, got %d", len(repos))
	}

	if repos[0].Path != "libs/lib2" {
		t.Fatalf("expected first repo path 'libs/lib2', got '%s'", repos[0].Path)
	}

	if repos[1].Path != "libs/lib1-updated" {
		t.Fatalf("expected second repo path 'libs/lib1-updated', got '%s'", repos[1].Path)
	}

	if repos[1].GitBranch == nil || *repos[1].GitBranch != "develop" {
		t.Fatalf("expected second repo branch 'develop', got '%v'", repos[1].GitBranch)
	}
}

func TestUpdateTemplateAdditionalRepositories_Empty(t *testing.T) {
	store := CreateTestStore()

	proj, err := store.CreateProject(db.Project{
		Created: tz.Now(),
		Name:    "TestProject",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	// Create access key first
	key, err := store.CreateAccessKey(db.AccessKey{
		Name:      "TestKey",
		Type:      "none",
		ProjectID: &proj.ID,
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	repo, err := store.CreateRepository(db.Repository{
		ProjectID: proj.ID,
		Name:      "TestRepo",
		GitURL:    "https://github.com/test/repo",
		GitBranch: "main",
		SSHKeyID:  key.ID,
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

	// Create initial additional repository
	addRepos := []db.TemplateAdditionalRepository{
		{
			TemplateID:   template.ID,
			RepositoryID: repo.ID,
			Path:         "libs/mylib",
			Position:     0,
		},
	}

	err = store.UpdateTemplateAdditionalRepositories(proj.ID, template.ID, addRepos)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Update with empty list to remove all
	err = store.UpdateTemplateAdditionalRepositories(proj.ID, template.ID, []db.TemplateAdditionalRepository{})
	if err != nil {
		t.Fatal(err.Error())
	}

	repos, err := store.GetTemplateAdditionalRepositories(proj.ID, template.ID)
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(repos) != 0 {
		t.Fatalf("expected 0 additional repositories after clearing, got %d", len(repos))
	}
}

func strPtr(s string) *string {
	return &s
}
