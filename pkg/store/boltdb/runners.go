package boltdb

import (
	"context"
	"errors"

	model "github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/pkg/store"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
)

func (db *BoltdbStore) ListGlobalRunners(_ context.Context, params model.RunnerParams) ([]*model.Runner, error) {
	records := make([]*model.Runner, 0)

	if err := db.handle.All(
		&records,
	); err != nil && !errors.Is(err, storm.ErrNotFound) {
		return nil, err
	}

	return records, nil
}

func (db *BoltdbStore) ShowGlobalRunner(_ context.Context, params model.RunnerParams) (*model.Runner, error) {
	record := &model.Runner{}

	if err := db.handle.Select(
		q.Eq("id", params.RunnerID),
	).First(record); err != nil {
		if err == storm.ErrNotFound {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	return record, nil
}

func (db *BoltdbStore) DeleteGlobalRunner(_ context.Context, params model.RunnerParams) error {
	tx, err := db.handle.Begin(true)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	record := &model.Runner{
		ID: params.RunnerID,
	}

	if err := tx.DeleteStruct(record); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *BoltdbStore) CreateGlobalRunner(_ context.Context, record *model.Runner) error {
	tx, err := db.handle.Begin(true)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := tx.Save(record); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (db *BoltdbStore) UpdateGlobalRunner(_ context.Context, record *model.Runner) error {
	tx, err := db.handle.Begin(true)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := tx.Save(record); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (db *BoltdbStore) ListProjectRunner(_ context.Context, params model.RunnerParams) ([]*model.Runner, error) {
	records := make([]*model.Runner, 0)

	if err := db.handle.Select(
		q.Eq("project_id", params.ProjectID),
	).Find(
		&records,
	); err != nil && !errors.Is(err, storm.ErrNotFound) {
		return nil, err
	}

	return records, nil
}

func (db *BoltdbStore) ShowProjectRunner(_ context.Context, params model.RunnerParams) (*model.Runner, error) {
	record := &model.Runner{}

	if err := db.handle.Select(
		q.And(
			q.Eq("id", params.RunnerID),
			q.Eq("project_id", params.ProjectID),
		),
	).First(record); err != nil {
		if err == storm.ErrNotFound {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	return record, nil
}

func (db *BoltdbStore) DeleteProjectRunner(_ context.Context, params model.RunnerParams) error {
	tx, err := db.handle.Begin(true)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	record := &model.Runner{
		ID:        params.RunnerID,
		ProjectID: &params.ProjectID,
	}

	if err := tx.DeleteStruct(record); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *BoltdbStore) CreateProjectRunner(_ context.Context, record *model.Runner) error {
	tx, err := db.handle.Begin(true)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := tx.Save(record); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (db *BoltdbStore) UpdateProjectRunner(_ context.Context, record *model.Runner) error {
	tx, err := db.handle.Begin(true)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := tx.Save(record); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
