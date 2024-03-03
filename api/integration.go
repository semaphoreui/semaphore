package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/gorilla/context"
	"io"
	"net/http"
	"strings"

	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	log "github.com/sirupsen/logrus"
	jsonq "github.com/thedevsaddam/gojsonq/v2"
	"golang.org/x/exp/slices"
)

// IsValidPayload checks if the github payload's hash fits with
// the hash computed by GitHub sent as a header
func IsValidPayload(secret, headerHash string, payload []byte) bool {
	hash := HashPayload(secret, payload)
	return hmac.Equal(
		[]byte(hash),
		[]byte(headerHash),
	)
}

// HashPayload computes the hash of payload's body according to the webhook's secret token
// see https://developer.github.com/webhooks/securing/#validating-payloads-from-github
// returning the hash as a hexadecimal string
func HashPayload(secret string, playloadBody []byte) string {
	hm := hmac.New(sha1.New, []byte(secret))
	hm.Write(playloadBody)
	sum := hm.Sum(nil)
	return fmt.Sprintf("%x", sum)
}

func ReceiveIntegration(w http.ResponseWriter, r *http.Request) {
	log.Info(fmt.Sprintf("Receiving Integration from: %s", r.RemoteAddr))

	var err error
	// var projects []db.Project
	var extractors []db.IntegrationExtractor

	if context.Get(r, "integration") != nil {
		integration := context.Get(r, "integration").(db.Integration)
		switch integration.AuthMethod {
		case db.IntegrationAuthHmac:
			var payload []byte
			_, err = r.Body.Read(payload)
			if err != nil {
				log.Error(err)
				return
			}

			if IsValidPayload(integration.AuthSecret.LoginPassword.Password, r.Header.Get(integration.AuthHeader), payload) {
				log.Error(err)
				return
			}
		case db.IntegrationAuthToken:
			if integration.AuthSecret.LoginPassword.Password != r.Header.Get(integration.AuthHeader) {
				log.Error("Invalid verification token")
				return
			}
		case db.IntegrationAuthNone:
		default:
			log.Error("Unknown verification method: " + integration.AuthMethod)
			return
		}

		extractors, err = helpers.Store(r).GetIntegrationExtractors(integration.ID, db.RetrieveQueryParams{})
	} else {
		// TODO: remove
		extractors, err = helpers.Store(r).GetAllIntegrationExtractors()
	}

	if err != nil {
		log.Error(err)
		return
	}

	var foundExtractors = make([]db.IntegrationExtractor, 0)
	for _, extractor := range extractors {
		var matchers []db.IntegrationMatcher
		matchers, err = helpers.Store(r).GetIntegrationMatchers(extractor.ID, db.RetrieveQueryParams{})
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
		// If all Matched...
		if matched {
			foundExtractors = append(foundExtractors, extractor)
		}
	}

	// Iterate over all Extractors that matched
	if len(foundExtractors) > 0 {
		var integrationIDs = make([]int, 0)
		var extractorIDs = make([]int, 0)

		for _, extractor := range foundExtractors {
			integrationIDs = append(integrationIDs, extractor.IntegrationID)
		}

		for _, extractor := range foundExtractors {
			extractorIDs = append(extractorIDs, extractor.ID)
		}

		var allIntegrationExtractorIDs = make([]int, 0)
		var integrations []db.Integration
		integrations, err = helpers.Store(r).GetAllIntegrations()
		if err != nil {
			log.Error(err)
			return
		}
		for _, id := range integrationIDs {
			var extractorsForIntegration []db.IntegrationExtractor
			extractorsForIntegration, err = helpers.Store(r).GetIntegrationExtractors(id, db.RetrieveQueryParams{})

			if err != nil {
				log.Error(err)
				return
			}

			for _, extractor := range extractorsForIntegration {
				allIntegrationExtractorIDs = append(allIntegrationExtractorIDs, extractor.ID)
			}

			var found = false
			for _, integrationExtractorID := range extractorIDs {
				if slices.Contains(allIntegrationExtractorIDs, integrationExtractorID) {
					found = true
					continue
				} else {
					found = false
					break
				}
			}

			// if all extractors for a integration matched during search
			if found {
				integration := FindIntegration(integrations, id)

				if integration.ID != id {
					log.Error(fmt.Sprintf("Could not find integration ID: %v", id))
					continue
				}
				RunIntegration(integration, r)
			}
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func FindIntegration(integrations []db.Integration, id int) (integration db.Integration) {
	for _, integration := range integrations {
		if integration.ID == id {
			return integration
		}
	}
	return db.Integration{}
}

func UniqueIntegrations(integrations []db.Integration) []db.Integration {
	var unique []db.Integration
integrationLoop:
	for _, v := range integrations {
		for i, u := range unique {
			if v.ID == u.ID {
				unique[i] = v
				continue integrationLoop
			}
		}
		unique = append(unique, v)
	}
	return unique
}

func UniqueExtractors(extractors []db.IntegrationExtractor) []db.IntegrationExtractor {
	var unique []db.IntegrationExtractor
integrationLoop:
	for _, v := range extractors {
		for i, u := range unique {
			if v.ID == u.ID {
				unique[i] = v
				continue integrationLoop
			}
		}
		unique = append(unique, v)
	}
	return unique
}

func Match(matcher db.IntegrationMatcher, r *http.Request) (matched bool) {

	switch matcher.MatchType {
	case db.IntegrationMatchHeader:
		var header_value = r.Header.Get(matcher.Key)
		return MatchCompare(header_value, matcher.Method, matcher.Value)
	case db.IntegrationMatchBody:
		bodyBytes, err := io.ReadAll(r.Body)

		if err != nil {
			log.Fatalln(err)
			return false
		}
		var body = string(bodyBytes)
		switch matcher.BodyDataType {
		case db.IntegrationBodyDataJSON:
			var jsonBytes bytes.Buffer
			jsonq.New().FromString(body).From(matcher.Key).Writer(&jsonBytes)
			var jsonString = jsonBytes.String()
			if err != nil {
				log.Error(fmt.Sprintf("Failed to marshal JSON contents of body. %v", err))
			}
			return MatchCompare(jsonString, matcher.Method, matcher.Value)
		case db.IntegrationBodyDataString:
			return MatchCompare(body, matcher.Method, matcher.Value)
		case db.IntegrationBodyDataXML:
			// XXX: TBI
			return false
		}
	}

	return false
}

func MatchCompare(value string, method db.IntegrationMatchMethodType, expected string) bool {
	switch method {
	case db.IntegrationMatchMethodEquals:
		return value == expected
	case db.IntegrationMatchMethodUnEquals:
		return value != expected
	case db.IntegrationMatchMethodContains:
		return strings.Contains(value, expected)
	default:
		return false
	}
}

func RunIntegration(integration db.Integration, r *http.Request) {
	extractors, err := helpers.Store(r).GetIntegrationExtractors(integration.ID, db.RetrieveQueryParams{})
	if err != nil {
		log.Error(err)
		return
	}

	if err != nil {
		log.Error(err)
		return
	}

	var extractValues = make([]db.IntegrationExtractValue, 0)
	for _, extractor := range extractors {
		extractValuesForExtractor, errextractValuesForExtractor := helpers.Store(r).GetIntegrationExtractValues(extractor.ID, db.RetrieveQueryParams{})
		if errextractValuesForExtractor != nil {
			log.Error(errextractValuesForExtractor)
			return
		}
		extractValues = append(extractValues, extractValuesForExtractor...)
	}

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

	var user db.User
	user, err = helpers.Store(r).GetUser(1)
	if err != nil {
		log.Error(err)
		return
	}

	helpers.TaskPool(r).AddTask(taskDefinition, &user.ID, integration.ProjectID)
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
				var jsonBytes bytes.Buffer
				jsonq.New().FromString(body).From(extractValue.Key).Writer(&jsonBytes)
				result[extractValue.Variable] = jsonBytes.String()
			case db.IntegrationBodyDataString:
				result[extractValue.Variable] = body
			case db.IntegrationBodyDataXML:
				// XXX: TBI
			}
		}
	}
	return
}
