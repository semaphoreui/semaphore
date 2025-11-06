package bolt

import (
	"time"

	"github.com/semaphoreui/semaphore/db"
	"go.etcd.io/bbolt"
)

func (d *BoltDb) CreateTaskStage(stage db.TaskStage) (db.TaskStage, error) {
	newOutput, err := d.createObject(stage.TaskID, db.TaskStageProps, stage)
	if err != nil {
		return db.TaskStage{}, err
	}
	return newOutput.(db.TaskStage), nil
}

func (d *BoltDb) GetTaskStages(projectID int, taskID int) (res []db.TaskStageWithResult, err error) {
	// check if task exists in the project
	_, err = d.GetTask(projectID, taskID)

	if err != nil {
		return
	}

	var stages []db.TaskStage
	err = d.getObjects(taskID, db.TaskStageProps, db.RetrieveQueryParams{}, nil, &stages)
	if err != nil {
		return
	}

	// Convert TaskStage to TaskStageWithResult
	res = make([]db.TaskStageWithResult, len(stages))
	for i, stage := range stages {
		res[i] = db.TaskStageWithResult{
			ID:     stage.ID,
			TaskID: stage.TaskID,
			Start:  stage.Start,
			End:    stage.End,
			Type:   stage.Type,
		}
	}

	return
}

func (d *BoltDb) clearTasks(projectID int, templateID int, maxTasks int) {
	tpl, err := d.GetTemplate(projectID, templateID)
	if err != nil {
		return
	}

	nTasks := tpl.Tasks

	if nTasks == 0 { // recalculate number of tasks for the template

		n, err := d.count(projectID, db.TaskProps, db.RetrieveQueryParams{}, func(item any) bool {
			task := item.(db.Task)

			return task.TemplateID == templateID
		})

		if err != nil {
			return
		}

		if n != nTasks {
			tpl.Tasks = n
			err = d.UpdateTemplate(tpl)

			if err != nil {
				return
			}
		}

		nTasks = n
	}

	if nTasks < maxTasks {
		return
	}

	i := 0

	_ = d.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(makeBucketId(db.TaskProps, projectID))
		if b == nil {
			return db.ErrNotFound
		}

		c := b.Cursor()

		return apply(c, db.TaskProps, db.RetrieveQueryParams{}, func(item any) bool {
			task := item.(db.Task)

			if task.TemplateID != templateID {
				return false
			}

			i++
			return i > maxTasks
		}, func(i any) error {
			task := i.(db.Task)
			return d.deleteTaskWithOutputs(projectID, task.ID, false, tx)
		})
	})
}

func (d *BoltDb) CreateTask(task db.Task, maxTasks int) (newTask db.Task, err error) {
	err = task.PreInsert(nil)
	if err != nil {
		return
	}

	task.ID = 0
	res, err := d.createObject(0, db.TaskProps, task)
	if err != nil {
		return
	}
	newTask = res.(db.Task)

	if maxTasks > 0 {
		d.clearTasks(task.ProjectID, task.TemplateID, maxTasks)
	}

	return
}

func (d *BoltDb) UpdateTask(task db.Task) error {
	return d.updateObject(0, db.TaskProps, task)
}

func (d *BoltDb) CreateTaskOutput(output db.TaskOutput) (db.TaskOutput, error) {
	newOutput, err := d.createObject(output.TaskID, db.TaskOutputProps, output)
	if err != nil {
		return db.TaskOutput{}, err
	}
	return newOutput.(db.TaskOutput), nil
}

func (d *BoltDb) InsertTaskOutputBatch(output []db.TaskOutput) error {
	if len(output) == 0 {
		return nil
	}

	return d.db.Update(func(tx *bbolt.Tx) error {
		for _, out := range output {
			_, err := d.createObjectTx(tx, out.TaskID, db.TaskOutputProps, out)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (d *BoltDb) getTasks(projectID int, templateID *int, params db.RetrieveQueryParams) (tasksWithTpl []db.TaskWithTpl, err error) {
	var tasks []db.Task

	err = d.getObjects(0, db.TaskProps, params, func(tsk any) bool {
		task := tsk.(db.Task)

		if task.ProjectID != projectID {
			return false
		}

		if templateID != nil && task.TemplateID != *templateID {
			return false
		}

		return true
	}, &tasks)

	if err != nil {
		return
	}

	var templates = make(map[int]db.Template)
	var users = make(map[int]db.User)

	tasksWithTpl = make([]db.TaskWithTpl, len(tasks))
	for i, task := range tasks {
		tpl, ok := templates[task.TemplateID]
		if !ok {
			if templateID == nil {
				tpl, _ = d.getRawTemplate(task.ProjectID, task.TemplateID)
			} else {
				tpl, _ = d.getRawTemplate(task.ProjectID, *templateID)
			}
			templates[task.TemplateID] = tpl
		}
		tasksWithTpl[i] = db.TaskWithTpl{Task: task}
		tasksWithTpl[i].TemplatePlaybook = tpl.Playbook
		tasksWithTpl[i].TemplateAlias = tpl.Name
		tasksWithTpl[i].TemplateType = tpl.Type
		tasksWithTpl[i].TemplateApp = tpl.App
		if task.UserID != nil {
			usr, ok := users[*task.UserID]
			if !ok {
				// trying to get user , but ignore error, because
				// user can be deleted, and it is ok
				usr, _ = d.GetUser(*task.UserID)
				users[*task.UserID] = usr
			}
			tasksWithTpl[i].UserName = &usr.Name
		}

		err = tasksWithTpl[i].Fill(d)
		if err != nil {
			return
		}
	}

	return
}

func (d *BoltDb) GetTask(projectID int, taskID int) (task db.Task, err error) {
	err = d.getObject(0, db.TaskProps, intObjectID(taskID), &task)
	if err != nil {
		return
	}

	if task.ProjectID != projectID {
		task = db.Task{}
		err = db.ErrNotFound
		return
	}

	return
}

func (d *BoltDb) GetTemplateTasks(projectID int, templateID int, params db.RetrieveQueryParams) ([]db.TaskWithTpl, error) {
	return d.getTasks(projectID, &templateID, params)
}

func (d *BoltDb) GetProjectTasks(projectID int, params db.RetrieveQueryParams) ([]db.TaskWithTpl, error) {
	return d.getTasks(projectID, nil, params)
}

func (d *BoltDb) deleteTaskWithOutputs(projectID int, taskID int, checkTaskExisting bool, tx *bbolt.Tx) (err error) {

	if checkTaskExisting {
		_, err = d.GetTask(projectID, taskID)
		if err != nil {
			return
		}
	}

	err = d.deleteObject(0, db.TaskProps, intObjectID(taskID), tx)
	if err != nil {
		return
	}

	err = tx.DeleteBucket(makeBucketId(db.TaskOutputProps, taskID))
	if err == bbolt.ErrBucketNotFound {
		err = nil
	}

	return
}

func (d *BoltDb) DeleteTaskWithOutputs(projectID int, taskID int) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		return d.deleteTaskWithOutputs(projectID, taskID, true, tx)
	})
}

func (d *BoltDb) GetTaskOutputs(projectID int, taskID int, params db.RetrieveQueryParams) (outputs []db.TaskOutput, err error) {
	// check if task exists in the project
	_, err = d.GetTask(projectID, taskID)

	if err != nil {
		return
	}

	err = d.getObjects(taskID, db.TaskOutputProps, params, nil, &outputs)

	return
}

func (d *BoltDb) EndTaskStage(taskID int, stageID int, end time.Time) error {
	return nil
}

func (d *BoltDb) CreateTaskStageResult(taskID int, stageID int, result map[string]any) error {
	return nil
}

func (d *BoltDb) GetTaskStageResult(projectID int, taskID int, stageID int) (res db.TaskStageResult, err error) {
	return
}

func (d *BoltDb) GetTaskStageOutputs(projectID int, taskID int, stageID int) (res []db.TaskOutput, err error) {
	return
}
