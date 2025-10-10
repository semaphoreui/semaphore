package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/semaphoreui/semaphore/pkg/conv"
	"github.com/semaphoreui/semaphore/services/server"
	task2 "github.com/semaphoreui/semaphore/services/tasks"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	log "github.com/sirupsen/logrus"
	"github.com/thedevsaddam/gojsonq/v2"
)

// isValidHmacPayload checks if the GitHub payload's hash fits with
// the hash computed by GitHub sent as a header
func isValidHmacPayload(secret, headerHash string, payload []byte, prefix string) bool {
	hash := hmacHashPayload(secret, payload)

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
func hmacHashPayload(secret string, payloadBody []byte) string {
	hm := hmac.New(sha256.New, []byte(secret))
	hm.Write(payloadBody)
	sum := hm.Sum(nil)
	return fmt.Sprintf("%x", sum)
}

type IntegrationController struct {
	integrationService server.IntegrationService
}

func NewIntegrationController(integrationService server.IntegrationService) *IntegrationController {
	return &IntegrationController{
		integrationService: integrationService,
	}
}

func (c *IntegrationController) ReceiveIntegration(w http.ResponseWriter, r *http.Request) {

	var err error

	integrationAlias, err := helpers.GetStrParam("integration_alias", w, r)

	if err != nil {
		log.Error(err)
		return
	}

	log.Info(fmt.Sprintf("Receiving Integration from: %s", r.RemoteAddr))

	store := helpers.Store(r)

	integrations, level, err := store.GetIntegrationsByAlias(integrationAlias)

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
			project, err = store.GetProject(integrations[0].ProjectID)
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
				"sha256=")

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
				"sha256=")

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
				"")

			if !ok {
				log.WithFields(log.Fields{
					"context": "integrations",
				}).Error("Invalid HMAC signature")
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
			matchers, err = store.GetIntegrationMatchers(integration.ProjectID, db.RetrieveQueryParams{}, integration.ID)
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

		RunIntegration(integration, project, r, payload)
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

func GetTaskDefinition(integration db.Integration, payload []byte, r *http.Request) (taskDefinition db.Task, err error) {

	var envValues = make([]db.IntegrationExtractValue, 0)
	var taskValues = make([]db.IntegrationExtractValue, 0)

	extractValuesForExtractor, err := helpers.Store(r).GetIntegrationExtractValues(integration.ProjectID, db.RetrieveQueryParams{}, integration.ID)
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

	var extractedEnvResults = Extract(envValues, r, payload)

	if integration.TaskParams != nil {
		taskDefinition = integration.TaskParams.CreateTask(integration.TemplateID)
	} else {
		taskDefinition = db.Task{
			ProjectID:  integration.ProjectID,
			TemplateID: integration.TemplateID,
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

	// Add extracted environment variables only if they don't conflict with
	// existing task definition variables (task definition has higher priority)
	for k, v := range extractedEnvResults {
		if _, exists := env[k]; !exists {
			env[k] = v
		}
	}

	envStr, err := json.Marshal(env)
	if err != nil {
		return
	}

	taskDefinition.Environment = string(envStr)

	extractedTaskResults := ExtractAsAnyForTaskParams(taskValues, r, payload)
	for k, v := range extractedTaskResults {
		taskDefinition.Params[k] = v
	}

	return
}

func RunIntegration(integration db.Integration, project db.Project, r *http.Request, payload []byte) {

	log.Info(fmt.Sprintf("Running integration %d", integration.ID))

	taskDefinition, err := GetTaskDefinition(integration, payload, r)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"context":        "integrations",
			"integration_id": integration.ID,
		}).Error("Failed to get task definition")
		return
	}

	tpl, err := helpers.Store(r).GetTemplate(integration.ProjectID, integration.TemplateID)
	if err != nil {
		log.Error(err)
		return
	}

	pool := helpers.GetFromContext(r, "task_pool").(*task2.TaskPool)

	_, err = pool.AddTask(taskDefinition, nil, "", integration.ProjectID, tpl.App.NeedTaskAlias())
	if err != nil {
		log.Error(err)
		return
	}
}

func Extract(extractValues []db.IntegrationExtractValue, r *http.Request, payload []byte) (result map[string]string) {
	result = make(map[string]string)

	for _, extractValue := range extractValues {
		switch extractValue.ValueSource {
		case db.IntegrationExtractHeaderValue:
			result[extractValue.Variable] = r.Header.Get(extractValue.Key)
		case db.IntegrationExtractBodyValue:
			switch extractValue.BodyDataType {
			case db.IntegrationBodyDataJSON:
				var extractedResult = fmt.Sprintf("%v", gojsonq.New().JSONString(string(payload)).Find(extractValue.Key))
				result[extractValue.Variable] = extractedResult
			case db.IntegrationBodyDataString:
				result[extractValue.Variable] = string(payload)
			}
		}
	}
	return
}

func ExtractAsAnyForTaskParams(extractValues []db.IntegrationExtractValue, r *http.Request, payload []byte) db.MapStringAnyField {
	// Create a result map that accepts any type
	result := make(db.MapStringAnyField)

	for _, extractValue := range extractValues {
		switch extractValue.ValueSource {
		case db.IntegrationExtractHeaderValue:
			// Extract the header value
			result[extractValue.Variable] = r.Header.Get(extractValue.Key)

		case db.IntegrationExtractBodyValue:
			switch extractValue.BodyDataType {
			case db.IntegrationBodyDataJSON:
				// Query the JSON payload for the key using gojsonq
				rawValue := gojsonq.New().JSONString(string(payload)).Find(extractValue.Key)
				result[extractValue.Variable] = rawValue

			case db.IntegrationBodyDataString:
				// Simply use the entire payload as a string
				result[extractValue.Variable] = string(payload)
			}
		}
	}
	return result
}
