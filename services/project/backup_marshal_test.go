package project

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_MarshalValue_NilPointer_ReturnsNil(t *testing.T) {
	var ptr *int
	result, err := marshalValue(reflect.ValueOf(ptr))
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func Test_MarshalValue_StructWithFields_ReturnsMap(t *testing.T) {
	type TestStruct struct {
		Field1 string `backup:"field1"`
		Field2 int    `backup:"field2"`
	}
	testStruct := TestStruct{Field1: "value1", Field2: 42}
	result, err := marshalValue(reflect.ValueOf(testStruct))
	assert.NoError(t, err)
	expected := map[string]any{"field1": "value1", "field2": 42}
	assert.Equal(t, expected, result)
}

func Test_MarshalValue_Slice_ReturnsSlice(t *testing.T) {
	slice := []int{1, 2, 3}
	result, err := marshalValue(reflect.ValueOf(slice))
	assert.NoError(t, err)
	expected := []any{1, 2, 3}
	assert.Equal(t, expected, result)
}

func Test_UnmarshalValueWithBackupTags_StructWithFields_SetsFields(t *testing.T) {
	type TestStruct struct {
		//Field1     string               `backup:"field1"`
		//Field2     int                  `backup:"field2"`
		TaskParams db.MapStringAnyField `backup:"task_params"`
	}
	data := map[string]any{
		//"field1": "value1",
		//"field2": 42,
		"task_params": map[string]any{
			"allow_debug": true,
			"skip_tags":   []string{"123"},
		},
	}
	var testStruct TestStruct
	err := unmarshalValueWithBackupTags(data, reflect.ValueOf(&testStruct).Elem())
	assert.NoError(t, err)
	//assert.Equal(t, "value1", testStruct.Field1)
	//assert.Equal(t, 42, testStruct.Field2)
}
func Test_UnmarshalValueWithBackupTags_Slice_SetsElements(t *testing.T) {
	data := []any{1, 2, 3}
	var slice []int
	err := unmarshalValueWithBackupTags(data, reflect.ValueOf(&slice).Elem())
	assert.NoError(t, err)
	expected := []int{1, 2, 3}
	assert.Equal(t, expected, slice)
}

func Test_UnmarshalValueWithBackupTags_Map_SetsEntries(t *testing.T) {
	data := map[string]any{"key1": "value1", "key2": "value2"}
	var m map[string]string
	err := unmarshalValueWithBackupTags(data, reflect.ValueOf(&m).Elem())
	assert.NoError(t, err)
	expected := map[string]string{"key1": "value1", "key2": "value2"}
	assert.Equal(t, expected, m)
}
func Test_SetBasicType_InvalidType_ReturnsError(t *testing.T) {
	var v reflect.Value
	err := setBasicType("string", v)
	assert.Error(t, err)
}

func Test_ToFloat64_ValidInt_ReturnsFloat64(t *testing.T) {
	result, ok := toFloat64(42)
	assert.True(t, ok)
	assert.Equal(t, 42.0, result)
}

func Test_ToFloat64_InvalidType_ReturnsFalse(t *testing.T) {
	_, ok := toFloat64("string")
	assert.False(t, ok)
}
