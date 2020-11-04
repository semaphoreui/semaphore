package projects

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"

	"github.com/ansible-semaphore/semaphore/db"

	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/context"
	"github.com/masterminds/squirrel"
)

func clearRepositoryCache(repository db.Repository) error {
	repoName := "repository_" + strconv.Itoa(repository.ID)
	repoPath := util.Config.TmpPath + "/" + repoName
	_, err := os.Stat(repoPath)
	if err == nil {
		return os.RemoveAll(repoPath)
	}
	return nil
}

// RepositoryMiddleware ensures a repository exists and loads it to the context
func RepositoryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := context.Get(r, "project").(db.Project)
		repositoryID, err := util.GetIntParam("repository_id", w, r)
		if err != nil {
			return
		}

		var repository db.Repository
		if err := db.Mysql.SelectOne(&repository, "select * from project__repository where project_id=? and id=?", project.ID, repositoryID); err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			panic(err)
		}

		context.Set(r, "repository", repository)
		next.ServeHTTP(w, r)
	})
}

// GetRepositories returns all repositories in a project sorted by type
func GetRepositories(w http.ResponseWriter, r *http.Request) {
	if repo := context.Get(r, "repository"); repo != nil {
		util.WriteJSON(w, http.StatusOK, repo.(db.Repository))
		return
	}

	project := context.Get(r, "project").(db.Project)
	var repos []db.Repository

	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")

	if order != asc && order != desc {
		order = asc
	}

	q := squirrel.Select("pr.id",
		"pr.name",
		"pr.project_id",
		"pr.git_url",
		"pr.ssh_key_id",
		"pr.removed").
		From("project__repository pr")

	switch sort {
	case "name", "git_url":
		q = q.Where("pr.project_id=?", project.ID).
			OrderBy("pr." + sort + " " + order)
	case "ssh_key":
		q = q.LeftJoin("access_key ak ON (pr.ssh_key_id = ak.id)").
			Where("pr.project_id=?", project.ID).
			OrderBy("ak.name " + order)
	default:
		q = q.Where("pr.project_id=?", project.ID).
			OrderBy("pr.name " + order)
	}

	query, args, err := q.ToSql()
	util.LogWarning(err)

	if _, err := db.Mysql.Select(&repos, query, args...); err != nil {
		panic(err)
	}

	util.WriteJSON(w, http.StatusOK, repos)
}

// AddRepository creates a new repository in the database
func AddRepository(w http.ResponseWriter, r *http.Request) {
	project := context.Get(r, "project").(db.Project)

	var repository struct {
		Name     string `json:"name" binding:"required"`
		GitURL   string `json:"git_url" binding:"required"`
		SSHKeyID int    `json:"ssh_key_id" binding:"required"`
	}
	if err := util.Bind(w, r, &repository); err != nil {
		return
	}

	res, err := db.Mysql.Exec("insert into project__repository set project_id=?, git_url=?, ssh_key_id=?, name=?", project.ID, repository.GitURL, repository.SSHKeyID, repository.Name)
	if err != nil {
		panic(err)
	}

	insertID, err := res.LastInsertId()
	util.LogWarning(err)
	insertIDInt := int(insertID)
	objType := "repository"

	desc := "Repository (" + repository.GitURL + ") created"
	if err := (db.Event{
		ProjectID:   &project.ID,
		ObjectType:  &objType,
		ObjectID:    &insertIDInt,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateRepository updates the values of a repository in the database
func UpdateRepository(w http.ResponseWriter, r *http.Request) {
	oldRepo := context.Get(r, "repository").(db.Repository)
	var repository struct {
		Name     string `json:"name" binding:"required"`
		GitURL   string `json:"git_url" binding:"required"`
		SSHKeyID int    `json:"ssh_key_id" binding:"required"`
	}
	if err := util.Bind(w, r, &repository); err != nil {
		return
	}

	if _, err := db.Mysql.Exec("update project__repository set name=?, git_url=?, ssh_key_id=? where id=?", repository.Name, repository.GitURL, repository.SSHKeyID, oldRepo.ID); err != nil {
		panic(err)
	}

	if oldRepo.GitURL != repository.GitURL {
		util.LogWarning(clearRepositoryCache(oldRepo))
	}

	desc := "Repository (" + repository.GitURL + ") updated"
	objType := "inventory"
	if err := (db.Event{
		ProjectID:   &oldRepo.ProjectID,
		Description: &desc,
		ObjectID:    &oldRepo.ID,
		ObjectType:  &objType,
	}.Insert()); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveRepository deletes a repository from a project in the database
func RemoveRepository(w http.ResponseWriter, r *http.Request) {
	repository := context.Get(r, "repository").(db.Repository)

	templatesC, err := db.Mysql.SelectInt("select count(1) from project__template where project_id=? and repository_id=?", repository.ProjectID, repository.ID)
	if err != nil {
		panic(err)
	}

	if templatesC > 0 {
		if len(r.URL.Query().Get("setRemoved")) == 0 {
			util.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error":        "Repository is in use by one or more templates",
				"templatesUse": true,
			})

			return
		}

		if _, err := db.Mysql.Exec("update project__repository set removed=1 where id=?", repository.ID); err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	if _, err := db.Mysql.Exec("delete from project__repository where id=?", repository.ID); err != nil {
		panic(err)
	}

	util.LogWarning(clearRepositoryCache(repository))

	desc := "Repository (" + repository.GitURL + ") deleted"
	if err := (db.Event{
		ProjectID:   &repository.ProjectID,
		Description: &desc,
	}.Insert()); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}
