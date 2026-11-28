package db

import (
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
)

type ViewType string

const (
	ViewTypeAll    ViewType = "all"
	ViewTypeCustom ViewType = ""
)

type View struct {
	ID        int      `db:"id" json:"id" backup:"-"`
	ProjectID int      `db:"project_id" json:"project_id" backup:"-"`
	Title     string   `db:"title" json:"title"`
	Position  int      `db:"position" json:"position"`
	Type      ViewType `db:"type" json:"type,omitempty"`
	Hidden    bool     `db:"hidden" json:"hidden,omitempty"`
	// ShowTemplateStats makes the tab of the view show the counts of its
	// templates by the status of their last task. Off by default: the counts
	// cost a query, so they are computed only for the views that ask for them.
	ShowTemplateStats bool               `db:"show_template_stats" json:"show_template_stats,omitempty"`
	Filter            *MapStringAnyField `db:"filter" json:"filter,omitempty"`
	SortColumn        *string            `db:"sort_column" json:"sort_column,omitempty"`
	SortReverse       bool               `db:"sort_reverse" json:"sort_reverse,omitempty"`
}

func (view *View) Validate() error {
	if view.Title == "" {
		return common_errors.NewValidationError("title can not be empty")
	}
	return nil
}

// ViewTemplateStats counts the templates of a view by the status of their last
// task. Every template is in exactly one of the counts.
type ViewTemplateStats struct {
	Success int `json:"success"`
	Failed  int `json:"failed"`
	// Stopped also counts the templates being stopped and the ones never launched.
	Stopped int `json:"stopped"`
	// Confirming counts the templates whose last task waits for a confirmation.
	Confirming int `json:"confirming"`
	// Running counts the templates whose last task is queued or executed:
	// waiting, starting, running, or going on after a confirmation was answered.
	Running int `json:"running"`
}

// add puts the templates whose last task has the status into their count; a
// nil status is a template that was never launched.
func (s *ViewTemplateStats) add(status *task_logger.TaskStatus, templates int) {
	if status == nil {
		s.Stopped += templates
		return
	}

	switch *status {
	case task_logger.TaskSuccessStatus:
		s.Success += templates
	case task_logger.TaskFailStatus:
		s.Failed += templates
	case task_logger.TaskStoppedStatus, task_logger.TaskStoppingStatus:
		s.Stopped += templates
	case task_logger.TaskWaitingConfirmation:
		s.Confirming += templates
	case task_logger.TaskWaitingStatus, task_logger.TaskStartingStatus, task_logger.TaskRunningStatus,
		task_logger.TaskConfirmed, task_logger.TaskRejected:
		s.Running += templates
	}
}

// ViewTemplateCount is the number of templates assigned to a view whose last
// task has the status.
type ViewTemplateCount struct {
	// ViewID is nil for the templates assigned to no view.
	ViewID *int `db:"view_id"`
	// Status is nil for the templates that were never launched.
	Status    *task_logger.TaskStatus `db:"status"`
	Templates int                     `db:"templates"`
}

// ViewTemplateCountScope names the templates that have to be counted.
type ViewTemplateCountScope struct {
	// All counts every template of the project.
	All bool
	// ViewIDs counts the templates assigned to these views, unless All is set.
	ViewIDs []int
}

// Empty is true when nothing has to be counted.
func (s ViewTemplateCountScope) Empty() bool {
	return !s.All && len(s.ViewIDs) == 0
}

// ViewTemplateCountScopeOf is the scope the views need: the templates of the
// Custom views that show the counts, or every template when an "All" view
// shows them. Views that do not show the counts add nothing to it.
func ViewTemplateCountScopeOf(views []View) (scope ViewTemplateCountScope) {
	for _, view := range views {
		if !view.ShowTemplateStats {
			continue
		}

		if view.Type == ViewTypeAll {
			scope.All = true
		} else {
			scope.ViewIDs = append(scope.ViewIDs, view.ID)
		}
	}

	return
}

// ViewWithTemplateStats is a view with the counts of the templates it lists.
type ViewWithTemplateStats struct {
	View
	// TemplateStats is nil for a view that does not show the counts.
	TemplateStats *ViewTemplateStats `json:"template_stats,omitempty"`
}

// ViewsWithTemplateStats attaches the template counts to the views that show
// them, the way the template list fills a view: a Custom view holds the
// templates assigned to it, the "All" view every template of the project.
func ViewsWithTemplateStats(views []View, counts []ViewTemplateCount) []ViewWithTemplateStats {
	res := make([]ViewWithTemplateStats, 0, len(views))

	for _, view := range views {
		item := ViewWithTemplateStats{View: view}

		if view.ShowTemplateStats {
			item.TemplateStats = &ViewTemplateStats{}

			for _, count := range counts {
				if view.Type == ViewTypeAll || (count.ViewID != nil && *count.ViewID == view.ID) {
					item.TemplateStats.add(count.Status, count.Templates)
				}
			}
		}

		res = append(res, item)
	}

	return res
}
