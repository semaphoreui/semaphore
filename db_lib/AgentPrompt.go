package db_lib

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/agentprompt"
)

// NewAgentPromptValues collects what the ${{ … }} expressions of an agent's
// prompt can name, from where the other apps take their variables:
//
//   - vars.extra: the extra variables of the Variable Groups and the survey
//     variables of the task (environment.JSON, merged by the task runner);
//   - vars.env: the environment variables of the Variable Groups;
//   - secrets.extra, secrets.env: the secrets of the Variable Groups, by type,
//     and the secret survey variables (surveySecrets, a JSON object).
//
// A survey variable delivered as an environment variable is an env one here
// too. One the task left unanswered is an empty string.
func NewAgentPromptValues(
	environment db.Environment,
	surveySecrets string,
	surveyVars []db.SurveyVar,
) (values agentprompt.Values, err error) {
	values = agentprompt.Values{
		ExtraVars:    map[string]string{},
		EnvVars:      map[string]string{},
		ExtraSecrets: map[string]string{},
		EnvSecrets:   map[string]string{},
	}

	if err = promptVarsFromJSON(environment.JSON, values.ExtraVars); err != nil {
		return values, fmt.Errorf("extra variables: %w", err)
	}
	delete(values.ExtraVars, "semaphore_vars")

	if environment.ENV != nil {
		if err = promptVarsFromJSON(*environment.ENV, values.EnvVars); err != nil {
			return values, fmt.Errorf("environment variables: %w", err)
		}
	}

	for _, secret := range environment.Secrets {
		switch secret.Type {
		case db.EnvironmentSecretVar:
			values.ExtraSecrets[secret.Name] = secret.Secret
		case db.EnvironmentSecretEnv:
			values.EnvSecrets[secret.Name] = secret.Secret
		}
	}

	// A secret survey variable is kept apart from the plain ones; it wins over
	// a secret of the Variable Groups with its name, as it does for a script.
	if err = promptVarsFromJSON(surveySecrets, values.ExtraSecrets); err != nil {
		return values, fmt.Errorf("survey secrets: %w", err)
	}

	for _, v := range surveyVars {
		if v.Target != db.SurveyVarTargetEnv {
			continue
		}
		if value, ok := values.ExtraVars[v.Name]; ok {
			values.EnvVars[v.Name] = value
			delete(values.ExtraVars, v.Name)
		}
		if v.Type != db.SurveyVarSecret {
			continue
		}
		if value, ok := values.ExtraSecrets[v.Name]; ok {
			values.EnvSecrets[v.Name] = value
			delete(values.ExtraSecrets, v.Name)
		}
	}

	// A survey variable left unanswered is an empty string where it would have
	// been, not a missing variable: the template declares it, the task did not
	// fill it.
	for _, v := range surveyVars {
		unanswered := values.ExtraVars
		switch {
		case v.Type == db.SurveyVarSecret && v.Target == db.SurveyVarTargetEnv:
			unanswered = values.EnvSecrets
		case v.Type == db.SurveyVarSecret:
			unanswered = values.ExtraSecrets
		case v.Target == db.SurveyVarTargetEnv:
			unanswered = values.EnvVars
		}
		if _, ok := unanswered[v.Name]; !ok {
			unanswered[v.Name] = ""
		}
	}

	return values, nil
}

// promptVarsFromJSON adds the members of a JSON object to vars, each as the
// text a prompt gets: a string as it is, a number as it is written, a list or
// an object as JSON.
func promptVarsFromJSON(object string, vars map[string]string) error {
	if object == "" {
		return nil
	}

	decoder := json.NewDecoder(bytes.NewReader([]byte(object)))
	decoder.UseNumber()

	parsed := map[string]any{}
	if err := decoder.Decode(&parsed); err != nil {
		return err
	}

	for name, value := range parsed {
		switch v := value.(type) {
		case string:
			vars[name] = v
		case nil:
			vars[name] = ""
		case json.Number:
			vars[name] = v.String()
		case bool:
			vars[name] = fmt.Sprintf("%t", v)
		default:
			encoded, err := json.Marshal(v)
			if err != nil {
				return err
			}
			vars[name] = string(encoded)
		}
	}

	return nil
}
