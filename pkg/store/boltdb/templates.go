package boltdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *BoltdbStore) TemplateRefs(ctx context.Context, params model.TemplateParams) (*model.ObjectReferrers, error) {
	return nil, nil
}

func (db *BoltdbStore) ListTemplates(ctx context.Context, params model.TemplateParams) ([]*model.Template, error) {
	return nil, nil
}

func (db *BoltdbStore) ShowTemplate(ctx context.Context, params model.TemplateParams) (*model.Template, error) {
	return nil, nil
}

func (db *BoltdbStore) DeleteTemplate(ctx context.Context, params model.TemplateParams) error {
	return nil
}

func (db *BoltdbStore) CreateTemplate(ctx context.Context, record *model.Template) error {
	return nil
}

func (db *BoltdbStore) UpdateTemplate(ctx context.Context, record *model.Template) error {
	return nil
}
