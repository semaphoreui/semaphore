package db

import "testing"

func TestObjectToJSON(t *testing.T) {
	v := &SurveyVar{
		Name:  "test",
		Title: "Test",
	}
	s := ObjectToJSON(v)
	if s == nil || *s != "{\"name\":\"test\",\"title\":\"Test\",\"required\":false,\"type\":\"\",\"description\":\"\",\"values\":null}" {
		t.Fail()
	}
}

func TestObjectToJSON2(t *testing.T) {
	var v *SurveyVar = nil
	s := ObjectToJSON(v)
	if s != nil {
		t.Fail()
	}
}

func TestObjectToJSON3(t *testing.T) {
	v := SurveyVar{
		Name:  "test",
		Title: "Test",
	}
	s := ObjectToJSON(v)
	if s == nil || *s != "{\"name\":\"test\",\"title\":\"Test\",\"required\":false,\"type\":\"\",\"description\":\"\",\"values\":null}" {
		t.Fail()
	}
}
