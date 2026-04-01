package export

import (
	"strconv"

	"github.com/semaphoreui/semaphore/db"
)

const taskOutputBatchSize = 1000

type TaskOutputExporter struct {
	ValueMap[db.TaskOutput]
	sourceStore  db.Store
	projectTasks map[string][]int
	totalTasks   int
}

// load saves the source store reference and collects the task IDs per project
// without loading any task output rows into memory.  The actual data transfer
// is deferred to restore() which streams rows in small batches.
func (e *TaskOutputExporter) load(store db.Store, exporter DataExporter, progress Progress) error {
	projs, err := exporter.getLoadedKeysInt(Project, GlobalScope)
	if err != nil {
		return err
	}

	e.sourceStore = store
	e.projectTasks = make(map[string][]int)
	e.totalTasks = 0

	for _, projId := range projs {
		tasks, err := exporter.getLoadedKeysInt(Task, strconv.Itoa(projId))
		if err != nil {
			return err
		}

		projScope := strconv.Itoa(projId)
		e.projectTasks[projScope] = tasks
		e.totalTasks += len(tasks)
	}

	return nil
}

// restore streams task output rows from the source store to the destination
// store in pages of taskOutputBatchSize rows, so that only a small number of
// rows are held in memory at any one time regardless of table size.
func (e *TaskOutputExporter) restore(store db.Store, exporter DataExporter, progress Progress) (err error) {
	if e.sourceStore == nil || e.projectTasks == nil {
		return nil
	}

	taskIndex := 0
	batch := make([]db.TaskOutput, 0, taskOutputBatchSize)

	for projScope, tasks := range e.projectTasks {
		// projScope was produced by strconv.Itoa in load(), so this conversion is always safe.
		projId, _ := strconv.Atoi(projScope)

		for _, taskID := range tasks {
			offset := 0

			for {
				outputs, loadErr := e.sourceStore.GetTaskOutputs(projId, taskID, db.RetrieveQueryParams{
					Count:  taskOutputBatchSize,
					Offset: offset,
				})
				if loadErr != nil {
					return loadErr
				}

				if len(outputs) == 0 {
					break
				}

				for _, old := range outputs {
					old.TaskID, err = exporter.getNewKeyInt(Task, projScope, old.TaskID)
					if err != nil {
						return err
					}

					// boltDb currently doesn't support task stages
					old.StageID = nil

					batch = append(batch, old)

					if len(batch) == taskOutputBatchSize {
						if err = store.InsertTaskOutputBatch(batch); err != nil {
							return err
						}
						batch = batch[:0]
					}
				}

				offset += len(outputs)
				if len(outputs) < taskOutputBatchSize {
					break
				}
			}

			taskIndex++
			if e.totalTasks > 0 {
				progress.update(float32(taskIndex)/float32(e.totalTasks), int64(taskIndex))
			}
		}
	}

	if len(batch) > 0 {
		err = store.InsertTaskOutputBatch(batch)
	}

	return err
}

func (e *TaskOutputExporter) clear() {
	e.ValueMap.clear()
	e.sourceStore = nil
	e.projectTasks = nil
	e.totalTasks = 0
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
