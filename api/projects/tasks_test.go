package projects

import (
	"net/url"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
)

func TestParseTasksPageParams(t *testing.T) {
	tests := []struct {
		name             string
		query            string
		expectedPageSize int
		expectedCount    int // params.Count == pageSize + 1
		expectedBeforeID int
	}{
		{"defaults", "", maxTasksPageSize, maxTasksPageSize + 1, 0},
		{"count and before", "count=20&before=100", 20, 21, 100},
		{"legacy limit", "limit=50", 50, 51, 0},
		{"count overrides limit", "count=10&limit=50", 10, 11, 0},
		{"page size capped at max", "count=10000", maxTasksPageSize, maxTasksPageSize + 1, 0},
		{"negative count ignored", "count=-5", maxTasksPageSize, maxTasksPageSize + 1, 0},
		{"zero count ignored", "count=0", maxTasksPageSize, maxTasksPageSize + 1, 0},
		{"invalid count ignored", "count=abc", maxTasksPageSize, maxTasksPageSize + 1, 0},
		{"negative before ignored", "count=20&before=-1", 20, 21, 0},
		{"invalid before ignored", "count=20&before=xyz", 20, 21, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := url.ParseQuery(tt.query)
			assert.NoError(t, err)

			params, pageSize := parseTasksPageParams(query, db.RetrieveQueryParams{})

			assert.Equal(t, tt.expectedPageSize, pageSize)
			assert.Equal(t, tt.expectedCount, params.Count)
			assert.Equal(t, tt.expectedBeforeID, params.BeforeID)
		})
	}
}

func TestParseTasksPageParams_PreservesBase(t *testing.T) {
	base := db.RetrieveQueryParams{SortBy: "id", SortInverted: true}

	params, pageSize := parseTasksPageParams(url.Values{}, base)

	assert.Equal(t, "id", params.SortBy)
	assert.True(t, params.SortInverted)
	assert.Equal(t, maxTasksPageSize, pageSize)
	assert.Equal(t, maxTasksPageSize+1, params.Count)
}
