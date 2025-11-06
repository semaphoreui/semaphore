package bolt

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"testing"
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
