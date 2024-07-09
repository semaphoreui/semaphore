package db

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/util"
	"reflect"
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

func AssignMapToStruct[P *S, S any](m map[string]interface{}, s P) error {
	v := reflect.ValueOf(s).Elem()
	return assignMapToStructRecursive(m, v)
}

func cloneStruct(origValue reflect.Value) reflect.Value {
	// Create a new instance of the same type as the original struct
	cloneValue := reflect.New(origValue.Type()).Elem()

	// Iterate over the fields of the struct
	for i := 0; i < origValue.NumField(); i++ {
		// Get the field value
		fieldValue := origValue.Field(i)
		// Set the field value in the clone
		cloneValue.Field(i).Set(fieldValue)
	}

	// Return the cloned struct
	return cloneValue
}

func assignMapToStructRecursive(m map[string]interface{}, structValue reflect.Value) error {
	structType := structValue.Type()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = field.Name
		} else {
			jsonTag = strings.Split(jsonTag, ",")[0]
		}

		if value, ok := m[jsonTag]; ok {
			fieldValue := structValue.FieldByName(field.Name)
			if fieldValue.CanSet() {

				val := reflect.ValueOf(value)

				switch fieldValue.Kind() {
				case reflect.Struct:

					if val.Kind() != reflect.Map {
						return fmt.Errorf("expected map for nested struct field %s but got %T", field.Name, value)
					}

					mapValue, ok := value.(map[string]interface{})
					if !ok {
						return fmt.Errorf("cannot assign value of type %T to field %s of type %s", value, field.Name, field.Type)
					}
					err := assignMapToStructRecursive(mapValue, fieldValue)
					if err != nil {
						return err
					}
				case reflect.Map:
					if fieldValue.IsNil() {
						mapValue := reflect.MakeMap(fieldValue.Type())
						fieldValue.Set(mapValue)
					}

					// Handle map
					if val.Kind() != reflect.Map {
						return fmt.Errorf("expected map for field %s but got %T", field.Name, value)
					}

					for _, key := range val.MapKeys() {
						mapElemValue := val.MapIndex(key)
						mapElemType := fieldValue.Type().Elem()

						srcVal := fieldValue.MapIndex(key)
						var mapElem reflect.Value
						if srcVal.IsValid() {
							mapElem = cloneStruct(srcVal)
						} else {
							mapElem = reflect.New(mapElemType).Elem()
						}

						if mapElemType.Kind() == reflect.Struct {
							if err := assignMapToStructRecursive(mapElemValue.Interface().(map[string]interface{}), mapElem); err != nil {
								return err
							}
						} else {
							if mapElemValue.Type().ConvertibleTo(mapElemType) {
								mapElem.Set(mapElemValue.Convert(mapElemType))
							} else {
								newVal, converted := util.CastValueToKind(mapElemValue.Interface(), mapElemType.Kind())
								if !converted {
									return fmt.Errorf("cannot assign value of type %s to map element of type %s",
										mapElemValue.Type(), mapElemType)
								}

								mapElem.Set(reflect.ValueOf(newVal))
							}

						}

						fieldValue.SetMapIndex(key, mapElem)
					}

				default:
					// Handle simple types
					if val.Type().ConvertibleTo(fieldValue.Type()) {
						fieldValue.Set(val.Convert(fieldValue.Type()))
					} else {

						newVal, converted := util.CastValueToKind(val.Interface(), fieldValue.Type().Kind())
						if !converted {
							return fmt.Errorf("cannot assign value of type %s to map element of type %s",
								val.Type(), val)
						}

						fieldValue.Set(reflect.ValueOf(newVal))
					}
				}
			}
		}
	}
	return nil
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

	err = AssignMapToStruct(options, util.Config)

	return
}
