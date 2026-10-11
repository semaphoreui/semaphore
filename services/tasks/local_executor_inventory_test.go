package tasks

import (
	"os"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticInventoryFileIsScopedToTask(t *testing.T) {
	previousConfig := util.Config
	t.Cleanup(func() { util.Config = previousConfig })
	util.Config = &util.ConfigType{TmpPath: t.TempDir()}
	projectID := 1
	require.NoError(t, os.MkdirAll(util.Config.GetProjectTmpDir(projectID), 0755))

	for _, inventoryType := range []db.InventoryType{db.InventoryStatic, db.InventoryStaticYaml} {
		t.Run(string(inventoryType), func(t *testing.T) {
			makeExecutor := func(taskID int) *LocalExecutor {
				return &LocalExecutor{
					Task:     db.Task{ID: taskID},
					Template: db.Template{ProjectID: projectID},
					Logger:   task_logger.NopLogger{},
					Inventory: db.Inventory{
						ID:        58,
						Type:      inventoryType,
						Inventory: "all:\n  hosts:\n    localhost:\n",
					},
				}
			}

			first := makeExecutor(101)
			second := makeExecutor(102)
			require.NotEqual(t, first.tmpInventoryFullPath(), second.tmpInventoryFullPath())
			require.NoError(t, first.installStaticInventory())
			require.NoError(t, second.installStaticInventory())

			first.destroyInventoryFile()
			assert.FileExists(t, second.tmpInventoryFullPath())
			second.destroyInventoryFile()
		})
	}
}
