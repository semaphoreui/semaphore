package db

const (
	// Special view ID for storing All tab settings
	AllTabViewID = -1
	AllTabViewTitle = "__all_tab_settings__"
)

type View struct {
	ID        int    `db:"id" json:"id" backup:"-"`
	ProjectID int    `db:"project_id" json:"project_id" backup:"-"`
	Title     string `db:"title" json:"title"`
	Position  int    `db:"position" json:"position"`
}

func (view *View) Validate() error {
	if view.Title == "" {
		return &ValidationError{"title can not be empty"}
	}
	return nil
}

// IsAllTabSettingsView returns true if this view represents All tab settings
func (view *View) IsAllTabSettingsView() bool {
	return view.Title == AllTabViewTitle
}

// ShouldAllTabBeAtEnd determines if All tab should be positioned at the end
// based on the position of the special All tab settings view
func ShouldAllTabBeAtEnd(views []View) bool {
	maxPosition := 0
	allTabPosition := 0
	hasAllTabSetting := false
	
	for _, view := range views {
		if view.IsAllTabSettingsView() {
			allTabPosition = view.Position
			hasAllTabSetting = true
		} else if view.Position > maxPosition {
			maxPosition = view.Position
		}
	}
	
	// If no special All tab setting exists, default to beginning (false)
	if !hasAllTabSetting {
		return false
	}
	
	// If All tab position is greater than all other views, it should be at end
	return allTabPosition > maxPosition
}
