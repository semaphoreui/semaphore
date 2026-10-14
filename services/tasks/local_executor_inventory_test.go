package tasks

import (
	"os"
	"path"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTmpInventoryFullPath_StaticIsScopedPerTask(t *testing.T) {
	setupExecutorConfig(t)

	repo := &db.Repository{ID: 3, GitURL: "git@example.com:org/inv.git", GitBranch: "main"}
	tmpDir := util.Config.GetProjectTmpDir(7)

	tests := []struct {
		name      string
		inventory db.Inventory
		expected  string
	}{
		{
			name:      "static without repository",
			inventory: db.Inventory{ID: 5, Type: db.InventoryStatic},
			expected:  path.Join(tmpDir, "inventory_5_task_42"),
		},
		{
			name:      "static yaml without repository",
			inventory: db.Inventory{ID: 5, Type: db.InventoryStaticYaml},
			expected:  path.Join(tmpDir, "inventory_5_task_42.yml"),
		},
		{
			// A file inventory converted to static in the UI keeps its repository_id.
			name:      "static with stale repository",
			inventory: db.Inventory{ID: 5, Type: db.InventoryStatic, RepositoryID: &repo.ID, Repository: repo},
			expected:  path.Join(tmpDir, "inventory_5_task_42"),
		},
		{
			name: "static with stale local repository is not written into the repository",
			inventory: db.Inventory{ID: 5, Type: db.InventoryStaticYaml, RepositoryID: &repo.ID,
				Repository: &db.Repository{ID: 3, GitURL: "/srv/repos/inv"}},
			expected: path.Join(tmpDir, "inventory_5_task_42.yml"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &LocalExecutor{
				Task:      db.Task{ID: 42},
				Template:  db.Template{ID: 11, ProjectID: 7},
				Inventory: tt.inventory,
			}
			assert.Equal(t, tt.expected, executor.tmpInventoryFullPath())
		})
	}
}

func TestTmpInventoryFullPath_ParallelTasksGetDistinctStaticFiles(t *testing.T) {
	setupExecutorConfig(t)

	repo := &db.Repository{ID: 3, GitURL: "git@example.com:org/inv.git", GitBranch: "main"}
	inventory := db.Inventory{ID: 5, Type: db.InventoryStatic, RepositoryID: &repo.ID, Repository: repo}
	template := db.Template{ID: 11, ProjectID: 7}

	first := &LocalExecutor{Task: db.Task{ID: 1}, Template: template, Inventory: inventory}
	second := &LocalExecutor{Task: db.Task{ID: 2}, Template: template, Inventory: inventory}

	assert.NotEqual(t, first.tmpInventoryFullPath(), second.tmpInventoryFullPath())
}

func TestTmpInventoryFullPath_FileInventoryKeepsSharedRepositoryPath(t *testing.T) {
	setupExecutorConfig(t)

	repo := &db.Repository{ID: 3, GitURL: "git@example.com:org/inv.git", GitBranch: "main"}
	template := db.Template{ID: 11, ProjectID: 7}
	inventory := db.Inventory{ID: 5, Type: db.InventoryFile, RepositoryID: &repo.ID, Repository: repo}

	first := &LocalExecutor{Task: db.Task{ID: 1}, Template: template, Inventory: inventory}
	second := &LocalExecutor{Task: db.Task{ID: 2}, Template: template, Inventory: inventory}

	expected := path.Join(util.Config.GetProjectTmpDir(7), repo.GetDirName(11)+"_inventory_5")
	assert.Equal(t, expected, first.tmpInventoryFullPath())
	assert.Equal(t, expected, second.tmpInventoryFullPath())

	local := &LocalExecutor{Task: db.Task{ID: 1}, Template: template, Inventory: db.Inventory{
		ID: 5, Type: db.InventoryFile, RepositoryID: &repo.ID,
		Repository: &db.Repository{ID: 3, GitURL: "/srv/repos/inv"},
	}}
	assert.Equal(t, "/srv/repos/inv", local.tmpInventoryFullPath())
}

func TestInstallStaticInventory_ParallelTasksDoNotClobberEachOther(t *testing.T) {
	setupExecutorConfig(t)

	repo := &db.Repository{ID: 3, GitURL: "git@example.com:org/inv.git", GitBranch: "main"}
	template := db.Template{ID: 11, ProjectID: 7}
	require.NoError(t, os.MkdirAll(util.Config.GetProjectTmpDir(7), 0755))

	first := &LocalExecutor{Task: db.Task{ID: 1}, Template: template, Logger: task_logger.NopLogger{}, Inventory: db.Inventory{
		ID: 5, Type: db.InventoryStatic, RepositoryID: &repo.ID, Repository: repo, Inventory: "host-a",
	}}
	second := &LocalExecutor{Task: db.Task{ID: 2}, Template: template, Logger: task_logger.NopLogger{}, Inventory: db.Inventory{
		ID: 5, Type: db.InventoryStatic, RepositoryID: &repo.ID, Repository: repo, Inventory: "host-b",
	}}

	require.NoError(t, first.installStaticInventory())
	require.NoError(t, second.installStaticInventory())

	// The second task finishing first must not remove the first task's inventory.
	second.destroyInventoryFile()

	content, err := os.ReadFile(first.tmpInventoryFullPath())
	require.NoError(t, err)
	assert.Equal(t, "host-a", string(content))

	_, err = os.Stat(second.tmpInventoryFullPath())
	assert.True(t, os.IsNotExist(err))
}
