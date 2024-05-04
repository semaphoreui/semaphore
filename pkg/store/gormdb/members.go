package gormdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *GormdbStore) ListMembers(ctx context.Context, params model.MemberParams) ([]*model.UserWithProjectRole, error) {
	return nil, nil
}

func (db *GormdbStore) ShowMember(ctx context.Context, params model.MemberParams) (*model.ProjectUser, error) {
	return nil, nil
}

func (db *GormdbStore) DeleteMember(ctx context.Context, params model.MemberParams) error {
	return nil
}

func (db *GormdbStore) CreateMember(ctx context.Context, record *model.ProjectUser) error {
	return nil
}

func (db *GormdbStore) UpdateMember(ctx context.Context, record *model.ProjectUser) error {
	return nil
}
