package boltdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *BoltdbStore) TemplateTasks(ctx context.Context, params model.TaskParams) ([]*model.TaskWithTpl, error) {
	return nil, nil
}

func (db *BoltdbStore) ProjectTasks(ctx context.Context, params model.TaskParams) ([]*model.TaskWithTpl, error) {
	return nil, nil
}

func (db *BoltdbStore) ShowTask(ctx context.Context, params model.TaskParams) (*model.Task, error) {
	return nil, nil
}

func (db *BoltdbStore) DeleteTask(ctx context.Context, params model.TaskParams) error {
	return nil
}

func (db *BoltdbStore) CreateTask(ctx context.Context, record *model.Task) error {
	return nil
}

func (db *BoltdbStore) UpdateTask(ctx context.Context, record *model.Task) error {
	return nil
}

func (db *BoltdbStore) PushOutput(ctx context.Context, record *model.TaskOutput) error {
	return nil
}

func (db *BoltdbStore) GetOutputs(ctx context.Context, params model.TaskParams) ([]*model.TaskOutput, error) {
	return nil, nil
}
