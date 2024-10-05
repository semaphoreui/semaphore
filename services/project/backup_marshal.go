package project

import (
	"fmt"
	"reflect"
)

func marshalValue(v reflect.Value) (interface{}, error) {
	// Handle pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, nil
		}
		return marshalValue(v.Elem())
	}

	// Handle structs
	if v.Kind() == reflect.Struct {
		typeOfV := v.Type()
		result := make(map[string]interface{})

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			fieldType := typeOfV.Field(i)

			// Handle anonymous fields (embedded structs)
			if fieldType.Anonymous {
				embeddedValue, err := marshalValue(fieldValue)
				if err != nil {
					return nil, err
				}
				if embeddedMap, ok := embeddedValue.(map[string]interface{}); ok {
					// Merge embedded struct fields into parent result map
					for k, v := range embeddedMap {
						result[k] = v
					}
				}
				continue
			}

			tag := fieldType.Tag.Get("backup")

			// Check if the field should be backed up
			if tag == "-" {
				continue // Skip fields with backup:"-"
			} else if tag == "" {
				// Get the field name from the "db" tag
				tag = fieldType.Tag.Get("db")
				if tag == "" || tag == "-" {
					continue // Skip if "db" tag is empty or "-"
				}
			}

			// Recursively process the field value
			value, err := marshalValue(fieldValue)
			if err != nil {
				return nil, err
			}

			result[tag] = value
		}
		return result, nil
	}

	// Handle slices and arrays
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		if v.IsNil() {
			return nil, nil
		}
		var result []interface{}
		for i := 0; i < v.Len(); i++ {
			elemValue, err := marshalValue(v.Index(i))
			if err != nil {
				return nil, err
			}
			result = append(result, elemValue)
		}
		return result, nil
	}

	// Handle maps
	if v.Kind() == reflect.Map {
		if v.IsNil() {
			return nil, nil
		}
		result := make(map[string]interface{})
		for _, key := range v.MapKeys() {
			// Assuming the key is a string
			mapKey := fmt.Sprintf("%v", key.Interface())
			mapValue, err := marshalValue(v.MapIndex(key))
			if err != nil {
				return nil, err
			}
			result[mapKey] = mapValue
		}
		return result, nil
	}

	// Handle other types (int, string, etc.)
	return v.Interface(), nil
}

func setBasicType(data interface{}, v reflect.Value) error {
	if !v.CanSet() {
		return fmt.Errorf("cannot set value of type %v", v.Type())
	}

	switch v.Kind() {
	case reflect.Bool:
		b, ok := data.(bool)
		if !ok {
			return fmt.Errorf("expected bool for field, got %T", data)
		}
		v.SetBool(b)
	case reflect.String:
		s, ok := data.(string)
		if !ok {
			return fmt.Errorf("expected string for field, got %T", data)
		}
		v.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, ok := toFloat64(data)
		if !ok {
			return fmt.Errorf("expected number for field, got %T", data)
		}
		v.SetInt(int64(n))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, ok := toFloat64(data)
		if !ok {
			return fmt.Errorf("expected number for field, got %T", data)
		}
		v.SetUint(uint64(n))
	case reflect.Float32, reflect.Float64:
		n, ok := toFloat64(data)
		if !ok {
			return fmt.Errorf("expected number for field, got %T", data)
		}
		v.SetFloat(n)
	default:
		return fmt.Errorf("unsupported kind %v", v.Kind())
	}
	return nil
}

func toFloat64(data interface{}) (float64, bool) {
	switch n := data.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	case int16:
		return float64(n), true
	case int8:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint64:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint8:
		return float64(n), true
	default:
		return 0, false
	}
}

func unmarshalValueWithBackupTags(data interface{}, v reflect.Value) error {
	// Handle pointers
	if v.Kind() == reflect.Ptr {
		// Initialize pointer if it's nil
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return unmarshalValueWithBackupTags(data, v.Elem())
	}

	// Handle structs
	if v.Kind() == reflect.Struct {
		// Data should be a map
		m, ok := data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected object for struct, got %T", data)
		}
		return unmarshalStructWithBackupTags(m, v)
	}

	// Handle slices and arrays
	if v.Kind() == reflect.Slice {
		dataSlice, ok := data.([]interface{})
		if !ok {
			return fmt.Errorf("expected array for slice, got %T", data)
		}
		slice := reflect.MakeSlice(v.Type(), len(dataSlice), len(dataSlice))
		for i := 0; i < len(dataSlice); i++ {
			elem := slice.Index(i)
			if err := unmarshalValueWithBackupTags(dataSlice[i], elem); err != nil {
				return err
			}
		}
		v.Set(slice)
		return nil
	}

	// Handle maps
	if v.Kind() == reflect.Map {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected object for map, got %T", data)
		}
		mapType := v.Type()
		mapValue := reflect.MakeMap(mapType)
		for key, value := range dataMap {
			keyVal := reflect.ValueOf(key).Convert(mapType.Key())
			valVal := reflect.New(mapType.Elem()).Elem()
			if err := unmarshalValueWithBackupTags(value, valVal); err != nil {
				return err
			}
			mapValue.SetMapIndex(keyVal, valVal)
		}
		v.Set(mapValue)
		return nil
	}

	// Handle basic types
	return setBasicType(data, v)
}

func unmarshalStructWithBackupTags(data map[string]interface{}, v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldType := t.Field(i)
		fieldValue := v.Field(i)

		// Handle anonymous fields (embedded structs)
		if fieldType.Anonymous {
			// Pass the entire data map to the embedded struct
			if err := unmarshalStructWithBackupTags(data, fieldValue); err != nil {
				return err
			}
			continue
		}

		// Skip fields with backup:"-"
		if backupTag := fieldType.Tag.Get("backup"); backupTag == "-" {
			continue
		}

		// Determine the JSON key to use
		var jsonKey string
		backupTag := fieldType.Tag.Get("backup")
		if backupTag != "" {
			jsonKey = backupTag
		} else {
			dbTag := fieldType.Tag.Get("db")
			if dbTag != "" {
				jsonKey = dbTag
			} else {
				continue // Skip if no backup or db tag
			}
		}

		// Check if the key exists in the data
		if value, ok := data[jsonKey]; ok {
			if !fieldValue.CanSet() {
				continue // Skip fields that cannot be set
			}
			if value == nil {
				continue
			}
			if err := unmarshalValueWithBackupTags(value, fieldValue); err != nil {
				return err
			}
		}
	}

	return nil
}
