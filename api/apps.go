package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
	"net/http"
	"reflect"
	"sort"
	"strings"
)

func structToFlatMap(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	val := reflect.ValueOf(obj)
	typ := reflect.TypeOf(obj)

	if typ.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return result
	}

	// Iterate over the struct fields
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		jsonTag := fieldType.Tag.Get("json")

		// Use the json tag if it is set, otherwise use the field name
		fieldName := jsonTag
		if fieldName == "" || fieldName == "-" {
			fieldName = fieldType.Name
		} else {
			// Handle the case where the json tag might have options like `json:"name,omitempty"`
			fieldName = strings.Split(fieldName, ",")[0]
		}

		// Check if the field is a struct itself
		if field.Kind() == reflect.Struct {
			// Convert nested struct to map
			nestedMap := structToFlatMap(field.Interface())
			// Add nested map to result with a prefixed key
			for k, v := range nestedMap {
				result[fieldName+"."+k] = v
			}
		} else if (field.Kind() == reflect.Ptr ||
			field.Kind() == reflect.Array ||
			field.Kind() == reflect.Slice ||
			field.Kind() == reflect.Map) && field.IsNil() {
			result[fieldName] = nil
		} else {
			result[fieldName] = field.Interface()
		}
	}

	return result
}

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

	delete(util.Config.Apps, appID)

	w.WriteHeader(http.StatusNoContent)
}

func setAppOption(store db.Store, appID string, field string, val interface{}) error {
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

	_ = db.AssignMapToStruct(options, util.Config)

	return nil
}

func setApp(w http.ResponseWriter, r *http.Request) {
	appID := context.Get(r, "app_id").(string)

	store := helpers.Store(r)

	var app util.App

	if !helpers.Bind(w, r, &app) {
		return
	}

	options := structToFlatMap(app)

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
	appID := context.Get(r, "app_id").(string)

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
