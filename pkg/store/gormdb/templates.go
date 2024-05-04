package gormdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *GormdbStore) TemplateRefs(ctx context.Context, params model.TemplateParams) (*model.ObjectReferrers, error) {
	return nil, nil
}

func (db *GormdbStore) ListTemplates(ctx context.Context, params model.TemplateParams) ([]*model.Template, error) {
	return nil, nil
}

func (db *GormdbStore) ShowTemplate(ctx context.Context, params model.TemplateParams) (*model.Template, error) {
	return nil, nil
}

func (db *GormdbStore) DeleteTemplate(ctx context.Context, params model.TemplateParams) error {
	return nil
}

func (db *GormdbStore) CreateTemplate(ctx context.Context, record *model.Template) error {
	return nil
}

func (db *GormdbStore) UpdateTemplate(ctx context.Context, record *model.Template) error {
	return nil
}
