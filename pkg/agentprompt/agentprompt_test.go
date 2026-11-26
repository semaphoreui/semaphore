package agentprompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testValues() Values {
	return Values{
		ExtraVars:    map[string]string{"repo": "semaphoreui/semaphore", "nested": "${{ secrets.env.API_KEY }}", "a.b": "dotted"},
		EnvVars:      map[string]string{"STAGE": "prod"},
		ExtraSecrets: map[string]string{"db_password": "p@ss", "empty": ""},
		EnvSecrets:   map[string]string{"API_KEY": "key-123"},
	}
}

func TestExpand(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		expected string
		secrets  []string
	}{
		{"no expressions", "Review the last commit.", "Review the last commit.", nil},
		{"extra variable", "Repo: ${{ vars.extra.repo }}.", "Repo: semaphoreui/semaphore.", nil},
		{"environment variable", "Stage ${{vars.env.STAGE}}", "Stage prod", nil},
		{"secret extra variable", "pw=${{ secrets.extra.db_password }}", "pw=p@ss", []string{"p@ss"}},
		{"secret environment variable", "${{  secrets.env.API_KEY  }}", "key-123", []string{"key-123"}},
		{"several", "${{ vars.env.STAGE }}/${{ secrets.env.API_KEY }}/${{ vars.env.STAGE }}", "prod/key-123/prod",
			[]string{"key-123"}},
		{"a name with a dot", "${{ vars.extra.a.b }}", "dotted", nil},
		{"a value is not expanded again", "${{ vars.extra.nested }}", "${{ secrets.env.API_KEY }}", nil},
		{"an empty secret is not masked", "[${{ secrets.extra.empty }}]", "[]", nil},
		{"another context stays", "sha: ${{ github.sha }} ${{ matrix.os }}", "sha: ${{ github.sha }} ${{ matrix.os }}", nil},
		{"escaped", "write $${{ secrets.TOKEN }} there", "write ${{ secrets.TOKEN }} there", nil},
		{"unclosed stays", "cost ${{ vars.extra.repo", "cost ${{ vars.extra.repo", nil},
		{"multi-line", "a\n${{ vars.env.STAGE }}\nb", "a\nprod\nb", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, secrets, err := Expand(tt.prompt, testValues())

			require.NoError(t, err)
			assert.Equal(t, tt.expected, text)
			assert.Equal(t, tt.secrets, secrets)
		})
	}
}

func TestExpand_Errors(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		expected string
	}{
		{"missing extra variable", "${{ vars.extra.nope }}",
			`${{ vars.extra.nope }}: no extra variable "nope" in the Variable Groups of the template`},
		{"missing environment variable", "${{ vars.env.NOPE }}", `no environment variable "NOPE"`},
		{"missing secret", "${{ secrets.env.NOPE }}", `no secret environment variable "NOPE"`},
		{"a secret is not a variable", "${{ vars.extra.db_password }}", `no extra variable "db_password"`},
		{"no kind", "${{ vars.repo }}",
			"invalid prompt expression ${{ vars.repo }}: expected vars.extra.NAME or vars.env.NAME"},
		{"github style secret", "${{ secrets.API_KEY }}", "expected secrets.extra.NAME or secrets.env.NAME"},
		{"no name", "${{ vars.env. }}", "invalid prompt expression"},
		{"only the context", "${{ secrets }}", "invalid prompt expression"},
		{"a name with a space", "${{ vars.env.A B }}", "invalid prompt expression"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, secrets, err := Expand(tt.prompt, testValues())

			assert.ErrorContains(t, err, tt.expected)
			assert.Empty(t, text)
			assert.Nil(t, secrets)
		})
	}
}

func TestExpand_ErrorDoesNotTellValues(t *testing.T) {
	_, _, err := Expand("${{ secrets.env.API_KEY }} ${{ vars.extra.nope }}", testValues())

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "key-123")
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		expected string
	}{
		{"plain", "Review the last commit.", ""},
		{"a variable that may not exist", "${{ vars.extra.anything }}", ""},
		{"a secret", "${{ secrets.extra.anything }}", ""},
		{"an escaped secret", "$${{ secrets.extra.anything }}", ""},
		{"malformed", "${{ vars.anything }}", "invalid prompt expression"},
		{"another context", "${{ github.sha }}", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.prompt)

			if tt.expected == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, tt.expected)
		})
	}
}
