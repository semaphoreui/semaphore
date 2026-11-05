package db

import (
	"encoding/json"
	"math"
)

// surveyVarSecret is the survey type the UI uses for secret variables; it has
// no constant in this branch.
const surveyVarSecret SurveyVarType = "secret"

// ValidateTaskSurveyValues checks the survey answers a user submitted with a
// new task against the template's survey. A key that is not declared in the
// survey is rejected: every key becomes an extra var / -var / CLI argument of
// the task and overrides the template's variable group, so an undeclared key
// would let a user who may only run tasks set arbitrary variables (for
// Ansible, connection variables such as ansible_ssh_common_args). An absent,
// null or empty answer is accepted.
//
// Both maps are checked because Secret is merged over Environment when the
// task runs (see LocalExecutor.getEnvironmentExtraVars): checking only
// Environment would let an answer bypass validation through Secret.
func ValidateTaskSurveyValues(vars []SurveyVar, environment string, secret string) error {
	env, err := parseSurveyAnswers(environment, "environment")
	if err != nil {
		return err
	}

	sec, err := parseSurveyAnswers(secret, "secret")
	if err != nil {
		return err
	}

	declared := make(map[string]struct{}, len(vars))
	for _, v := range vars {
		declared[v.Name] = struct{}{}
	}
	for _, answers := range []map[string]any{env, sec} {
		for name := range answers {
			if _, ok := declared[name]; !ok {
				return NewValidationError(
					"variable \"" + name + "\" is not declared in the template survey")
			}
		}
	}

	for _, v := range vars {
		if val, ok := env[v.Name]; ok {
			if v.Type == surveyVarSecret && !isEmptySurveyAnswer(val) {
				// The environment is stored on the task in clear.
				return NewValidationError(
					"survey variable \"" + v.Name + "\" is a secret and must be passed in the secret field")
			}
			if err = validateSurveyAnswer(v, val); err != nil {
				return err
			}
		}

		if val, ok := sec[v.Name]; ok {
			if err = validateSurveyAnswer(v, val); err != nil {
				return err
			}
		}
	}

	return nil
}

func parseSurveyAnswers(raw string, field string) (map[string]any, error) {
	res := make(map[string]any)
	if raw == "" {
		return res, nil
	}
	if err := json.Unmarshal([]byte(raw), &res); err != nil {
		return nil, NewValidationError("task " + field + " must be a JSON object")
	}
	return res, nil
}

func isEmptySurveyAnswer(val any) bool {
	switch v := val.(type) {
	case nil:
		return true
	case string:
		return v == ""
	case []any:
		return len(v) == 0
	}
	return false
}

// isSurveyInt reports whether s is an optionally negative decimal integer.
func isSurveyInt(s string) bool {
	if len(s) > 0 && s[0] == '-' {
		s = s[1:]
	}
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func surveyAnswerAllowed(v SurveyVar, val any) bool {
	s, ok := val.(string)
	if !ok {
		return false
	}
	for _, ev := range v.Values {
		if ev.Value == s {
			return true
		}
	}
	return false
}

func validateSurveyAnswer(v SurveyVar, val any) error {
	if isEmptySurveyAnswer(val) {
		return nil
	}

	invalid := func(msg string) error {
		return NewValidationError("survey variable \"" + v.Name + "\": " + msg)
	}

	// The SurveyVar* constants are declared as TemplateType in this branch.
	switch string(v.Type) {
	case string(SurveyVarInt):
		switch n := val.(type) {
		case string:
			if !isSurveyInt(n) {
				return invalid("value must be an integer")
			}
		case float64:
			if n != math.Trunc(n) {
				return invalid("value must be an integer")
			}
		default:
			return invalid("value must be an integer")
		}

	case string(SurveyVarEnum):
		if !surveyAnswerAllowed(v, val) {
			return invalid("value is not in values list")
		}

	default:
		// String, text and secret: scalar values only.
		switch val.(type) {
		case []any, map[string]any:
			return invalid("value must be a string")
		}
	}

	return nil
}
