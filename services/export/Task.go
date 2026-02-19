package export

import (
	"math"
	"slices"
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type TaskExporter struct {
	ValueMap[db.Task]
}

func (e *TaskExporter) load(store db.Store, exporter DataExporter, progress Progress) error {
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

		slices.Reverse(tasks)

		err = e.appendValues(tasks, strconv.Itoa(proj))
		if err != nil {
			return err
		}
	}

	return nil

}

func (e *TaskExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {

	for _, val := range e.values {
		old := val.value

		old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID, e)

		old.TemplateID, err = exporter.getNewKeyInt(Template, val.scope, old.TemplateID, e)

		old.InventoryID, err = exporter.getNewKeyIntRef(Inventory, val.scope, old.InventoryID, e)

		old.ScheduleID, err = exporter.getNewKeyIntRef(Schedule, val.scope, old.ScheduleID, e)

		old.UserID, err = exporter.getNewKeyIntRef(User, GlobalScope, old.UserID, e)

		old.IntegrationID, err = exporter.getNewKeyIntRef(Integration, val.scope, old.IntegrationID, e)

		old.BuildTaskID, err = exporter.getNewKeyIntRef(Task, val.scope, old.BuildTaskID, e)

		if err != nil {
			return err
		}

		newVault, err := store.CreateTask(old, math.MaxInt)
		if err != nil {
			return err
		}

		err = exporter.mapIntKeys(e.getName(), val.scope, old.ID, newVault.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *TaskExporter) getName() string {
	return Task
}

func (e *TaskExporter) exportDependsOn() []string {
	return []string{Project}
}

func (e *TaskExporter) importDependsOn() []string {
	return []string{Project, Template, Inventory, Integration, Schedule, User}
}
