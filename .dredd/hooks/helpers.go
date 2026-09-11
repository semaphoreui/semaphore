package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/semaphoreui/semaphore/pkg/tz"

	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db/factory"
	"github.com/semaphoreui/semaphore/db/sql"
	"github.com/semaphoreui/semaphore/pkg/random"
	proFactory "github.com/semaphoreui/semaphore/pro/db/factory"
	proFeatures "github.com/semaphoreui/semaphore/pro/pkg/features"
	"github.com/semaphoreui/semaphore/util"
	"github.com/snikch/goodman/transaction"
)

// Test Runner User
func addTestRunnerUser() {
	uid := getUUID()
	testRunnerUser = &db.User{
		Username: "ITU-" + uid,
		Name:     "ITU-" + uid,
		Email:    uid + "@semaphore.test",
		Created:  db.GetParsedTime(tz.Now()),
		Admin:    true,
	}

	dbConnect()
	defer store.Close()

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
		"project__workflow_approval",
		"project__workflow_run",
		"project__workflow_edge",
		"project__workflow_node",
		"project__workflow_template",
		"runner",
	}

	switch store.(type) {
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
	defer store.Close()
	_ = store.DeleteAPIToken(testRunnerUser.ID, adminToken)
	_ = store.DeleteUser(testRunnerUser.ID)
}

// Parameter Substitution
func setupObjectsAndPaths(t *transaction.Transaction) {
	// Skipped transactions never set up their fixtures, so path/body
	// substitution would dereference nil objects.
	if t.Skip {
		return
	}
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
	key, err := store.CreateAccessKey(db.AccessKey{
		Name:      "ITK-" + uid,
		Type:      "ssh",
		Secret:    new("5up3r53cr3t\n"),
		ProjectID: pid,
	})

	if err != nil {
		panic(err)
	}
	return &key
}

func addProject() *db.Project {
	uid := getUUID()
	project := db.Project{
		Name:      "ITP-" + uid,
		Created:   tz.Now(),
		AlertChat: new("Test"),
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
		Created:  tz.Now(),
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

func addInvite() *db.ProjectInvite {
	invite, err := store.CreateProjectInvite(db.ProjectInvite{
		ProjectID:     userProject.ID,
		UserID:        &userPathTestUser.ID,
		Email:         &userPathTestUser.Email,
		Role:          "owner",
		Status:        db.ProjectInvitePending,
		Token:         getUUID(),
		InviterUserID: testRunnerUser.ID,
		Created:       tz.Now(),
		ExpiresAt:     nil, // No expiration for this test
		AcceptedAt:    nil,
	})

	if err != nil {
		panic(err)
	}

	return &invite
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
		Created:    db.GetParsedTime(tz.Now()),
	}

	t, err := store.CreateTask(t, 0)

	if err != nil {
		fmt.Println("error during insertion of task:")
		if j, e := json.Marshal(t); e == nil {
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
		VariableType:  db.IntegrationVariableEnvironment,
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

func addRunner() *db.Runner {
	runner, err := store.CreateRunner(db.Runner{
		Token:            db.GenerateRunnerToken(),
		ProjectID:        &userProject.ID,
		Name:             "ITRN-" + getUUID(),
		Active:           true,
		MaxParallelTasks: 1,
	})

	if err != nil {
		panic(err)
	}

	return &runner
}

func addGlobalRunner() *db.Runner {
	runner, err := store.CreateRunner(db.Runner{
		Token:            db.GenerateRunnerToken(),
		ProjectID:        nil,
		Name:             "ITGRN-" + getUUID(),
		Active:           true,
		MaxParallelTasks: 1,
	})

	if err != nil {
		panic(err)
	}

	return &runner
}

func addWorkflow() *db.WorkflowTemplate {
	input := db.WorkflowTemplate{
		ProjectID: userProject.ID,
		Name:      "ITW-" + getUUID(),
		Nodes: []db.WorkflowNode{
			{ID: 1, TemplateID: templateID},
			{ID: 2, Kind: db.WorkflowNodeApprovalKind, ApprovalMessage: new("approve")},
		},
		Edges: []db.WorkflowEdge{
			{
				SourceNodeID:      2,
				DestinationNodeID: 1,
				Condition:         db.WorkflowEdgeOnSuccess,
			},
		},
	}
	wf, err := workflowStore.CreateWorkflowTemplate(input)
	if err != nil {
		panic(err)
	}
	// When the workflow store is a no-op stub (OSS build), it returns a
	// zero-value struct with no Nodes. Copy the input nodes so that
	// addWorkflowApproval() can still find the required approval node.
	if len(wf.Nodes) == 0 {
		wf.Nodes = input.Nodes
	}
	return &wf
}

func addWorkflowRun() *db.WorkflowRun {
	run, err := workflowStore.CreateWorkflowRun(db.WorkflowRun{
		ProjectID:          userProject.ID,
		WorkflowTemplateID: workflowID,
		Status:             db.WorkflowRunRunning,
		Start:              new(tz.Now()),
	})
	if err != nil {
		panic(err)
	}
	return &run
}

func addWorkflowApproval() *db.WorkflowApproval {
	if workflow == nil {
		panic("workflow fixture is nil; ensure addWorkflow() is called before addWorkflowApproval()")
	}

	approvalNodeID := 0
	for _, node := range workflow.Nodes {
		if node.EffectiveKind() == db.WorkflowNodeApprovalKind {
			approvalNodeID = node.ID
			break
		}
	}
	if approvalNodeID == 0 {
		panic("no approval node found in workflow.Nodes; workflow must include at least one approval node")
	}

	approval, err := workflowStore.CreateWorkflowApproval(db.WorkflowApproval{
		ProjectID:      userProject.ID,
		WorkflowRunID:  workflowRunID,
		WorkflowNodeID: approvalNodeID,
		Status:         db.WorkflowApprovalPending,
		Created:        tz.Now(),
	})
	if err != nil {
		panic(err)
	}
	return &approval
}

// Token Handling
func addToken(tok string, user int) {
	_, err := store.CreateAPIToken(db.APIToken{
		ID:      tok,
		Created: tz.Now(),
		UserID:  user,
		Expired: false,
	})
	if err != nil {
		panic(err)
	}
}

// isProBuild reports whether the hooks binary was built with the PRO
// implementation (go.work + pro_impl). The open-source stub of
// pro/pkg/features returns an empty Features struct, while the PRO
// implementation enables Workflows even for the standard (empty) plan.
func isProBuild() bool {
	return proFeatures.GetFeatures(&db.User{}, "").Workflows
}

// HELPERS
var randSetup = false

func getUUID() string {
	if !randSetup {
		randSetup = true
	}
	return random.String(8)
}

func strPtr(v string) *string {
	return &v
}

func loadConfig() {
	cwd, _ := os.Getwd()
	file, _ := os.Open(cwd + "/.dredd/config.json")
	if err := json.NewDecoder(file).Decode(&util.Config); err != nil {
		fmt.Println("Could not decode configuration!")
		panic(err)
	}

	// The hooks decode the config file directly instead of going through
	// util.ConfigInit, so util.Config.Apps stays empty and Template.Validate
	// rejects every fixture template with "invalid app: <app>". The server under
	// test gets the same whitelist from SEMAPHORE_APPS in server-wrapper.sh,
	// which does not reach this process because dredd spawns it separately.
	if util.Config.Apps == nil {
		util.Config.Apps = make(map[string]util.App)
	}
	// Apps the fixtures create templates for. The hooks never execute a task, so
	// the corresponding binaries do not have to be installed.
	for _, app := range []db.TemplateApp{db.AppAnsible} {
		if _, ok := util.Config.Apps[string(app)]; !ok {
			util.Config.Apps[string(app)] = util.App{Active: true}
		}
	}
}

var store db.Store
var workflowStore db.WorkflowManager

func dbConnect() {
	store = factory.CreateStore()

	store.Connect()

	workflowStore = proFactory.NewWorkflowStore(store)
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
