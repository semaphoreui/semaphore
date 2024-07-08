package db

import "testing"

func TestConfig_assignMapToStruct(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type Detail struct {
		Value       string `json:"value"`
		Description string `json:"description"`
	}

	type User struct {
		//Name    string            `json:"name"`
		//Age     int               `json:"age"`
		//Email   string            `json:"email"`
		//Address Address           `json:"address"`
		Details map[string]Detail `json:"details"`
	}

	johnData := map[string]interface{}{
		//"name":  "John Doe",
		//"age":   30,
		//"email": "john.doe@example.com",
		//"address": map[string]interface{}{
		//	"street": "123 Main St",
		//	"city":   "Anytown",
		//},
		"details": map[string]interface{}{
			//"occupation": map[string]interface{}{
			//	"value":       "engineer",
			//	"description": "Works with computers",
			//},
			//"hobby": map[string]interface{}{
			//	"value":       "hiking",
			//	"description": "Enjoys the outdoors",
			//},
			"interests": map[string]interface{}{
				"description": "Ho ho ho",
			},
		},
	}

	var john User
	john.Details = make(map[string]Detail)
	john.Details["interests"] = Detail{
		Value:       "politics",
		Description: "Follows current events",
	}

	err := AssignMapToStruct(johnData, &john)

	if err != nil {
		t.Fatal(err)
	}

	//if john.Name != "John Doe" {
	//	t.Errorf("Expected name to be John Doe but got %s", john.Name)
	//}

	if john.Details["interests"].Description != "Ho ho ho" {
		t.Errorf("Expected interests description to be 'Ho ho ho' but got %s", john.Details["interests"].Description)
	}

	if john.Details["interests"].Value != "politics" {
		t.Errorf("Expected interests to be politics but got '%s'", john.Details["interests"].Value)
	}

	//if john.Details["occupation"].Value != "engineer" {
	//	t.Errorf("Expected occupation to be engineer but got %s", john.Details["occupation"].Value)
	//}
}
