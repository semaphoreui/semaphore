package export

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type TaskStageResultExporter struct {
	ValueMap[db.TaskStageResult]
}

func (e *TaskStageResultExporter) load(store db.Store, exporter DataExporter, progress Progress) error {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	for _, projId := range projs {

		tasks, err := exporter.getLoadedKeysInt(Task, strconv.Itoa(projId))
		if err != nil {
			return err
		}

		allValues := make([]db.TaskStageResult, 0)
		for _, task := range tasks {
			stagesRes, err := store.GetTaskStages(projId, task)
			if err != nil {
				return err
			}

			allValues = append(allValues, getStageResults(stagesRes)...)
		}

		err = e.appendValues(allValues, strconv.Itoa(projId))
		if err != nil {
			return err
		}
	}

	return nil
}

func getStageResults(vals []db.TaskStageWithResult) []db.TaskStageResult {
	values := make([]db.TaskStageResult, 0)

	for _, val := range vals {
		values = append(values, db.TaskStageResult{
			ID:     val.ID,
			TaskID: val.TaskID,
			JSON:   val.JSON,
		})
	}

	return values
}

func (e *TaskStageResultExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	return e.restoreValues(store, exporter, progress, e)
}

func (e *TaskStageResultExporter) restoreValue(val EntityObject[db.TaskStageResult], store db.Store, exporter DataExporter) (err error) {

	old := val.value

	old.TaskID, err = exporter.getNewKeyInt(Task, val.scope, old.TaskID)
	if err != nil {
		return err
	}

	res := make(map[string]any)
	err = json.Unmarshal([]byte(old.JSON), &res)
	if err != nil {
		fmt.Println("Unable to parse TaskStageResult " + old.JSON)
	}

	return store.CreateTaskStageResult(old.TaskID, old.StageID, res)
}

func (e *TaskStageResultExporter) getName() string {
	return TaskStageResult
}

func (e *TaskStageResultExporter) exportDependsOn() []string {
	return []string{Task, Project}
}

func (e *TaskStageResultExporter) importDependsOn() []string {
	return []string{Task, TaskStage}
}
