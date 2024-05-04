package gormdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/pkg/store"
	"gorm.io/gorm"
)

func (db *GormdbStore) ListGlobalRunners(ctx context.Context, params model.RunnerParams) ([]*model.Runner, error) {
	records := make([]*model.Runner, 0)

	err := db.handle.WithContext(ctx).Order(
		"id ASC",
	).Find(
		&records,
	).Error

	return records, err
}

func (db *GormdbStore) ShowGlobalRunner(ctx context.Context, params model.RunnerParams) (*model.Runner, error) {
	record := &model.Runner{}

	err := db.handle.WithContext(ctx).Where(
		"id = ?",
		params.RunnerID,
	).First(
		record,
	).Error

	if err == gorm.ErrRecordNotFound {
		return record, store.ErrRecordNotFound
	}

	return record, err
}

func (db *GormdbStore) DeleteGlobalRunner(ctx context.Context, params model.RunnerParams) error {
	tx := db.handle.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := tx.Where(
		"id = ?",
		params.RunnerID,
	).Delete(
		&model.Runner{},
	).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return store.ErrRecordNotFound
		}

		return err
	}

	return tx.Commit().Error
}

func (db *GormdbStore) CreateGlobalRunner(ctx context.Context, record *model.Runner) error {
	tx := db.handle.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := tx.Create(record).Error; err != nil {
		return err
	}

	return tx.Commit().Error
}

func (db *GormdbStore) UpdateGlobalRunner(ctx context.Context, record *model.Runner) error {
	tx := db.handle.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := tx.Save(record).Error; err != nil {
		return err
	}

	return tx.Commit().Error
}

func (db *GormdbStore) ListProjectRunner(ctx context.Context, params model.RunnerParams) ([]*model.Runner, error) {
	records := make([]*model.Runner, 0)

	err := db.handle.WithContext(ctx).Order(
		"id ASC",
	).Where(
		"project_id = ?",
		params.ProjectID,
	).Find(
		&records,
	).Error

	return records, err
}

func (db *GormdbStore) ShowProjectRunner(ctx context.Context, params model.RunnerParams) (*model.Runner, error) {
	record := &model.Runner{}

	err := db.handle.WithContext(ctx).Where(
		"id = ?",
		params.RunnerID,
	).Where(
		"project_id = ?",
		params.ProjectID,
	).First(
		record,
	).Error

	if err == gorm.ErrRecordNotFound {
		return record, store.ErrRecordNotFound
	}

	return record, err
}

func (db *GormdbStore) DeleteProjectRunner(ctx context.Context, params model.RunnerParams) error {
	tx := db.handle.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := tx.Where(
		"id = ?",
		params.RunnerID,
	).Where(
		"project_id = ?",
		params.ProjectID,
	).Delete(
		&model.Runner{},
	).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return store.ErrRecordNotFound
		}

		return err
	}

	return tx.Commit().Error
}

func (db *GormdbStore) CreateProjectRunner(ctx context.Context, record *model.Runner) error {
	tx := db.handle.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := tx.Create(record).Error; err != nil {
		return err
	}

	return tx.Commit().Error
}

func (db *GormdbStore) UpdateProjectRunner(ctx context.Context, record *model.Runner) error {
	tx := db.handle.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := tx.Save(record).Error; err != nil {
		return err
	}

	return tx.Commit().Error
}
