package projects

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Digital-Data-Co/forge/api/helpers"
	"github.com/Digital-Data-Co/forge/db"
	"github.com/Digital-Data-Co/forge/db_lib"
	"github.com/Digital-Data-Co/forge/util"
)

// RepositoryMiddleware ensures a repository exists and loads it to the context
func RepositoryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		repositoryID, err := helpers.GetIntParam("repository_id", w, r)
		if err != nil {
			return
		}

		repository, err := helpers.Store(r).GetRepository(project.ID, repositoryID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "repository", repository)
		next.ServeHTTP(w, r)
	})
}

func GetRepositoryRefs(w http.ResponseWriter, r *http.Request) {
	repo := helpers.GetFromContext(r, "repository").(db.Repository)
	refs, err := helpers.Store(r).GetRepositoryRefs(repo.ProjectID, repo.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

type RepositoryController struct {
	keyInstaller db_lib.AccessKeyInstaller
}

func NewRepositoryController(keyInstaller db_lib.AccessKeyInstaller) *RepositoryController {
	return &RepositoryController{
		keyInstaller: keyInstaller,
	}
}

func (c *RepositoryController) GetRepositoryBranches(w http.ResponseWriter, r *http.Request) {
	repo := helpers.GetFromContext(r, "repository").(db.Repository)

	if repo.GetType() == db.RepositoryLocal || repo.GetType() == db.RepositoryFile {
		helpers.WriteJSON(w, http.StatusBadRequest, "Wrong repository type: "+repo.GetType())
		return
	}

	git := db_lib.GitRepository{
		Repository: repo,
		Client:     db_lib.CreateDefaultGitClient(c.keyInstaller),
	}

	branches, err := git.GetRemoteBranches()

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, branches)
}

type repositoryModule struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// GetRepositoryModules lists Terraform modules from an internal (local) repository
// under the "modules" directory. Optional filters:
//   - provider: e.g. "AWS", "Azure", "Google Cloud" (case-insensitive; spaces -> '-')
//   - kubernetes: "Self-Managed Kubernetes" or "Managed Kubernetes"
func (c *RepositoryController) GetRepositoryModules(w http.ResponseWriter, r *http.Request) {
	repo := helpers.GetFromContext(r, "repository").(db.Repository)

	if repo.GetType() != db.RepositoryLocal {
		helpers.WriteJSON(w, http.StatusBadRequest, "Wrong repository type: "+repo.GetType())
		return
	}

	basePath := repo.GetGitURL(true)
	startDir := filepath.Join(basePath, "modules")

	// Optional filtering by provider and kubernetes type
	provider := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("provider")))
	if provider != "" {
		provider = strings.ReplaceAll(provider, " ", "-")
		startDir = filepath.Join(startDir, provider)
	}

	kubernetes := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("kubernetes")))
	if kubernetes != "" {
		var k8sKey string
		if strings.HasPrefix(kubernetes, "self") {
			k8sKey = "self"
		} else if strings.HasPrefix(kubernetes, "managed") {
			k8sKey = "managed"
		}
		if k8sKey != "" {
			startDir = filepath.Join(startDir, k8sKey)
		}
	}

	// If startDir doesn't exist, return empty list
	if fi, err := os.Stat(startDir); err != nil || !fi.IsDir() {
		helpers.WriteJSON(w, http.StatusOK, []repositoryModule{})
		return
	}

	var modules []repositoryModule

	// Walk directories and collect any directory that contains at least one .tf file
	err := filepath.WalkDir(startDir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip paths that error out
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		entries, readErr := os.ReadDir(p)
		if readErr != nil {
			return nil
		}
		hasTf := false
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".tf") {
				hasTf = true
				break
			}
		}
		if hasTf {
			rel, _ := filepath.Rel(basePath, p)
			modules = append(modules, repositoryModule{
				Name: filepath.Base(p),
				Path: rel,
			})
			// Do not descend further inside a found module directory
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, modules)
}

// GetRepositories returns all repositories in a project sorted by type
func GetRepositories(w http.ResponseWriter, r *http.Request) {
	if repo := helpers.GetFromContext(r, "repository"); repo != nil {
		helpers.WriteJSON(w, http.StatusOK, repo.(db.Repository))
		return
	}

	project := helpers.GetFromContext(r, "project").(db.Project)

	params := helpers.QueryParamsForProps(r.URL, db.RepositoryProps)

	repos, err := helpers.Store(r).GetRepositories(project.ID, params)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, repos)
}

// AddRepository creates a new repository in the database
func AddRepository(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	var repository db.Repository

	if !helpers.Bind(w, r, &repository) {
		return
	}

	if repository.ProjectID != project.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
	}

	if err := db.ValidateRepository(helpers.Store(r), &repository); err != nil {
		helpers.WriteError(w, err)
		return
	}

	newRepo, err := helpers.Store(r).CreateRepository(repository)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   newRepo.ProjectID,
		ObjectType:  db.EventRepository,
		ObjectID:    newRepo.ID,
		Description: fmt.Sprintf("Repository %s created", repository.GitURL),
	})

	helpers.WriteJSON(w, http.StatusCreated, newRepo)
}

// UpdateRepository updates the values of a repository in the database
func UpdateRepository(w http.ResponseWriter, r *http.Request) {
	oldRepo := helpers.GetFromContext(r, "repository").(db.Repository)
	var repository db.Repository

	if !helpers.Bind(w, r, &repository) {
		return
	}

	if repository.ID != oldRepo.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Repository ID in body and URL must be the same",
		})
		return
	}

	if repository.ProjectID != oldRepo.ProjectID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}

	if err := db.ValidateRepository(helpers.Store(r), &repository); err != nil {
		helpers.WriteError(w, err)
		return
	}

	if err := helpers.Store(r).UpdateRepository(repository); err != nil {
		helpers.WriteError(w, err)
		return
	}

	if oldRepo.GitURL != repository.GitURL {
		util.LogWarning(oldRepo.ClearCache())
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   oldRepo.ProjectID,
		ObjectType:  db.EventRepository,
		ObjectID:    oldRepo.ID,
		Description: fmt.Sprintf("Repository %s updated", repository.GitURL),
	})

	w.WriteHeader(http.StatusNoContent)
}

// RemoveRepository deletes a repository from a project in the database
func RemoveRepository(w http.ResponseWriter, r *http.Request) {
	repository := helpers.GetFromContext(r, "repository").(db.Repository)

	var err error = helpers.Store(r).DeleteRepository(repository.ProjectID, repository.ID)
	if errors.Is(err, db.ErrInvalidOperation) {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Repository is in use by one or more templates",
			"inUse": true,
		})
		return
	}

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	util.LogWarning(repository.ClearCache())

	helpers.EventLog(r, helpers.EventLogDelete, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   repository.ProjectID,
		ObjectType:  db.EventRepository,
		ObjectID:    repository.ID,
		Description: fmt.Sprintf("Repository %s deleted", repository.GitURL),
	})

	w.WriteHeader(http.StatusNoContent)
}
