package projects

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
)

func Test_mergeProjectRunnersList(t *testing.T) {
	projectID := 1

	projectRunners := []db.Runner{
		{ID: 1, ProjectID: &projectID, Name: "project-runner", Active: true},
	}

	globalRunners := []db.Runner{
		{ID: 2, Name: "active-tagged", Active: true, Tags: []string{"prod"}},
		{ID: 3, Name: "inactive-tagged", Active: false, Tags: []string{"prod"}},
		{ID: 4, Name: "active-untagged", Active: true, Tags: nil},
	}

	result := mergeProjectRunnersList(projectRunners, globalRunners)

	assert.Len(t, result, 2)
	assert.Equal(t, 1, result[0].ID)
	assert.Equal(t, 2, result[1].ID)
}
