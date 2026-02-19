package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

type TaskOutputExporter struct {
	ValueMap[db.TaskOutput]
}

func (e *TaskOutputExporter) load(store db.Store, exporter DataExporter, progress Progress) error {
	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	taskCount, err := taskCount(exporter)
	if err != nil {
		return err
	}

	taskIndex := 0

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

			taskIndex = taskIndex + 1
			progress.update(float32(taskIndex) / float32(taskCount))
		}

		err = e.appendValuesAndCheck(allValues, strconv.Itoa(projId), false)
		if err != nil {
			return err
		}
	}

	return nil
}

func taskCount(exporter DataExporter) (int, error) {

	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return 0, err
	}

	count := 0

	for _, projId := range projs {

		tasks, err := exporter.getLoadedKeysInt(Task, strconv.Itoa(projId))
		if err != nil {
			return 0, err
		}
		count = count + len(tasks)
	}

	return count, nil
}

func (e *TaskOutputExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {

	outputs := make([]db.TaskOutput, 0)

	size := len(e.values)

	for index, val := range e.values {
		old := val.value

		old.TaskID, err = exporter.getNewKeyInt(Task, val.scope, old.TaskID, e)
		if err != nil {
			return err
		}

		old.StageID, err = exporter.getNewKeyIntRef(TaskStage, val.scope, old.StageID, e)
		if err != nil {
			return err
		}

		outputs = append(outputs, old)

		if len(outputs) >= 1000 {
			err = store.InsertTaskOutputBatch(outputs)
			if err != nil {
				return err
			}

			outputs = make([]db.TaskOutput, 0)
		}

		progress.update(float32(index) / float32(size))
	}

	if len(outputs) > 0 {
		err = store.InsertTaskOutputBatch(outputs)
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
