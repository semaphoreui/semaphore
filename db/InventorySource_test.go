package db

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInventorySource_Validate(t *testing.T) {
	valid := InventorySource{Name: "git", Type: InventorySourceGit, Params: MapStringAnyField{"path": "inventory/prod.yml"}}
	assert.NoError(t, valid.Validate())

	key := 1
	assert.NoError(t, InventorySource{Name: "api", Type: "other", KeyID: &key, Params: MapStringAnyField{"url": "https://cmdb.example.com"}}.Validate())

	invalid := map[string]InventorySource{
		"no path":          {Name: "git", Type: InventorySourceGit},
		"root path":        {Name: "git", Type: InventorySourceGit, Params: MapStringAnyField{"path": "./"}},
		"path is not text": {Name: "git", Type: InventorySourceGit, Params: MapStringAnyField{"path": 5}},
		"key":              {Name: "git", Type: InventorySourceGit, KeyID: &key, Params: valid.Params},
		"large params":     {Name: "api", Type: "other", Params: MapStringAnyField{"filter": strings.Repeat("a", 17<<10)}},
		"empty name":       {Type: InventorySourceGit, Params: valid.Params},
		"empty type":       {Name: "git"},
		"parent path":      {Name: "git", Type: InventorySourceGit, Params: MapStringAnyField{"path": "../x"}},
		"absolute path":    {Name: "git", Type: InventorySourceGit, Params: MapStringAnyField{"path": "/etc/hosts"}},
		"blank name":       {Name: "  ", Type: InventorySourceGit, Params: valid.Params},
		"long path":        {Name: "git", Type: InventorySourceGit, Params: MapStringAnyField{"path": strings.Repeat("a", 1001)}},
		"drive path":       {Name: "git", Type: InventorySourceGit, Params: MapStringAnyField{"path": "C:\\inventory\\prod.yml"}},
		"drive file":       {Name: "git", Type: InventorySourceGit, Params: MapStringAnyField{"path": "c:prod.yml"}},
		"nul in path":      {Name: "git", Type: InventorySourceGit, Params: MapStringAnyField{"path": "a\x00b"}},
	}
	for name, source := range invalid {
		t.Run(name, func(t *testing.T) {
			assert.Error(t, source.Validate())
		})
	}
}
