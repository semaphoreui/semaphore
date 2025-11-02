package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObjectToJSON(t *testing.T) {
	v := &SurveyVar{
		Name:  "test",
		Title: "Test",
	}
	s := ObjectToJSON(v)
	assert.NotNil(t, s)
	assert.Equal(t, "{\"name\":\"test\",\"title\":\"Test\"}", *s)
}

func TestObjectToJSON2(t *testing.T) {
	var v *SurveyVar = nil
	s := ObjectToJSON(v)
	assert.Nil(t, s)
}

func TestObjectToJSON3(t *testing.T) {
	v := SurveyVar{
		Name:  "test",
		Title: "Test",
	}
	s := ObjectToJSON(v)
	assert.NotNil(t, s)
	assert.Equal(t, "{\"name\":\"test\",\"title\":\"Test\"}", *s)
}

func TestMapStringAnyField_ScanEmptyString(t *testing.T) {
	var m MapStringAnyField
	
	// Test empty string
	err := m.Scan("")
	require.NoError(t, err, "Scan empty string should not return error")
	require.NotNil(t, m, "Scan empty string should initialize map, not set to nil")
	assert.Equal(t, 0, len(m), "Scan empty string should initialize empty map")
}

func TestMapStringAnyField_ScanEmptyBytes(t *testing.T) {
	var m MapStringAnyField
	
	// Test empty byte slice
	err := m.Scan([]byte{})
	require.NoError(t, err, "Scan empty bytes should not return error")
	require.NotNil(t, m, "Scan empty bytes should initialize map, not set to nil")
	assert.Equal(t, 0, len(m), "Scan empty bytes should initialize empty map")
}

func TestMapStringAnyField_ScanNil(t *testing.T) {
	var m MapStringAnyField
	
	// Test nil value
	err := m.Scan(nil)
	require.NoError(t, err, "Scan nil should not return error")
	assert.Nil(t, m, "Scan nil should set map to nil")
}

func TestMapStringAnyField_ScanValidJSON(t *testing.T) {
	var m MapStringAnyField
	
	// Test valid JSON string
	err := m.Scan(`{"key": "value", "number": 42}`)
	require.NoError(t, err, "Scan valid JSON should not return error")
	require.NotNil(t, m, "Scan valid JSON should initialize map")
	assert.Equal(t, "value", m["key"], "Expected key='value'")
	assert.Equal(t, float64(42), m["number"], "Expected number=42")
}
