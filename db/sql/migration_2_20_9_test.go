package sql

import (
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
)

func TestMigration_2_20_9_Charset(t *testing.T) {
	tests := []struct {
		dialect string
		tables  int
	}{
		{"mysql", 3},
		{util.DbDriverSQLite, 0},
		{"postgres", 0},
	}
	for _, tt := range tests {
		t.Run(tt.dialect, func(t *testing.T) {
			queries := getVersionSQL(tt.dialect, "v2.20.9.sql", false)
			assert.Equal(t, tt.tables, strings.Count(strings.Join(queries, "\n"), "default charset=utf8mb4"))
		})
	}
}
