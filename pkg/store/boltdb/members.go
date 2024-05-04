package boltdb

import (
	"context"

	model "github.com/ansible-semaphore/semaphore/db"
)

func (db *BoltdbStore) ListMembers(ctx context.Context, params model.MemberParams) ([]*model.UserWithProjectRole, error) {
	return nil, nil
}

func (db *BoltdbStore) ShowMember(ctx context.Context, params model.MemberParams) (*model.ProjectUser, error) {
	return nil, nil
}

func (db *BoltdbStore) DeleteMember(ctx context.Context, params model.MemberParams) error {
	return nil
}

func (db *BoltdbStore) CreateMember(ctx context.Context, record *model.ProjectUser) error {
	return nil
}

func (db *BoltdbStore) UpdateMember(ctx context.Context, record *model.ProjectUser) error {
	return nil
}
