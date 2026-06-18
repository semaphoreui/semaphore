package main

import (
	"strconv"
	"strings"

	"github.com/snikch/goodman/hooks"
	trans "github.com/snikch/goodman/transaction"
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
	"runner > /api/project/{project_id}/runners > Add project runner > 201 > application/json",
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
		transaction.Request.Body = "{ \"user_id\": " + strconv.Itoa(userPathTestUser.ID) + ",\"role\": \"owner\"}"
	})

	h.Before("project > /api/project/{project_id}/invites > Get invitations for project > 200 > application/json", capabilityWrapper("invite"))
	h.Before("project > /api/project/{project_id}/invites > Create project invitation > 201 > application/json", capabilityWrapper("invite"))
	h.Before("project > /api/project/{project_id}/invites/{invite_id} > Get specific project invitation > 200 > application/json", capabilityWrapper("invite"))
	h.Before("project > /api/project/{project_id}/invites/{invite_id} > Update project invitation status > 204 > application/json", capabilityWrapper("invite"))
	h.Before("project > /api/project/{project_id}/invites/{invite_id} > Delete project invitation > 204 > application/json", capabilityWrapper("invite"))

	h.Before("integration > /api/project/{project_id}/integrations > get all integrations > 200 > application/json", capabilityWrapper("integration"))
	h.Before("integration > /api/project/{project_id}/integrations/{integration_id} > Get Integration > 200 > application/json", capabilityWrapper("integration"))
	h.Before("integration > /api/project/{project_id}/integrations/{integration_id} > Update Integration > 204 > application/json", capabilityWrapper("integration"))
	h.Before("integration > /api/project/{project_id}/integrations/{integration_id} > Remove integration > 204 > application/json", capabilityWrapper("integration"))

	h.Before("integration > /api/project/{project_id}/integrations/{integration_id}/values > Get Integration Extracted Values linked to integration extractor > 200 > application/json", capabilityWrapper("integrationextractvalue"))
	h.Before("integration > /api/project/{project_id}/integrations/{integration_id}/values > Add Integration Extracted Value > 204 > application/json", capabilityWrapper("integrationextractvalue"))
	h.Before("integration > /api/project/{project_id}/integrations/{integration_id}/values/{extractvalue_id} > Removes integration extract value > 204 > application/json", capabilityWrapper("integrationextractvalue"))
	h.Before("integration > /api/project/{project_id}/integrations/{integration_id}/values > Add Integration Extracted Value > 204 > application/json", capabilityWrapper("integration"))
	h.Before("integration > /api/project/{project_id}/integrations/{integration_id}/values/{extractvalue_id} > Updates Integration ExtractValue > 204 > application/json", capabilityWrapper("integrationextractvalue"))
	h.Before("integration > /api/project/{project_id}/integrations/{integration_id}/matchers > Get Integration Matcher linked to integration extractor > 200 > application/json", capabilityWrapper("integration"))
	h.Before("integration > /api/project/{project_id}/integrations/{integration_id}/matchers > Add Integration Matcher > 204 > application/json", capabilityWrapper("integration"))
	h.Before("integration > /api/project/{project_id}/integrations/{integration_id}/matchers/{matcher_id} > Updates Integration Matcher > 204 > application/json", capabilityWrapper("integrationmatcher"))

	h.Before("key-store > /api/project/{project_id}/keys > Add access key > 201 > application/json", capabilityWrapper("access_key"))
	h.Before("key-store > /api/project/{project_id}/keys/{key_id} > Updates access key > 204 > application/json", capabilityWrapper("access_key"))
	h.Before("key-store > /api/project/{project_id}/keys/{key_id} > Removes access key > 204 > application/json", capabilityWrapper("access_key"))

	h.Before("repository > /api/project/{project_id}/repositories > Add repository > 201 > application/json", capabilityWrapper("access_key"))
	h.Before("repository > /api/project/{project_id}/repositories/{repository_id} > Get repository > 200 > application/json", capabilityWrapper("repository"))
	h.Before("repository > /api/project/{project_id}/repositories/{repository_id} > Updates repository > 204 > application/json", capabilityWrapper("repository"))
	h.Before("repository > /api/project/{project_id}/repositories/{repository_id} > Removes repository > 204 > application/json", capabilityWrapper("repository"))

	h.Before("inventory > /api/project/{project_id}/inventory > create inventory > 201 > application/json", capabilityWrapper("inventory"))
	h.Before("inventory > /api/project/{project_id}/inventory/{inventory_id} > Get inventory > 200 > application/json", capabilityWrapper("inventory"))
	h.Before("inventory > /api/project/{project_id}/inventory/{inventory_id} > Updates inventory > 204 > application/json", capabilityWrapper("inventory"))
	h.Before("inventory > /api/project/{project_id}/inventory/{inventory_id} > Removes inventory > 204 > application/json", capabilityWrapper("inventory"))

	h.Before("variable-group > /api/project/{project_id}/environment > Add environment > 201 > application/json", capabilityWrapper("environment"))
	h.Before("variable-group > /api/project/{project_id}/environment/{environment_id} > Get environment > 200 > application/json", capabilityWrapper("environment"))
	h.Before("variable-group > /api/project/{project_id}/environment/{environment_id} > Update environment > 204 > application/json", capabilityWrapper("environment"))
	h.Before("variable-group > /api/project/{project_id}/environment/{environment_id} > Removes environment > 204 > application/json", capabilityWrapper("environment"))

	h.Before("template > /api/project/{project_id}/templates > create template > 201 > application/json", func(t *trans.Transaction) {
		addCapabilities([]string{"repository", "inventory", "environment", "view"})
	})

	h.Before("template > /api/project/{project_id}/templates/{template_id}/stop_all_tasks > Stop all active tasks of template > 204 > application/json", capabilityWrapper("template"))
	h.Before("template > /api/project/{project_id}/templates/{template_id} > Get template > 200 > application/json", capabilityWrapper("template"))
	h.Before("template > /api/project/{project_id}/templates/{template_id} > Updates template > 204 > application/json", capabilityWrapper("template"))
	h.Before("template > /api/project/{project_id}/templates/{template_id} > Removes template > 204 > application/json", capabilityWrapper("template"))

	h.Before("workflow > /api/project/{project_id}/workflows > Get workflows > 200 > application/json", capabilityWrapper("workflow"))
	h.Before("workflow > /api/project/{project_id}/workflows > Add workflow > 201 > application/json", capabilityWrapper("template"))
	h.Before("workflow > /api/project/{project_id}/workflows > Add workflow > 201 > application/json", func(t *trans.Transaction) {
		t.Request.Body = "{\"name\":\"workflow-doc-test\",\"nodes\":[{\"id\":1,\"template_id\":" + strconv.Itoa(templateID) + "},{\"id\":2,\"kind\":\"approval\"}],\"edges\":[{\"source_node_id\":1,\"destination_node_id\":2,\"condition\":\"on_success\"}]}"
	})
	h.Before("workflow > /api/project/{project_id}/workflows/{workflow_id} > Get workflow > 200 > application/json", capabilityWrapper("workflow"))
	h.Before("workflow > /api/project/{project_id}/workflows/{workflow_id} > Update workflow > 204 > application/json", capabilityWrapper("workflow"))
	h.Before("workflow > /api/project/{project_id}/workflows/{workflow_id} > Update workflow > 204 > application/json", func(t *trans.Transaction) {
		t.Request.Body = "{\"id\":" + strconv.Itoa(workflowID) + ",\"project_id\":" + strconv.Itoa(userProject.ID) + ",\"name\":\"workflow-updated\",\"nodes\":[{\"id\":1,\"template_id\":" + strconv.Itoa(templateID) + "},{\"id\":2,\"kind\":\"approval\",\"approval_timeout\":120}],\"edges\":[{\"source_node_id\":1,\"destination_node_id\":2,\"condition\":\"on_success\"}]}"
	})
	h.Before("workflow > /api/project/{project_id}/workflows/{workflow_id} > Remove workflow > 204 > application/json", capabilityWrapper("workflow"))
	h.Before("workflow > /api/project/{project_id}/workflows/{workflow_id}/run > Run workflow > 201 > application/json", capabilityWrapper("workflow"))
	h.Before("workflow > /api/project/{project_id}/workflows/{workflow_id}/runs > Get workflow runs > 200 > application/json", capabilityWrapper("workflow_run"))
	h.Before("workflow > /api/project/{project_id}/workflows/{workflow_id}/runs/{run_id} > Get workflow run details > 200 > application/json", capabilityWrapper("workflow_run"))
	h.Before("workflow > /api/project/{project_id}/workflows/{workflow_id}/runs/{run_id}/approvals > Get workflow run approvals > 200 > application/json", capabilityWrapper("workflow_approval"))
	h.Before("workflow > /api/project/{project_id}/workflows/{workflow_id}/runs/{run_id}/approvals/{node_id} > Resolve workflow approval > 200 > application/json", capabilityWrapper("workflow_approval"))
	h.Before("workflow > /api/project/{project_id}/workflows/{workflow_id}/runs/{run_id}/approvals/{node_id} > Resolve workflow approval > 200 > application/json", func(t *trans.Transaction) {
		t.Request.Body = "{\"status\":\"approved\"}"
	})

	h.Before("task > /api/project/{project_id}/tasks > Starts a job > 201 > application/json", capabilityWrapper("template"))
	h.Before("task > /api/project/{project_id}/tasks/last > Get last 200 Tasks related to current project > 200 > application/json", capabilityWrapper("template"))

	h.Before("task > /api/project/{project_id}/tasks/{task_id} > Get a single task > 200 > application/json", capabilityWrapper("task"))
	h.Before("task > /api/project/{project_id}/tasks/{task_id} > Deletes task (including output) > 204 > application/json", capabilityWrapper("task"))
	h.Before("task > /api/project/{project_id}/tasks/{task_id}/output > Get task output > 200 > application/json", capabilityWrapper("task"))
	h.Before("task > /api/project/{project_id}/tasks/{task_id}/raw_output > Get task raw output > 200 > text/plain; charset=utf-8", capabilityWrapper("task"))
	h.Before("task > /api/project/{project_id}/tasks/{task_id}/stop > Stop a job > 204 > application/json", capabilityWrapper("task"))

	h.Before("schedule > /api/project/{project_id}/schedules/{schedule_id} > Get schedule > 200 > application/json", capabilityWrapper("schedule"))
	h.Before("schedule > /api/project/{project_id}/schedules/{schedule_id} > Updates schedule > 204 > application/json", capabilityWrapper("schedule"))
	h.Before("schedule > /api/project/{project_id}/schedules/{schedule_id} > Deletes schedule > 204 > application/json", capabilityWrapper("schedule"))

	h.Before("project > /api/project/{project_id}/views/{view_id} > Get view > 200 > application/json", capabilityWrapper("view"))
	h.Before("project > /api/project/{project_id}/views/{view_id} > Updates view > 204 > application/json", capabilityWrapper("view"))
	h.Before("project > /api/project/{project_id}/views/{view_id} > Removes view > 204 > application/json", capabilityWrapper("view"))

	h.Before("project > /api/project/{project_id}/backup > Get backup > 200 > application/json", func(t *trans.Transaction) {
		addCapabilities([]string{"repository", "inventory", "environment", "view", "template"})
	})

	// project runners
	h.Before("runner > /api/project/{project_id}/runners > Get project runners > 200 > application/json", capabilityWrapper("project"))
	//h.Before("runner > /api/project/{project_id}/runners > Add project runner > 201 > application/json", capabilityWrapper("project"))
	h.Before("runner > /api/project/{project_id}/runner_tags > Get project runner tags > 200 > application/json", capabilityWrapper("project"))
	h.Before("runner > /api/project/{project_id}/runners/{runner_id} > Get project runner > 200 > application/json", capabilityWrapper("runner"))
	h.Before("runner > /api/project/{project_id}/runners/{runner_id} > Update project runner > 204 > application/json", capabilityWrapper("runner"))
	h.Before("runner > /api/project/{project_id}/runners/{runner_id} > Delete project runner > 204 > application/json", capabilityWrapper("runner"))
	h.Before("runner > /api/project/{project_id}/runners/{runner_id}/active > Set project runner active state > 204 > application/json", capabilityWrapper("runner"))
	h.Before("runner > /api/project/{project_id}/runners/{runner_id}/cache > Clear project runner cache > 204 > application/json", capabilityWrapper("runner"))

	// global runners (admin)
	h.Before("runner > /api/runners/{runner_id} > Get global runner > 200 > application/json", capabilityWrapper("global_runner"))
	h.Before("runner > /api/runners/{runner_id} > Update global runner > 204 > application/json", capabilityWrapper("global_runner"))
	h.Before("runner > /api/runners/{runner_id} > Delete global runner > 204 > application/json", capabilityWrapper("global_runner"))
	h.Before("runner > /api/runners/{runner_id}/active > Set global runner active state > 204 > application/json", capabilityWrapper("global_runner"))
	h.Before("runner > /api/runners/{runner_id}/cache > Clear global runner cache > 204 > application/json", capabilityWrapper("global_runner"))

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
