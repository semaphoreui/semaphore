package boltdb

import (
	"context"
	"errors"

	model "github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/pkg/store"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
)

func (db *BoltdbStore) AccessKeyRefs(ctx context.Context, params model.AccessKeyParams) (*model.ObjectReferrers, error) {
	if _, err := db.ShowAccessKey(ctx, params); err != nil {
		return nil, err
	}

	result := &model.ObjectReferrers{
		Repositories: make([]model.ObjectReferrer, 0),
		Inventories:  make([]model.ObjectReferrer, 0),
		Templates:    make([]model.ObjectReferrer, 0),
	}

	{
		rows := make([]*model.Repository, 0)

		if err := db.handle.Select(
			q.Eq("project_id", params.ProjectID),
			q.Eq("ssh_key_id", params.AccessKeyID),
		).Find(
			&rows,
		); err != nil && !errors.Is(err, storm.ErrNotFound) {
			return nil, err
		}

		for _, row := range rows {
			result.Repositories = append(result.Repositories, model.ObjectReferrer{
				ID:   row.ID,
				Name: row.Name,
			})
		}
	}

	{
		rows := make([]*model.Inventory, 0)

		if err := db.handle.Select(
			q.Eq("project_id", params.ProjectID),
			q.Eq("ssh_key_id", params.AccessKeyID),
		).Find(
			&rows,
		); err != nil && !errors.Is(err, storm.ErrNotFound) {
			return nil, err
		}

		for _, row := range rows {
			result.Inventories = append(result.Inventories, model.ObjectReferrer{
				ID:   row.ID,
				Name: row.Name,
			})
		}
	}

	{
		rows := make([]*model.Inventory, 0)

		if err := db.handle.Select(
			q.Eq("project_id", params.ProjectID),
			q.Eq("become_key_id", params.AccessKeyID),
		).Find(
			&rows,
		); err != nil && !errors.Is(err, storm.ErrNotFound) {
			return nil, err
		}

		for _, row := range rows {
			result.Inventories = append(result.Inventories, model.ObjectReferrer{
				ID:   row.ID,
				Name: row.Name,
			})
		}
	}

	{
		rows := make([]*model.Inventory, 0)

		if err := db.handle.Select(
			q.Eq("project_id", params.ProjectID),
			q.Eq("vault_key_id", params.AccessKeyID),
		).Find(
			&rows,
		); err != nil && !errors.Is(err, storm.ErrNotFound) {
			return nil, err
		}

		for _, row := range rows {
			result.Templates = append(result.Templates, model.ObjectReferrer{
				ID:   row.ID,
				Name: row.Name,
			})
		}
	}

	return result, nil
}

func (db *BoltdbStore) ListAccessKeys(ctx context.Context, params model.AccessKeyParams) ([]*model.AccessKey, error) {
	records := make([]*model.AccessKey, 0)

	if err := db.sortQuery(
		db.handle.Select(
			q.Eq("project_id", params.ProjectID),
		),
		model.AccessKeyProps,
		params.Query,
	).Find(
		&records,
	); err != nil && !errors.Is(err, storm.ErrNotFound) {
		return nil, err
	}

	return records, nil
}

func (db *BoltdbStore) ShowAccessKey(ctx context.Context, params model.AccessKeyParams) (*model.AccessKey, error) {
	record := &model.AccessKey{}

	if err := db.handle.Select(
		q.And(
			q.Eq("id", params.AccessKeyID),
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

func (db *BoltdbStore) DeleteAccessKey(ctx context.Context, params model.AccessKeyParams) error {
	tx, err := db.handle.Begin(true)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	record := &model.AccessKey{
		ID:        params.AccessKeyID,
		ProjectID: &params.ProjectID,
	}

	if err := tx.DeleteStruct(record); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *BoltdbStore) RekeyAccessKey(ctx context.Context, params model.AccessKeyParams) error {
	for i := 0; ; i++ {
		records := make([]*model.AccessKey, 0)

		if err := db.sortQuery(
			db.handle.Select(),
			model.AccessKeyProps,
			model.RetrieveQueryParams{Count: 100, Offset: i * 100},
		).Find(
			&records,
		); err != nil && !errors.Is(err, storm.ErrNotFound) {
			return err
		}

		if len(records) == 0 {
			break
		}

		for _, record := range records {
			if err := record.DeserializeSecret2(
				params.OldKey,
			); err != nil {
				return err
			}

			record.OverrideSecret = true

			if err := db.UpdateAccessKey(
				ctx,
				record,
			); err != nil && !errors.Is(err, store.ErrRecordNotFound) {
				return err
			}
		}
	}

	return nil
}

func (db *BoltdbStore) CreateAccessKey(ctx context.Context, record *model.AccessKey) error {
	tx, err := db.handle.Begin(true)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := record.Validate(true); err != nil {
		return err
	}

	if err := record.SerializeSecret(); err != nil {
		return err
	}

	if err := tx.Save(record); err != nil {
		return err
	}

	return tx.Commit()
}

func (db *BoltdbStore) UpdateAccessKey(ctx context.Context, record *model.AccessKey) error {
	tx, err := db.handle.Begin(true)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := record.Validate(
		record.OverrideSecret,
	); err != nil {
		return err
	}

	if err := record.SerializeSecret(); err != nil {
		return err
	}

	if err := tx.Save(record); err != nil {
		return err
	}

	return tx.Commit()
}
