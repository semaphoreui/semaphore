package db

import (
	"strconv"
	"strings"
)

type WebhookMatchType string

const (
	WebhookMatchHeader WebhookMatchType = "header"
	WebhookMatchBody   WebhookMatchType = "body"
)

type WebhookMatchMethodType string

const (
	WebhookMatchMethodEquals   WebhookMatchMethodType = "equals"
	WebhookMatchMethodUnEquals WebhookMatchMethodType = "unequals"
	WebhookMatchMethodContains WebhookMatchMethodType = "contains"
)

type WebhookBodyDataType string

const (
	WebhookBodyDataJSON   WebhookBodyDataType = "json"
	WebhookBodyDataXML    WebhookBodyDataType = "xml"
	WebhookBodyDataString WebhookBodyDataType = "string"
)

type WebhookMatcher struct {
	ID           int                    `db:"id" json:"id"`
	Name         string                 `db:"name" json:"name"`
	ExtractorID  int                    `db:"extractor_id" json:"extractor_id"`
	MatchType    WebhookMatchType       `db:"match_type" json:"match_type"`
	Method       WebhookMatchMethodType `db:"method" json:"method"`
	BodyDataType WebhookBodyDataType    `db:"body_data_type" json:"body_data_type"`
	Key          string                 `db:"key" json:"key"`
	Value        string                 `db:"value" json:"value"`
}

type WebhookExtractValueSource string

const (
	WebhookExtractBodyValue   WebhookExtractValueSource = "body"
	WebhookExtractHeaderValue WebhookExtractValueSource = "header"
)

type WebhookExtractValue struct {
	ID           int                       `db:"id" json:"id"`
	Name         string                    `db:"name" json:"name"`
	ExtractorID  int                       `db:"extractor_id" json:"extractor_id"`
	ValueSource  WebhookExtractValueSource `db:"value_source" json:"value_source"`
	BodyDataType WebhookBodyDataType       `db:"body_data_type" json:"body_data_type"`
	Key          string                    `db:"key" json:"key"`
	Variable     string                    `db:"variable" json:"variable"`
}

type WebhookExtractor struct {
	ID        int    `db:"id" json:"id"`
	Name      string `db:"name" json:"name"`
	WebhookID int    `db:"webhook_id" json:"webhook_id"`
}

type Webhook struct {
	ID         int    `db:"id" json:"id"`
	Name       string `db:"name" json:"name"`
	ProjectID  int    `db:"project_id" json:"project_id"`
	TemplateID int    `db:"template_id" json:"template_id"`
}

func (env *Webhook) Validate() error {
	if env.Name == "" {
		return &ValidationError{"No Name set for webhook"}
	}
	return nil
}

func (env *WebhookMatcher) Validate() error {
	if env.MatchType == "" {
		return &ValidationError{"No Match Type set"}
	} else {
		if env.Key == "" {
			return &ValidationError{"No key set"}
		}
		if env.Value == "" {
			return &ValidationError{"No value set"}
		}

	}

	if env.Name == "" {
		return &ValidationError{"No Name set for webhook"}
	}

	return nil
}

func (env *WebhookExtractor) Validate() error {
	if env.Name == "" {
		return &ValidationError{"No Name set for webhook"}
	}

	return nil
}

func (env *WebhookExtractValue) Validate() error {
	if env.ValueSource == "" {
		return &ValidationError{"No Value Source defined"}
	}

	if env.Name == "" {
		return &ValidationError{"No Name set for webhook"}
	}

	if env.ValueSource == WebhookExtractBodyValue {
		if env.BodyDataType == "" {
			return &ValidationError{"Value Source but no body data type set"}
		}

		if env.BodyDataType == WebhookBodyDataJSON {
			if env.Key == "" {
				return &ValidationError{"No Key set for JSON Body Data extraction."}
			}
		}
	}

	if env.ValueSource == WebhookExtractHeaderValue {
		if env.Key == "" {
			return &ValidationError{"Value Source set but no Key set"}
		}
	}

	return nil
}

func (matcher *WebhookMatcher) String() string {
	var builder strings.Builder
	// ID:1234 body/json key == value on Extractor: 1234
	builder.WriteString("ID:" + strconv.Itoa(matcher.ID) + " " + string(matcher.MatchType))

	if matcher.MatchType == WebhookMatchBody {
		builder.WriteString("/" + string(matcher.BodyDataType))
	}

	builder.WriteString(" " + matcher.Key + " ")

	switch matcher.Method {
	case WebhookMatchMethodEquals:
		builder.WriteString("==")
	case WebhookMatchMethodUnEquals:
		builder.WriteString("!=")
	case WebhookMatchMethodContains:
		builder.WriteString(" contains ")
	default:

	}

	builder.WriteString(matcher.Value + ", on Extractor: " + strconv.Itoa(matcher.ExtractorID))

	return builder.String()
}

func (value *WebhookExtractValue) String() string {
	var builder strings.Builder

	// ID:1234 body/json from key as argument
	builder.WriteString("ID:" + strconv.Itoa(value.ID) + " " + string(value.ValueSource))

	if value.ValueSource == WebhookExtractBodyValue {
		builder.WriteString("/" + string(value.BodyDataType))
	}

	builder.WriteString(" from " + value.Key + " as " + value.Variable)

	return builder.String()
}
