package projects

import (
	"fmt"
	"net/http"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"

	"github.com/gorilla/context"
)

// ViewMiddleware ensures a key exists and loads it to the context
func ViewMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := context.Get(r, "project").(db.Project)
		viewID, err := helpers.GetIntParam("view_id", w, r)
		if err != nil {
			return
		}

		view, err := helpers.Store(r).GetView(project.ID, viewID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		context.Set(r, "view", view)
		next.ServeHTTP(w, r)
	})
}

func GetViewTemplates(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
	view := context.Get(r, "view").(db.View)

	templates, err := helpers.Store(r).GetTemplates(project.ID, db.TemplateFilter{ViewID: &view.ID}, helpers.QueryParams(r.URL))

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, templates)
}

// GetViews retrieves sorted keys from the database
func GetViews(w http.ResponseWriter, r *http.Request) {
	if view := context.Get(r, "view"); view != nil {
		k := view.(db.View)
		helpers.WriteJSON(w, http.StatusOK, k)
		return
	}

	project := context.Get(r, "project").(db.Project)
	var views []db.View

	views, err := helpers.Store(r).GetViews(project.ID)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, views)
}

// AddView adds a new key to the database
func AddView(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)
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

	project := context.Get(r, "project").(db.Project)

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
	oldView := context.Get(r, "view").(db.View)

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
	view := context.Get(r, "view").(db.View)

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
