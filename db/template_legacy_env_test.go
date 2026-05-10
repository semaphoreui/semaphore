package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplate_ApplyLegacyEnvironmentField(t *testing.T) {
	t.Run("fills from deprecated environment_id when list omitted", func(t *testing.T) {
		tpl := Template{EnvironmentID: 42}
		tpl.ApplyLegacyEnvironmentField()
		assert.Equal(t, []int{42}, tpl.EnvironmentIDs)
	})

	t.Run("does not override explicit environment_ids", func(t *testing.T) {
		tpl := Template{EnvironmentID: 99, EnvironmentIDs: []int{1, 2, 3}}
		tpl.ApplyLegacyEnvironmentField()
		assert.Equal(t, []int{1, 2, 3}, tpl.EnvironmentIDs)
	})

	t.Run("explicit empty slice means clear groups", func(t *testing.T) {
		tpl := Template{EnvironmentID: 7, EnvironmentIDs: []int{}}
		tpl.ApplyLegacyEnvironmentField()
		assert.Equal(t, []int{}, tpl.EnvironmentIDs)
	})

	t.Run("zero environment_id with nil list is unchanged", func(t *testing.T) {
		tpl := Template{}
		tpl.ApplyLegacyEnvironmentField()
		assert.Nil(t, tpl.EnvironmentIDs)
	})
}
