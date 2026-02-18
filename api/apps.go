package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
)

func getApps(w http.ResponseWriter, r *http.Request) {
	type app struct {
		util.App
		ID string `json:"id"`
	}

	apps := make([]app, 0)

	for k, a := range util.Config.Apps {
		apps = append(apps, app{
			App: a,
			ID:  k,
		})
	}

	sort.Slice(apps, func(i, j int) bool {
		return apps[i].Priority > apps[j].Priority
	})

	helpers.WriteJSON(w, http.StatusOK, apps)
}

func getApp(w http.ResponseWriter, r *http.Request) {
	appID := helpers.GetFromContext(r, "app_id").(string)

	app, ok := util.Config.Apps[appID]
	if !ok {
		helpers.WriteErrorStatus(w, "app not found", http.StatusNotFound)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, app)
}

func deleteApp(w http.ResponseWriter, r *http.Request) {
	appID := helpers.GetFromContext(r, "app_id").(string)
	store := helpers.Store(r)

	err := store.DeleteApp(appID)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	delete(util.Config.Apps, appID)
	w.WriteHeader(http.StatusNoContent)
}

func setApp(w http.ResponseWriter, r *http.Request) {
	appID := helpers.GetFromContext(r, "app_id").(string)
	store := helpers.Store(r)

	var body util.App
	if !helpers.Bind(w, r, &body) {
		return
	}

	existing, err := store.GetApp(appID)

	if errors.Is(err, db.ErrNotFound) {
		_, err = store.CreateApp(db.App{
			ID:        appID,
			Title:     body.Title,
			Icon:      body.Icon,
			Color:     body.Color,
			DarkColor: body.DarkColor,
			Active:    body.Active,
			Priority:  body.Priority,
		})
	} else if err == nil {
		existing.Title = body.Title
		existing.Icon = body.Icon
		existing.Color = body.Color
		existing.DarkColor = body.DarkColor
		existing.Active = body.Active
		existing.Priority = body.Priority
		err = store.UpdateApp(existing)
	}

	if err != nil {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	versions, _ := store.GetAppVersions(appID)
	var argsJSON *string
	if len(body.AppArgs) > 0 {
		b, _ := json.Marshal(body.AppArgs)
		s := string(b)
		argsJSON = &s
	}

	if len(versions) == 0 {
		_, err = store.CreateAppVersion(db.AppVersion{
			AppID:  appID,
			Path:   body.AppPath,
			Args:   argsJSON,
			Active: true,
		})
	} else {
		versions[0].Path = body.AppPath
		versions[0].Args = argsJSON
		err = store.UpdateAppVersion(versions[0])
	}

	if err != nil {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	util.Config.Apps[appID] = body

	w.WriteHeader(http.StatusNoContent)
}

func setAppActive(w http.ResponseWriter, r *http.Request) {
	appID := helpers.GetFromContext(r, "app_id").(string)
	store := helpers.Store(r)

	var body struct {
		Active bool `json:"active"`
	}

	if !helpers.Bind(w, r, &body) {
		helpers.WriteErrorStatus(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	existing, err := store.GetApp(appID)
	if errors.Is(err, db.ErrNotFound) {
		_, err = store.CreateApp(db.App{
			ID:     appID,
			Active: body.Active,
		})
	} else if err == nil {
		existing.Active = body.Active
		err = store.UpdateApp(existing)
	}

	if err != nil {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if a, ok := util.Config.Apps[appID]; ok {
		a.Active = body.Active
		util.Config.Apps[appID] = a
	}

	w.WriteHeader(http.StatusNoContent)
}

func getAppVersions(w http.ResponseWriter, r *http.Request) {
	appID, err := helpers.GetStrParam("app_id", w, r)
	if err != nil {
		helpers.WriteErrorStatus(w, "invalid app", http.StatusBadRequest)
		return
	}

	store := helpers.Store(r)

	versions, err := store.GetAppVersions(appID)
	if err != nil {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, versions)
}

func getAppVersion(w http.ResponseWriter, r *http.Request) {
	appID, err := helpers.GetStrParam("app_id", w, r)
	if err != nil {
		helpers.WriteErrorStatus(w, "invalid app", http.StatusBadRequest)
		return
	}

	versionID, err := helpers.GetIntParam("version_id", w, r)
	if err != nil {
		helpers.WriteErrorStatus(w, "invalid version id", http.StatusBadRequest)
		return
	}

	store := helpers.Store(r)

	version, err := store.GetAppVersion(appID, versionID)
	if errors.Is(err, db.ErrNotFound) {
		helpers.WriteErrorStatus(w, "version not found", http.StatusNotFound)
		return
	}
	if err != nil {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, version)
}

func createAppVersion(w http.ResponseWriter, r *http.Request) {
	appID, err := helpers.GetStrParam("app_id", w, r)
	if err != nil {
		helpers.WriteErrorStatus(w, "invalid app", http.StatusBadRequest)
		return
	}

	store := helpers.Store(r)

	var version db.AppVersion
	if !helpers.Bind(w, r, &version) {
		return
	}

	version.AppID = appID
	created, err := store.CreateAppVersion(version)
	if err != nil {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, created)
}

func updateAppVersion(w http.ResponseWriter, r *http.Request) {
	appID, err := helpers.GetStrParam("app_id", w, r)
	if err != nil {
		helpers.WriteErrorStatus(w, "invalid app", http.StatusBadRequest)
		return
	}

	versionID, err := helpers.GetIntParam("version_id", w, r)
	if err != nil {
		helpers.WriteErrorStatus(w, "invalid version id", http.StatusBadRequest)
		return
	}

	store := helpers.Store(r)

	var version db.AppVersion
	if !helpers.Bind(w, r, &version) {
		return
	}

	version.ID = versionID
	version.AppID = appID

	err = store.UpdateAppVersion(version)
	if err != nil {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func setAppVersionOrder(w http.ResponseWriter, r *http.Request) {
	appID, err := helpers.GetStrParam("app_id", w, r)
	if err != nil {
		helpers.WriteErrorStatus(w, "invalid app", http.StatusBadRequest)
		return
	}

	store := helpers.Store(r)

	var order map[int]int
	if !helpers.Bind(w, r, &order) {
		return
	}

	err = store.SetAppVersionOrder(appID, order)
	if err != nil {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func deleteAppVersion(w http.ResponseWriter, r *http.Request) {
	appID, err := helpers.GetStrParam("app_id", w, r)
	if err != nil {
		helpers.WriteErrorStatus(w, "invalid app", http.StatusBadRequest)
		return
	}

	versionID, err := helpers.GetIntParam("version_id", w, r)
	if err != nil {
		helpers.WriteErrorStatus(w, "invalid version id", http.StatusBadRequest)
		return
	}

	store := helpers.Store(r)

	err = store.DeleteAppVersion(appID, versionID)
	if err != nil {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
