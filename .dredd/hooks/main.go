package main

import (
	"github.com/snikch/goodman/hooks"
	trans "github.com/snikch/goodman/transaction"
	"strconv"
	"strings"
)

const (
	adminToken   = "h4a_i4qslpnxyyref71rk5nqbwxccrs7enwvggx0vfs="
	expiredToken = "kwofd61g93-yuqvex8efmhjkgnbxlo8mp1tin6spyhu="
)

var skipTests = []string{
	// TODO - dredd seems not to like the text response from this endpoint
	"/api/ping > PING test > 200 > text/plain; charset=utf-8",
	"/api/ws > Websocket handler > 200 > application/json",
	"authentication > /api/auth/login > Performs Login > 204 > application/json",
	"authentication > /api/auth/logout > Destroys current session > 204 > application/json",
	//"/api/upgrade > Upgrade the server > 200 > application/json",
	// TODO - Skipping this while we work out how to get a 204 response from the api for testing
	//"/api/upgrade > Check if new updates available and fetch /info > 204 > application/json",
}

// Dredd expects that you have already set up the database and run all migrations before it begins.
// It will NOT initialize the database, only insert its test data.
// It does this in a way which ignores errors, which is fine on the ci, but might be an issue locally
// so look at the logs carefully if these tests fail and if in doubt re-init the db
// These hooks do NOT clean up after themselves and they produce a lot of database writes,
// so don't run this in production
func main() {

	h := hooks.NewHooks()
	server := hooks.NewServer(hooks.NewHooksRunner(h))

	//Get database connection info and create an admin who's token is used to execute the tests
	h.BeforeAll(func(t []*trans.Transaction) {
		loadConfig()
		addTestRunnerUser()
	})

	for _, v := range skipTests {
		h.Before(v, skipTest)
	}

	h.BeforeEach(func(t *trans.Transaction) {
		if strings.HasPrefix(t.Name, "user") {
			addCapabilities([]string{"user"})
		} else if strings.HasPrefix(t.Name, "project") || strings.HasPrefix(t.Name, "projects") {
			addCapabilities([]string{"project"})
		}
	})

	h.Before("user > /api/user/tokens/{api_token_id} > Expires API token > 204 > application/json", func(transaction *trans.Transaction) {
		dbConnect()
		defer store.Close()
		addToken(expiredToken, testRunnerUser.ID)
	})
	h.After("user > /api/user/tokens/{api_token_id} > Expires API token > 204 > application/json", func(transaction *trans.Transaction) {
		dbConnect()
		defer store.Close()
		//tokens are expired and not deleted so we need to clean up
		_ = store.DeleteAPIToken(testRunnerUser.ID, expiredToken)
	})

	// This one seems to need some manual value setting in the body
	h.Before("user > /api/users/{user_id}/password > Updates user password > 204 > application/json", func(transaction *trans.Transaction) {
		transaction.Request.Body = "{\"password\":\"staub\"}"
	})

	// delete the auto generated association and insert the user id into the query
	h.Before("project > /api/project/{project_id}/users > Link user to project > 204 > application/json", func(transaction *trans.Transaction) {
		dbConnect()
		defer store.Close()
		deleteUserProjectRelation(userProject.ID, userPathTestUser.ID)
		transaction.Request.Body = "{ \"user_id\": " + strconv.Itoa(userPathTestUser.ID) + ",\"admin\": true}"
	})

	h.Before("project > /api/project/{project_id}/keys/{key_id} > Updates access key > 204 > application/json", capabilityWrapper("access_key"))
	h.Before("project > /api/project/{project_id}/keys/{key_id} > Removes access key > 204 > application/json", capabilityWrapper("access_key"))

	h.Before("project > /api/project/{project_id}/repositories > Add repository > 204 > application/json", capabilityWrapper("access_key"))
	h.Before("project > /api/project/{project_id}/repositories/{repository_id} > Removes repository > 204 > application/json", capabilityWrapper("repository"))

	h.Before("project > /api/project/{project_id}/inventory > create inventory > 201 > application/json", capabilityWrapper("inventory"))
	h.Before("project > /api/project/{project_id}/inventory/{inventory_id} > Updates inventory > 204 > application/json", capabilityWrapper("inventory"))
	h.Before("project > /api/project/{project_id}/inventory/{inventory_id} > Removes inventory > 204 > application/json", capabilityWrapper("inventory"))

	h.Before("project > /api/project/{project_id}/environment/{environment_id} > Update environment > 204 > application/json", capabilityWrapper("environment"))
	h.Before("project > /api/project/{project_id}/environment/{environment_id} > Removes environment > 204 > application/json", capabilityWrapper("environment"))

	h.Before("project > /api/project/{project_id}/templates > create template > 201 > application/json", func(t *trans.Transaction) {
		addCapabilities([]string{"repository", "inventory", "environment", "view"})
	})

	h.Before("project > /api/project/{project_id}/templates/{template_id} > Get template > 200 > application/json", capabilityWrapper("template"))
	h.Before("project > /api/project/{project_id}/templates/{template_id} > Updates template > 204 > application/json", capabilityWrapper("template"))
	h.Before("project > /api/project/{project_id}/templates/{template_id} > Removes template > 204 > application/json", capabilityWrapper("template"))

	h.Before("project > /api/project/{project_id}/tasks > Starts a job > 201 > application/json", capabilityWrapper("template"))
	h.Before("project > /api/project/{project_id}/tasks/last > Get last 200 Tasks related to current project > 200 > application/json", capabilityWrapper("template"))

	h.Before("project > /api/project/{project_id}/tasks/{task_id} > Get a single task > 200 > application/json", capabilityWrapper("task"))
	h.Before("project > /api/project/{project_id}/tasks/{task_id} > Deletes task (including output) > 204 > application/json", capabilityWrapper("task"))
	h.Before("project > /api/project/{project_id}/tasks/{task_id}/output > Get task output > 200 > application/json", capabilityWrapper("task"))

	h.Before("schedule > /api/project/{project_id}/schedules/{schedule_id} > Get schedule > 200 > application/json", capabilityWrapper("schedule"))
	h.Before("schedule > /api/project/{project_id}/schedules/{schedule_id} > Updates schedule > 204 > application/json", capabilityWrapper("schedule"))
	h.Before("schedule > /api/project/{project_id}/schedules/{schedule_id} > Deletes schedule > 204 > application/json", capabilityWrapper("schedule"))

	h.Before("project > /api/project/{project_id}/views/{view_id} > Get view > 200 > application/json", capabilityWrapper("view"))
	h.Before("project > /api/project/{project_id}/views/{view_id} > Updates view > 204 > application/json", capabilityWrapper("view"))
	h.Before("project > /api/project/{project_id}/views/{view_id} > Removes view > 204 > application/json", capabilityWrapper("view"))

	//Add these last as they normalize the requests and path values after hook processing
	h.BeforeAll(func(transactions []*trans.Transaction) {
		for _, t := range transactions {
			h.Before(t.Name, setupObjectsAndPaths)
		}
	})

	// Delete the test runner user so adding him next time does not result in errors
	h.AfterAll(removeTestRunnerUser)

	server.Serve()
	defer server.Listener.Close()
}
