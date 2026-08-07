package project

import (
	"reflect"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
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

func Test_MarshalValue_Runner_ExcludesSecrets(t *testing.T) {
	regToken := "reg-secret"
	runner := db.Runner{
		ID:                    7,
		Name:                  "runner-1",
		Token:                 "super-secret-token",
		RegistrationTokenHash: &regToken,
		Webhook:               "https://example.com/hook",
	}

	result, err := marshalValue(reflect.ValueOf(runner))
	assert.NoError(t, err)

	m, ok := result.(map[string]any)
	assert.True(t, ok)

	// Authentication secrets must never be present in a project backup.
	assert.NotContains(t, m, "token")
	assert.NotContains(t, m, "registration_token")
	assert.NotContains(t, m, "registration_token_expires_at")

	// Non-secret fields are still exported.
	assert.Equal(t, "runner-1", m["name"])
	assert.Equal(t, "https://example.com/hook", m["webhook"])
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

func Test_MarshalUnmarshal_SurveyVarDefaultValue_ScalarRoundTrip(t *testing.T) {
	// Verify that SurveyVarDefaultValue (a json.Marshaler/Unmarshaler type)
	// is handled correctly by the backup marshaller — not serialized as "{}".
	type Wrapper struct {
		DefaultValue *db.SurveyVarDefaultValue `backup:"default_value"`
	}

	dv := &db.SurveyVarDefaultValue{}
	_ = dv.UnmarshalJSON([]byte(`"hello"`))
	original := Wrapper{DefaultValue: dv}

	// Marshal
	marshaled, err := marshalValue(reflect.ValueOf(original))
	assert.NoError(t, err)
	m, ok := marshaled.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "hello", m["default_value"], "scalar default_value must survive backup marshal")

	// Unmarshal
	var restored Wrapper
	err = unmarshalValueWithBackupTags(m, reflect.ValueOf(&restored).Elem())
	assert.NoError(t, err)
	assert.NotNil(t, restored.DefaultValue)
	assert.Equal(t, []string{"hello"}, restored.DefaultValue.Values)
	assert.False(t, restored.DefaultValue.IsArray())
}

func Test_MarshalUnmarshal_SurveyVarDefaultValue_ArrayRoundTrip(t *testing.T) {
	type Wrapper struct {
		DefaultValue *db.SurveyVarDefaultValue `backup:"default_value"`
	}

	dv := &db.SurveyVarDefaultValue{}
	_ = dv.UnmarshalJSON([]byte(`["a","b"]`))
	original := Wrapper{DefaultValue: dv}

	// Marshal
	marshaled, err := marshalValue(reflect.ValueOf(original))
	assert.NoError(t, err)
	m, ok := marshaled.(map[string]any)
	assert.True(t, ok)
	arr, ok := m["default_value"].([]any)
	assert.True(t, ok, "array default_value must survive backup marshal as []any")
	assert.Equal(t, []any{"a", "b"}, arr)

	// Unmarshal
	var restored Wrapper
	err = unmarshalValueWithBackupTags(m, reflect.ValueOf(&restored).Elem())
	assert.NoError(t, err)
	assert.NotNil(t, restored.DefaultValue)
	assert.Equal(t, []string{"a", "b"}, restored.DefaultValue.Values)
	assert.True(t, restored.DefaultValue.IsArray())
}
