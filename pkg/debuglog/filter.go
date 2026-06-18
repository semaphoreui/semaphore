// Package debuglog implements Node.js-`debug`-style selective debug logging on
// top of logrus. The value of the SEMAPHORE_DEBUG_FILTER environment variable is
// parsed into a Filter that decides which debug namespaces (the `context` field
// attached via log.WithFields) may emit output. The filter only narrows
// DEBUG-level entries and only when the log level is already DEBUG.
package debuglog

import (
	"regexp"
	"strings"
)

// Filter compiles a Node-`debug`-style namespace pattern list and reports which
// namespaces are enabled. A namespace is enabled when it matches at least one
// include pattern and no exclude pattern (exclusions always win).
type Filter struct {
	includes []*regexp.Regexp
	excludes []*regexp.Regexp
}

// Parse builds a Filter from a SEMAPHORE_DEBUG_FILTER spec. Tokens are separated
// by commas or whitespace. A token starting with '-' is an exclusion; '*' is a
// wildcard matching any sequence of characters within a namespace. An empty spec
// yields a Filter whose Enabled always returns false.
func Parse(spec string) *Filter {
	f := &Filter{}

	tokens := strings.FieldsFunc(spec, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
	})

	for _, tok := range tokens {
		exclude := false
		if strings.HasPrefix(tok, "-") {
			exclude = true
			tok = tok[1:]
		}

		if tok == "" {
			continue
		}

		re := compilePattern(tok)
		if exclude {
			f.excludes = append(f.excludes, re)
		} else {
			f.includes = append(f.includes, re)
		}
	}

	return f
}

// compilePattern converts a glob-like token (where '*' is a wildcard) into an
// anchored regular expression. All non-`*` characters are matched literally.
func compilePattern(tok string) *regexp.Regexp {
	parts := strings.Split(tok, "*")
	for i, p := range parts {
		parts[i] = regexp.QuoteMeta(p)
	}
	return regexp.MustCompile("^" + strings.Join(parts, ".*") + "$")
}

// Enabled reports whether the given namespace (the `context` field) should emit
// debug output under this filter. A namespace with no `context` field is passed
// as "" and is only enabled by an explicit "*" pattern.
func (f *Filter) Enabled(namespace string) bool {
	matched := false
	for _, re := range f.includes {
		if re.MatchString(namespace) {
			matched = true
			break
		}
	}

	if !matched {
		return false
	}

	for _, re := range f.excludes {
		if re.MatchString(namespace) {
			return false
		}
	}

	return true
}

// Active reports whether any pattern (include or exclude) was configured.
func (f *Filter) Active() bool {
	return len(f.includes) > 0 || len(f.excludes) > 0
}
