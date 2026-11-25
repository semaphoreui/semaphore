package sql

import (
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) GetView(projectID int, viewID int) (view db.View, err error) {
	err = d.getObject(projectID, db.ViewProps, viewID, &view)
	return
}

func (d *SqlDb) GetViews(projectID int, types []db.ViewType) (views []db.View, err error) {
	err = d.getObjects(projectID, db.ViewProps, db.RetrieveQueryParams{}, func(builder squirrel.SelectBuilder) squirrel.SelectBuilder {
		if len(types) == 0 {
			return builder
		}

		return builder.Where(squirrel.Eq{"type": types})
	}, &views)
	return
}

// GetViewTemplateCounts costs one index lookup per template, whatever the
// number of tasks: the last task of a template is the first entry of the
// task (template_id, id) index read backwards. Ordering by both columns, not
// by id alone, is what makes MySQL read the index in reverse instead of
// sorting every task of the template. PostgreSQL needs the two-column index
// of migration 2.21.6, MySQL and SQLite keep the primary key in every index.
func (d *SqlDb) GetViewTemplateCounts(projectID int, scope db.ViewTemplateCountScope, viewer *db.Viewer) (counts []db.ViewTemplateCount, err error) {
	counts = make([]db.ViewTemplateCount, 0)

	if scope.Empty() {
		return
	}

	query := "select s.view_id, s.status, count(*) as templates from (" +
		"select pt.view_id as view_id, " +
		"(select t.status from task t where t.template_id = pt.id order by t.template_id desc, t.id desc limit 1) as status " +
		"from project__template pt where pt.project_id = ?"
	args := []any{projectID}

	if !scope.All {
		query += " and pt.view_id in (?" + strings.Repeat(", ?", len(scope.ViewIDs)-1) + ")"
		for _, viewID := range scope.ViewIDs {
			args = append(args, viewID)
		}
	}

	if scope.Name != "" {
		cond, arg := nameContains(scope.Name)
		query += " and " + cond
		args = append(args, arg)
	}

	if filtered(viewer) {
		where, whereArgs := visibleTemplatesWhere(projectID, viewer.UserID, "pt")
		query += " and " + where
		args = append(args, whereArgs...)
	}

	query += ") s group by s.view_id, s.status"

	_, err = d.selectAll(&counts, query, args...)
	return
}

// UpdateView never moves a view: the project is part of the match, not of
// what is written. A nil filter is left out, so the stored one stays as it is
// at the time of the update.
func (d *SqlDb) UpdateView(view db.View) error {
	columns := map[string]any{
		"title":               view.Title,
		"position":            view.Position,
		"`type`":              view.Type,
		"sort_column":         view.SortColumn,
		"sort_reverse":        view.SortReverse,
		"hidden":              view.Hidden,
		"show_template_stats": view.ShowTemplateStats,
	}

	if view.Filter != nil {
		columns["filter"] = view.Filter
	}

	query, args, err := squirrel.Update("project__view").
		SetMap(columns).
		Where("id=? and project_id=?", view.ID, view.ProjectID).
		ToSql()
	if err != nil {
		return err
	}

	_, err = d.exec(query, args...)

	return err
}

func (d *SqlDb) CreateView(view db.View) (newView db.View, err error) {
	insertID, err := d.insert(
		"id",
		"insert into project__view (project_id, title, position, `type`, filter, sort_column, sort_reverse, "+
			"hidden, show_template_stats) values (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		view.ProjectID,
		view.Title,
		view.Position,
		view.Type,
		view.Filter,
		view.SortColumn,
		view.SortReverse,
		view.Hidden,
		view.ShowTemplateStats,
	)

	if err != nil {
		return
	}

	newView = view
	newView.ID = insertID
	return
}

func (d *SqlDb) DeleteView(projectID int, viewID int) error {
	return d.deleteObject(projectID, db.ViewProps, viewID)
}

func (d *SqlDb) SetViewPositions(projectID int, positions map[int]int) error {
	for id, position := range positions {
		_, err := d.exec("update project__view set position=? where project_id=? and id=?",
			position,
			projectID,
			id)
		if err != nil {
			return err
		}
	}
	return nil
}
