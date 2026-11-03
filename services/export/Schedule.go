package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
	log "github.com/sirupsen/logrus"
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

// getSchedules unwraps the list rows. This exporter does not carry workflows,
// so a schedule that starts one has no target on the other side and is left
// out, like an integration with a workflow target.
func getSchedules(vals []db.ScheduleWithTpl) []db.Schedule {
	values := make([]db.Schedule, 0)

	for _, val := range vals {
		if val.RunsWorkflow() {
			log.WithFields(log.Fields{
				"context":     "export",
				"project_id":  val.ProjectID,
				"schedule_id": val.ID,
			}).Warn("schedule starts a workflow and is skipped: workflows are not exported")
			continue
		}
		values = append(values, val.Schedule)
	}

	return values
}
func (e *ScheduleExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *ScheduleExporter) restoreValue(val EntityObject[db.Schedule], store db.Store, exporter DataExporter) (err error) {
	old := val.value

	if old.TaskParamsID != nil {
		old.TaskParams.InventoryID, err = exporter.getNewKeyIntRef(Inventory, val.scope, old.TaskParams.InventoryID, e)
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

	old.RepositoryID, err = exporter.getNewKeyIntRef(Repository, val.scope, old.RepositoryID, e)
	if err != nil {
		return err
	}

	old.TemplateID, err = exporter.getNewKeyIntRef(Template, val.scope, old.TemplateID, e)
	if err != nil {
		return err
	}

	newObj, err := store.CreateSchedule(old)
	if err != nil {
		return err
	}

	return exporter.mapKeys(e.getName(), val.scope, old.GetDbKey(), newObj.GetDbKey())
}

func (e *ScheduleExporter) getName() string {
	return Schedule
}

func (e *ScheduleExporter) exportDependsOn() []string {
	return []string{Project}
}

func (e *ScheduleExporter) importDependsOn() []string {
	return []string{Repository, Project, Inventory, Template}
}
