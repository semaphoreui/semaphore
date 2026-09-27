package export

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapAlertIDs(t *testing.T) {
	mapper := NewKeyMapper()
	require.NoError(t, mapper.mapKeys(Alert, "1", "10", "100"))
	require.NoError(t, mapper.mapKeys(Alert, "1", "11", "101"))

	mapped, err := mapAlertIDs(mapper, "1", []int{10, 11})
	require.NoError(t, err)
	assert.Equal(t, []int{100, 101}, mapped)

	empty, err := mapAlertIDs(mapper, "1", []int{})
	require.NoError(t, err)
	assert.Empty(t, empty)

	var nilIDs []int
	got, err := mapAlertIDs(mapper, "1", nilIDs)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestMapAlertIDs_Missing(t *testing.T) {
	mapper := NewKeyMapper()

	_, err := mapAlertIDs(mapper, "1", []int{10})
	require.Error(t, err)
}

func TestMapAlertIDsOpt_SkipsMissing(t *testing.T) {
	mapper := NewKeyMapper()
	require.NoError(t, mapper.mapKeys(Alert, "1", "10", "100"))

	mapped, err := mapAlertIDsOpt(mapper, "1", []int{10, 99}, true)
	require.NoError(t, err)
	assert.Equal(t, []int{100}, mapped)
}
