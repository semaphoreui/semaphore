package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type TaskStageExporter struct {
	ValueMap[db.TaskStage]
}

func (e *TaskStageExporter) load(store db.Store, exporter DataExporter, progress Progress) error {
	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, projId := range projs {

		tasks, err := exporter.getLoadedKeysInt(Task, strconv.Itoa(projId))
		if err != nil {
			return err
		}

		allValues := make([]db.TaskStage, 0)
		for _, task := range tasks {

			stagesRes, err := store.GetTaskStages(projId, task)
			if err != nil {
				return err
			}

			allValues = append(allValues, getStages(stagesRes)...)
		}

		err = e.appendValues(allValues, strconv.Itoa(projId))
		if err != nil {
			return err
		}
	}

	return nil
}

func getStages(vals []db.TaskStageWithResult) []db.TaskStage {
	values := make([]db.TaskStage, 0)

	for _, val := range vals {
		values = append(values, db.TaskStage{
			ID:     val.ID,
			TaskID: val.TaskID,
			Start:  val.Start,
			End:    val.End,
			Type:   val.Type,
		})
	}

	return values
}

func (e *TaskStageExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	for _, val := range e.values {
		old := val.value

		old.TaskID, err = exporter.getNewKeyInt(Task, val.scope, old.TaskID, e)
		if err != nil {
			return err
		}

		newVault, err := store.CreateTaskStage(old)
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

func (e *TaskStageExporter) getName() string {
	return TaskStage
}

func (e *TaskStageExporter) exportDependsOn() []string {
	return []string{Task}
}

func (e *TaskStageExporter) importDependsOn() []string {
	return []string{Task}
}
