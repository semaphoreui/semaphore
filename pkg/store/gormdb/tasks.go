package gormdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *GormdbStore) TemplateTasks(ctx context.Context, params model.TaskParams) ([]*model.TaskWithTpl, error) {
	return nil, nil
}

func (db *GormdbStore) ProjectTasks(ctx context.Context, params model.TaskParams) ([]*model.TaskWithTpl, error) {
	return nil, nil
}

func (db *GormdbStore) ShowTask(ctx context.Context, params model.TaskParams) (*model.Task, error) {
	return nil, nil
}

func (db *GormdbStore) DeleteTask(ctx context.Context, params model.TaskParams) error {
	return nil
}

func (db *GormdbStore) CreateTask(ctx context.Context, record *model.Task) error {
	return nil
}

func (db *GormdbStore) UpdateTask(ctx context.Context, record *model.Task) error {
	return nil
}

func (db *GormdbStore) PushOutput(ctx context.Context, record *model.TaskOutput) error {
	return nil
}

func (db *GormdbStore) GetOutputs(ctx context.Context, params model.TaskParams) ([]*model.TaskOutput, error) {
	return nil, nil
}
