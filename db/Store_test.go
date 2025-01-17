package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectToJSON(t *testing.T) {
	v := &SurveyVar{
		Name:  "test",
		Title: "Test",
	}
	s := ObjectToJSON(v)
	assert.NotNil(t, s)
	assert.Equal(t, "{\"name\":\"test\",\"title\":\"Test\",\"required\":false,\"type\":\"\",\"description\":\"\",\"values\":null}", *s)
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
	assert.Equal(t, "{\"name\":\"test\",\"title\":\"Test\",\"required\":false,\"type\":\"\",\"description\":\"\",\"values\":null}", *s)
}
