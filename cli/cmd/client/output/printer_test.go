package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type row struct {
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Status string   `json:"status"`
	Tags   []string `json:"tags,omitempty"`
	Nested struct {
		Count float64 `json:"count"`
	} `json:"nested"`
	Ptr *int `json:"ptr"`
}

func fixture() []row {
	a := row{ID: 1, Name: "alpha", Status: "success", Tags: []string{"x", "z"}}
	a.Nested.Count = 2
	b := row{ID: 22, Name: "beta", Status: "error"}
	b.Nested.Count = 1.5
	return []row{a, b}
}

func render(t *testing.T, opts Options, v any, cols []Column) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, New(&buf, opts).Print(v, cols))
	return buf.String()
}

func TestPrint_TableNoTTY_TabSeparated(t *testing.T) {
	out := render(t, Options{Format: FormatTable}, fixture(), Columns("id", "name", "status"))
	assert.Equal(t, "ID\tNAME\tSTATUS\n1\talpha\tsuccess\n22\tbeta\terror\n", out)
}

func TestPrint_TableTTY_Aligned(t *testing.T) {
	out := render(t, Options{Format: FormatTable, TTY: true}, fixture(), Columns("id", "name", "status"))
	assert.Equal(t, "ID  NAME   STATUS\n1   alpha  success\n22  beta   error\n", out)
}

func TestPrint_TableNoHeaders(t *testing.T) {
	out := render(t, Options{Format: FormatTable, NoHeaders: true}, fixture(), Columns("id", "name"))
	assert.Equal(t, "1\talpha\n22\tbeta\n", out)
}

func TestPrint_TableFieldsOverrideColumnsAndNested(t *testing.T) {
	out := render(t, Options{Format: FormatTable, Fields: []string{"name", "nested.count", "tags", "ptr"}}, fixture(), Columns("id"))
	assert.Equal(t, "NAME\tNESTED COUNT\tTAGS\tPTR\nalpha\t2\tx,z\t\nbeta\t1.5\t\t\n", out)
}

func TestPrint_TableInfersColumnsWhenNoneGiven(t *testing.T) {
	out := render(t, Options{Format: FormatTable}, []map[string]any{{"b": 1, "id": 7, "a": "x"}}, nil)
	assert.Equal(t, "ID\tA\tB\n7\tx\t1\n", out, "id first, then alphabetical")
}

func TestPrint_EmptyList(t *testing.T) {
	assert.Equal(t, "ID\tNAME\n", render(t, Options{Format: FormatTable}, []row{}, Columns("id", "name")))
	assert.Equal(t, "[]\n", render(t, Options{Format: FormatJSON}, []row{}, nil))
	assert.Equal(t, "[]\n", render(t, Options{Format: FormatJSON}, []row(nil), nil), "nil slice is still an array")
	assert.Equal(t, "", render(t, Options{Format: FormatTable, Quiet: true}, []row{}, nil))
}

func TestPrint_Quiet(t *testing.T) {
	out := render(t, Options{Format: FormatJSON, Quiet: true}, fixture(), nil)
	assert.Equal(t, "1\n22\n", out, "quiet wins over the format")
	assert.Equal(t, "1\n", render(t, Options{Quiet: true}, fixture()[0], nil))
}

func TestPrint_JSON(t *testing.T) {
	out := render(t, Options{Format: FormatJSON}, fixture(), nil)
	var back []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &back))
	require.Len(t, back, 2)
	assert.Equal(t, "alpha", back[0]["name"])
	assert.Contains(t, out, "  \"id\": 1", "indented")
	assert.NotContains(t, out, "\x1b[", "no colors in machine output")

	out = render(t, Options{Format: FormatJSON, Fields: []string{"id", "nested.count"}}, fixture(), nil)
	var picked []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &picked))
	assert.Equal(t, map[string]any{"id": float64(1), "nested.count": float64(2)}, picked[0])

	out = render(t, Options{Format: FormatJSON}, fixture()[1], nil)
	var one map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &one))
	assert.Equal(t, "beta", one["name"], "single object stays an object")
}

func TestPrint_YAML(t *testing.T) {
	out := render(t, Options{Format: FormatYAML}, fixture()[0], nil)
	assert.Contains(t, out, "id: 1\n")
	assert.Contains(t, out, "name: alpha\n")
	assert.Contains(t, out, "tags:\n  - x\n  - z\n")

	out = render(t, Options{Format: FormatYAML}, fixture(), nil)
	assert.True(t, bytes.HasPrefix([]byte(out), []byte("- id: 1\n")), out)
}

func TestPrint_SingleObjectTable(t *testing.T) {
	out := render(t, Options{Format: FormatTable}, map[string]any{"name": "n", "id": 3, "active": true}, nil)
	assert.Equal(t, "ID\t3\nACTIVE\ttrue\nNAME\tn\n", out)

	out = render(t, Options{Format: FormatTable, Fields: []string{"name"}}, map[string]any{"name": "n", "id": 3}, nil)
	assert.Equal(t, "NAME\tn\n", out)
}

func TestPrint_ColorOnlyOnTTY(t *testing.T) {
	cols := Columns("id", "status")
	plain := render(t, Options{Format: FormatTable, Color: true}, fixture(), cols)
	assert.NotContains(t, plain, "\x1b[", "color requested but no TTY: none emitted")

	colored := render(t, Options{Format: FormatTable, Color: true, TTY: true}, fixture(), cols)
	assert.Contains(t, colored, ansiGreen+"success"+ansiReset)
	assert.Contains(t, colored, ansiRed+"error"+ansiReset)
	assert.Contains(t, colored, ansiBold+"ID"+ansiReset)

	noColor := render(t, Options{Format: FormatTable, Color: false, TTY: true}, fixture(), cols)
	assert.NotContains(t, noColor, "\x1b[")
}

func TestPrint_NonObjectIsAnError(t *testing.T) {
	var buf bytes.Buffer
	assert.Error(t, New(&buf, Options{}).Print(42, nil))
	assert.Error(t, New(&buf, Options{}).Print([]int{1, 2}, nil))
}

func TestCellString(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"nil", nil, ""},
		{"string", "s", "s"},
		{"bool", true, "true"},
		{"whole float", float64(42), "42"},
		{"fraction", 1.25, "1.25"},
		{"list", []any{"a", float64(2), nil}, "a,2,"},
		{"object", map[string]any{"k": "v"}, `{"k":"v"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, cellString(tt.in))
		})
	}
}

func TestColorEnabled(t *testing.T) {
	env := func(m map[string]string) func(string) string {
		return func(k string) string { return m[k] }
	}
	assert.True(t, ColorEnabled(env(map[string]string{"TERM": "xterm-256color"})))
	assert.False(t, ColorEnabled(env(map[string]string{"NO_COLOR": "1"})))
	assert.False(t, ColorEnabled(env(map[string]string{"TERM": "dumb"})))
}

func TestIsTerminal(t *testing.T) {
	assert.False(t, IsTerminal(&bytes.Buffer{}))
	assert.False(t, IsTerminal(nil))
}
