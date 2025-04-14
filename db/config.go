package db

import (
	"github.com/semaphoreui/semaphore/util"
	"strings"
)

func ConvertFlatToNested(flatMap map[string]string) map[string]interface{} {
	nestedMap := make(map[string]interface{})

	for key, value := range flatMap {
		parts := strings.Split(key, ".")
		currentMap := nestedMap

		for i, part := range parts {
			if i == len(parts)-1 {
				currentMap[part] = value
			} else {
				if _, exists := currentMap[part]; !exists {
					currentMap[part] = make(map[string]interface{})
				}
				currentMap = currentMap[part].(map[string]interface{})
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

	if options["apps"] == nil {
		options["apps"] = make(map[string]interface{})
	}

	err = util.AssignMapToStruct(options, util.Config)

	return
}
