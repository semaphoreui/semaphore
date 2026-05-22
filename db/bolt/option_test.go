package bolt

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetOptions_PrefixCollision(t *testing.T) {
	store := CreateTestStore()

	require.NoError(t, store.SetOption("user1.nav.pinnedItems", "a"))
	require.NoError(t, store.SetOption("user10.nav.pinnedItems", "b"))

	res, err := store.GetOptions(db.RetrieveQueryParams{Filter: "user1"})
	require.NoError(t, err)

	// filter "user1" must match only "user1.*" and not the sibling "user10.*"
	assert.Equal(t, map[string]string{"user1.nav.pinnedItems": "a"}, res)
}

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
