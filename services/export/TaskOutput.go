package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type TaskOutputExporter struct {
	ValueMap[db.TaskOutput]
}

func (e *TaskOutputExporter) load(store db.Store, exporter DataExporter) error {
	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, projId := range projs {

		tasks, err := exporter.getLoadedKeysInt(Task, strconv.Itoa(projId))
		if err != nil {
			return err
		}

		allValues := make([]db.TaskOutput, 0)
		for _, task := range tasks {

			outputRes, err := store.GetTaskOutputs(projId, task, db.RetrieveQueryParams{})
			if err != nil {
				return err
			}

			allValues = append(allValues, outputRes...)
		}

		err = e.appendValuesAndCheck(allValues, strconv.Itoa(projId), false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *TaskOutputExporter) restore(store db.Store, exporter DataExporter) (err error) {
	for _, val := range e.values {
		old := val.value

		old.TaskID, err = exporter.getNewKeyInt(Task, val.scope, old.TaskID)
		if err != nil {
			return err
		}

		old.StageID, err = exporter.getNewKeyIntRef(TaskStage, val.scope, old.StageID)
		if err != nil {
			return err
		}

		newVault, err := store.CreateTaskOutput(old)
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

func (e *TaskOutputExporter) getName() string {
	return TaskOutput
}

func (e *TaskOutputExporter) exportDependsOn() []string {
	return []string{Task}
}

func (e *TaskOutputExporter) importDependsOn() []string {
	return []string{Task, TaskStage}
}
