package db

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInventoryGroupValidate(t *testing.T) {
	tests := []struct {
		name      string
		group     InventoryGroup
		wantError bool
	}{
		{"plain", InventoryGroup{Name: "web"}, false},
		{"hyphen", InventoryGroup{Name: "web-eu"}, false},
		{"upper and underscore", InventoryGroup{Name: "Web_1"}, false},
		{"max length", InventoryGroup{Name: strings.Repeat("a", 250)}, false},
		{"empty", InventoryGroup{}, true},
		{"space", InventoryGroup{Name: "a b"}, true},
		{"comma", InventoryGroup{Name: "a,b"}, true},
		{"brackets", InventoryGroup{Name: "a[1]"}, true},
		{"colon", InventoryGroup{Name: "a:b"}, true},
		{"reserved all", InventoryGroup{Name: "all"}, true},
		{"reserved ungrouped", InventoryGroup{Name: "ungrouped"}, true},
		{"too long", InventoryGroup{Name: strings.Repeat("a", 251)}, true},
		{"variables mapping", InventoryGroup{Name: "web", Variables: "a: 1"}, false},
		{"variables list", InventoryGroup{Name: "web", Variables: "- a"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.group.Validate()

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInventoryGroupDescendants_Diamond(t *testing.T) {
	// 1 web -> 2 app, 3 db -> 4 base
	children := map[int][]int{1: {2, 3}, 2: {4}, 3: {4}}

	assert.ElementsMatch(t, []int{2, 3, 4}, InventoryGroupDescendants(children, 1))
	assert.Empty(t, InventoryGroupDescendants(children, 4))
	assert.Nil(t, InventoryGroupCycle(children))
}

func TestInventoryGroupCycle(t *testing.T) {
	children := map[int][]int{1: {2}, 2: {3}, 3: {1}}

	assert.ElementsMatch(t, []int{1, 2, 3}, InventoryGroupCycle(children))
	assert.ElementsMatch(t, []int{2, 3}, InventoryGroupDescendants(children, 1))
}
