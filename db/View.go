package db

const (
	ViewTypeCustom = "custom"
	ViewTypeAll    = "all"
	ViewTypeFailed = "failed"
)

type View struct {
	ID        int    `db:"id" json:"id" backup:"-"`
	ProjectID int    `db:"project_id" json:"project_id" backup:"-"`
	Title     string `db:"title" json:"title"`
	Position  int    `db:"position" json:"position"`
	Hidden    bool   `db:"hidden" json:"hidden"`
	Type      string `db:"type" json:"type"`
}

func (view *View) Validate() error {
	if view.Title == "" {
		return &ValidationError{"title can not be empty"}
	}
	
	// Validate type field
	if view.Type != ViewTypeCustom && view.Type != ViewTypeAll && view.Type != ViewTypeFailed {
		return &ValidationError{"type must be one of: custom, all, failed"}
	}
	
	return nil
}

// IsAllView returns true if this view is the "All" view type
func (view *View) IsAllView() bool {
	return view.Type == ViewTypeAll
}

// IsFailedView returns true if this view is the "Failed" view type  
func (view *View) IsFailedView() bool {
	return view.Type == ViewTypeFailed
}

// IsCustomView returns true if this view is a custom user-created view
func (view *View) IsCustomView() bool {
	return view.Type == ViewTypeCustom
}

// ShouldAllTabBeAtEnd determines if All tab should be positioned at the end
// based on the position of the All view relative to other visible views
func ShouldAllTabBeAtEnd(views []View) bool {
	var allView *View
	maxCustomPosition := -1
	
	for i, view := range views {
		if view.IsAllView() && !view.Hidden {
			allView = &views[i]
		} else if !view.Hidden && view.Position > maxCustomPosition {
			maxCustomPosition = view.Position
		}
	}
	
	// If no All view exists or it's hidden, default to beginning
	if allView == nil {
		return false
	}
	
	// If All view position is greater than all other visible views, it should be at end
	return allView.Position > maxCustomPosition
}
