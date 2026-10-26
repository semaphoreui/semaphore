package sql

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type environmentSecretFixture struct {
	store    *SqlDb
	project  db.Project
	env      db.Environment
	otherEnv db.Environment
	owned    db.AccessKey
	shared   db.AccessKey
}

func newEnvironmentSecretFixture(t *testing.T) environmentSecretFixture {
	store := InitConfigCreateTestStore()
	project, err := store.CreateProject(db.Project{Name: "P"})
	require.NoError(t, err)
	env, err := store.CreateEnvironment(db.Environment{ProjectID: project.ID, Name: "E", JSON: "{}"})
	require.NoError(t, err)
	otherEnv, err := store.CreateEnvironment(db.Environment{ProjectID: project.ID, Name: "F", JSON: "{}"})
	require.NoError(t, err)
	secret := "x"
	owned, err := store.CreateAccessKey(db.AccessKey{
		Name: "owned", Type: db.AccessKeyString, ProjectID: &project.ID,
		Owner: db.AccessKeyVariable, EnvironmentID: &env.ID, Secret: &secret,
	})
	require.NoError(t, err)
	shared, err := store.CreateAccessKey(db.AccessKey{
		Name: "shared", Type: db.AccessKeyString, ProjectID: &project.ID, Secret: &secret,
	})
	require.NoError(t, err)
	return environmentSecretFixture{store: store, project: project, env: env, otherEnv: otherEnv, owned: owned, shared: shared}
}

func (f environmentSecretFixture) link(t *testing.T, env db.Environment, typ db.EnvironmentSecretType, name string, key db.AccessKey) db.EnvironmentSecret {
	s, err := f.store.CreateEnvironmentSecret(db.EnvironmentSecret{
		ProjectID: f.project.ID, EnvironmentID: env.ID, Type: typ, Name: name, KeyID: key.ID,
	})
	require.NoError(t, err)
	assert.NotZero(t, s.ID)
	return s
}

func TestEnvironmentSecret_CreateAndRead(t *testing.T) {
	f := newEnvironmentSecretFixture(t)
	f.link(t, f.env, db.EnvironmentSecretVar, "token", f.owned)
	f.link(t, f.env, db.EnvironmentSecretEnv, "DB_PASSWORD", f.shared)

	secrets, err := f.store.GetEnvironmentSecretKeys(f.project.ID, f.env.ID)
	require.NoError(t, err)
	require.Len(t, secrets, 2)

	assert.Equal(t, "DB_PASSWORD", secrets[0].Name)
	assert.Equal(t, "shared", secrets[0].KeyName)
	assert.False(t, secrets[0].Owned, "a Key Store key is linked, not owned")
	assert.Equal(t, f.shared.ID, secrets[0].Key.ID)
	assert.Empty(t, secrets[0].Secret)

	assert.Equal(t, "token", secrets[1].Name)
	assert.Equal(t, "owned", secrets[1].KeyName)
	assert.True(t, secrets[1].Owned)

	other, err := f.store.GetEnvironmentSecretKeys(f.project.ID, f.otherEnv.ID)
	require.NoError(t, err)
	assert.Empty(t, other)

	one, err := f.store.GetEnvironmentSecret(f.project.ID, f.env.ID, secrets[1].ID)
	require.NoError(t, err)
	assert.Equal(t, f.owned.ID, one.KeyID)

	_, err = f.store.GetEnvironmentSecret(f.project.ID, f.otherEnv.ID, secrets[1].ID)
	assert.ErrorIs(t, err, db.ErrNotFound, "a secret is read only through its own group")
}

func TestEnvironmentSecret_OwnedKeyLinkedFromAnotherGroupIsNotOwnedThere(t *testing.T) {
	f := newEnvironmentSecretFixture(t)
	f.link(t, f.otherEnv, db.EnvironmentSecretVar, "token", f.owned)

	secrets, err := f.store.GetEnvironmentSecretKeys(f.project.ID, f.otherEnv.ID)
	require.NoError(t, err)
	require.Len(t, secrets, 1)
	assert.False(t, secrets[0].Owned)
	assert.Equal(t, "owned", secrets[0].KeyName)
}

func TestEnvironmentSecret_DuplicateNameIsValidationError(t *testing.T) {
	f := newEnvironmentSecretFixture(t)
	f.link(t, f.env, db.EnvironmentSecretVar, "token", f.owned)

	_, err := f.store.CreateEnvironmentSecret(db.EnvironmentSecret{
		ProjectID: f.project.ID, EnvironmentID: f.env.ID, Type: db.EnvironmentSecretVar, Name: "token", KeyID: f.shared.ID,
	})
	var validation *common_errors.ValidationError
	require.ErrorAs(t, err, &validation)
	assert.Contains(t, err.Error(), "token")

	// The same name as an env secret, or in another group, is fine.
	f.link(t, f.env, db.EnvironmentSecretEnv, "token", f.shared)
	f.link(t, f.otherEnv, db.EnvironmentSecretVar, "token", f.shared)
}

func TestEnvironmentSecret_UpdateAndDelete(t *testing.T) {
	f := newEnvironmentSecretFixture(t)
	s := f.link(t, f.env, db.EnvironmentSecretVar, "token", f.owned)
	f.link(t, f.env, db.EnvironmentSecretVar, "taken", f.shared)

	s.Name = "renamed"
	s.Type = db.EnvironmentSecretEnv
	s.KeyID = f.shared.ID
	require.NoError(t, f.store.UpdateEnvironmentSecret(s))

	got, err := f.store.GetEnvironmentSecret(f.project.ID, f.env.ID, s.ID)
	require.NoError(t, err)
	assert.Equal(t, "renamed", got.Name)
	assert.Equal(t, db.EnvironmentSecretEnv, got.Type)
	assert.Equal(t, f.shared.ID, got.KeyID)

	s.Name = "taken"
	s.Type = db.EnvironmentSecretVar
	var validation *common_errors.ValidationError
	assert.ErrorAs(t, f.store.UpdateEnvironmentSecret(s), &validation, "renaming onto an existing name is refused")

	s.EnvironmentID = f.otherEnv.ID
	assert.ErrorIs(t, f.store.UpdateEnvironmentSecret(s), db.ErrNotFound, "another group can not update the row")

	assert.ErrorIs(t, f.store.DeleteEnvironmentSecret(f.project.ID, f.otherEnv.ID, s.ID), db.ErrNotFound)
	require.NoError(t, f.store.DeleteEnvironmentSecret(f.project.ID, f.env.ID, s.ID))
	_, err = f.store.GetEnvironmentSecret(f.project.ID, f.env.ID, s.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)

	key, err := f.store.GetAccessKey(f.project.ID, f.shared.ID)
	require.NoError(t, err)
	assert.Equal(t, "shared", key.Name, "deleting the link leaves the key")
}

func TestEnvironmentSecret_KeyReferences(t *testing.T) {
	f := newEnvironmentSecretFixture(t)
	f.link(t, f.env, db.EnvironmentSecretVar, "a", f.shared)
	f.link(t, f.env, db.EnvironmentSecretEnv, "A", f.shared)
	f.link(t, f.otherEnv, db.EnvironmentSecretVar, "b", f.shared)

	n, err := f.store.CountEnvironmentSecretsByKey(f.project.ID, f.shared.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, n)

	refs, err := f.store.GetAccessKeyEnvironmentRefs(f.project.ID, f.shared.ID)
	require.NoError(t, err)
	assert.Equal(t, []db.ObjectReferrer{{ID: f.env.ID, Name: "E"}, {ID: f.otherEnv.ID, Name: "F"}}, refs, "a group is listed once")

	all, err := f.store.GetAccessKeyRefs(f.project.ID, f.shared.ID)
	require.NoError(t, err)
	assert.Equal(t, refs, all.Environments, "the key refs dialog lists the groups")
	assert.Empty(t, all.Templates)

	none, err := f.store.GetAccessKeyRefs(f.project.ID, f.owned.ID)
	require.NoError(t, err)
	assert.Empty(t, none.Environments)

	err = f.store.DeleteAccessKey(f.project.ID, f.shared.ID)
	assert.ErrorIs(t, err, db.ErrInvalidOperation, "the foreign key refuses to delete a key groups still use")

	require.NoError(t, f.store.DeleteEnvironmentSecretsByKey(f.project.ID, f.shared.ID))
	n, err = f.store.CountEnvironmentSecretsByKey(f.project.ID, f.shared.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	require.NoError(t, f.store.DeleteAccessKey(f.project.ID, f.shared.ID))
}

func TestEnvironmentSecret_DeletingTheGroupRemovesItsLinks(t *testing.T) {
	f := newEnvironmentSecretFixture(t)
	f.link(t, f.env, db.EnvironmentSecretVar, "a", f.shared)
	f.link(t, f.env, db.EnvironmentSecretVar, "own", f.owned)

	require.NoError(t, f.store.DeleteEnvironment(f.project.ID, f.env.ID))

	n, err := f.store.CountEnvironmentSecretsByKey(f.project.ID, f.shared.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, n, "links go with the group")
	_, err = f.store.GetAccessKey(f.project.ID, f.owned.ID)
	assert.ErrorIs(t, err, db.ErrNotFound, "the owned key goes with the group")
	_, err = f.store.GetAccessKey(f.project.ID, f.shared.ID)
	assert.NoError(t, err, "the shared key stays")
}

func TestEnvironmentSecret_OwnedKeyLinkedElsewhereBlocksGroupDeleteAtDbLevel(t *testing.T) {
	f := newEnvironmentSecretFixture(t)
	f.link(t, f.otherEnv, db.EnvironmentSecretVar, "borrowed", f.owned)

	// The cascade on access_key.environment_id would delete a key the other group points to; the
	// foreign key without on delete is the safety net — the service refuses earlier with a message.
	err := f.store.DeleteEnvironment(f.project.ID, f.env.ID)
	assert.ErrorIs(t, err, db.ErrInvalidOperation)
	_, err = f.store.GetEnvironment(f.project.ID, f.env.ID)
	assert.NoError(t, err, "nothing was deleted")
}
