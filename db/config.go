package db

import (
	"encoding/json"
	"strings"

	"github.com/semaphoreui/semaphore/util"
)

func ConvertFlatToNested(flatMap map[string]string) map[string]any {
	nestedMap := make(map[string]any)

	for key, value := range flatMap {
		parts := strings.Split(key, ".")
		currentMap := nestedMap

		for i, part := range parts {
			if i == len(parts)-1 {
				currentMap[part] = value
			} else {
				if _, exists := currentMap[part]; !exists {
					currentMap[part] = make(map[string]any)
				}
				currentMap = currentMap[part].(map[string]any)
			}
		}
	}

	return nestedMap
}

func FillConfigFromDB(store Store) (err error) {

	opts, err := store.GetOptions(RetrieveQueryParams{})

	if err != nil {
		return
	}

	options := ConvertFlatToNested(opts)

	err = util.AssignMapToStruct(options, util.Config)

	if err != nil {
		return
	}

	err = FillAppsFromDB(store)

	return
}

func FillAppsFromDB(store Store) error {
	apps, err := store.GetApps()
	if err != nil {
		return err
	}

	if util.Config.Apps == nil {
		util.Config.Apps = make(map[string]util.App)
	}

	util.Config.AppVersions = make(map[int]util.AppVersionRuntime)

	for _, a := range apps {
		versions, err := store.GetAppVersions(a.ID)
		if err != nil {
			return err
		}

		app := util.App{
			Active:    a.Active,
			Priority:  a.Priority,
			Title:     a.Title,
			Icon:      a.Icon,
			Color:     a.Color,
			DarkColor: a.DarkColor,
		}

		if len(versions) > 0 {
			v := versions[0]
			app.AppPath = v.Path
			if v.Args != nil {
				_ = json.Unmarshal([]byte(*v.Args), &app.AppArgs)
			}
		}

		util.Config.Apps[a.ID] = app

		for _, v := range versions {
			rt := util.AppVersionRuntime{
				AppPath: v.Path,
			}
			if v.Args != nil {
				_ = json.Unmarshal([]byte(*v.Args), &rt.AppArgs)
			}
			util.Config.AppVersions[v.ID] = rt
		}
	}

	return nil
}
