package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A doc comment ends up in a `.md` page that Docusaurus parses as MDX, where a
// bare `<` opens a JSX tag and a bare `{` opens an expression. Either one fails
// the whole documentation build, so the generator has to escape them — but not
// inside code spans, where MDX does not parse JSX and a backslash would show up
// verbatim in the rendered page.
func TestEscapeMDX(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"angle bracket outside code", "the /auth/oidc/<id>/login URL", `the /auth/oidc/\<id>/login URL`},
		{"angle bracket inside code", "run `kill -HUP <pid>` now", "run `kill -HUP <pid>` now"},
		{"brace outside code", "HOME is {template_dir}", `HOME is \{template_dir}`},
		{"brace inside code", "set `{\"a\": 1}`", "set `{\"a\": 1}`"},
		{"mixed", "a {x} and `a {y} one`", "a \\{x} and `a {y} one`"},
		{"nothing to escape", "plain sentence", "plain sentence"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, escapeMDX(tt.input))
		})
	}
}
