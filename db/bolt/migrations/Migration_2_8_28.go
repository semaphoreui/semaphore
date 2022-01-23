package migrations

import (
	"encoding/json"
	"go.etcd.io/bbolt"
	"strings"
)

type Migration_2_8_28 struct {
	DB *bbolt.DB
}

func (d Migration_2_8_28) getProjectRepositories(projectID string) (map[string]map[string]interface{}, error) {
	repos := make(map[string]map[string]interface{})
	err := d.DB.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("project__repository_" + projectID))
		return b.ForEach(func(id, body []byte) error {
			r := make(map[string]interface{})
			repos[string(id)] = r
			return json.Unmarshal(body, &r)
		})
	})
	return repos, err
}

func (d Migration_2_8_28) setProjectRepository(projectID string, repoID string, repo map[string]interface{}) error {
	return d.DB.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("project__repository_" + projectID))
		j, err := json.Marshal(repo)
		if err != nil {
			return err
		}
		return b.Put([]byte(repoID), j)
	})
}

func (d Migration_2_8_28) Apply() (err error) {
	var projectIDs []string

	err = d.DB.View(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte("project")).ForEach(func(id, _ []byte) error {
			projectIDs = append(projectIDs, string(id))
			return nil
		})
	})

	if err != nil {
		return
	}

	projectsRepositories := make(map[string]map[string]map[string]interface{})

	for _, projectID := range projectIDs {
		var err2 error
		projectsRepositories[projectID], err2 = d.getProjectRepositories(projectID)
		if err2 != nil {
			return err2
		}
	}

	for projectID, repositories := range projectsRepositories {
		for repoID, repo := range repositories {
			branch := "master"
			url := repo["git_url"].(string)
			parts := strings.Split(url, "#")
			if len(parts) > 1 {
				url, branch = parts[0], parts[1]
			}
			repo["git_url"] = url
			repo["git_branch"] = branch
			err = d.setProjectRepository(projectID, repoID, repo)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
