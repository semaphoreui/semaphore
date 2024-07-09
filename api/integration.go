package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/ansible-semaphore/semaphore/util"
	"io"
	"net/http"
	"strings"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
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

func ReceiveIntegration(w http.ResponseWriter, r *http.Request) {

	var err error

	integrationAlias, err := helpers.GetStrParam("integration_alias", w, r)

	if err != nil {
		log.Error(err)
		return
	}

	log.Info(fmt.Sprintf("Receiving Integration from: %s", r.RemoteAddr))

	var integrations []db.Integration

	if util.Config.IntegrationAlias != "" && integrationAlias == util.Config.IntegrationAlias {
		integrations, err = helpers.Store(r).GetAllSearchableIntegrations()
	} else {
		integrations, err = helpers.Store(r).GetIntegrationsByAlias(integrationAlias)
	}

	if err != nil {
		log.Error(err)
		return
	}

	log.Info(fmt.Sprintf("%d integrations found for alias %s", len(integrations), integrationAlias))

	projects := make(map[int]db.Project)

	for _, integration := range integrations {

		project, ok := projects[integration.ProjectID]
		if !ok {
			project, err = helpers.Store(r).GetProject(integrations[0].ProjectID)
			if err != nil {
				log.Error(err)
				return
			}
			projects[integration.ProjectID] = project
		}

		if integration.ProjectID != project.ID {
			panic("")
		}

		err = db.FillIntegration(helpers.Store(r), &integration)
		if err != nil {
			log.Error(err)
			return
		}

		var payload []byte

		payload, err = io.ReadAll(r.Body)

		if err != nil {
			log.Error(err)
			continue
		}

		switch integration.AuthMethod {
		case db.IntegrationAuthGitHub:
			ok := isValidHmacPayload(
				integration.AuthSecret.LoginPassword.Password,
				r.Header.Get("X-Hub-Signature-256"),
				payload,
				"sha256=")

			if !ok {
				log.Error("Invalid HMAC signature")
				continue
			}
		case db.IntegrationAuthHmac:
			ok := isValidHmacPayload(
				integration.AuthSecret.LoginPassword.Password,
				r.Header.Get(integration.AuthHeader),
				payload,
				"")

			if !ok {
				log.Error("Invalid HMAC signature")
				continue
			}
		case db.IntegrationAuthToken:
			if integration.AuthSecret.LoginPassword.Password != r.Header.Get(integration.AuthHeader) {
				log.Error("Invalid verification token")
				continue
			}
		case db.IntegrationAuthNone:
			// Do nothing
		default:
			log.Error("Unknown verification method: " + integration.AuthMethod)
			continue
		}

		var matchers []db.IntegrationMatcher
		matchers, err = helpers.Store(r).GetIntegrationMatchers(integration.ProjectID, db.RetrieveQueryParams{}, integration.ID)
		if err != nil {
			log.Error(err)
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

func convertFloatToIntIfPossible(v interface{}) (int64, bool) {

	switch v.(type) {
	case float64:
		f := v.(float64)
		i := int64(f)
		if float64(i) == f {
			return i, true
		}
	case float32:
		f := v.(float32)
		i := int64(f)
		if float32(i) == f {
			return i, true
		}
	}

	return 0, false
}

func MatchCompare(value interface{}, method db.IntegrationMatchMethodType, expected string) bool {

	if intValue, ok := convertFloatToIntIfPossible(value); ok {
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

func RunIntegration(integration db.Integration, project db.Project, r *http.Request, payload []byte) {

	log.Info(fmt.Sprintf("Running integration %d", integration.ID))

	var extractValues = make([]db.IntegrationExtractValue, 0)

	extractValuesForExtractor, err := helpers.Store(r).GetIntegrationExtractValues(project.ID, db.RetrieveQueryParams{}, integration.ID)
	if err != nil {
		log.Error(err)
		return
	}

	extractValues = append(extractValues, extractValuesForExtractor...)

	var extractedResults = Extract(extractValues, r, payload)

	environmentJSONBytes, err := json.Marshal(extractedResults)
	if err != nil {
		log.Error(err)
		return
	}

	var environmentJSONString = string(environmentJSONBytes)
	var taskDefinition = db.Task{
		TemplateID:    integration.TemplateID,
		ProjectID:     integration.ProjectID,
		Environment:   environmentJSONString,
		IntegrationID: &integration.ID,
	}

	_, err = helpers.TaskPool(r).AddTask(taskDefinition, nil, integration.ProjectID)
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
