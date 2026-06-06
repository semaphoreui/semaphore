package tasks

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
)

func TestFilterRegisteredRunners(t *testing.T) {
	runners := []db.Runner{
		{ID: 1, Token: "auth-token", Active: true},
		{ID: 2, Token: "", Active: true},
		{ID: 3, Token: "other-token", Active: false},
	}

	filtered := filterRegisteredRunners(runners)

	assert.Len(t, filtered, 2)
	assert.Equal(t, 1, filtered[0].ID)
	assert.Equal(t, 3, filtered[1].ID)
}

func TestFilterRegisteredRunners_EmptyWhenNoneRegistered(t *testing.T) {
	runners := []db.Runner{
		{ID: 1, Token: "", Active: true},
		{ID: 2, Token: "", Active: false},
	}

	assert.Empty(t, filterRegisteredRunners(runners))
}
