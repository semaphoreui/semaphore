package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type ScheduleExporter struct {
	ValueMap[db.Schedule]
}

func (a *ScheduleExporter) load(store db.Store, exporter DataExporter) error {

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
		err = a.appendValues(envs, strconv.Itoa(proj))
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

func (a *ScheduleExporter) restore(store db.Store, exporter DataExporter) (err error) {

	for _, val := range a.values {
		old := val.value

		if old.TaskParamsID != nil {
			old.TaskParams.InventoryID, err = exporter.getNewKeyIntRef(Inventory, val.scope, old.TaskParams.InventoryID)
			if err != nil {
				return err
			}

			old.TaskParams.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
			if err != nil {
				return err
			}
		}

		old.ProjectID, err = exporter.getNewKeyInt(Project, GlobalScope, old.ProjectID)
		if err != nil {
			return err
		}

		old.RepositoryID, err = exporter.getNewKeyIntRef(Repository, val.scope, old.RepositoryID)
		if err != nil {
			return err
		}

		newVault, err := store.CreateSchedule(old)
		if err != nil {
			return err
		}

		err = exporter.mapIntKeys(a.getName(), val.scope, old.ID, newVault.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *ScheduleExporter) getName() string {
	return Schedule
}

func (a *ScheduleExporter) exportDependsOn() []string {
	return []string{Project}
}

func (a *ScheduleExporter) importDependsOn() []string {
	return []string{SecretStorage, Repository, Project, Inventory}
}
