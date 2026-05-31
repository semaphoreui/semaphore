package bolt

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
)

func Test_GetRunnerByToken_ReturnsGlobalRunnerWhenTokenExists(t *testing.T) {
	store := CreateTestStore()

	testRunner, err := store.CreateRunner(db.Runner{})
	assert.NoError(t, err)

	_, err = store.GetRunnerByToken(testRunner.Token)
	assert.NoError(t, err)
}

func Test_GetRunnerByToken_ReturnsRunnerWhenTokenExists(t *testing.T) {
	store := CreateTestStore()

	project, err := store.CreateProject(db.Project{})
	assert.NoError(t, err)

	testRunner, err := store.CreateRunner(db.Runner{ProjectID: &project.ID})
	assert.NoError(t, err)

	_, err = store.GetRunnerByToken(testRunner.Token)
	assert.NoError(t, err)
}

func Test_GetGlobalRunner_ReturnsErrorWhenTryingGetProjectRunner(t *testing.T) {
	store := CreateTestStore()

	project, err := store.CreateProject(db.Project{})
	assert.NoError(t, err)

	testRunner, err := store.CreateRunner(db.Runner{ProjectID: &project.ID})
	assert.NoError(t, err)

	_, err = store.GetGlobalRunner(testRunner.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}

func Test_CreateRunner_UnregisteredHasEmptyToken(t *testing.T) {
	store := CreateTestStore()

	testRunner, err := store.CreateRunner(db.Runner{Unregistered: true})
	assert.NoError(t, err)
	assert.Empty(t, testRunner.Token)

	stored, err := store.GetGlobalRunner(testRunner.ID)
	assert.NoError(t, err)
	assert.Empty(t, stored.Token)
	assert.False(t, stored.IsRegistered())
}

func Test_RegisterRunner_SetsTokenAndActivates(t *testing.T) {
	store := CreateTestStore()

	testRunner, err := store.CreateRunner(db.Runner{Unregistered: true})
	assert.NoError(t, err)

	publicKey := "test-public-key"
	registered, err := store.RegisterRunner(testRunner.ID, &publicKey, true)
	assert.NoError(t, err)
	assert.NotEmpty(t, registered.Token)
	assert.True(t, registered.Active)
	assert.Equal(t, publicKey, *registered.PublicKey)

	stored, err := store.GetGlobalRunner(testRunner.ID)
	assert.NoError(t, err)
	assert.Equal(t, registered.Token, stored.Token)
	assert.True(t, stored.Active)
}

func Test_RegisterRunner_FailsWhenAlreadyRegistered(t *testing.T) {
	store := CreateTestStore()

	testRunner, err := store.CreateRunner(db.Runner{})
	assert.NoError(t, err)
	assert.NotEmpty(t, testRunner.Token)

	_, err = store.RegisterRunner(testRunner.ID, nil, true)
	assert.Error(t, err)
}
