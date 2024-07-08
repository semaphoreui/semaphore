package api

import (
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/util"
	"net/http"
)

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

}

func setApp(w http.ResponseWriter, r *http.Request) {

}
