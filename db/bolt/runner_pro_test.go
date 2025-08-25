package bolt

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_DeleteRunner_DeletesProjectRunner(t *testing.T) {
	store := CreateTestStore()

	project, err := store.CreateProject(db.Project{})
	assert.NoError(t, err)

	testRunner, err := store.CreateRunner(db.Runner{ProjectID: &project.ID})
	assert.NoError(t, err)

	err = store.DeleteRunner(project.ID, testRunner.ID)
	assert.NoError(t, err)

	_, err = store.GetRunner(project.ID, testRunner.ID)
	assert.ErrorIs(t, err, db.ErrNotFound)
}
