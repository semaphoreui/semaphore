package db

type View struct {
	ID        int    `db:"id" json:"id" backup:"-"`
	ProjectID int    `db:"project_id" json:"project_id" backup:"-"`
	Title     string `db:"title" json:"title"`
	Position  int    `db:"position" json:"position"`
	
	// Optional fields that may exist in some database schemas (e.g., pro version)
	// These are ignored if not present in the database
	Hidden       *bool   `db:"hidden" json:"hidden,omitempty"`
	Type         *string `db:"type" json:"type,omitempty"`
	Filter       *string `db:"filter" json:"filter,omitempty"`
	SortColumn   *string `db:"sort_column" json:"sort_column,omitempty"`
	SortReverse  *bool   `db:"sort_reverse" json:"sort_reverse,omitempty"`
}

func (view *View) Validate() error {
	if view.Title == "" {
		return &ValidationError{"title can not be empty"}
	}
	return nil
}
