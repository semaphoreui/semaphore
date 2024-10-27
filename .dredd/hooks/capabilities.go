package main

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/semaphoreui/semaphore/db"
	trans "github.com/snikch/goodman/transaction"
)

// STATE
// Runtime created objects we need to reference in test setups
var testRunnerUser *db.User
var userPathTestUser *db.User
var userProject *db.Project
var userKey *db.AccessKey
var task *db.Task
var schedule *db.Schedule
var view *db.View
var integration *db.Integration
var integrationextractvalue *db.IntegrationExtractValue
var integrationmatch *db.IntegrationMatcher

// Runtime created simple ID values for some items we need to reference in other objects
var repoID int
var inventoryID int
var environmentID int
var templateID int
var integrationID int
var integrationExtractValueID int
var integrationMatchID int

var capabilities = map[string][]string{
	"user":                    {},
	"project":                 {"user"},
	"repository":              {"access_key"},
	"inventory":               {"repository"},
	"environment":             {"repository"},
	"template":                {"repository", "inventory", "environment", "view"},
	"task":                    {"template"},
	"schedule":                {"template"},
	"view":                    {},
	"integration":             {"project", "template"},
	"integrationextractvalue": {"integration"},
	"integrationmatcher":      {"integration"},
}

func capabilityWrapper(cap string) func(t *trans.Transaction) {
	return func(t *trans.Transaction) {
		addCapabilities([]string{cap})
	}
}

func addCapabilities(caps []string) {
	dbConnect()
	defer store.Close("")
	resolved := make([]string, 0)
	uid := getUUID()
	resolveCapability(caps, resolved, uid)
}

func resolveCapability(caps []string, resolved []string, uid string) {
	for _, v := range caps {

		//if cap has deps resolve them
		if val, ok := capabilities[v]; ok {
			resolveCapability(val, resolved, uid)
		}

		//skip if already resolved
		if _, exists := stringInSlice(v, resolved); exists {
			continue
		}

		//Add dep specific stuff
		switch v {
		case "schedule":
			schedule = addSchedule()
		case "view":
			view = addView()
		case "user":
			userPathTestUser = addUser()
		case "project":
			userProject = addProject()
			//allow the admin user (test executor) to manipulate the project
			addUserProjectRelation(userProject.ID, testRunnerUser.ID)
			addUserProjectRelation(userProject.ID, userPathTestUser.ID)
		case "access_key":
			userKey = addAccessKey(&userProject.ID)
		case "repository":
			pRepo, err := store.CreateRepository(db.Repository{
				ProjectID: userProject.ID,
				GitURL:    "git@github.com/ansible,semaphore/semaphore",
				GitBranch: "develop",
				SSHKeyID:  userKey.ID,
				Name:      "ITR-" + uid,
			})
			printError(err)
			repoID = pRepo.ID
		case "inventory":
			res, err := store.CreateInventory(db.Inventory{
				ProjectID:   userProject.ID,
				Name:        "ITI-" + uid,
				Type:        "static",
				SSHKeyID:    &userKey.ID,
				BecomeKeyID: &userKey.ID,
				Inventory:   "Test Inventory",
			})
			printError(err)
			inventoryID = res.ID
		case "environment":
			pwd := "test-pass"
			env := "{}"
			res, err := store.CreateEnvironment(db.Environment{
				ProjectID: userProject.ID,
				Name:      "ITI-" + uid,
				JSON:      "{}",
				Password:  &pwd,
				ENV:       &env,
			})
			printError(err)
			environmentID = res.ID
		case "template":
			args := "[]"
			desc := "Hello, World!"
			branch := "main"
			res, err := store.CreateTemplate(db.Template{
				ProjectID:               userProject.ID,
				InventoryID:             &inventoryID,
				RepositoryID:            repoID,
				EnvironmentID:           &environmentID,
				Name:                    "Test-" + uid,
				Playbook:                "test-playbook.yml",
				Arguments:               &args,
				AllowOverrideArgsInTask: false,
				Description:             &desc,
				ViewID:                  &view.ID,
				App:                     db.AppAnsible,
				GitBranch:               &branch,
			})

			printError(err)
			templateID = res.ID
		case "task":
			task = addTask()
		case "integration":
			integration = addIntegration()
			integrationID = integration.ID
		case "integrationextractvalue":
			integrationextractvalue = addIntegrationExtractValue()
			integrationExtractValueID = integrationextractvalue.ID
		case "integrationmatcher":
			integrationmatch = addIntegrationMatcher()
			integrationMatchID = integrationmatch.ID
		default:
			panic("unknown capability " + v)
		}
		resolved = append(resolved, v)
	}
}

// HOOKS
var skipTest = func(t *trans.Transaction) {
	t.Skip = true
}

// Contains all the substitutions for paths under test
// The parameter example value in the api-doc should respond to the index+1 of the function in this slice
// ie the project id, with example value 1, will be replaced by the return value of pathSubPatterns[0]
var pathSubPatterns = []func() string{
	func() string { return strconv.Itoa(userProject.ID) },
	func() string { return strconv.Itoa(userPathTestUser.ID) },
	func() string { return strconv.Itoa(userKey.ID) },
	func() string { return strconv.Itoa(repoID) },
	func() string { return strconv.Itoa(inventoryID) },
	func() string { return strconv.Itoa(environmentID) },
	func() string { return strconv.Itoa(templateID) },
	func() string { return strconv.Itoa(task.ID) },
	func() string { return strconv.Itoa(schedule.ID) },
	func() string { return strconv.Itoa(view.ID) },
	func() string { return strconv.Itoa(integration.ID) },
	func() string { return strconv.Itoa(integrationextractvalue.ID) },
	func() string { return strconv.Itoa(integrationmatch.ID) },
}

// alterRequestPath with the above slice of functions
func alterRequestPath(t *trans.Transaction) {
	pathArgs := strings.Split(t.FullPath, "/")
	exploded := make([]string, len(pathArgs))
	copy(exploded, pathArgs)
	for k, v := range pathSubPatterns {

		pos, exists := stringInSlice(strconv.Itoa(k+1), exploded)
		if exists {
			pathArgs[pos] = v()
		}
	}
	t.FullPath = strings.Join(pathArgs, "/")

	t.Request.URI = t.FullPath
}

func alterRequestBody(t *trans.Transaction) {
	var request map[string]interface{}
	json.Unmarshal([]byte(t.Request.Body), &request)

	if userProject != nil {
		bodyFieldProcessor("project_id", userProject.ID, &request)
	}
	bodyFieldProcessor("json", "{}", &request)
	if userKey != nil {
		bodyFieldProcessor("ssh_key_id", userKey.ID, &request)
		bodyFieldProcessor("become_key_id", userKey.ID, &request)
	}
	bodyFieldProcessor("environment_id", environmentID, &request)
	bodyFieldProcessor("inventory_id", inventoryID, &request)
	bodyFieldProcessor("repository_id", repoID, &request)
	bodyFieldProcessor("template_id", templateID, &request)
	if task != nil {
		bodyFieldProcessor("task_id", task.ID, &request)
	}
	if schedule != nil {
		bodyFieldProcessor("schedule_id", schedule.ID, &request)
	}
	if view != nil {
		bodyFieldProcessor("view_id", view.ID, &request)
	}

	if integration != nil {
		bodyFieldProcessor("integration_id", integration.ID, &request)
	}
	if integrationextractvalue != nil {
		bodyFieldProcessor("value_id", integrationextractvalue.ID, &request)
	}
	if integrationmatch != nil {
		bodyFieldProcessor("matcher_id", integrationmatch.ID, &request)
	}

	// Inject object ID to body for PUT requests
	if strings.ToLower(t.Request.Method) == "put" {

		putRequestPathRE := regexp.MustCompile(`\w+/(\d+)/?$`)
		m := putRequestPathRE.FindStringSubmatch(t.FullPath)
		if len(m) > 0 {
			objectID, err := strconv.Atoi(m[1])
			if err != nil {
				panic("Invalid object ID in PUT request " + t.FullPath)
			}
			request["id"] = objectID

		} else {
			panic("Unexpected PUT request " + t.FullPath)
		}
	}

	out, _ := json.Marshal(request)
	t.Request.Body = string(out)
}

func bodyFieldProcessor(id string, sub interface{}, request *map[string]interface{}) {
	if _, ok := (*request)[id]; ok {
		(*request)[id] = sub
	}
}
