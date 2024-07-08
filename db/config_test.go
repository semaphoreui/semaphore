package db

import "testing"

func TestConfig_assignMapToStruct(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type User struct {
		Name    string            `json:"name"`
		Age     int               `json:"age"`
		Email   string            `json:"email"`
		Address Address           `json:"address"`
		Details map[string]string `json:"details"`
	}

	johnData := map[string]interface{}{
		"name":  "John Doe",
		"age":   30,
		"email": "john.doe@example.com",
		"address": map[string]interface{}{
			"street": "123 Main St",
			"city":   "Anytown",
		},
		"details": map[string]string{
			"occupation": "engineer",
			"hobby":      "hiking",
		},
	}

	var john User
	john.Details = make(map[string]string)
	john.Details["interests"] = "politics"

	err := assignMapToStruct(johnData, &john)

	if err != nil {
		t.Fatal(err)
	}

	if john.Name != "John Doe" {
		t.Errorf("Expected name to be John Doe but got %s", john.Name)
	}

	if john.Details["interests"] != "politics" {
		t.Errorf("Expected interests to be politics but got %s", john.Details["interests"])
	}
}
