package bolt

import (
	"strings"
)

type migration_2_8_28 struct {
	migration
}

func (d migration_2_8_28) Apply() (err error) {
	projectIDs, err := d.getProjectIDs()

	if err != nil {
		return
	}

	repos := make(map[string]map[string]map[string]interface{})

	for _, projectID := range projectIDs {
		var err2 error
		repos[projectID], err2 = d.getObjects(projectID, "repository")
		if err2 != nil {
			return err2
		}
	}

	for projectID, projectRepos := range repos {
		for repoID, repo := range projectRepos {
			branch := "master"
			url := repo["git_url"].(string)
			parts := strings.Split(url, "#")
			if len(parts) > 1 {
				url, branch = parts[0], parts[1]
			}
			repo["git_url"] = url
			repo["git_branch"] = branch
			err = d.setObject(projectID, "repository", repoID, repo)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
