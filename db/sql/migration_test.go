package sql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVersionSQL_MissingErrFileWithIgnoreErrors(t *testing.T) {
	assert.NotPanics(t, func() {
		queries := getVersionSQL("sqlite", "v9999.9.9.err.sql", true)
		assert.Nil(t, queries)
	})
}

func TestGetVersionSQL_MissingErrFileWithoutIgnoreErrorsPanics(t *testing.T) {
	assert.Panics(t, func() {
		getVersionSQL("sqlite", "v9999.9.9.err.sql", false)
	})
}
