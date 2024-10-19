package projects

import (
	"net/http"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	log "github.com/sirupsen/logrus"

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

func createDemoProject(projectID int, noneKeyID int, emptyEnvID int, store db.Store) (err error) {
	var demoRepo db.Repository

	var buildInv db.Inventory
	var devInv db.Inventory
	var prodInv db.Inventory

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
		SSHKeyID:  noneKeyID,
	})

	if err != nil {
		return
	}

	buildInv, err = store.CreateInventory(db.Inventory{
		Name:      "Build",
		ProjectID: projectID,
		Inventory: "[builder]\nlocalhost ansible_connection=local",
		Type:      "static",
		SSHKeyID:  &noneKeyID,
	})

	if err != nil {
		return
	}

	devInv, err = store.CreateInventory(db.Inventory{
		Name:      "Dev",
		ProjectID: projectID,
		Inventory: "invs/dev/hosts",
		Type:      "file",
		SSHKeyID:  &noneKeyID,
	})

	if err != nil {
		return
	}

	prodInv, err = store.CreateInventory(db.Inventory{
		Name:      "Prod",
		ProjectID: projectID,
		Inventory: "invs/prod/hosts",
		Type:      "file",
		SSHKeyID:  &noneKeyID,
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
		InventoryID:   &prodInv.ID,
		EnvironmentID: &emptyEnvID,
		RepositoryID:  demoRepo.ID,
		App:           db.AppAnsible,
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
		InventoryID:   &buildInv.ID,
		EnvironmentID: &emptyEnvID,
		RepositoryID:  demoRepo.ID,
		StartVersion:  &startVersion,
		App:           db.AppAnsible,
	})

	if err != nil {
		return
	}

	var template db.Template
	template, err = store.CreateTemplate(db.Template{
		Name:            "Deploy to Dev",
		Type:            db.TemplateDeploy,
		Playbook:        "deploy.yml",
		ProjectID:       projectID,
		InventoryID:     &devInv.ID,
		EnvironmentID:   &emptyEnvID,
		RepositoryID:    demoRepo.ID,
		BuildTemplateID: &buildTpl.ID,
		Autorun:         true,
		App:             db.AppAnsible,
	})

	if err != nil {
		return
	}

	_, err = store.CreateTemplateVault(db.TemplateVault{
		ProjectID:  projectID,
		TemplateID: template.ID,
		VaultKeyID: vaultKey.ID,
		Name:       nil,
	})

	if err != nil {
		return
	}

	template, err = store.CreateTemplate(db.Template{
		Name:            "Deploy to Production",
		Type:            db.TemplateDeploy,
		Playbook:        "deploy.yml",
		ProjectID:       projectID,
		InventoryID:     &prodInv.ID,
		EnvironmentID:   &emptyEnvID,
		RepositoryID:    demoRepo.ID,
		BuildTemplateID: &buildTpl.ID,
		App:             db.AppAnsible,
	})

	if err != nil {
		return
	}

	_, err = store.CreateTemplateVault(db.TemplateVault{
		ProjectID:  projectID,
		TemplateID: template.ID,
		VaultKeyID: vaultKey.ID,
		Name:       nil,
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

	//_, err = store.CreateInventory(db.Inventory{
	//	Name:      "None",
	//	ProjectID: body.ID,
	//	Type:      "none",
	//	SSHKeyID:  &noneKey.ID,
	//})

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	emptyEnv, err := store.CreateEnvironment(db.Environment{
		Name:      "Empty",
		ProjectID: body.ID,
		JSON:      "{}",
	})

	if err != nil {
		return
	}

	if bodyWithDemo.Demo {
		err = createDemoProject(body.ID, noneKey.ID, emptyEnv.ID, store)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}
	}

	helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   body.ID,
		ObjectType:  db.EventProject,
		ObjectID:    body.ID,
		Description: "Project created",
	})

	helpers.WriteJSON(w, http.StatusCreated, body)
}
