package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/conv"
	"github.com/semaphoreui/semaphore/util"
	"net/http"
	"reflect"
	"sort"
)

func validateAppID(str string) error {
	return nil
}

func appMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appID, err := helpers.GetStrParam("app_id", w, r)
		if err != nil {
			helpers.WriteErrorStatus(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := validateAppID(appID); err != nil {
			helpers.WriteErrorStatus(w, err.Error(), http.StatusBadRequest)
			return
		}

		r = helpers.SetContextValue(r, "app_id", appID)
		next.ServeHTTP(w, r)
	})
}

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

	err := store.DeleteOptions("apps." + appID)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	delete(util.Config.Apps, appID)

	w.WriteHeader(http.StatusNoContent)
}

func setAppOption(store db.Store, appID string, field string, val any) error {
	key := "apps." + appID + "." + field

	if val == nil {
		return store.DeleteOptions(key)
	}

	v := fmt.Sprintf("%v", val)

	if err := store.SetOption(key, v); err != nil {
		return err
	}

	opts := make(map[string]string)
	opts[key] = v

	options := db.ConvertFlatToNested(opts)

	_ = util.AssignMapToStruct(options, util.Config)

	return nil
}

func setApp(w http.ResponseWriter, r *http.Request) {
	appID := helpers.GetFromContext(r, "app_id").(string)

	store := helpers.Store(r)

	var app util.App

	if !helpers.Bind(w, r, &app) {
		return
	}

	options := conv.StructToFlatMap(app)

	for k, v := range options {
		t := reflect.TypeOf(v)

		if v != nil {
			switch t.Kind() {
			case reflect.String:
				if v == "" {
					v = nil
				}
			case reflect.Slice, reflect.Array:
				newV, err := json.Marshal(v)
				if err != nil {
					helpers.WriteErrorStatus(w, err.Error(), http.StatusInternalServerError)
					return
				}
				v = string(newV)
				if v == "[]" {
					v = nil
				}
			default:
			}
		}

		if err := setAppOption(store, appID, k, v); err != nil {
			helpers.WriteErrorStatus(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

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

	if err := setAppOption(store, appID, "active", body.Active); err != nil {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
