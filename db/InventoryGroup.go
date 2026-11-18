package db

import (
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

const inventoryGroupNameMaxLength = 250

// InventoryGroup is a group of hosts of a managed inventory
type InventoryGroup struct {
	ID          int       `db:"id" json:"id" backup:"-"`
	InventoryID int       `db:"inventory_id" json:"inventory_id" backup:"-"`
	ProjectID   int       `db:"project_id" json:"project_id" backup:"-"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Variables   string    `db:"variables" json:"variables"`
	Created     time.Time `db:"created" json:"created" backup:"-"`
	Children    []int     `db:"-" json:"children" backup:"-"`
	HostCount   int       `db:"-" json:"host_count" backup:"-"`

	// VariableCount is the number of top-level variables, set when the group is saved
	VariableCount int `db:"variable_count" json:"variable_count" backup:"-"`
}

// InventoryGroupHost is a membership of a host in a group
type InventoryGroupHost struct {
	GroupID int `db:"group_id" json:"group_id"`
	HostID  int `db:"host_id" json:"host_id"`
}

func (g InventoryGroup) Validate() error {
	if g.Name == "" {
		return common_errors.NewValidationError("group name must not be empty")
	}

	if strings.ContainsFunc(g.Name, unicode.IsSpace) || strings.ContainsAny(g.Name, ",[]:") {
		return common_errors.NewValidationError("group name must not contain spaces, commas, brackets or colons")
	}

	if utf8.RuneCountInString(g.Name) > inventoryGroupNameMaxLength {
		return common_errors.NewValidationError("group name is too long")
	}

	if g.Name == "all" || g.Name == "ungrouped" {
		return common_errors.NewValidationError("group name is reserved")
	}

	_, err := ParseInventoryVariables(g.Variables)
	return err
}

// InventoryGroupDescendants returns every group below id once, without id itself
func InventoryGroupDescendants(children map[int][]int, id int) []int {
	seen := map[int]bool{id: true}
	var result []int
	pending := slices.Clone(children[id])

	for len(pending) > 0 {
		next := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if seen[next] {
			continue
		}
		seen[next] = true
		result = append(result, next)
		pending = append(pending, children[next]...)
	}

	return result
}

// InventoryGroupCycle returns the groups of one cycle in order, or nil when there is none
func InventoryGroupCycle(children map[int][]int) []int {
	type frame struct{ id, next int }

	done := map[int]bool{}
	ids := make([]int, 0, len(children))
	for id := range children {
		ids = append(ids, id)
	}
	slices.Sort(ids)

	for _, start := range ids {
		if done[start] {
			continue
		}

		stack := []frame{{id: start}}
		onPath := map[int]int{start: 0}

		for len(stack) > 0 {
			top := &stack[len(stack)-1]
			if top.next == len(children[top.id]) {
				done[top.id] = true
				delete(onPath, top.id)
				stack = stack[:len(stack)-1]
				continue
			}

			child := children[top.id][top.next]
			top.next++

			if pos, ok := onPath[child]; ok {
				cycle := make([]int, 0, len(stack)-pos)
				for _, f := range stack[pos:] {
					cycle = append(cycle, f.id)
				}
				return cycle
			}
			if !done[child] {
				onPath[child] = len(stack)
				stack = append(stack, frame{id: child})
			}
		}
	}

	return nil
}
