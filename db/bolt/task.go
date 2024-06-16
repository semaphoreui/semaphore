package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
	"go.etcd.io/bbolt"
	"time"
)

func (d *BoltDb) CreateTaskStage(stage db.TaskStage) (db.TaskStage, error) {
	newOutput, err := d.createObject(stage.TaskID, db.TaskStageProps, stage)
	if err != nil {
		return db.TaskStage{}, err
	}
	return newOutput.(db.TaskStage), nil
}

func (d *BoltDb) GetTaskStages(projectID int, taskID int) (res []db.TaskStage, err error) {
	// check if task exists in the project
	_, err = d.GetTask(projectID, taskID)

	if err != nil {
		return
	}

	err = d.getObjects(taskID, db.TaskStageProps, db.RetrieveQueryParams{}, nil, &res)

	return
}

func (d *BoltDb) CreateTask(task db.Task) (newTask db.Task, err error) {
	task.Created = time.Now()
	res, err := d.createObject(0, db.TaskProps, task)
	if err != nil {
		return
	}
	newTask = res.(db.Task)
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

func (d *BoltDb) getTasks(projectID int, templateID *int, params db.RetrieveQueryParams) (tasksWithTpl []db.TaskWithTpl, err error) {
	var tasks []db.Task

	err = d.getObjects(0, db.TaskProps, params, func(tsk interface{}) bool {
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

func (d *BoltDb) deleteTaskWithOutputs(projectID int, taskID int, tx *bbolt.Tx) (err error) {
	// check if task exists in the project
	_, err = d.GetTask(projectID, taskID)
	if err != nil {
		return
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
		return d.deleteTaskWithOutputs(projectID, taskID, tx)
	})
}

func (d *BoltDb) GetTaskOutputs(projectID int, taskID int) (outputs []db.TaskOutput, err error) {
	// check if task exists in the project
	_, err = d.GetTask(projectID, taskID)

	if err != nil {
		return
	}

	err = d.getObjects(taskID, db.TaskOutputProps, db.RetrieveQueryParams{}, nil, &outputs)

	return
}
