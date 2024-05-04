package gormdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/pkg/store"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func (db *GormdbStore) AccessKeyRefs(ctx context.Context, params model.AccessKeyParams) (*model.ObjectReferrers, error) {
	record := &model.AccessKey{}

	if err := db.handle.WithContext(ctx).Where(
		"id = ?",
		params.AccessKeyID,
	).Where(
		"project_id = ?",
		params.ProjectID,
	).Preload(
		"RepositorySSHKeys",
	).Preload(
		"InventorySSHKeys",
	).Preload(
		"InventoryBecomeKeys",
	).Preload(
		"TemplateVaultKeys",
	).First(
		record,
	).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, store.ErrRecordNotFound
		}

		return nil, err
	}

	result := &model.ObjectReferrers{
		Repositories: make([]model.ObjectReferrer, 0),
		Inventories:  make([]model.ObjectReferrer, 0),
		Templates:    make([]model.ObjectReferrer, 0),
	}

	for _, row := range record.RepositorySSHKeys {
		result.Repositories = append(result.Repositories, model.ObjectReferrer{
			ID:   row.ID,
			Name: row.Name,
		})
	}

	for _, row := range record.InventorySSHKeys {
		result.Inventories = append(result.Inventories, model.ObjectReferrer{
			ID:   row.ID,
			Name: row.Name,
		})
	}

	for _, row := range record.InventoryBecomeKeys {
		result.Inventories = append(result.Inventories, model.ObjectReferrer{
			ID:   row.ID,
			Name: row.Name,
		})
	}

	for _, row := range record.TemplateVaultKeys {
		result.Templates = append(result.Templates, model.ObjectReferrer{
			ID:   row.ID,
			Name: row.Name,
		})
	}

	return result, nil
}

func (db *GormdbStore) ListAccessKeys(ctx context.Context, params model.AccessKeyParams) ([]*model.AccessKey, error) {
	records := make([]*model.AccessKey, 0)

	err := db.sortQuery(
		db.handle.WithContext(ctx),
		model.AccessKeyProps,
		params.Query,
	).Where(
		"project_id = ?",
		params.ProjectID,
	).Find(
		&records,
	).Error

	return records, err
}

func (db *GormdbStore) ShowAccessKey(ctx context.Context, params model.AccessKeyParams) (*model.AccessKey, error) {
	record := &model.AccessKey{}

	err := db.handle.WithContext(ctx).Where(
		"id = ?",
		params.AccessKeyID,
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

func (db *GormdbStore) DeleteAccessKey(ctx context.Context, params model.AccessKeyParams) error {
	tx := db.handle.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := tx.Where(
		"id = ?",
		params.AccessKeyID,
	).Where(
		"project_id = ?",
		params.ProjectID,
	).Delete(
		&model.AccessKey{},
	).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return store.ErrRecordNotFound
		}

		return err
	}

	return tx.Commit().Error
}

func (db *GormdbStore) RekeyAccessKey(ctx context.Context, params model.AccessKeyParams) error {
	for i := 0; ; i++ {
		records := make([]*model.AccessKey, 0)

		if err := db.sortQuery(
			db.handle.WithContext(ctx),
			model.AccessKeyProps,
			model.RetrieveQueryParams{Count: 100, Offset: i * 100},
		).Where(
			"project_id = ?",
			params.ProjectID,
		).Find(
			&records,
		).Error; err != nil {
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

func (db *GormdbStore) CreateAccessKey(ctx context.Context, record *model.AccessKey) error {
	tx := db.handle.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := record.Validate(true); err != nil {
		return err
	}

	if err := record.SerializeSecret(); err != nil {
		return err
	}

	if err := tx.Create(record).Error; err != nil {
		return err
	}

	return tx.Commit().Error
}

func (db *GormdbStore) UpdateAccessKey(ctx context.Context, record *model.AccessKey) error {
	tx := db.handle.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := record.Validate(
		record.OverrideSecret,
	); err != nil {
		return err
	}

	if err := record.SerializeSecret(); err != nil {
		return err
	}

	if err := tx.Save(record).Error; err != nil {
		return err
	}

	return tx.Commit().Error
}
