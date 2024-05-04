package boltdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *BoltdbStore) TemplateSchedules(ctx context.Context, params model.ScheduleParams) ([]*model.Schedule, error) {
	return nil, nil
}

func (db *BoltdbStore) ListSchedules(ctx context.Context, params model.ScheduleParams) ([]*model.Schedule, error) {
	return nil, nil
}

func (db *BoltdbStore) ShowSchedule(ctx context.Context, params model.ScheduleParams) (*model.Schedule, error) {
	return nil, nil
}

func (db *BoltdbStore) DeleteSchedule(ctx context.Context, params model.ScheduleParams) error {
	return nil
}

func (db *BoltdbStore) HashSchedule(ctx context.Context, params model.ScheduleParams) error {
	return nil
}

func (db *BoltdbStore) CreateSchedule(ctx context.Context, record *model.Schedule) error {
	return nil
}

func (db *BoltdbStore) UpdateSchedule(ctx context.Context, record *model.Schedule) error {
	return nil
}
