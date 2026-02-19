package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type ScheduleExporter struct {
	ValueMap[db.Schedule]
}

func (e *ScheduleExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, proj := range projs {
		vals, err := store.GetProjectSchedules(proj, true, true)
		if err != nil {
			return err
		}
		envs := getSchedules(vals)
		err = e.appendValues(envs, strconv.Itoa(proj))
		if err != nil {
			return err
		}
	}

	return nil
}

func getSchedules(vals []db.ScheduleWithTpl) []db.Schedule {
	values := make([]db.Schedule, 0)

	for _, val := range vals {
		values = append(values, val.Schedule)
	}

	return values
}

func (e *ScheduleExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {

	for _, val := range e.values {
		old := val.value

		if old.TaskParamsID != nil {
			old.TaskParams.InventoryID, err = exporter.getNewKeyIntRef(Inventory, val.scope, old.TaskParams.InventoryID, e)
			if err != nil {
				return err
			}

			old.TaskParams.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID, e)
			if err != nil {
				return err
			}
		}

		old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID, e)
		if err != nil {
			return err
		}

		old.RepositoryID, err = exporter.getNewKeyIntRef(Repository, val.scope, old.RepositoryID, e)
		if err != nil {
			return err
		}

		old.TemplateID, err = exporter.getNewKeyInt(Template, val.scope, old.TemplateID, e)
		if err != nil {
			return err
		}

		newVault, err := store.CreateSchedule(old)
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

func (e *ScheduleExporter) getName() string {
	return Schedule
}

func (e *ScheduleExporter) exportDependsOn() []string {
	return []string{Project}
}

func (e *ScheduleExporter) importDependsOn() []string {
	return []string{SecretStorage, Repository, Project, Inventory, Template}
}
