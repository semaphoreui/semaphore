package project

import (
	"reflect"
	"strings"
)

// backupTag reads a field's backup tag: the key it is stored under, and
// whether the field is legacy — read from files of older versions but never
// written, so the format sheds it while old files still restore.
func backupTag(field reflect.StructField) (name string, legacy bool) {
	parts := strings.Split(field.Tag.Get("backup"), ",")
	name = parts[0]
	for _, option := range parts[1:] {
		if option == "legacy" {
			legacy = true
		}
	}
	return
}
