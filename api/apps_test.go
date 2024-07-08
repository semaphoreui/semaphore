package api

import (
	"fmt"
	"testing"
)

func TestStructToMap(t *testing.T) {
	type Address struct {
		City  string `json:"city"`
		State string `json:"state"`
	}

	type Person struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Email   string  `json:"email"`
		Active  bool    `json:"active"`
		Address Address `json:"address"`
	}

	// Create an instance of the struct
	p := Person{
		Name:   "John Doe",
		Age:    30,
		Email:  "johndoe@example.com",
		Active: true,
		Address: Address{
			City:  "New York",
			State: "NY",
		},
	}

	// Convert the struct to a flat map
	flatMap := structToFlatMap(&p)

	if flatMap["address.city"] != "New York" {
		t.Fail()
	}
	// Print the map
	fmt.Println(flatMap)
}
