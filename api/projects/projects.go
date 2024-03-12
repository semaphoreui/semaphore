package projects

import (
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	log "github.com/sirupsen/logrus"
	"net/http"

	"github.com/gorilla/context"
)

// GetProjects returns all projects in this users context
func GetProjects(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "user").(*db.User)

	var err error
	var projects []db.Project
	if user.Admin {
		projects, err = helpers.Store(r).GetAllProjects()
	} else {
		projects, err = helpers.Store(r).GetProjects(user.ID)
	}

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, projects)
}

func createDemoProject(projectID int, store db.Store) (err error) {
	var noneKey db.AccessKey
	var demoRepo db.Repository
	var emptyEnv db.Environment

	var buildInv db.Inventory
	var devInv db.Inventory
	var prodInv db.Inventory

	noneKey, err = store.CreateAccessKey(db.AccessKey{
		Name:      "None",
		Type:      db.AccessKeyNone,
		ProjectID: &projectID,
	})

	if err != nil {
		return
	}

	vaultKey, err := store.CreateAccessKey(db.AccessKey{
		Name:      "Vault Password",
		Type:      db.AccessKeyLoginPassword,
		ProjectID: &projectID,
		LoginPassword: db.LoginPassword{
			Password: "RAX6yKN7sBn2qDagRPls",
		},
	})

	if err != nil {
		return
	}

	demoRepo, err = store.CreateRepository(db.Repository{
		Name:      "Demo",
		ProjectID: projectID,
		GitURL:    "https://github.com/semaphoreui/demo-project.git",
		GitBranch: "main",
		SSHKeyID:  noneKey.ID,
	})

	if err != nil {
		return
	}

	emptyEnv, err = store.CreateEnvironment(db.Environment{
		Name:      "Empty",
		ProjectID: projectID,
		JSON:      "{}",
	})

	if err != nil {
		return
	}

	buildInv, err = store.CreateInventory(db.Inventory{
		Name:      "Build",
		ProjectID: projectID,
		Inventory: "[builder]\nlocalhost ansible_connection=local",
		Type:      "static",
		SSHKeyID:  &noneKey.ID,
	})

	if err != nil {
		return
	}

	devInv, err = store.CreateInventory(db.Inventory{
		Name:      "Dev",
		ProjectID: projectID,
		Inventory: "invs/dev/hosts",
		Type:      "file",
		SSHKeyID:  &noneKey.ID,
	})

	if err != nil {
		return
	}

	prodInv, err = store.CreateInventory(db.Inventory{
		Name:      "Prod",
		ProjectID: projectID,
		Inventory: "invs/prod/hosts",
		Type:      "file",
		SSHKeyID:  &noneKey.ID,
	})

	var desc string

	if err != nil {
		return
	}

	desc = "This task pings the website to provide real word example of using Semaphore."
	_, err = store.CreateTemplate(db.Template{
		Name:          "Ping Site",
		Playbook:      "ping.yml",
		Description:   &desc,
		ProjectID:     projectID,
		InventoryID:   prodInv.ID,
		EnvironmentID: &emptyEnv.ID,
		RepositoryID:  demoRepo.ID,
	})

	if err != nil {
		return
	}

	desc = "Creates artifact and store it in the cache."

	var startVersion = "1.0.0"
	buildTpl, err := store.CreateTemplate(db.Template{
		Name:          "Build",
		Playbook:      "build.yml",
		Type:          db.TemplateBuild,
		ProjectID:     projectID,
		InventoryID:   buildInv.ID,
		EnvironmentID: &emptyEnv.ID,
		RepositoryID:  demoRepo.ID,
		StartVersion:  &startVersion,
	})

	if err != nil {
		return
	}

	_, err = store.CreateTemplate(db.Template{
		Name:            "Deploy to Dev",
		Type:            db.TemplateDeploy,
		Playbook:        "deploy.yml",
		ProjectID:       projectID,
		InventoryID:     devInv.ID,
		EnvironmentID:   &emptyEnv.ID,
		RepositoryID:    demoRepo.ID,
		BuildTemplateID: &buildTpl.ID,
		Autorun:         true,
		VaultKeyID:      &vaultKey.ID,
	})

	if err != nil {
		return
	}

	_, err = store.CreateTemplate(db.Template{
		Name:            "Deploy to Production",
		Type:            db.TemplateDeploy,
		Playbook:        "deploy.yml",
		ProjectID:       projectID,
		InventoryID:     prodInv.ID,
		EnvironmentID:   &emptyEnv.ID,
		RepositoryID:    demoRepo.ID,
		BuildTemplateID: &buildTpl.ID,
		VaultKeyID:      &vaultKey.ID,
	})

	return
}

// AddProject adds a new project to the database
func AddProject(w http.ResponseWriter, r *http.Request) {

	user := context.Get(r, "user").(*db.User)

	if !user.Admin && !util.Config.NonAdminCanCreateProject {
		log.Warn(user.Username + " is not permitted to edit users")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var bodyWithDemo struct {
		db.Project
		Demo bool `json:"demo"`
	}

	if !helpers.Bind(w, r, &bodyWithDemo) {
		return
	}

	body := bodyWithDemo.Project

	store := helpers.Store(r)

	body, err := store.CreateProject(body)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	_, err = store.CreateProjectUser(db.ProjectUser{ProjectID: body.ID, UserID: user.ID, Role: db.ProjectOwner})
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	noneKey, err := store.CreateAccessKey(db.AccessKey{
		Name:      "None",
		Type:      db.AccessKeyNone,
		ProjectID: &body.ID,
	})

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	_, err = store.CreateInventory(db.Inventory{
		Name:      "None",
		ProjectID: body.ID,
		Type:      "none",
		SSHKeyID:  &noneKey.ID,
	})

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	if bodyWithDemo.Demo {
		err = createDemoProject(body.ID, store)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}
	}

	desc := "Project Created"
	oType := db.EventProject
	_, err = store.CreateEvent(db.Event{
		UserID:      &user.ID,
		ProjectID:   &body.ID,
		Description: &desc,
		ObjectType:  &oType,
		ObjectID:    &body.ID,
	})

	if err != nil {
		log.Error(err)
	}

	helpers.WriteJSON(w, http.StatusCreated, body)
}
