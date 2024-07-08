package api

import (
	"errors"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
	"net/http"
)

func validateAppID(str string) error {
	return nil
}

func appMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appID, err := helpers.GetStrParam("app_id", w, r)
		if err != nil {
			helpers.WriteErrorStatus(w, err.Error(), http.StatusBadRequest)
		}

		if err := validateAppID(appID); err != nil {
			helpers.WriteErrorStatus(w, err.Error(), http.StatusBadRequest)
			return
		}

		context.Set(r, "app_id", appID)
		next.ServeHTTP(w, r)
	})
}

func getApps(w http.ResponseWriter, r *http.Request) {

	type app struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Icon      string `json:"icon"`
		Color     string `json:"color"`
		DarkColor string `json:"dark_color"`
		Active    bool   `json:"active"`
	}

	apps := make([]app, 0)

	for k, a := range util.Config.Apps {

		apps = append(apps, app{
			ID:        k,
			Title:     a.Title,
			Icon:      a.Icon,
			Color:     a.Color,
			DarkColor: a.DarkColor,
			Active:    a.Active,
		})
	}

	helpers.WriteJSON(w, http.StatusOK, apps)
}

func getApp(w http.ResponseWriter, r *http.Request) {
	appID := context.Get(r, "app_id").(string)

	app, ok := util.Config.Apps[appID]
	if !ok {
		helpers.WriteErrorStatus(w, "app not found", http.StatusNotFound)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, app)
}

func deleteApp(w http.ResponseWriter, r *http.Request) {
	appID := context.Get(r, "app_id").(string)

	store := helpers.Store(r)

	err := store.DeleteOptions("apps." + appID)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func setApp(w http.ResponseWriter, r *http.Request) {
}

func setAppActive(w http.ResponseWriter, r *http.Request) {

}
