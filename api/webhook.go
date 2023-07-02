package api

import (
	log "github.com/Sirupsen/logrus"
	"fmt"
	"encoding/json"
	"bytes"
	"strings"
	"golang.org/x/exp/slices"
	"github.com/ansible-semaphore/semaphore/api/helpers"
	"github.com/ansible-semaphore/semaphore/db"
	"net/http"
	"io"
	jsonq "github.com/thedevsaddam/gojsonq/v2"
)

func ReceiveWebhook(w http.ResponseWriter, r *http.Request) {
	log.Info(fmt.Sprintf("Receiving Webhook from: %s", r.RemoteAddr))

	var err error
	// var projects []db.Project
	var extractors []db.WebhookExtractor
	extractors, err = helpers.Store(r).GetAllWebhookExtractors()

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

		var allWebhookExtractorIDs []int = make([]int, 0)
		var webhooks []db.Webhook
		webhooks, err = helpers.Store(r).GetAllWebhooks()
		for _, id := range webhookIDs {
			var extractorsForWebhook []db.WebhookExtractor
			extractorsForWebhook, err = helpers.Store(r).GetWebhookExtractors(id, db.RetrieveQueryParams{})

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
				var webhook db.Webhook
				webhook = FindWebhook(webhooks, id)
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
	if matcher.MatchType == db.WebhookMatchHeader {
		var header_value string = r.Header.Get(matcher.Key)
		return MatchCompare(header_value, matcher.Method, matcher.Value)
	} else if matcher.MatchType == db.WebhookMatchBody {
		bodyBytes, err := io.ReadAll(r.Body)

		if err != nil {
			log.Fatalln(err)
			return false
		}
		var body = string(bodyBytes)
		if matcher.BodyDataType == db.WebhookBodyDataJSON {
			var jsonBytes bytes.Buffer
			jsonq.New().FromString(body).From(matcher.Key).Writer(&jsonBytes)
			var jsonString = jsonBytes.String()
			if err != nil {
				log.Error(fmt.Sprintf("Failed to marshal JSON contents of body. %v", err))
			}
			return MatchCompare(jsonString, matcher.Method, matcher.Value)
		} else if matcher.BodyDataType == db.WebhookBodyDataString {
			return MatchCompare(body, matcher.Method, matcher.Value)
		} else if matcher.BodyDataType == db.WebhookBodyDataXML {
			// XXX: TBI
			return false
		}
	}
	return false
}

func MatchCompare(value string, method db.WebhookMatchMethodType, expected string) (bool) {
	if method == db.WebhookMatchMethodEquals {
		return value == expected
	} else if method == db.WebhookMatchMethodEquals {
		return value != expected
	} else if method == db.WebhookMatchMethodContains {
		return strings.Contains(value, expected)
	}
	return false
}

func RunWebhook(webhook db.Webhook, r *http.Request) {
	extractors, err := helpers.Store(r).GetWebhookExtractors(webhook.ID, db.RetrieveQueryParams{});
	if err != nil {
		log.Error(err)
		return
	}

	if err != nil {
		log.Error(err)
		return
	}

	var extractValues []db.WebhookExtractValue = make([]db.WebhookExtractValue, 0)
	for _, extractor := range extractors {
		extractValuesForExtractor, err := helpers.Store(r).GetWebhookExtractValues(extractor.ID, db.RetrieveQueryParams{})
		if err != nil {
			log.Error(err)
		}
		for _, extraExtractor := range extractValuesForExtractor {
			extractValues = append(extractValues, extraExtractor)
		}
	}

	var extractedResults = Extract(extractValues, r)

	// XXX: LOG AN EVENT HERE
	environmentJSONBytes, err := json.Marshal(extractedResults)
	var environmentJSONString = string(environmentJSONBytes)
	var taskDefinition = db.Task{
		TemplateID: webhook.TemplateID,
		ProjectID: webhook.ProjectID,
		Debug: true,
		Environment: environmentJSONString,
	}

	var user db.User
	user, err = helpers.Store(r).GetUser(1)

	helpers.TaskPool(r).AddTask(taskDefinition, &user.ID, webhook.ProjectID)
}

func Extract(extractValues []db.WebhookExtractValue, r *http.Request) (result map[string]string) {
	result = make(map[string]string)

	for _, extractValue := range extractValues {
		if extractValue.ValueSource == db.WebhookExtractHeaderValue {
			result[extractValue.Variable] = r.Header.Get(extractValue.Key)
		} else if extractValue.ValueSource == db.WebhookExtractBodyValue {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				log.Fatalln(err)
				return
			}
			var body = string(bodyBytes)

			if extractValue.BodyDataType == db.WebhookBodyDataJSON {
				var jsonBytes bytes.Buffer
				jsonq.New().FromString(body).From(extractValue.Key).Writer(&jsonBytes)
				result[extractValue.Variable] = jsonBytes.String()
			} else if extractValue.BodyDataType == db.WebhookBodyDataString {
				result[extractValue.Variable] = body
			}	else if extractValue.BodyDataType == db.WebhookBodyDataXML {
				// XXX: TBI
			}
		}
	}
	return
}
