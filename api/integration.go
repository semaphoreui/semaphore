package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
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

// verifyIntegrationAuth checks the request against the integration's auth
// method and returns why it was rejected, empty when it passed.
func verifyIntegrationAuth(integration db.Integration, r *http.Request, payload []byte) string {
	secret := integration.AuthSecret.LoginPassword

	switch integration.AuthMethod {
	case db.IntegrationAuthGitHub:
		if !isValidHmacPayload(secret.Password, r.Header.Get("X-Hub-Signature-256"), payload, "sha256=", sha256.New) {
			return "Invalid GitHub/HMAC signature"
		}
	case db.IntegrationAuthBitbucket:
		if !isValidHmacPayload(secret.Password, r.Header.Get("x-hub-signature"), payload, "sha256=", sha256.New) {
			return "Invalid Bitbucket/HMAC signature"
		}
	case db.IntegrationAuthHmac:
		if !isValidHmacPayload(secret.Password, r.Header.Get(integration.AuthHeader), payload, "", sha256.New) {
			return "Invalid HMAC signature"
		}
	case db.IntegrationAuthHmacSha512:
		if !isValidHmacPayload(secret.Password, r.Header.Get(integration.AuthHeader), payload, "", sha512.New) {
			return "Invalid HMAC-SHA512 signature"
		}
	case db.IntegrationAuthToken:
		if secret.Password != r.Header.Get(integration.AuthHeader) {
			return "Invalid verification token"
		}
	case db.IntegrationAuthBasic:
		login, password, ok := r.BasicAuth()
		if !ok || secret.Password != password || secret.Login != login {
			return "Invalid BasicAuth: incorrect login or password"
		}
	case db.IntegrationAuthNone:
		// Do nothing
	default:
		return "Unknown verification method: " + string(integration.AuthMethod)
	}

	return ""
}

// recordCall stores one call in the project's integration history. A failure to
// store it is logged and otherwise ignored: the integration has already run.
func (c *IntegrationController) recordCall(
	call db.IntegrationCall, started time.Time, outcome db.IntegrationCallOutcome, reason string,
) {
	call.Created = started
	call.Duration = int(time.Since(started).Milliseconds())
	call.Outcome = outcome
	call.Reason = reason

	if err := c.store.CreateIntegrationCall(call); err != nil {
		log.WithFields(log.Fields{
			"context": "integrations",
			"alias":   call.Alias,
		}).WithError(err).Error("Failed to record the integration call")
	}
}

func (c *IntegrationController) ReceiveIntegration(w http.ResponseWriter, r *http.Request) {

	integrationAlias, ok := helpers.GetStrParamOrAbort("integration_alias", w, r)

	if !ok {
		return
	}

	received := time.Now()

	log.Info(fmt.Sprintf("Receiving Integration from: %s", r.RemoteAddr))

	integrations, level, projectID, err := c.store.GetIntegrationsByAlias(integrationAlias)

	if err != nil {
		log.Error(err)
		return
	}

	log.Info(fmt.Sprintf("%d integrations found for alias %s", len(integrations), integrationAlias))

	// Read only once the alias resolved: an unknown alias never gets its body read.
	payload, err := io.ReadAll(r.Body)

	if err != nil {
		log.Error(err)
		return
	}

	body := string(payload)
	base := db.IntegrationCall{
		ProjectID:   projectID,
		Alias:       integrationAlias,
		SourceIP:    sourceIP(r),
		Method:      r.Method,
		RequestSize: len(payload),
		RequestBody: &body,
	}

	if len(integrations) == 0 {
		// A real alias that starts nothing is the case a webhook author most needs to see.
		// An unknown alias resolves to no project and is deliberately left unrecorded, so a
		// public endpoint cannot be used to fill the table from the internet.
		c.recordCall(base, received, db.IntegrationCallSkipped, "No integration matched the alias")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	projects := make(map[int]db.Project)

	for _, integration := range integrations {

		started := time.Now()

		logger := log.WithFields(log.Fields{
			"context":        "integrations",
			"integration_id": integration.ID,
			"alias":          integrationAlias,
		})

		call := base
		call.ProjectID = integration.ProjectID
		call.IntegrationID = &integration.ID
		call.IntegrationName = integration.Name

		project, ok := projects[integration.ProjectID]
		if !ok {
			project, err = c.store.GetProject(integration.ProjectID)
			if err != nil {
				logger.WithError(err).Error("Could not retrieve the project")
				c.recordCall(call, started, db.IntegrationCallError, "Could not retrieve the project")
				continue
			}
			projects[integration.ProjectID] = project
		}

		if integration.ProjectID != project.ID {
			logger.WithField("project_id", project.ID).Error("integration project mismatch")
			c.recordCall(call, started, db.IntegrationCallError, "Integration project mismatch")
			continue
		}

		err = c.integrationService.FillIntegration(&integration)
		if err != nil {
			logger.WithError(err).Error("Could not load the integration secret")
			c.recordCall(call, started, db.IntegrationCallError, "Could not load the integration secret")
			continue
		}

		if reason := verifyIntegrationAuth(integration, r, payload); reason != "" {
			logger.Error(reason)
			c.recordCall(call, started, db.IntegrationCallUnauthorized, reason)
			continue
		}

		if level != db.IntegrationAliasSingle {
			var matchers []db.IntegrationMatcher
			matchers, err = c.store.GetIntegrationMatchers(integration.ProjectID, db.RetrieveQueryParams{}, integration.ID)
			if err != nil {
				logger.WithError(err).Error("Could not retrieve matchers")
				c.recordCall(call, started, db.IntegrationCallError, "Could not retrieve matchers")
				continue
			}

			var matched = false

			for _, matcher := range matchers {
				if Match(matcher, r.Header, payload) {
					matched = true
					continue
				} else {
					matched = false
					break
				}
			}

			if !matched {
				c.recordCall(call, started, db.IntegrationCallSkipped, "No matcher matched the request")
				continue
			}
		}

		result := c.RunIntegration(integration, project, r, payload)

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

			call.TaskID = &task.ID
		}

		if run := result.Run; run != nil {
			w.Header().Add("X-Semaphore-Workflow-ID", strconv.Itoa(run.WorkflowTemplateID))
			w.Header().Add("X-Semaphore-Workflow-Run-ID", strconv.Itoa(run.ID))
			w.Header().Add("X-Semaphore-Project-ID", strconv.Itoa(run.ProjectID))
			w.Header().Add("X-Semaphore-Integration-ID", strconv.Itoa(integration.ID))

			if run.RootTaskID != nil {
				w.Header().Add("X-Semaphore-Task-ID", strconv.Itoa(*run.RootTaskID))
			}

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

	w.WriteHeader(http.StatusNoContent)
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
