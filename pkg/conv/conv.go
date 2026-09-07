package conv

import (
	"reflect"
	"strings"
	"unicode/utf8"
)


const commitMessageMaxRunes = 100

// TruncateValidUTF8 sanitizes s so it can be stored in the task.commit_message
// column (VARCHAR(100)): NUL bytes are removed, invalid UTF-8 byte sequences are
// replaced with U+FFFD, and the result is capped at commitMessageMaxRunes runes.
//
// This is not a general-purpose string helper: the hard cap is tied to the
// commit-message column width, so reusing it elsewhere would truncate at 100
// runes unexpectedly.
//
// It guards two failure modes that both make PostgreSQL reject the value with
// `invalid byte sequence for encoding "UTF8"` and abort the transaction:
//   - Byte slicing can split a multi-byte character; slicing by runes cannot.
//   - Text originating from a non-UTF-8 encoding (e.g. latin-1) can already
//     contain invalid bytes even when short, so it is sanitized regardless of
//     length.
func TruncateValidUTF8(s string) string {
	// Decode at most the first commitMessageMaxRunes runes we intend to keep.
	// Converting the whole (possibly very large) message to []rune just to
	// discard all but the leading runes wastes CPU and allocations, since
	// callers no longer truncate before reaching here.
	var b strings.Builder
	// Cap the initial allocation: the kept runes never exceed
	// commitMessageMaxRunes*utf8.UTFMax bytes, and shorter inputs need even less.
	grow := commitMessageMaxRunes * utf8.UTFMax
	if len(s) < grow {
		grow = len(s)
	}
	b.Grow(grow)

	count := 0
	// Ranging over a string decodes runes and yields U+FFFD for invalid UTF-8
	// bytes (matching a []rune conversion), so the result is always storable as
	// UTF-8 without splitting a multi-byte character.
	for _, r := range s {
		// PostgreSQL text/varchar cannot store NUL, even though it is valid UTF-8.
		if r == 0 {
			continue
		}
		b.WriteRune(r)
		count++
		if count == commitMessageMaxRunes {
			break
		}
	}
	return b.String()
}

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
