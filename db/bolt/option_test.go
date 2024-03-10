package bolt

import (
	"testing"
)

func TestGetOption(t *testing.T) {
	store := CreateTestStore()

	val, err := store.GetOption("unknown_option")

	if err != nil && val != "" {
		t.Fatal("Result must be empty string for non-existent option")
	}
}

func TestGetSetOption(t *testing.T) {
	store := CreateTestStore()

	err := store.SetOption("age", "33")

	if err != nil {
		t.Fatal("Can not save option")
	}

	val, err := store.GetOption("age")

	if err != nil {
		t.Fatal("Can not get option")
	}

	if val != "33" {
		t.Fatal("Invalid option value")
	}

	err = store.SetOption("age", "22")

	if err != nil {
		t.Fatal("Can not save option")
	}

	val, err = store.GetOption("age")

	if err != nil {
		t.Fatal("Can not get option")
	}

	if val != "22" {
		t.Fatal("Invalid option value")
	}

}
