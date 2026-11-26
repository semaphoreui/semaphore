// Package agentprompt replaces the ${{ … }} expressions of an AI agent's
// prompt with the variables and secrets of the template's Variable Groups:
//
//	${{ vars.extra.NAME }}     an extra variable
//	${{ vars.env.NAME }}       an environment variable
//	${{ secrets.extra.NAME }}  a secret extra variable
//	${{ secrets.env.NAME }}    a secret environment variable
//
// An expression of another context (${{ github.sha }} in a workflow the agent
// is asked to write) is not ours and stays as it is; $${{ … }} gives a literal
// ${{ … }}. A value put into the prompt is not looked through again.
package agentprompt

import (
	"fmt"
	"strings"
)

const (
	open  = "${{"
	shut  = "}}"
	vars  = "vars"
	secrs = "secrets"
	extra = "extra"
	env   = "env"
)

// Values are what the expressions of a prompt can name.
type Values struct {
	// ExtraVars and EnvVars are the plain extra and environment variables.
	ExtraVars map[string]string
	EnvVars   map[string]string
	// ExtraSecrets and EnvSecrets are the secret ones.
	ExtraSecrets map[string]string
	EnvSecrets   map[string]string
}

// reference is a parsed vars.… or secrets.… expression.
type reference struct {
	secret bool
	env    bool
	name   string
}

func (r reference) String() string {
	context, kind := vars, extra
	if r.secret {
		context = secrs
	}
	if r.env {
		kind = env
	}
	return fmt.Sprintf("%s %s.%s.%s %s", open, context, kind, r.name, shut)
}

// parse reads the inside of ${{ … }}. ours is false for an expression of
// another context, which is left alone.
func parse(expr string) (ref reference, ours bool, err error) {
	expr = strings.TrimSpace(expr)
	context, rest, _ := strings.Cut(expr, ".")

	switch context {
	case vars:
	case secrs:
		ref.secret = true
	default:
		return ref, false, nil
	}

	malformed := fmt.Errorf("invalid prompt expression %s %s %s: expected %s.extra.NAME or %s.env.NAME",
		open, expr, shut, context, context)

	kind, name, _ := strings.Cut(rest, ".")
	switch kind {
	case extra:
	case env:
		ref.env = true
	default:
		return ref, true, malformed
	}

	if name == "" || strings.ContainsAny(name, " \t\r\n") {
		return ref, true, malformed
	}

	ref.name = name
	return ref, true, nil
}

// walk copies prompt to the result, giving every expression of ours to
// replace.
func walk(prompt string, replace func(reference) (string, error)) (string, error) {
	var b strings.Builder

	rest := prompt
	for {
		i := strings.Index(rest, open)
		if i < 0 {
			break
		}

		end := strings.Index(rest[i+len(open):], shut)
		if end < 0 {
			break
		}
		exprEnd := i + len(open) + end
		after := exprEnd + len(shut)

		// $${{ … }} is a literal ${{ … }}.
		if i > 0 && rest[i-1] == '$' {
			b.WriteString(rest[:i-1])
			b.WriteString(rest[i:after])
			rest = rest[after:]
			continue
		}

		ref, ours, err := parse(rest[i+len(open) : exprEnd])
		if err != nil {
			return "", err
		}

		b.WriteString(rest[:i])
		if ours {
			value, err := replace(ref)
			if err != nil {
				return "", err
			}
			b.WriteString(value)
		} else {
			b.WriteString(rest[i:after])
		}
		rest = rest[after:]
	}

	b.WriteString(rest)
	return b.String(), nil
}

// Validate checks the expressions of prompt without their values.
func Validate(prompt string) error {
	_, err := walk(prompt, func(reference) (string, error) {
		return "", nil
	})
	return err
}

// Expand returns prompt with its expressions replaced, and the secret values
// it put there: whoever logs the agent's output masks them. An expression
// naming a variable that does not exist is an error, so a prompt never runs
// with a hole in it.
func Expand(prompt string, values Values) (string, []string, error) {
	var secrets []string

	text, err := walk(prompt, func(ref reference) (string, error) {
		var (
			source map[string]string
			what   string
		)
		switch {
		case ref.secret && ref.env:
			source, what = values.EnvSecrets, "secret environment variable"
		case ref.secret:
			source, what = values.ExtraSecrets, "secret extra variable"
		case ref.env:
			source, what = values.EnvVars, "environment variable"
		default:
			source, what = values.ExtraVars, "extra variable"
		}

		value, ok := source[ref.name]
		if !ok {
			return "", fmt.Errorf("%s: no %s %q in the Variable Groups of the template", ref, what, ref.name)
		}

		if ref.secret && value != "" {
			secrets = append(secrets, value)
		}
		return value, nil
	})
	if err != nil {
		return "", nil, err
	}

	return text, secrets, nil
}
