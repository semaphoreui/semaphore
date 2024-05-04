package gormdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *GormdbStore) TemplateSchedules(ctx context.Context, params model.ScheduleParams) ([]*model.Schedule, error) {
	return nil, nil
}

func (db *GormdbStore) ListSchedules(ctx context.Context, params model.ScheduleParams) ([]*model.Schedule, error) {
	return nil, nil
}

func (db *GormdbStore) ShowSchedule(ctx context.Context, params model.ScheduleParams) (*model.Schedule, error) {
	return nil, nil
}

func (db *GormdbStore) DeleteSchedule(ctx context.Context, params model.ScheduleParams) error {
	return nil
}

func (db *GormdbStore) HashSchedule(ctx context.Context, params model.ScheduleParams) error {
	return nil
}

func (db *GormdbStore) CreateSchedule(ctx context.Context, record *model.Schedule) error {
	return nil
}

func (db *GormdbStore) UpdateSchedule(ctx context.Context, record *model.Schedule) error {
	return nil
}
