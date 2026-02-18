package db

type ViewType string

const (
	ViewTypeAll    ViewType = "all"
	ViewTypeCustom ViewType = ""
)

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
		return &ValidationError{"title can not be empty"}
	}
	return nil
}
