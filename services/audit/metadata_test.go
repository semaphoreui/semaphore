package audit

import (
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
)

func TestNewRoleRef_BoundsSlug(t *testing.T) {
	ref := NewRoleRef(db.Role{ID: 4, Slug: strings.Repeat("s", 5000)})

	assert.Equal(t, 4, ref.ID)
	assert.Len(t, ref.Slug, MaxNameBytes)
}
