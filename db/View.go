package db

import (
	"encoding/json"
	"strings"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

type ViewType string

const (
	ViewTypeAll    ViewType = "all"
	ViewTypeCustom ViewType = ""
)

// ViewFilterName is the only key a view filter accepts: an "all" view shows the
// templates whose name contains its value, ignoring case.
const ViewFilterName = "name"

// maxViewFilterBytes is the width of project__view.filter, which holds the
// filter as JSON. The encoded form is what has to fit, not the name alone:
// json.Marshal writes <, > and & as six bytes each.
const maxViewFilterBytes = 1000

type View struct {
	ID          int                `db:"id" json:"id" backup:"-"`
	ProjectID   int                `db:"project_id" json:"project_id" backup:"-"`
	Title       string             `db:"title" json:"title"`
	Position    int                `db:"position" json:"position"`
	Type        ViewType           `db:"type" json:"type,omitempty"`
	Hidden      bool               `db:"hidden" json:"hidden,omitempty"`
	Filter      *MapStringAnyField `db:"filter" json:"filter,omitempty"`
	SortColumn  *string            `db:"sort_column" json:"sort_column,omitempty"`
	SortReverse bool               `db:"sort_reverse" json:"sort_reverse,omitempty"`
}

func (view *View) Validate() error {
	if view.Title == "" {
		return common_errors.NewValidationError("title can not be empty")
	}

	if view.Filter == nil {
		return nil
	}

	// A key the backend does not apply would be stored and silently do nothing.
	for key, value := range *view.Filter {
		if key != ViewFilterName {
			return common_errors.NewValidationError("unknown view filter: " + key)
		}

		if _, ok := value.(string); !ok {
			return common_errors.NewValidationError("view filter name must be a string")
		}
	}

	encoded, err := json.Marshal(view.Filter)
	if err != nil {
		return err
	}

	if len(encoded) > maxViewFilterBytes {
		return common_errors.NewValidationError("view filter is too long")
	}

	return nil
}

// NameFilter is what an "all" view matches template names against, empty when
// the view has no name filter.
func (view *View) NameFilter() string {
	if view.Filter == nil {
		return ""
	}

	name, _ := (*view.Filter)[ViewFilterName].(string)

	return strings.TrimSpace(name)
}
