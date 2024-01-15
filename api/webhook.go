package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	jsonq "github.com/thedevsaddam/gojsonq/v2"
	"golang.org/x/exp/slices"
)

func ReceiveWebhook(w http.ResponseWriter, r *http.Request) {
	log.Info(fmt.Sprintf("Receiving Webhook from: %s", r.RemoteAddr))

	var err error
	// var projects []db.Project
	var extractors []db.WebhookExtractor
	extractors, err = helpers.Store(r).GetAllWebhookExtractors()

	if err != nil {
		log.Error(err)
		return
	}

	var foundExtractors = make([]db.WebhookExtractor, 0)
	for _, extractor := range extractors {
		var matchers []db.WebhookMatcher
		matchers, err = helpers.Store(r).GetWebhookMatchers(extractor.ID, db.RetrieveQueryParams{})
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
		var webhookIDs = make([]int, 0)
		var extractorIDs = make([]int, 0)

		for _, extractor := range foundExtractors {
			webhookIDs = append(webhookIDs, extractor.WebhookID)
		}

		for _, extractor := range foundExtractors {
			extractorIDs = append(extractorIDs, extractor.ID)
		}

		var allWebhookExtractorIDs = make([]int, 0)
		var webhooks []db.Webhook
		webhooks, err = helpers.Store(r).GetAllWebhooks()
		if err != nil {
			log.Error(err)
			return
		}
		for _, id := range webhookIDs {
			var extractorsForWebhook []db.WebhookExtractor
			extractorsForWebhook, err = helpers.Store(r).GetWebhookExtractors(id, db.RetrieveQueryParams{})

			if err != nil {
				log.Error(err)
				return
			}

			for _, extractor := range extractorsForWebhook {
				allWebhookExtractorIDs = append(allWebhookExtractorIDs, extractor.ID)
			}

			var found = false
			for _, webhookExtractorID := range extractorIDs {
				if slices.Contains(allWebhookExtractorIDs, webhookExtractorID) {
					found = true
					continue
				} else {
					found = false
					break
				}
			}

			// if all extractors for a webhook matched during search
			if found {
				webhook := FindWebhook(webhooks, id)

				if webhook.ID != id {
					log.Error(fmt.Sprintf("Could not find webhook ID: %v", id))
					continue
				}
				RunWebhook(webhook, r)
			}
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func FindWebhook(webhooks []db.Webhook, id int) (webhook db.Webhook) {
	for _, webhook := range webhooks {
		if webhook.ID == id {
			return webhook
		}
	}
	return db.Webhook{}
}

func UniqueWebhooks(webhooks []db.Webhook) []db.Webhook {
	var unique []db.Webhook
webhookLoop:
	for _, v := range webhooks {
		for i, u := range unique {
			if v.ID == u.ID {
				unique[i] = v
				continue webhookLoop
			}
		}
		unique = append(unique, v)
	}
	return unique
}

func UniqueExtractors(extractors []db.WebhookExtractor) []db.WebhookExtractor {
	var unique []db.WebhookExtractor
webhookLoop:
	for _, v := range extractors {
		for i, u := range unique {
			if v.ID == u.ID {
				unique[i] = v
				continue webhookLoop
			}
		}
		unique = append(unique, v)
	}
	return unique
}

func Match(matcher db.WebhookMatcher, r *http.Request) (matched bool) {

	switch matcher.MatchType {
	case db.WebhookMatchHeader:
		var header_value = r.Header.Get(matcher.Key)
		return MatchCompare(header_value, matcher.Method, matcher.Value)
	case db.WebhookMatchBody:
		bodyBytes, err := io.ReadAll(r.Body)

		if err != nil {
			log.Fatalln(err)
			return false
		}
		var body = string(bodyBytes)
		switch matcher.BodyDataType {
		case db.WebhookBodyDataJSON:
			var jsonBytes bytes.Buffer
			jsonq.New().FromString(body).From(matcher.Key).Writer(&jsonBytes)
			var jsonString = jsonBytes.String()
			if err != nil {
				log.Error(fmt.Sprintf("Failed to marshal JSON contents of body. %v", err))
			}
			return MatchCompare(jsonString, matcher.Method, matcher.Value)
		case db.WebhookBodyDataString:
			return MatchCompare(body, matcher.Method, matcher.Value)
		case db.WebhookBodyDataXML:
			// XXX: TBI
			return false
		}
	}

	return false
}

func MatchCompare(value string, method db.WebhookMatchMethodType, expected string) bool {
	switch method {
	case db.WebhookMatchMethodEquals:
		return value == expected
	case db.WebhookMatchMethodUnEquals:
		return value != expected
	case db.WebhookMatchMethodContains:
		return strings.Contains(value, expected)
	default:
		return false
	}
}

func RunWebhook(webhook db.Webhook, r *http.Request) {
	extractors, err := helpers.Store(r).GetWebhookExtractors(webhook.ID, db.RetrieveQueryParams{})
	if err != nil {
		log.Error(err)
		return
	}

	if err != nil {
		log.Error(err)
		return
	}

	var extractValues = make([]db.WebhookExtractValue, 0)
	for _, extractor := range extractors {
		extractValuesForExtractor, errextractValuesForExtractor := helpers.Store(r).GetWebhookExtractValues(extractor.ID, db.RetrieveQueryParams{})
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
		TemplateID:  webhook.TemplateID,
		ProjectID:   webhook.ProjectID,
		Debug:       true,
		Environment: environmentJSONString,
	}

	var user db.User
	user, err = helpers.Store(r).GetUser(1)
	if err != nil {
		log.Error(err)
		return
	}

	helpers.TaskPool(r).AddTask(taskDefinition, &user.ID, webhook.ProjectID)
}

func Extract(extractValues []db.WebhookExtractValue, r *http.Request) (result map[string]string) {
	result = make(map[string]string)

	for _, extractValue := range extractValues {
		switch extractValue.ValueSource {
		case db.WebhookExtractHeaderValue:
			result[extractValue.Variable] = r.Header.Get(extractValue.Key)
		case db.WebhookExtractBodyValue:
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				log.Fatal(err)
				return
			}
			var body = string(bodyBytes)

			switch extractValue.BodyDataType {
			case db.WebhookBodyDataJSON:
				var jsonBytes bytes.Buffer
				jsonq.New().FromString(body).From(extractValue.Key).Writer(&jsonBytes)
				result[extractValue.Variable] = jsonBytes.String()
			case db.WebhookBodyDataString:
				result[extractValue.Variable] = body
			case db.WebhookBodyDataXML:
				// XXX: TBI
			}
		}
	}
	return
}
