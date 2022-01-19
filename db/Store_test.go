package db

import "testing"

func TestObjectToJSON(t *testing.T) {
	v := &SurveyVar{
		Name:  "test",
		Title: "Test",
	}
	s := ObjectToJSON(v)
	if s == nil {
		t.Fail()
	}
	if *s != "{\"name\":\"test\",\"title\":\"Test\"}" {
		t.Fail()
	}
}

func TestObjectToJSON2(t *testing.T) {
	var v *SurveyVar
	v = nil
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
	if s == nil {
		t.Fail()
	}
	if *s != "{\"name\":\"test\",\"title\":\"Test\"}" {
		t.Fail()
	}
}
