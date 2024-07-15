package db

import (
	"strconv"
	"strings"
)

type IntegrationAuthMethod string

const (
	IntegrationAuthNone   = ""
	IntegrationAuthGitHub = "github"
	IntegrationAuthToken  = "token"
	IntegrationAuthHmac   = "hmac"
)

type IntegrationMatchType string

const (
	IntegrationMatchHeader IntegrationMatchType = "header"
	IntegrationMatchBody   IntegrationMatchType = "body"
)

type IntegrationMatchMethodType string

const (
	IntegrationMatchMethodEquals   IntegrationMatchMethodType = "equals"
	IntegrationMatchMethodUnEquals IntegrationMatchMethodType = "unequals"
	IntegrationMatchMethodContains IntegrationMatchMethodType = "contains"
)

type IntegrationBodyDataType string

const (
	IntegrationBodyDataJSON   IntegrationBodyDataType = "json"
	IntegrationBodyDataString IntegrationBodyDataType = "string"
)

type IntegrationMatcher struct {
	ID            int                        `db:"id" json:"id"`
	Name          string                     `db:"name" json:"name"`
	IntegrationID int                        `db:"integration_id" json:"integration_id"`
	MatchType     IntegrationMatchType       `db:"match_type" json:"match_type"`
	Method        IntegrationMatchMethodType `db:"method" json:"method"`
	BodyDataType  IntegrationBodyDataType    `db:"body_data_type" json:"body_data_type"`
	Key           string                     `db:"key" json:"key"`
	Value         string                     `db:"value" json:"value"`
}

type IntegrationExtractValueSource string

const (
	IntegrationExtractBodyValue   IntegrationExtractValueSource = "body"
	IntegrationExtractHeaderValue IntegrationExtractValueSource = "header"
)

type IntegrationExtractValue struct {
	ID            int                           `db:"id" json:"id"`
	Name          string                        `db:"name" json:"name"`
	IntegrationID int                           `db:"integration_id" json:"integration_id"`
	ValueSource   IntegrationExtractValueSource `db:"value_source" json:"value_source"`
	BodyDataType  IntegrationBodyDataType       `db:"body_data_type" json:"body_data_type"`
	Key           string                        `db:"key" json:"key"`
	Variable      string                        `db:"variable" json:"variable"`
}

type IntegrationAlias struct {
	ID            int    `db:"id" json:"-"`
	Alias         string `db:"alias" json:"alias"`
	ProjectID     int    `db:"project_id" json:"project_id"`
	IntegrationID *int   `db:"integration_id" json:"integration_id"`
}

type Integration struct {
	ID           int                   `db:"id" json:"id"`
	Name         string                `db:"name" json:"name"`
	ProjectID    int                   `db:"project_id" json:"project_id"`
	TemplateID   int                   `db:"template_id" json:"template_id"`
	AuthMethod   IntegrationAuthMethod `db:"auth_method" json:"auth_method"`
	AuthSecretID *int                  `db:"auth_secret_id" json:"auth_secret_id"`
	AuthHeader   string                `db:"auth_header" json:"auth_header"`
	AuthSecret   AccessKey             `db:"-" json:"-"`
	Searchable   bool                  `db:"searchable" json:"searchable"`
	TaskParams   MapStringAnyField     `db:"task_params" json:"task_params"`
}

func (env *Integration) Validate() error {
	if env.Name == "" {
		return &ValidationError{"No Name set for integration"}
	}
	return nil
}

func (env *IntegrationMatcher) Validate() error {
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
		return &ValidationError{"No Name set for integration"}
	}

	return nil
}

func (env *IntegrationExtractValue) Validate() error {
	if env.ValueSource == "" {
		return &ValidationError{"No Value Source defined"}
	}

	if env.Name == "" {
		return &ValidationError{"No Name set for integration"}
	}

	if env.ValueSource == IntegrationExtractBodyValue {
		if env.BodyDataType == "" {
			return &ValidationError{"Value Source but no body data type set"}
		}

		if env.BodyDataType == IntegrationBodyDataJSON {
			if env.Key == "" {
				return &ValidationError{"No Key set for JSON Body Data extraction."}
			}
		}
	}

	if env.ValueSource == IntegrationExtractHeaderValue {
		if env.Key == "" {
			return &ValidationError{"Value Source set but no Key set"}
		}
	}

	return nil
}

func (matcher *IntegrationMatcher) String() string {
	var builder strings.Builder
	// ID:1234 body/json key == value on Extractor: 1234
	builder.WriteString("ID:" + strconv.Itoa(matcher.ID) + " " + string(matcher.MatchType))

	if matcher.MatchType == IntegrationMatchBody {
		builder.WriteString("/" + string(matcher.BodyDataType))
	}

	builder.WriteString(" " + matcher.Key + " ")

	switch matcher.Method {
	case IntegrationMatchMethodEquals:
		builder.WriteString("==")
	case IntegrationMatchMethodUnEquals:
		builder.WriteString("!=")
	case IntegrationMatchMethodContains:
		builder.WriteString(" contains ")
	default:

	}

	builder.WriteString(matcher.Value + ", on Extractor: " + strconv.Itoa(matcher.IntegrationID))

	return builder.String()
}

func (value *IntegrationExtractValue) String() string {
	var builder strings.Builder

	// ID:1234 body/json from key as argument
	builder.WriteString("ID:" + strconv.Itoa(value.ID) + " " + string(value.ValueSource))

	if value.ValueSource == IntegrationExtractBodyValue {
		builder.WriteString("/" + string(value.BodyDataType))
	}

	builder.WriteString(" from " + value.Key + " as " + value.Variable)

	return builder.String()
}

func FillIntegration(d Store, inventory *Integration) (err error) {
	if inventory.AuthSecretID != nil {
		inventory.AuthSecret, err = d.GetAccessKey(inventory.ProjectID, *inventory.AuthSecretID)
	}

	if err != nil {
		return
	}

	err = inventory.AuthSecret.DeserializeSecret()

	return
}
