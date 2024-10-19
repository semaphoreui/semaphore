package db

import (
	"math/rand"
	"os"
	"path"
	"testing"

	"github.com/ansible-semaphore/semaphore/util"
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

func TestRepository_GetGitURL(t *testing.T) {
	for _, v := range []struct {
		Repository     Repository
		ExpectedGitUrl string
	}{
		{
			Repository: Repository{GitURL: "https://github.com/user/project.git", SSHKey: AccessKey{
				Type: AccessKeyLoginPassword,
				LoginPassword: LoginPassword{
					Login:    "login",
					Password: "password",
				},
			},
			},
			ExpectedGitUrl: "https://login:password@github.com/user/project.git",
		},
		{
			Repository: Repository{GitURL: "https://github.com/user/project.git", SSHKey: AccessKey{
				Type: AccessKeyLoginPassword,
				LoginPassword: LoginPassword{
					Password: "password",
				},
			},
			},
			ExpectedGitUrl: "https://password@github.com/user/project.git",
		},
	} {
		gitUrl := v.Repository.GetGitURL()
		if gitUrl != v.ExpectedGitUrl {
			t.Error("wrong gitUrl", "expected: ", v.ExpectedGitUrl, " got: ", gitUrl)
		}
	}
}
