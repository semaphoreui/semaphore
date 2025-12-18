package sql

import (
	"strconv"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

const templateAdditionalRepoColumns = "id, template_id, repository_id, path, git_branch, position"

func (d *SqlDb) GetTemplateAdditionalRepositories(projectID int, templateID int) (repos []db.TemplateAdditionalRepository, err error) {
	query, args, err := squirrel.Select(templateAdditionalRepoColumns).
		From("project__template_additional_repository").
		Where("template_id = ?", templateID).
		OrderBy("position ASC, id ASC").
		ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(&repos, query, args...)

	if err != nil {
		return
	}

	// Expand Repository data
	for i := range repos {
		var repo db.Repository
		repo, err = d.GetRepository(projectID, repos[i].RepositoryID)
		if err != nil {
			return
		}
		repos[i].Repository = &repo
	}

	return
}

func (d *SqlDb) UpdateTemplateAdditionalRepositories(projectID int, templateID int, repos []db.TemplateAdditionalRepository) (err error) {
	if repos == nil {
		repos = []db.TemplateAdditionalRepository{}
	}

	var repoIDs []string
	for _, repo := range repos {
		// Load repository data for validation
		var repository db.Repository
		repository, err = d.GetRepository(projectID, repo.RepositoryID)
		if err != nil {
			return
		}
		repo.Repository = &repository

		err = repo.Validate()
		if err != nil {
			return
		}

		repo.TemplateID = templateID

		if repo.ID == 0 {
			// Insert new additional repository
			var repoID int
			repoID, err = d.insert("id",
				"insert into project__template_additional_repository "+
					"(template_id, repository_id, path, git_branch, position) "+
					"values (?, ?, ?, ?, ?)",
				templateID, repo.RepositoryID, repo.Path, repo.GitBranch, repo.Position)
			if err != nil {
				return
			}
			repoIDs = append(repoIDs, strconv.Itoa(repoID))
		} else {
			// Update existing additional repository
			_, err = d.exec(
				"update project__template_additional_repository set "+
					"repository_id=?, path=?, git_branch=?, position=? "+
					"where id=? and template_id=?",
				repo.RepositoryID, repo.Path, repo.GitBranch, repo.Position, repo.ID, templateID)
			repoIDs = append(repoIDs, strconv.Itoa(repo.ID))
		}
		if err != nil {
			return
		}
	}

	// Delete removed additional repositories
	if len(repoIDs) == 0 {
		_, err = d.exec("delete from project__template_additional_repository where template_id=?", templateID)
	} else {
		_, err = d.exec("delete from project__template_additional_repository where template_id=? and id not in ("+strings.Join(repoIDs, ",")+")", templateID)
	}

	return
}
