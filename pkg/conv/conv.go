package conv

import (
	"reflect"
	"strings"
)

func ConvertFloatToIntIfPossible(v any) (int64, bool) {

	switch v := v.(type) {
	case float64:
		f := v
		i := int64(f)
		if float64(i) == f {
			return i, true
		}
	case float32:
		f := v
		i := int64(f)
		if float32(i) == f {
			return i, true
		}
	}

	return 0, false
}

func StructToFlatMap(obj any) map[string]any {
	result := make(map[string]any)
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
			nestedMap := StructToFlatMap(field.Interface())
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
