package export

import (
	"math"
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type TaskExporter struct {
	ValueMap[db.Task]
}

func (t *TaskExporter) load(store db.Store, exporter DataExporter) error {
	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {
		tasksTmpl, err := store.GetProjectTasks(proj, db.RetrieveQueryParams{})
		if err != nil {
			return err
		}

		tasks := make([]db.Task, len(tasksTmpl))
		for i, task := range tasksTmpl {
			tasks[i] = task.Task
		}
		err = t.appendValues(tasks, strconv.Itoa(proj))
		if err != nil {
			return err
		}
	}

	return nil

}

func (t *TaskExporter) restore(store db.Store, exporter DataExporter) (err error) {
	for _, val := range t.values {
		old := val.value

		old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)

		old.TemplateID, err = exporter.getNewKeyInt(Template, val.scope, old.TemplateID)

		old.InventoryID, err = exporter.getNewKeyIntRef(Inventory, val.scope, old.InventoryID)

		old.ScheduleID, err = exporter.getNewKeyIntRef(Schedule, val.scope, old.ScheduleID)

		old.UserID, err = exporter.getNewKeyIntRef(User, GlobalScope, old.UserID)

		old.IntegrationID, err = exporter.getNewKeyIntRef(Integration, val.scope, old.IntegrationID)

		old.BuildTaskID, err = exporter.getNewKeyIntRef(Task, val.scope, old.BuildTaskID)

		if err != nil {
			return err
		}

		newVault, err := store.CreateTask(old, math.MaxInt)
		if err != nil {
			return err
		}

		err = exporter.mapIntKeys(t.getName(), val.scope, old.ID, newVault.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *TaskExporter) getName() string {
	return Task
}

func (t *TaskExporter) exportDependsOn() []string {
	return []string{Project}
}

func (t *TaskExporter) importDependsOn() []string {
	return []string{Project, Template, Inventory, Integration, Schedule, User}
}
