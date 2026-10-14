package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createInventoryValidationFixture(t *testing.T, store *SqlDb) (db.Project, db.Repository) {
	t.Helper()

	project, err := store.CreateProject(db.Project{Name: "p"})
	require.NoError(t, err)

	key, err := store.CreateAccessKey(db.AccessKey{Name: "none", Type: db.AccessKeyNone, ProjectID: &project.ID})
	require.NoError(t, err)

	repo, err := store.CreateRepository(db.Repository{
		Name:      "repo",
		ProjectID: project.ID,
		GitURL:    "git@example.com:org/inv.git",
		GitBranch: "main",
		SSHKeyID:  key.ID,
	})
	require.NoError(t, err)

	return project, repo
}

func TestValidateInventory_RepositoryID(t *testing.T) {
	store := InitConfigCreateTestStore()
	project, repo := createInventoryValidationFixture(t, store)

	otherProject, err := store.CreateProject(db.Project{Name: "other"})
	require.NoError(t, err)

	missingID := repo.ID + 1000

	tests := []struct {
		name         string
		inventory    db.Inventory
		wantErr      bool
		wantRepoID   *int
		wantNotFound bool
	}{
		{
			name:       "file keeps own repository",
			inventory:  db.Inventory{ProjectID: project.ID, Type: db.InventoryFile, RepositoryID: &repo.ID},
			wantRepoID: &repo.ID,
		},
		{
			name:      "file without repository",
			inventory: db.Inventory{ProjectID: project.ID, Type: db.InventoryFile},
		},
		{
			name:      "static drops stale repository",
			inventory: db.Inventory{ProjectID: project.ID, Type: db.InventoryStatic, RepositoryID: &repo.ID},
		},
		{
			name:      "static yaml drops stale repository",
			inventory: db.Inventory{ProjectID: project.ID, Type: db.InventoryStaticYaml, RepositoryID: &repo.ID},
		},
		{
			name:      "terraform workspace drops repository",
			inventory: db.Inventory{ProjectID: project.ID, Type: db.InventoryTerraformWorkspace, RepositoryID: &repo.ID},
		},
		{
			name:      "static with missing repository is not an error",
			inventory: db.Inventory{ProjectID: project.ID, Type: db.InventoryStatic, RepositoryID: &missingID},
		},
		{
			name:      "file with missing repository",
			inventory: db.Inventory{ProjectID: project.ID, Type: db.InventoryFile, RepositoryID: &missingID},
			wantErr:   true,
		},
		{
			name:      "file with repository of another project",
			inventory: db.Inventory{ProjectID: otherProject.ID, Type: db.InventoryFile, RepositoryID: &repo.ID},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := tt.inventory
			err := db.ValidateInventory(store, &inv)
			if tt.wantErr {
				assert.ErrorIs(t, err, db.ErrNotFound)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantRepoID, inv.RepositoryID)
		})
	}
}
