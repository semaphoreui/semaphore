package bolt

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"testing"
)

func TestApiPing(t *testing.T) {
	objects := []db.Inventory{
		{
			ID: 1,
			Name: "x",
		},
		{
			ID: 2,
			Name: "a",
		},
		{
			ID: 3,
			Name: "d",
		},
		{
			ID: 4,
			Name: "b",
		},
		{
			ID: 5,
			Name: "r",
		},
	}

	err := sortObjects(&objects, "name", false)
	if err != nil {
		t.Fatal(err)
	}

	expected := objects[0].Name == "a" &&
		objects[1].Name == "b" &&
		objects[2].Name == "d" &&
		objects[3].Name == "r" &&
		objects[4].Name == "x"


	if !expected {
		t.Fatal(fmt.Errorf("objects not sorted"))
	}
}
