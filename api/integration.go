package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"maps"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/semaphoreui/semaphore/pkg/conv"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/server"
	task2 "github.com/semaphoreui/semaphore/services/tasks"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	log "github.com/sirupsen/logrus"
	"github.com/thedevsaddam/gojsonq/v2"
)

// isValidHmacPayload checks if the payload's hash fits with
// the hash computed by the sender and sent as a header
func isValidHmacPayload(secret, headerHash string, payload []byte, prefix string, hashFunc func() hash.Hash) bool {
	hash := hmacHashPayload(secret, payload, hashFunc)

	if !strings.HasPrefix(headerHash, prefix) {
		return false
	}

	headerHash = headerHash[len(prefix):]

	return hmac.Equal(
		[]byte(hash),
		[]byte(headerHash),
	)
}

// hmacHashPayload computes the hash of payload's body according to the webhook's secret token
// see https://developer.github.com/webhooks/securing/#validating-payloads-from-github
// returning the hash as a hexadecimal string
func hmacHashPayload(secret string, payloadBody []byte, hashFunc func() hash.Hash) string {
	hm := hmac.New(hashFunc, []byte(secret))
	hm.Write(payloadBody)
	sum := hm.Sum(nil)
	return fmt.Sprintf("%x", sum)
}

type IntegrationController struct {
	store              db.Store
	integrationService server.IntegrationService
	workflowRepo       db.WorkflowManager
	workflowService    pro_interfaces.WorkflowService
}

func NewIntegrationController(
	store db.Store,
	integrationService server.IntegrationService,
	workflowRepo db.WorkflowManager,
	workflowService pro_interfaces.WorkflowService,
) *IntegrationController {
	return &IntegrationController{
		store:              store,
		integrationService: integrationService,
		workflowRepo:       workflowRepo,
		workflowService:    workflowService,
	}
}

// IntegrationRunResult is what one matching integration started: a task for
// a template integration, a workflow run for a workflow integration, neither
// when the start failed.
type IntegrationRunResult struct {
	Task *db.Task
	Run  *db.WorkflowRun
	// Reason is why nothing started, empty when a task or a run did. It is
	// logged and stored with the call so a failed webhook can be debugged.
	Reason string
}

// sourceIP is the address the call came from: the one the audit middleware
// resolved against the trusted proxies when audit logging is on, otherwise the
// peer address of the connection.
func sourceIP(r *http.Request) string {
	if info, ok := audit.RequestFrom(r.Context()); ok && info.IP != "" {
		return info.IP
	}

	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}

	return r.RemoteAddr
}

// recordCall stores one call in the project's integration history. A failure to
// store it is logged and otherwise ignored: the integration has already run.
// The alias is a bearer secret and stays out of the log.
func (c *IntegrationController) recordCall(
	call db.IntegrationCall, started time.Time, outcome db.IntegrationCallOutcome, reason string,
) {
	call.Created = started
	call.Duration = int(time.Since(started).Milliseconds())
	call.Outcome = outcome
	call.Reason = reason

	if err := c.store.CreateIntegrationCall(call); err != nil {
		log.WithFields(log.Fields{
			"context":    "integrations",
			"project_id": call.ProjectID,
		}).WithError(err).Error("Failed to record the integration call")
	}
}

// verifyAuth reports whether the request passes auth, whose key must already
// be loaded (AuthSecret). The reason names what failed for the log and the
// call history; it is never sent to the caller, who always gets 204. Every
// secret comparison is constant-time, so a wrong token or password costs the
// same as a right one.
func verifyAuth(r *http.Request, payload []byte, auth db.IntegrationAuth) (ok bool, reason string) {
	// A key that cannot provide the shared secret refuses the request here,
	// before the method-specific check.
	secret, ok := auth.VerificationSecret()
	if !ok {
		return false, "Auth secret is empty or its key type does not fit the auth method"
	}

	switch auth.AuthMethod {
	case db.IntegrationAuthGitHub:
		if !isValidHmacPayload(secret, r.Header.Get("X-Hub-Signature-256"), payload, "sha256=", sha256.New) {
			return false, "Invalid GitHub signature"
		}
	case db.IntegrationAuthBitbucket:
		if !isValidHmacPayload(secret, r.Header.Get("x-hub-signature"), payload, "sha256=", sha256.New) {
			return false, "Invalid Bitbucket signature"
		}
	case db.IntegrationAuthHmac:
		if !isValidHmacPayload(secret, r.Header.Get(auth.AuthHeader), payload, "", sha256.New) {
			return false, "Invalid HMAC-SHA256 signature"
		}
	case db.IntegrationAuthHmacSha512:
		if !isValidHmacPayload(secret, r.Header.Get(auth.AuthHeader), payload, "", sha512.New) {
			return false, "Invalid HMAC-SHA512 signature"
		}
	case db.IntegrationAuthToken:
		if !hmac.Equal([]byte(secret), []byte(r.Header.Get(auth.AuthHeader))) {
			return false, "Invalid verification token"
		}
	case db.IntegrationAuthBasic:
		if auth.AuthSecret.Type != db.AccessKeyLoginPassword {
			return false, "BasicAuth needs a login/password credential"
		}
		if auth.AuthSecret.LoginPassword.Password == "" {
			return false, "BasicAuth credential has no password"
		}
		login, password, present := r.BasicAuth()
		if !present {
			return false, "No BasicAuth credentials in the request"
		}
		// Both are compared, whatever the first says, so the response time
		// does not tell a wrong login from a wrong password.
		loginOk := subtle.ConstantTimeCompare([]byte(login), []byte(auth.AuthSecret.LoginPassword.Login))
		passwordOk := subtle.ConstantTimeCompare([]byte(password), []byte(auth.AuthSecret.LoginPassword.Password))
		if loginOk&passwordOk != 1 {
			return false, "Invalid BasicAuth: incorrect login or password"
		}
	case db.IntegrationAuthNone:
		// Nothing to check.
	default:
		return false, "Unknown verification method: " + string(auth.AuthMethod)
	}

	return true, ""
}

// matchAll reports whether every matcher accepts the request. An integration
// without matchers reacts to every request its group lets through.
func matchAll(matchers []db.IntegrationMatcher, header http.Header, payload []byte) bool {
	for _, matcher := range matchers {
		if !Match(matcher, header, payload) {
			return false
		}
	}

	return true
}

// authCheck is what checking the request against one auth configuration
// gave. When the request did not pass, outcome and reason are what the call
// history records for every integration behind that configuration.
type authCheck struct {
	ok      bool
	outcome db.IntegrationCallOutcome
	reason  string
}

// authorize loads the key of auth and checks the request against it. A
// failure is logged and recorded in the call history, never reported to the
// caller.
func (c *IntegrationController) authorize(r *http.Request, payload []byte, projectID int, auth db.IntegrationAuth, logger *log.Entry) authCheck {
	if err := c.integrationService.FillAuthSecret(projectID, &auth); err != nil {
		logger.WithError(err).Error("Could not load the auth secret")
		return authCheck{outcome: db.IntegrationCallError, reason: "Could not load the auth secret"}
	}

	ok, reason := verifyAuth(r, payload, auth)
	if !ok {
		logger.Warn("Webhook refused: " + reason)
		return authCheck{outcome: db.IntegrationCallUnauthorized, reason: reason}
	}

	return authCheck{ok: true}
}

// ReceiveIntegration is the public webhook: the alias names an endpoint,
// every integration group of that endpoint checks the request against its
// effective auth (the endpoint's, verified once per request, or its own), and
// each integration of an authorized group whose matchers all accept the
// request runs. Whatever happens the caller gets 204: the alias is a bearer
// secret and the response must not help guessing anything behind it.
//
// Every integration the call reached gets one row in the project's call
// history, whatever it did with the call. A call to an unknown alias is not
// recorded: the route is public, and recording it would let anyone fill the
// table. The request body is kept only for a call that passed authorization.
func (c *IntegrationController) ReceiveIntegration(w http.ResponseWriter, r *http.Request) {

	alias, ok := helpers.GetStrParamOrAbort("integration_alias", w, r)
	if !ok {
		return
	}

	received := time.Now()

	// The alias itself is never logged.
	logger := log.WithFields(log.Fields{
		"context":     "integrations",
		"remote_addr": r.RemoteAddr,
	})
	logger.Info("Receiving integration webhook")

	endpoint, err := c.store.GetEndpointByAlias(alias)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			logger.Warn("Unknown integration alias")
		} else {
			logger.WithError(err).Error("Could not resolve the integration alias")
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	logger = logger.WithFields(log.Fields{
		"project_id":  endpoint.ProjectID,
		"endpoint_id": endpoint.ID,
	})

	// Read only once the alias resolved: an unknown alias never gets its body read.
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		logger.WithError(err).Error("Could not read the webhook body")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// unverified is a call nothing has authorized yet, so it carries no body:
	// anyone who knows the URL can send one.
	unverified := db.IntegrationCall{
		ProjectID:   endpoint.ProjectID,
		Alias:       alias,
		SourceIP:    sourceIP(r),
		Method:      r.Method,
		RequestSize: len(payload),
	}

	body := string(payload)
	verified := unverified
	verified.RequestBody = &body

	project, err := c.store.GetProject(endpoint.ProjectID)
	if err != nil {
		logger.WithError(err).Error("Could not load the endpoint's project")
		c.recordCall(unverified, received, db.IntegrationCallError, "Could not retrieve the project")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	groups, err := c.store.GetIntegrationGroupsByEndpoint(endpoint.ProjectID, endpoint.ID)
	if err != nil {
		logger.WithError(err).Error("Could not load the endpoint's integration groups")
		c.recordCall(unverified, received, db.IntegrationCallError, "Could not retrieve the integration groups")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// The endpoint's own auth is checked at most once per request, when the
	// first thing that needs it comes up.
	var endpointAuth *authCheck
	checkEndpointAuth := func() authCheck {
		if endpointAuth == nil {
			check := c.authorize(r, payload, endpoint.ProjectID, endpoint.IntegrationAuth, logger)
			endpointAuth = &check
		}
		return *endpointAuth
	}

	recorded := false

	for _, group := range groups {
		groupStarted := time.Now()
		groupLogger := logger.WithField("integration_group_id", group.ID)

		integrations, err := c.store.GetIntegrationsByGroup(group.ProjectID, group.ID, true)
		if err != nil {
			groupLogger.WithError(err).Error("Could not load the group's integrations")
			c.recordCall(unverified, groupStarted, db.IntegrationCallError, "Could not retrieve the integrations of an integration group")
			recorded = true
			continue
		}

		if len(integrations) == 0 {
			// Nothing can run, so there is nothing to authorize or to record.
			continue
		}

		var auth authCheck
		if group.AuthOverride {
			auth = c.authorize(r, payload, group.ProjectID, group.IntegrationAuth, groupLogger)
		} else {
			auth = checkEndpointAuth()
		}

		if !auth.ok {
			for _, integration := range integrations {
				call := unverified
				call.IntegrationID = &integration.ID
				call.IntegrationName = integration.Name
				c.recordCall(call, groupStarted, auth.outcome, auth.reason)
			}
			recorded = true
			continue
		}

		for _, integration := range integrations {
			started := time.Now()

			call := verified
			call.IntegrationID = &integration.ID
			call.IntegrationName = integration.Name
			recorded = true

			matchers, err := c.store.GetIntegrationMatchers(integration.ProjectID, db.RetrieveQueryParams{}, integration.ID)
			if err != nil {
				groupLogger.WithField("integration_id", integration.ID).WithError(err).Error("Could not retrieve matchers")
				c.recordCall(call, started, db.IntegrationCallError, "Could not retrieve matchers")
				continue
			}

			if !matchAll(matchers, r.Header, payload) {
				c.recordCall(call, started, db.IntegrationCallSkipped, "No matcher matched the request")
				continue
			}

			result := c.RunIntegration(integration, project, r, payload)
			writeRunHeaders(w, endpoint, group, integration, result)

			if task := result.Task; task != nil {
				call.TaskID = &task.ID
			}

			if run := result.Run; run != nil {
				call.RunID = &run.ID
				call.WorkflowID = &run.WorkflowTemplateID
				call.TaskID = run.RootTaskID
			}

			if result.Reason != "" {
				c.recordCall(call, started, db.IntegrationCallError, result.Reason)
			} else {
				c.recordCall(call, started, db.IntegrationCallOK, "")
			}
		}
	}

	if !recorded {
		// A real alias that starts nothing is the case a webhook author most
		// needs to see. The endpoint's auth still decides whether the body is
		// kept and how the call is named.
		if auth := checkEndpointAuth(); auth.ok {
			c.recordCall(verified, received, db.IntegrationCallSkipped, "The endpoint has no integrations")
		} else {
			c.recordCall(unverified, received, auth.outcome, auth.reason)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// writeRunHeaders tells the caller what one integration started. Nothing is
// written for an integration that did not run.
func writeRunHeaders(w http.ResponseWriter, endpoint db.Endpoint, group db.IntegrationGroup, integration db.Integration, result IntegrationRunResult) {
	if result.Task == nil && result.Run == nil {
		return
	}

	w.Header().Set("X-Semaphore-Endpoint-ID", strconv.Itoa(endpoint.ID))
	w.Header().Add("X-Semaphore-Integration-Group-ID", strconv.Itoa(group.ID))

	if task := result.Task; task != nil {
		w.Header().Add("X-Semaphore-Task-ID", strconv.Itoa(task.ID))
		w.Header().Add("X-Semaphore-Template-ID", strconv.Itoa(task.TemplateID))
		w.Header().Add("X-Semaphore-Project-ID", strconv.Itoa(task.ProjectID))

		if task.IntegrationID != nil {
			w.Header().Add("X-Semaphore-Integration-ID", strconv.Itoa(*task.IntegrationID))
		}

		if task.InventoryID != nil {
			w.Header().Add("X-Semaphore-Inventory-ID", strconv.Itoa(*task.InventoryID))
		}
	}

	if run := result.Run; run != nil {
		w.Header().Add("X-Semaphore-Workflow-ID", strconv.Itoa(run.WorkflowTemplateID))
		w.Header().Add("X-Semaphore-Workflow-Run-ID", strconv.Itoa(run.ID))
		w.Header().Add("X-Semaphore-Project-ID", strconv.Itoa(run.ProjectID))
		w.Header().Add("X-Semaphore-Integration-ID", strconv.Itoa(integration.ID))

		if run.RootTaskID != nil {
			w.Header().Add("X-Semaphore-Task-ID", strconv.Itoa(*run.RootTaskID))
		}
	}
}

func Match(matcher db.IntegrationMatcher, header http.Header, bodyBytes []byte) (matched bool) {

	switch matcher.MatchType {
	case db.IntegrationMatchHeader:
		return MatchCompare(header.Get(matcher.Key), matcher.Method, matcher.Value)
	case db.IntegrationMatchBody:
		var body = string(bodyBytes)
		switch matcher.BodyDataType {
		case db.IntegrationBodyDataJSON:
			value := gojsonq.New().JSONString(body).Find(matcher.Key)

			return MatchCompare(value, matcher.Method, matcher.Value)
		case db.IntegrationBodyDataString:
			return MatchCompare(body, matcher.Method, matcher.Value)
		}
	}

	return false
}

func MatchCompare(value any, method db.IntegrationMatchMethodType, expected string) bool {

	if intValue, ok := conv.ConvertFloatToIntIfPossible(value); ok {
		value = intValue
	}

	strValue := fmt.Sprintf("%v", value)

	switch method {
	case db.IntegrationMatchMethodEquals:
		return strValue == expected
	case db.IntegrationMatchMethodUnEquals:
		return strValue != expected
	case db.IntegrationMatchMethodContains:
		return strings.Contains(fmt.Sprintf("%v", value), expected)
	default:
		return false
	}
}

func GetTaskDefinition(
	integration db.Integration,
	payload []byte,
	h http.Header,
	extractorCreator func(projectID, integrationID int) ([]db.IntegrationExtractValue, error),
) (taskDefinition db.Task, err error) {

	var envValues = make([]db.IntegrationExtractValue, 0)
	var taskValues = make([]db.IntegrationExtractValue, 0)

	extractValuesForExtractor, err := extractorCreator(integration.ProjectID, integration.ID)
	if err != nil {
		return
	}

	for _, val := range extractValuesForExtractor {
		switch val.VariableType {
		case "", db.IntegrationVariableEnvironment: // "" handles null/empty for backward compatibility
			envValues = append(envValues, val)
		case db.IntegrationVariableTaskParam:
			taskValues = append(taskValues, val)
		}
	}

	var extractedEnvResults = Extract(envValues, h, payload)

	// Only a template integration gets here (a workflow one goes through
	// GetWorkflowRunParams); a missing template id is left as zero.
	templateID := 0
	if integration.TemplateID != nil {
		templateID = *integration.TemplateID
	}

	if integration.TaskParams != nil {
		taskDefinition = integration.TaskParams.CreateTask(templateID)
	} else {
		taskDefinition = db.Task{
			ProjectID:  integration.ProjectID,
			TemplateID: templateID,
			Params:     make(db.MapStringAnyField),
		}
	}

	taskDefinition.IntegrationID = &integration.ID

	env := make(map[string]any)

	if taskDefinition.Environment != "" {
		err = json.Unmarshal([]byte(taskDefinition.Environment), &env)
		if err != nil {
			return
		}
	}

	maps.Copy(env, extractedEnvResults)

	envStr, err := json.Marshal(env)
	if err != nil {
		return
	}

	taskDefinition.Environment = string(envStr)

	if taskDefinition.Params == nil {
		taskDefinition.Params = make(db.MapStringAnyField)
	}

	extractedTaskResults := Extract(taskValues, h, payload)
	maps.Copy(taskDefinition.Params, extractedTaskResults)

	return
}

// applyGitBranchParam moves the reserved `git_branch` task parameter onto
// task.GitBranch, so a webhook sets the branch the same way the task dialog
// does. The key is removed from Params so it never reaches the app-specific
// params. Callers must check the template's AllowOverrideBranchInTask first.
// The value is validated by ValidateNewTask like any task branch.
func applyGitBranchParam(task *db.Task, allowOverrideGitBranch bool) {
	raw, ok := task.Params[db.TaskParamGitBranch]
	if !ok {
		return
	}
	delete(task.Params, db.TaskParamGitBranch)

	if !allowOverrideGitBranch {
		log.WithFields(log.Fields{
			"context":       "integration",
			"project_id":    task.ProjectID,
			"template_name": task.TemplateName,
			"task_id":       task.ID,
		}).Debug("Skipping Git branch from params because doesn't allow override git branch")
		return
	}

	branch, _ := raw.(string)
	branch = strings.TrimPrefix(strings.TrimSpace(branch), "refs/heads/")
	if branch == "" {
		return
	}

	task.GitBranch = &branch
}

// GetWorkflowRunParams builds the start parameters of a workflow run from a
// webhook: every extracted value, whatever its variable type, becomes an
// answer to the workflow's survey under the extractor's variable name. The
// engine validates the answers against the survey (defaults, required,
// types) and delivers them to nodes through their explicit bindings.
func GetWorkflowRunParams(
	integration db.Integration,
	payload []byte,
	h http.Header,
	extractorCreator func(projectID, integrationID int) ([]db.IntegrationExtractValue, error),
) (params db.WorkflowRunParams, err error) {
	extractValues, err := extractorCreator(integration.ProjectID, integration.ID)
	if err != nil {
		return
	}

	params.IntegrationID = &integration.ID

	extracted := Extract(extractValues, h, payload)
	if len(extracted) > 0 {
		params.Environment = db.MapStringAnyField(extracted)
	}

	return
}

func (c *IntegrationController) RunIntegration(integration db.Integration, project db.Project, r *http.Request, payload []byte) (result IntegrationRunResult) {
	log.Info(fmt.Sprintf("Running integration %d", integration.ID))

	if integration.RunsWorkflow() {
		result.Run, result.Reason = c.runWorkflowIntegration(integration, r, payload)
		return
	}

	result.Task, result.Reason = c.runTemplateIntegration(integration, r, payload)
	return
}

func (c *IntegrationController) runTemplateIntegration(integration db.Integration, r *http.Request, payload []byte) (*db.Task, string) {
	logger := log.WithFields(log.Fields{
		"context":        "integrations",
		"integration_id": integration.ID,
	})

	if integration.TemplateID == nil {
		logger.Error("integration has no template")
		return nil, "Integration has no template"
	}

	taskDefinition, err := GetTaskDefinition(
		integration, payload, r.Header, func(projectID, integrationID int) ([]db.IntegrationExtractValue, error) {
			return c.store.GetIntegrationExtractValues(projectID, db.RetrieveQueryParams{}, integrationID)
		})
	if err != nil {
		logger.WithError(err).Error("Failed to get task definition")
		return nil, "Failed to build the task definition"
	}

	tpl, err := c.store.GetTemplate(integration.ProjectID, *integration.TemplateID)
	if err != nil {
		logger.WithError(err).Error("Failed to get template")
		return nil, "Failed to retrieve the template"
	}

	applyGitBranchParam(&taskDefinition, tpl.AllowOverrideBranchInTask)

	pool := helpers.GetFromContext(r, "task_pool").(*task2.TaskPool)

	// The webhook caller authenticates as the integration, and its address stays in the request context.
	ctx := audit.WithActor(r.Context(), audit.IntegrationActor(integration.ID, integration.Name))
	task, err := pool.AddTaskFrom(ctx, audit.TriggerIntegration, taskDefinition, nil, "", integration.ProjectID, tpl.App.NeedTaskAlias())
	if err != nil {
		logger.WithError(err).Error("Failed to add task")
		return nil, "Failed to start the task: " + err.Error()
	}

	return &task, ""
}

// runWorkflowIntegration starts a run of the integration's workflow. The
// start is attributed to the integration (audit actor, task.integration_id
// of the root task); a start the engine refuses — a required survey variable
// the webhook did not provide, a workflow without a root — is logged, and
// the webhook still answers 204 like every other silently skipped
// integration.
func (c *IntegrationController) runWorkflowIntegration(integration db.Integration, r *http.Request, payload []byte) (*db.WorkflowRun, string) {
	logger := log.WithFields(log.Fields{
		"context":        "integrations",
		"integration_id": integration.ID,
		"workflow_id":    *integration.WorkflowID,
	})

	workflow, err := c.workflowRepo.GetWorkflowTemplate(integration.ProjectID, *integration.WorkflowID)
	if err != nil {
		logger.WithError(err).Error("Failed to load the integration's workflow")
		return nil, "Failed to load the workflow"
	}
	if workflow.ID == 0 {
		// The open-source workflow store answers every lookup with an empty
		// template and no error: workflows are a Pro feature.
		logger.Error("Workflows are not available in this build")
		return nil, "Workflows are not available in this build"
	}

	params, err := GetWorkflowRunParams(
		integration, payload, r.Header, func(projectID, integrationID int) ([]db.IntegrationExtractValue, error) {
			return c.store.GetIntegrationExtractValues(projectID, db.RetrieveQueryParams{}, integrationID)
		})
	if err != nil {
		logger.WithError(err).Error("Failed to get workflow run params")
		return nil, "Failed to build the workflow run parameters"
	}

	ctx := audit.WithActor(r.Context(), audit.IntegrationActor(integration.ID, integration.Name))
	run, err := c.workflowService.StartWorkflow(ctx, workflow, params, nil)
	if err != nil {
		logger.WithError(err).Error("Failed to start workflow")
		return nil, "Failed to start the workflow: " + err.Error()
	}

	return &run, ""
}

func Extract(extractValues []db.IntegrationExtractValue, h http.Header, payload []byte) (result map[string]any) {
	result = make(map[string]any)

	for _, extractValue := range extractValues {
		switch extractValue.ValueSource {
		case db.IntegrationExtractHeaderValue:
			result[extractValue.Variable] = h.Get(extractValue.Key)
		case db.IntegrationExtractBodyValue:
			switch extractValue.BodyDataType {
			case db.IntegrationBodyDataJSON:
				val := gojsonq.New().JSONString(string(payload)).Find(extractValue.Key)
				if val != nil {
					result[extractValue.Variable] = val
				}
			case db.IntegrationBodyDataString:
				result[extractValue.Variable] = string(payload)
			}
		}
	}
	return
}
