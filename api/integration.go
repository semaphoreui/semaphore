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
	"net/http"
	"strconv"
	"strings"

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
// when the start failed (the reason is in the log).
type IntegrationRunResult struct {
	Task *db.Task
	Run  *db.WorkflowRun
}

func (c *IntegrationController) ReceiveIntegration(w http.ResponseWriter, r *http.Request) {

	integrationAlias, ok := helpers.GetStrParamOrAbort("integration_alias", w, r)

	if !ok {
		return
	}

	log.Info(fmt.Sprintf("Receiving Integration from: %s", r.RemoteAddr))

	integrations, level, err := c.store.GetIntegrationsByAlias(integrationAlias)

	if err != nil {
		log.Error(err)
		return
	}

	log.Info(fmt.Sprintf("%d integrations found for alias %s", len(integrations), integrationAlias))

	projects := make(map[int]db.Project)

	var payload []byte

	payload, err = io.ReadAll(r.Body)

	if err != nil {
		log.Error(err)
		return
	}

	for _, integration := range integrations {

		project, ok := projects[integration.ProjectID]
		if !ok {
			project, err = c.store.GetProject(integrations[0].ProjectID)
			if err != nil {
				log.Error(err)
				return
			}
			projects[integration.ProjectID] = project
		}

		if integration.ProjectID != project.ID {
			log.WithFields(log.Fields{
				"context":       "integrations",
				"project_id":    project.ID,
				"integrationId": integration.ID,
			}).Error("integration project mismatch")
			continue
		}

		err = c.integrationService.FillIntegration(&integration)
		if err != nil {
			log.Error(err)
			return
		}

		switch integration.AuthMethod {
		case db.IntegrationAuthGitHub:
			ok := isValidHmacPayload(
				integration.AuthSecret.LoginPassword.Password,
				r.Header.Get("X-Hub-Signature-256"),
				payload,
				"sha256=",
				sha256.New)

			if !ok {
				log.WithFields(log.Fields{
					"context": "integrations",
				}).Error("Invalid GitHub/HMAC signature")
				continue
			}
		case db.IntegrationAuthBitbucket:
			ok := isValidHmacPayload(
				integration.AuthSecret.LoginPassword.Password,
				r.Header.Get("x-hub-signature"),
				payload,
				"sha256=",
				sha256.New)

			if !ok {
				log.WithFields(log.Fields{
					"context": "integrations",
				}).Error("Invalid Bitbucket/HMAC signature")
				continue
			}
		case db.IntegrationAuthHmac:
			ok := isValidHmacPayload(
				integration.AuthSecret.LoginPassword.Password,
				r.Header.Get(integration.AuthHeader),
				payload,
				"",
				sha256.New)

			if !ok {
				log.WithFields(log.Fields{
					"context": "integrations",
				}).Error("Invalid HMAC signature")
				continue
			}
		case db.IntegrationAuthHmacSha512:
			ok := isValidHmacPayload(
				integration.AuthSecret.LoginPassword.Password,
				r.Header.Get(integration.AuthHeader),
				payload,
				"",
				sha512.New)

			if !ok {
				log.WithFields(log.Fields{
					"context": "integrations",
				}).Error("Invalid HMAC-SHA512 signature")
				continue
			}
		case db.IntegrationAuthToken:
			if integration.AuthSecret.LoginPassword.Password != r.Header.Get(integration.AuthHeader) {
				log.WithFields(log.Fields{
					"context": "integrations",
				}).Error("Invalid verification token")
				continue
			}
		case db.IntegrationAuthBasic:
			var username, password, auth = r.BasicAuth()
			if !auth || integration.AuthSecret.LoginPassword.Password != password || integration.AuthSecret.LoginPassword.Login != username {
				log.WithFields(log.Fields{
					"context": "integrations",
				}).Error("Invalid BasicAuth: incorrect login or password")
				continue
			}
		case db.IntegrationAuthNone:
			// Do nothing
		default:
			log.WithFields(log.Fields{
				"context": "integrations",
			}).Error("Unknown verification method: " + integration.AuthMethod)
			continue
		}

		if level != db.IntegrationAliasSingle {
			var matchers []db.IntegrationMatcher
			matchers, err = c.store.GetIntegrationMatchers(integration.ProjectID, db.RetrieveQueryParams{}, integration.ID)
			if err != nil {
				log.WithFields(log.Fields{
					"context": "integrations",
				}).WithError(err).Error("Could not retrieve matchers")
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
		result.Run = c.runWorkflowIntegration(integration, r, payload)
		return
	}

	result.Task = c.runTemplateIntegration(integration, r, payload)
	return
}

func (c *IntegrationController) runTemplateIntegration(integration db.Integration, r *http.Request, payload []byte) *db.Task {
	logger := log.WithFields(log.Fields{
		"context":        "integrations",
		"integration_id": integration.ID,
	})

	if integration.TemplateID == nil {
		logger.Error("integration has no template")
		return nil
	}

	taskDefinition, err := GetTaskDefinition(
		integration, payload, r.Header, func(projectID, integrationID int) ([]db.IntegrationExtractValue, error) {
			return c.store.GetIntegrationExtractValues(projectID, db.RetrieveQueryParams{}, integrationID)
		})
	if err != nil {
		logger.WithError(err).Error("Failed to get task definition")
		return nil
	}

	tpl, err := c.store.GetTemplate(integration.ProjectID, *integration.TemplateID)
	if err != nil {
		log.Error(err)
		return nil
	}

	applyGitBranchParam(&taskDefinition, tpl.AllowOverrideBranchInTask)

	pool := helpers.GetFromContext(r, "task_pool").(*task2.TaskPool)

	// The webhook caller authenticates as the integration, and its address stays in the request context.
	ctx := audit.WithActor(r.Context(), audit.IntegrationActor(integration.ID, integration.Name))
	task, err := pool.AddTaskFrom(ctx, audit.TriggerIntegration, taskDefinition, nil, "", integration.ProjectID, tpl.App.NeedTaskAlias())
	if err != nil {
		log.Error(err)
		return nil
	}

	return &task
}

// runWorkflowIntegration starts a run of the integration's workflow. The
// start is attributed to the integration (audit actor, task.integration_id
// of the root task); a start the engine refuses — a required survey variable
// the webhook did not provide, a workflow without a root — is logged, and
// the webhook still answers 204 like every other silently skipped
// integration.
func (c *IntegrationController) runWorkflowIntegration(integration db.Integration, r *http.Request, payload []byte) *db.WorkflowRun {
	logger := log.WithFields(log.Fields{
		"context":        "integrations",
		"integration_id": integration.ID,
		"workflow_id":    *integration.WorkflowID,
	})

	workflow, err := c.workflowRepo.GetWorkflowTemplate(integration.ProjectID, *integration.WorkflowID)
	if err != nil {
		logger.WithError(err).Error("Failed to load the integration's workflow")
		return nil
	}
	if workflow.ID == 0 {
		// The open-source workflow store answers every lookup with an empty
		// template and no error: workflows are a Pro feature.
		logger.Error("Workflows are not available in this build")
		return nil
	}

	params, err := GetWorkflowRunParams(
		integration, payload, r.Header, func(projectID, integrationID int) ([]db.IntegrationExtractValue, error) {
			return c.store.GetIntegrationExtractValues(projectID, db.RetrieveQueryParams{}, integrationID)
		})
	if err != nil {
		logger.WithError(err).Error("Failed to get workflow run params")
		return nil
	}

	ctx := audit.WithActor(r.Context(), audit.IntegrationActor(integration.ID, integration.Name))
	run, err := c.workflowService.StartWorkflow(ctx, workflow, params, nil)
	if err != nil {
		logger.WithError(err).Error("Failed to start workflow")
		return nil
	}

	return &run
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
