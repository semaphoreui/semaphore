package projects

import (
	"fmt"
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
)

// ViewMiddleware ensures a key exists and loads it to the context
func ViewMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		viewID, err := helpers.GetIntParam("view_id", w, r)
		if err != nil {
			return
		}

		view, err := helpers.Store(r).GetView(project.ID, viewID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "view", view)
		next.ServeHTTP(w, r)
	})
}

func GetViewTemplates(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	view := helpers.GetFromContext(r, "view").(db.View)

	templates, err := helpers.Store(r).GetTemplates(project.ID, db.TemplateFilter{ViewID: &view.ID}, helpers.QueryParams(r.URL))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, templates)
}

// GetViews retrieves sorted views from the database
func GetViews(w http.ResponseWriter, r *http.Request) {
	if view := helpers.GetFromContext(r, "view"); view != nil {
		k := view.(db.View)
		helpers.WriteJSON(w, http.StatusOK, k)
		return
	}

	project := helpers.GetFromContext(r, "project").(db.Project)
	views, err := helpers.Store(r).GetViews(project.ID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	// Filter out hidden views from normal views response unless explicitly requested
	showHidden := r.URL.Query().Get("show_hidden") == "true"
	var filteredViews []db.View
	
	for _, view := range views {
		if !view.Hidden || showHidden {
			filteredViews = append(filteredViews, view)
		}
	}

	helpers.WriteJSON(w, http.StatusOK, filteredViews)
}

// AddView adds a new key to the database
func AddView(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	var view db.View

	if !helpers.Bind(w, r, &view) {
		return
	}

	if view.ProjectID != project.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}

	// Set default values for new fields if not provided
	if view.Type == "" {
		view.Type = db.ViewTypeCustom
	}

	if err := view.Validate(); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	newView, err := helpers.Store(r).CreateView(view)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   newView.ProjectID,
		ObjectType:  db.EventView,
		ObjectID:    newView.ID,
		Description: fmt.Sprintf("View %s created", view.Title),
	})

	helpers.WriteJSON(w, http.StatusCreated, newView)
}

func SetViewPositions(w http.ResponseWriter, r *http.Request) {
	var positions map[int]int

	project := helpers.GetFromContext(r, "project").(db.Project)

	if !helpers.Bind(w, r, &positions) {
		return
	}

	err := helpers.Store(r).SetViewPositions(project.ID, positions)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateView updates key in database
// nolint: gocyclo
func UpdateView(w http.ResponseWriter, r *http.Request) {
	var view db.View
	oldView := helpers.GetFromContext(r, "view").(db.View)

	if !helpers.Bind(w, r, &view) {
		return
	}

	if view.ID != oldView.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "View ID in URL and in body must be the same",
		})
		return
	}

	if err := view.Validate(); err != nil {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if err := helpers.Store(r).UpdateView(view); err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   oldView.ProjectID,
		ObjectType:  db.EventView,
		ObjectID:    oldView.ID,
		Description: fmt.Sprintf("View %s updated", view.Title),
	})

	w.WriteHeader(http.StatusNoContent)
}

// RemoveView deletes a view from the database
func RemoveView(w http.ResponseWriter, r *http.Request) {
	view := helpers.GetFromContext(r, "view").(db.View)

	err := helpers.Store(r).DeleteView(view.ProjectID, view.ID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogDelete, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   view.ProjectID,
		ObjectType:  db.EventView,
		ObjectID:    view.ID,
		Description: fmt.Sprintf("View %s deleted", view.Title),
	})

	w.WriteHeader(http.StatusNoContent)
}

// GetAllTabSettings retrieves the All tab position setting for a project
func GetAllTabSettings(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	views, err := helpers.Store(r).GetViews(project.ID)
	
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	
	allTabAtEnd := db.ShouldAllTabBeAtEnd(views)
	
	helpers.WriteJSON(w, http.StatusOK, map[string]bool{
		"allTabAtEnd": allTabAtEnd,
	})
}

// SetAllTabSettings sets the All tab position setting for a project
func SetAllTabSettings(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	
	var settings struct {
		AllTabAtEnd bool `json:"allTabAtEnd"`
	}
	
	if !helpers.Bind(w, r, &settings) {
		return
	}
	
	views, err := helpers.Store(r).GetViews(project.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	
	// Find the All view
	var allView *db.View
	maxPosition := -1
	
	for i, view := range views {
		if view.IsAllView() {
			allView = &views[i]
		} else if !view.Hidden && view.Position > maxPosition {
			maxPosition = view.Position
		}
	}
	
	if allView == nil {
		// Create All view if it doesn't exist
		newPosition := -1
		if settings.AllTabAtEnd && maxPosition >= 0 {
			newPosition = maxPosition + 1
		}
		
		allView = &db.View{
			ProjectID: project.ID,
			Title:     "All",
			Position:  newPosition,
			Hidden:    false,
			Type:      db.ViewTypeAll,
		}
		
		createdView, err := helpers.Store(r).CreateView(*allView)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}
		allView = &createdView
	} else {
		// Update existing All view position
		var newPosition int
		if settings.AllTabAtEnd {
			newPosition = maxPosition + 1
		} else {
			newPosition = -1
		}
		
		if allView.Position != newPosition {
			allView.Position = newPosition
			err = helpers.Store(r).UpdateView(*allView)
			if err != nil {
				helpers.WriteError(w, err)
				return
			}
		}
	}
	
	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   project.ID,
		ObjectType:  db.EventView,
		ObjectID:    allView.ID,
		Description: fmt.Sprintf("All tab position updated to %s", map[bool]string{true: "end", false: "beginning"}[settings.AllTabAtEnd]),
	})
	
	w.WriteHeader(http.StatusNoContent)
}
