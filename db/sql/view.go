package sql

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) GetView(projectID int, viewID int) (view db.View, err error) {
	err = d.getObject(projectID, db.ViewProps, viewID, &view)
	return
}

func (d *SqlDb) GetViews(projectID int) (views []db.View, err error) {
	err = d.getObjects(projectID, db.ViewProps, db.RetrieveQueryParams{}, nil, &views)
	return
}

// UpdateView never moves a view: the project is part of the match, not of
// what is written. A nil filter is left out, so the stored one stays as it is
// at the time of the update.
func (d *SqlDb) UpdateView(view db.View) error {
	columns := map[string]any{
		"title":        view.Title,
		"position":     view.Position,
		"`type`":       view.Type,
		"sort_column":  view.SortColumn,
		"sort_reverse": view.SortReverse,
		"hidden":       view.Hidden,
	}

	if view.Filter != nil {
		columns["filter"] = view.Filter
	}

	query, args, err := sq.Update("project__view").
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
			"hidden) values (?, ?, ?, ?, ?, ?, ?, ?)",
		view.ProjectID,
		view.Title,
		view.Position,
		view.Type,
		view.Filter,
		view.SortColumn,
		view.SortReverse,
		view.Hidden,
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
