package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/bolt"
	"github.com/semaphoreui/semaphore/db/factory"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pkg/random"
	"github.com/semaphoreui/semaphore/util"
	"github.com/go-gorp/gorp/v3"
	"github.com/snikch/goodman/transaction"
)

// Test Runner User
func addTestRunnerUser() {
	uid := getUUID()
	testRunnerUser = &db.User{
		Username: "ITU-" + uid,
		Name:     "ITU-" + uid,
		Email:    uid + "@semaphore.test",
		Created:  db.GetParsedTime(time.Now()),
		Admin:    true,
	}

	dbConnect()
	defer store.Close("")

	truncateAll()

	newUser, err := store.CreateUserWithoutPassword(*testRunnerUser)

	if err != nil {
		panic(err)
	}

	testRunnerUser.ID = newUser.ID

	addToken(adminToken, testRunnerUser.ID)
}

func truncateAll() {
	var tablesShouldBeTruncated = [...]string{
		"access_key",
		"event",
		"user__token",
		"project",
		"task__output",
		"task",
		"session",
		"project__environment",
		"project__inventory",
		"project__repository",
		"project__template",
		"project__template_vault",
		"project__schedule",
		"project__user",
		"user",
		"project__view",
		"project__integration",
		"project__integration_extract_value",
		"project__integration_matcher",
	}

	switch store.(type) {
	case *bolt.BoltDb:
		// Do nothing
	case *sql.SqlDb:
		switch store.(*sql.SqlDb).Sql().Dialect.(type) {
		case gorp.PostgresDialect:
			// Do nothing
		case gorp.MySQLDialect:
			tx, err := store.(*sql.SqlDb).Sql().Begin()
			if err != nil {
				panic(err)
			}

			_, err = tx.Exec("SET FOREIGN_KEY_CHECKS = 0")
			if err == nil {
				for _, tableName := range tablesShouldBeTruncated {
					tx.Exec("TRUNCATE TABLE " + tableName)
				}
				tx.Exec("SET FOREIGN_KEY_CHECKS = 1")
			}

			if err := tx.Commit(); err != nil {
				panic(err)
			}
		}
	}
}

func removeTestRunnerUser(transactions []*transaction.Transaction) {
	dbConnect()
	defer store.Close("")
	_ = store.DeleteAPIToken(testRunnerUser.ID, adminToken)
	_ = store.DeleteUser(testRunnerUser.ID)
}

// Parameter Substitution
func setupObjectsAndPaths(t *transaction.Transaction) {
	alterRequestPath(t)
	alterRequestBody(t)
}

// Object Lifecycle
func addUserProjectRelation(pid int, user int) {
	_, err := store.CreateProjectUser(db.ProjectUser{
		ProjectID: pid,
		UserID:    user,
		Role:      db.ProjectOwner,
	})
	if err != nil {
		panic(err)
	}
}

func deleteUserProjectRelation(pid int, user int) {
	err := store.DeleteProjectUser(pid, user)
	if err != nil {
		panic(err)
	}
}

func addAccessKey(pid *int) *db.AccessKey {
	uid := getUUID()
	secret := "5up3r53cr3t\n"

	key, err := store.CreateAccessKey(db.AccessKey{
		Name:      "ITK-" + uid,
		Type:      "ssh",
		Secret:    &secret,
		ProjectID: pid,
	})

	if err != nil {
		panic(err)
	}
	return &key
}

func addProject() *db.Project {
	uid := getUUID()
	chat := "Test"
	project := db.Project{
		Name:      "ITP-" + uid,
		Created:   time.Now(),
		AlertChat: &chat,
	}
	project, err := store.CreateProject(project)
	if err != nil {
		panic(err)
	}

	err = store.UpdateProject(project)
	if err != nil {
		panic(err)
	}

	return &project
}

func addUser() *db.User {
	uid := getUUID()
	user := db.User{
		Created:  time.Now(),
		Username: "ITU-" + uid,
		Email:    "test@semaphore." + uid,
		Name:     "ITU-" + uid,
	}

	user, err := store.CreateUserWithoutPassword(user)

	if err != nil {
		panic(err)
	}
	return &user
}

func addView() *db.View {
	view, err := store.CreateView(db.View{
		ProjectID: userProject.ID,
		Title:     "Test",
		Position:  1,
	})

	if err != nil {
		panic(err)
	}

	return &view
}

func addSchedule() *db.Schedule {
	schedule, err := store.CreateSchedule(db.Schedule{
		TemplateID: int(templateID),
		CronFormat: "* * * 1 *",
		ProjectID:  userProject.ID,
	})

	if err != nil {
		panic(err)
	}

	return &schedule
}

func addTask() *db.Task {
	t := db.Task{
		ProjectID:  userProject.ID,
		TemplateID: templateID,
		Status:     "testing",
		UserID:     &userPathTestUser.ID,
		Created:    db.GetParsedTime(time.Now()),
	}
	t, err := store.CreateTask(t, 0)
	if err != nil {
		fmt.Println("error during insertion of task:")
		if j, err := json.Marshal(t); err == nil {
			fmt.Println(string(j))
		} else {
			fmt.Println("can not stringify task object")
		}
		panic(err)
	}
	return &t
}

func addIntegration() *db.Integration {
	integration, err := store.CreateIntegration(db.Integration{
		ProjectID:  userProject.ID,
		Name:       "Test Integration",
		TemplateID: templateID,
	})
	if err != nil {
		panic(err)
	}

	return &integration
}

func addIntegrationExtractValue() *db.IntegrationExtractValue {
	integrationextractvalue, err := store.CreateIntegrationExtractValue(userProject.ID, db.IntegrationExtractValue{
		Name:          "Value",
		IntegrationID: integrationID,
		ValueSource:   db.IntegrationExtractBodyValue,
		BodyDataType:  db.IntegrationBodyDataJSON,
		Key:           "key",
		Variable:      "var",
	})

	if err != nil {
		panic(err)
	}

	return &integrationextractvalue
}

func addIntegrationMatcher() *db.IntegrationMatcher {
	integrationmatch, err := store.CreateIntegrationMatcher(userProject.ID, db.IntegrationMatcher{
		Name:          "matcher",
		IntegrationID: integrationID,
		MatchType:     "body",
		Method:        "equals",
		BodyDataType:  "json",
		Key:           "key",
		Value:         "value",
	})

	if err != nil {
		panic(err)
	}

	return &integrationmatch
}

// Token Handling
func addToken(tok string, user int) {
	_, err := store.CreateAPIToken(db.APIToken{
		ID:      tok,
		Created: time.Now(),
		UserID:  user,
		Expired: false,
	})
	if err != nil {
		panic(err)
	}
}

// HELPERS
var randSetup = false

func getUUID() string {
	if !randSetup {
		randSetup = true
	}
	return random.String(8)
}

func loadConfig() {
	cwd, _ := os.Getwd()
	file, _ := os.Open(cwd + "/.dredd/config.json")
	if err := json.NewDecoder(file).Decode(&util.Config); err != nil {
		fmt.Println("Could not decode configuration!")
		panic(err)
	}
}

var store db.Store

func dbConnect() {
	store = factory.CreateStore()

	store.Connect("")
}

func stringInSlice(a string, list []string) (int, bool) {
	for k, b := range list {
		if b == a {
			return k, true
		}
	}
	return 0, false
}

func printError(err error) {
	if err != nil {
		//fmt.Println(err)
		panic(err)
	}
}
