package db

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInventoryHostValidate_Name(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		wantErr bool
	}{
		{"plain", "web1", false},
		{"fqdn", "Web-1.example.com", false},
		{"ip", "10.0.0.5", false},
		{"max length", strings.Repeat("a", 250), false},
		{"empty", "", true},
		{"space", "web 1", true},
		{"tab", "web\t1", true},
		{"comma", "a,b", true},
		{"range", "web[1:3]", true},
		{"port", "db1:2222", true},
		{"too long", strings.Repeat("a", 251), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InventoryHost{Name: tt.host}.Validate()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

const aliasBomb = "&a [1,1,1,1,1,1,1,1,1,1], &b [*a,*a,*a,*a,*a,*a,*a,*a,*a,*a], " +
	"&c [*b,*b,*b,*b,*b,*b,*b,*b,*b,*b], &d [*c,*c,*c,*c,*c,*c,*c,*c,*c,*c], " +
	"&e [*d,*d,*d,*d,*d,*d,*d,*d,*d,*d], &f [*e,*e,*e,*e,*e,*e,*e,*e,*e,*e]"

func TestParseInventoryVariables(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantLen int
		wantKey bool
		wantErr bool
	}{
		{"empty", "", 0, false, false},
		{"comment", "# only a comment", 0, false, false},
		{"document marker", "---", 0, false, false},
		{"yaml", "a: 1\nb:\n  c: x", 2, true, false},
		{"json", `{"a": 1}`, 1, true, false},
		{"list", "- a\n- b", 0, false, true},
		{"scalar", "just text", 0, false, true},
		{"bad syntax", "a: [1", 0, false, true},
		{"leading document marker", "---\na: 1", 1, true, false},
		{"two documents", "a: 1\n---\nb: 2", 0, false, true},
		{"cycle under a null key", "&loop {null: *loop}", 0, false, true},
		{"alias to itself", "a: &x\n  b: *x", 0, false, true},
		{"null value", "a: ~", 1, true, false},
		{"cycle in an overridden merge", "a: 1\n<<: {a: &loop [*loop]}", 0, false, true},
		{"alias bomb in an overridden merge", "a: 1\n<<: {a: [" + aliasBomb + "]}", 0, false, true},
		{"repeated alias", "a: &x [1, 2]\nb: *x\nc: *x", 3, true, false},
		{"merge", "<<: {a: 1}\nb: 2", 2, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars, err := ParseInventoryVariables(tt.text)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, vars, tt.wantLen)
			if tt.wantKey {
				assert.Contains(t, vars, "a")
			}
		})
	}
}
