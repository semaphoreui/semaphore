package bolt

import (
	"errors"
	"github.com/ansible-semaphore/semaphore/db"
	"testing"
)

func TestGetOption(t *testing.T) {
	store := CreateTestStore()

	_, err := store.GetOption("unknown_option")

	if !errors.Is(err, db.ErrNotFound) {
		t.Fatal("Result must be nil for non-existent option")
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
