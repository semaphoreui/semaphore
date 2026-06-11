package projects

import (
	"net/url"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
)

func TestParseTasksPageParams(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedCount  int
		expectedOffset int
	}{
		{"defaults", "", maxTasksPageSize, 0},
		{"count and offset", "count=20&offset=40", 20, 40},
		{"legacy limit", "limit=50", 50, 0},
		{"count overrides limit", "count=10&limit=50", 10, 0},
		{"count capped at max", "count=10000", maxTasksPageSize, 0},
		{"negative count ignored", "count=-5", maxTasksPageSize, 0},
		{"zero count ignored", "count=0", maxTasksPageSize, 0},
		{"invalid count ignored", "count=abc", maxTasksPageSize, 0},
		{"negative offset ignored", "count=20&offset=-1", 20, 0},
		{"invalid offset ignored", "count=20&offset=xyz", 20, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := url.ParseQuery(tt.query)
			assert.NoError(t, err)

			params := parseTasksPageParams(query, db.RetrieveQueryParams{})

			assert.Equal(t, tt.expectedCount, params.Count)
			assert.Equal(t, tt.expectedOffset, params.Offset)
		})
	}
}

func TestParseTasksPageParams_PreservesBase(t *testing.T) {
	base := db.RetrieveQueryParams{SortBy: "id", SortInverted: true}

	params := parseTasksPageParams(url.Values{}, base)

	assert.Equal(t, "id", params.SortBy)
	assert.True(t, params.SortInverted)
	assert.Equal(t, maxTasksPageSize, params.Count)
}
