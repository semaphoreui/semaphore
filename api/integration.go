package api

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/json"
	"fmt"
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
	hash := hmacHashPayload(secret, payload, prefix)
	return hmac.Equal(
		[]byte(hash),
		[]byte(headerHash),
	)
}

// hmacHashPayload computes the hash of payload's body according to the webhook's secret token
// see https://developer.github.com/webhooks/securing/#validating-payloads-from-github
// returning the hash as a hexadecimal string
func hmacHashPayload(secret string, payloadBody []byte, prefix string) string {
	hm := hmac.New(sha1.New, []byte(secret))
	hm.Write(payloadBody)
	sum := hm.Sum(nil)
	return fmt.Sprintf("%s%x", prefix, sum)
}

func ReceiveIntegration(w http.ResponseWriter, r *http.Request) {

	var err error

	integrationAlias, err := helpers.GetStrParam("integration_alias", w, r)

	if err != nil {
		log.Error(err)
		return
	}

	log.Info(fmt.Sprintf("Receiving Integration from: %s", r.RemoteAddr))

	integrations, err := helpers.Store(r).GetIntegrationsByAlias(integrationAlias)

	if err != nil {
		log.Error(err)
		return
	}

	log.Info(fmt.Sprintf("%d integrations found for alias %s", len(integrations), integrationAlias))

	var project db.Project
	if len(integrations) > 0 {
		project, err = helpers.Store(r).GetProject(integrations[0].ProjectID)
		if err != nil {
			log.Error(err)
			return
		}
	}
	for _, integration := range integrations {
		if integration.ProjectID != project.ID {
			panic("")
		}

		err = db.FillIntegration(helpers.Store(r), &integration)
		if err != nil {
			log.Error(err)
			return
		}

		switch integration.AuthMethod {
		case db.IntegrationAuthGitHub:
			var payload []byte
			_, err = r.Body.Read(payload)
			if err != nil {
				log.Error(err)
				continue
			}

			ok := isValidHmacPayload(
				integration.AuthSecret.LoginPassword.Password,
				r.Header.Get("X-Hub-Signature-256"),
				payload,
				"sha256=")

			if !ok {
				log.Error(err)
				continue
			}
		case db.IntegrationAuthHmac:
			var payload []byte
			_, err = r.Body.Read(payload)
			if err != nil {
				log.Error(err)
				continue
			}

			ok := isValidHmacPayload(
				integration.AuthSecret.LoginPassword.Password,
				r.Header.Get(integration.AuthHeader),
				payload,
				"")

			if !ok {
				log.Error(err)
				continue
			}
		case db.IntegrationAuthToken:
			if integration.AuthSecret.LoginPassword.Password != r.Header.Get(integration.AuthHeader) {
				log.Error("Invalid verification token")
				continue
			}
		case db.IntegrationAuthNone:
			// TODO: do nothing
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
			if Match(matcher, r) {
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

		RunIntegration(integration, project, r)
	}

	w.WriteHeader(http.StatusNoContent)
}

func Match(matcher db.IntegrationMatcher, r *http.Request) (matched bool) {

	switch matcher.MatchType {
	case db.IntegrationMatchHeader:
		var headerValue = r.Header.Get(matcher.Key)
		return MatchCompare(headerValue, matcher.Method, matcher.Value)
	case db.IntegrationMatchBody:
		bodyBytes, err := io.ReadAll(r.Body)

		if err != nil {
			log.Fatalln(err)
			return false
		}
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

func MatchCompare(value interface{}, method db.IntegrationMatchMethodType, expected string) bool {
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

func RunIntegration(integration db.Integration, project db.Project, r *http.Request) {

	log.Info(fmt.Sprintf("Running integration %d", integration.ID))

	var extractValues = make([]db.IntegrationExtractValue, 0)

	extractValuesForExtractor, err := helpers.Store(r).GetIntegrationExtractValues(project.ID, db.RetrieveQueryParams{}, integration.ID)
	if err != nil {
		log.Error(err)
		return
	}

	extractValues = append(extractValues, extractValuesForExtractor...)

	var extractedResults = Extract(extractValues, r)

	// XXX: LOG AN EVENT HERE
	environmentJSONBytes, err := json.Marshal(extractedResults)
	if err != nil {
		log.Error(err)
		return
	}

	var environmentJSONString = string(environmentJSONBytes)
	var taskDefinition = db.Task{
		TemplateID:  integration.TemplateID,
		ProjectID:   integration.ProjectID,
		Debug:       true,
		Environment: environmentJSONString,
	}

	//var user db.User
	//user, err = helpers.Store(r).GetUser(1)
	//if err != nil {
	//	log.Error(err)
	//	return
	//}

	_, err = helpers.TaskPool(r).AddTask(taskDefinition, nil, integration.ProjectID)
	if err != nil {
		log.Error(err)
		return
	}
}

func Extract(extractValues []db.IntegrationExtractValue, r *http.Request) (result map[string]string) {
	result = make(map[string]string)

	for _, extractValue := range extractValues {
		switch extractValue.ValueSource {
		case db.IntegrationExtractHeaderValue:
			result[extractValue.Variable] = r.Header.Get(extractValue.Key)
		case db.IntegrationExtractBodyValue:
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				log.Fatal(err)
				return
			}
			var body = string(bodyBytes)

			switch extractValue.BodyDataType {
			case db.IntegrationBodyDataJSON:
				result[extractValue.Variable] =
					fmt.Sprintf("%v", gojsonq.New().JSONString(body).Find(extractValue.Key))
			case db.IntegrationBodyDataString:
				result[extractValue.Variable] = body
			}
		}
	}
	return
}
