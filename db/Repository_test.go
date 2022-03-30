package db

import (
	"github.com/ansible-semaphore/semaphore/util"
	"math/rand"
	"os"
	"path"
	"testing"
)

func TestRepository_GetSchema(t *testing.T) {
	repo := Repository{GitURL: "https://example.com/hello/world"}
	schema := repo.GetType()
	if schema != "https" {
		t.Fatal()
	}
}

func TestRepository_ClearCache(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: path.Join(os.TempDir(), util.RandString(rand.Intn(10-4)+4)),
	}
	repoDir := path.Join(util.Config.TmpPath, "repository_123_55")
	err := os.MkdirAll(repoDir, 0755)
	if err != nil {
		t.Fatal(err)
	}
	repo := Repository{ID: 123}
	err = repo.ClearCache()
	if err != nil {
		t.Fatal(err)
	}
	_, err = os.Stat(repoDir)
	if err == nil {
		t.Fatal("repo directory not deleted")
	}
	if !os.IsNotExist(err) {
		t.Fatal(err)
	}
}
