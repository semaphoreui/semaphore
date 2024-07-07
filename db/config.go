package db

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/util"
	"reflect"
	"strings"
)

func assignMapToStruct[P *S, S any](m map[string]interface{}, s P) error {
	v := reflect.ValueOf(s).Elem()
	return assignMapToStructRecursive(m, v)
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

		value, ok := m[jsonTag]
		if !ok {
			continue
		}

		fieldValue := structValue.FieldByName(field.Name)
		if !fieldValue.CanSet() {
			continue
		}

		val := reflect.ValueOf(value)
		switch fieldValue.Kind() {
		case reflect.Struct:
			// Handle nested struct
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

			if val.Kind() != reflect.Map {
				return fmt.Errorf("expected map for field %s but got %T", field.Name, value)
			}

			mapValue := reflect.MakeMap(fieldValue.Type())
			for _, key := range val.MapKeys() {
				mapElemValue := val.MapIndex(key)
				mapValue.SetMapIndex(key, mapElemValue)
			}

			fieldValue.Set(mapValue)
		default:
			// Handle simple types
			if val.Type().ConvertibleTo(fieldValue.Type()) {
				fieldValue.Set(val.Convert(fieldValue.Type()))
			} else {
				return fmt.Errorf("cannot assign value of type %s to field %s of type %s",
					val.Type(), field.Name, fieldValue.Type())
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

	err = assignMapToStruct(options, util.Config)

	return
}
