// Package output renders command results for people (a table) and for
// programs (JSON, YAML). The rule is the one gh, kubectl and doctl follow: data
// goes to stdout in the requested format, everything else goes to stderr, and
// nothing is reformatted silently when stdout is not a terminal.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// Formats accepted by -o/--output.
const (
	FormatTable = "table"
	FormatJSON  = "json"
	FormatYAML  = "yaml"
)

// Column describes one table column: the header and the JSON field it shows.
// Field is a dot path into the JSON form of a row ("id", "task.status").
type Column struct {
	Header string
	Field  string
}

// Columns builds a column list from JSON field names; the header is the field
// name upper-cased with dots and underscores turned into spaces.
func Columns(fields ...string) []Column {
	cols := make([]Column, 0, len(fields))
	for _, f := range fields {
		cols = append(cols, Column{Header: headerFor(f), Field: f})
	}
	return cols
}

func headerFor(field string) string {
	return strings.ToUpper(strings.NewReplacer(".", " ", "_", " ").Replace(field))
}

// Options select how Print renders.
type Options struct {
	Format string
	// Fields overrides the default columns (table) or filters the keys of
	// every object (json, yaml). Dot paths address nested fields.
	Fields []string
	// Quiet prints only the "id" field of every row, one per line.
	Quiet bool
	// NoHeaders drops the header line of a table.
	NoHeaders bool
	// TTY says whether the destination is a terminal: a table is then aligned
	// with spaces; otherwise cells are separated with a tab so that cut and awk
	// can split them.
	TTY bool
	// Color enables ANSI colors; only honoured when TTY is true.
	Color bool
}

// Printer writes results to one destination.
type Printer struct {
	out  io.Writer
	opts Options
}

// New returns a Printer writing to out.
func New(out io.Writer, opts Options) *Printer {
	if opts.Format == "" {
		opts.Format = FormatTable
	}
	return &Printer{out: out, opts: opts}
}

// Options returns the printer's settings.
func (p *Printer) Options() Options { return p.opts }

// Print renders v. v is a struct, a slice of structs, a map or a slice of
// maps — anything that marshals to a JSON object or an array of objects.
// columns are the default table columns for a list; a single object is shown
// as a two-column FIELD / VALUE table instead.
func (p *Printer) Print(v any, columns []Column) error {
	rows, isList, err := toRows(v)
	if err != nil {
		return err
	}

	if p.opts.Quiet {
		return p.printQuiet(rows)
	}

	switch p.opts.Format {
	case FormatJSON:
		return p.printJSON(rows, isList)
	case FormatYAML:
		return p.printYAML(rows, isList)
	case FormatTable:
		if isList {
			return p.printTable(rows, p.resolveColumns(columns))
		}
		return p.printObject(rows[0])
	default:
		return fmt.Errorf("unknown output format %q", p.opts.Format)
	}
}

// PrintRaw writes text as-is; for logs and other content that is not a record.
func (p *Printer) PrintRaw(text string) error {
	_, err := io.WriteString(p.out, text)
	return err
}

// Printf writes formatted text to the data stream.
func (p *Printer) Printf(format string, args ...any) {
	_, _ = fmt.Fprintf(p.out, format, args...)
}

func (p *Printer) printQuiet(rows []map[string]any) error {
	for _, row := range rows {
		if _, err := fmt.Fprintln(p.out, cellString(lookup(row, "id"))); err != nil {
			return err
		}
	}
	return nil
}

func (p *Printer) printJSON(rows []map[string]any, isList bool) error {
	enc := json.NewEncoder(p.out)
	enc.SetIndent("", "  ")
	return enc.Encode(p.filtered(rows, isList))
}

func (p *Printer) printYAML(rows []map[string]any, isList bool) error {
	enc := yaml.NewEncoder(p.out)
	enc.SetIndent(2)
	if err := enc.Encode(p.filtered(rows, isList)); err != nil {
		return err
	}
	return enc.Close()
}

// filtered applies --fields to machine output and keeps the list/object shape.
func (p *Printer) filtered(rows []map[string]any, isList bool) any {
	if len(p.opts.Fields) > 0 {
		out := make([]map[string]any, 0, len(rows))
		for _, row := range rows {
			picked := make(map[string]any, len(p.opts.Fields))
			for _, f := range p.opts.Fields {
				picked[f] = lookup(row, f)
			}
			out = append(out, picked)
		}
		rows = out
	}
	if isList {
		if rows == nil {
			return []map[string]any{}
		}
		return rows
	}
	return rows[0]
}

func (p *Printer) resolveColumns(defaults []Column) []Column {
	if len(p.opts.Fields) > 0 {
		return Columns(p.opts.Fields...)
	}
	return defaults
}

func (p *Printer) printTable(rows []map[string]any, columns []Column) error {
	if len(columns) == 0 {
		columns = inferColumns(rows)
	}

	var w io.Writer = p.out
	var tw *tabwriter.Writer
	sep := "\t"
	if p.opts.TTY {
		tw = tabwriter.NewWriter(p.out, 0, 0, 2, ' ', 0)
		w = tw
	}

	if !p.opts.NoHeaders {
		headers := make([]string, len(columns))
		for i, c := range columns {
			headers[i] = p.bold(c.Header)
		}
		if _, err := fmt.Fprintln(w, strings.Join(headers, sep)); err != nil {
			return err
		}
	}
	for _, row := range rows {
		cells := make([]string, len(columns))
		for i, c := range columns {
			cells[i] = p.colorCell(c.Field, cellString(lookup(row, c.Field)))
		}
		if _, err := fmt.Fprintln(w, strings.Join(cells, sep)); err != nil {
			return err
		}
	}
	if tw != nil {
		return tw.Flush()
	}
	return nil
}

// printObject shows one record vertically: FIELD then VALUE, every field.
func (p *Printer) printObject(row map[string]any) error {
	fields := p.opts.Fields
	if len(fields) == 0 {
		fields = sortedKeys(row)
	}
	var w io.Writer = p.out
	var tw *tabwriter.Writer
	sep := "\t"
	if p.opts.TTY {
		tw = tabwriter.NewWriter(p.out, 0, 0, 2, ' ', 0)
		w = tw
	}
	for _, f := range fields {
		if _, err := fmt.Fprintf(w, "%s%s%s\n", p.bold(headerFor(f)), sep, p.colorCell(f, cellString(lookup(row, f)))); err != nil {
			return err
		}
	}
	if tw != nil {
		return tw.Flush()
	}
	return nil
}

// toRows marshals v through JSON so that the printer sees exactly what the API
// returns, with the API field names.
func toRows(v any) ([]map[string]any, bool, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, false, err
	}
	data := strings.TrimSpace(string(raw))
	switch {
	case len(data) == 0 || data == "null":
		return nil, true, nil
	case data[0] == '[':
		var rows []map[string]any
		if err := json.Unmarshal([]byte(data), &rows); err != nil {
			return nil, true, fmt.Errorf("output is a list of non-objects: %w", err)
		}
		return rows, true, nil
	case data[0] == '{':
		var row map[string]any
		if err := json.Unmarshal([]byte(data), &row); err != nil {
			return nil, false, err
		}
		return []map[string]any{row}, false, nil
	default:
		return nil, false, fmt.Errorf("cannot print a %T as a table", v)
	}
}

// lookup follows a dot path through nested maps.
func lookup(row map[string]any, path string) any {
	var cur any = row
	for _, part := range strings.Split(path, ".") {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur, ok = m[part]
		if !ok {
			return nil
		}
	}
	return cur
}

// cellString renders a JSON value for a table cell.
func cellString(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case bool:
		return strconv.FormatBool(x)
	case float64:
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	case []any:
		parts := make([]string, 0, len(x))
		for _, e := range x {
			parts = append(parts, cellString(e))
		}
		return strings.Join(parts, ",")
	default:
		b, err := json.Marshal(x)
		if err != nil {
			return fmt.Sprint(x)
		}
		return string(b)
	}
}

func inferColumns(rows []map[string]any) []Column {
	if len(rows) == 0 {
		return nil
	}
	return Columns(sortedKeys(rows[0])...)
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// id first, the rest alphabetical: that is how people read a record.
	for i, k := range keys {
		if k == "id" {
			copy(keys[1:i+1], keys[:i])
			keys[0] = "id"
			break
		}
	}
	return keys
}

// IsTerminal reports whether w is a character device (an interactive
// terminal). It uses only the standard library.
func IsTerminal(w any) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// ColorEnabled applies the conventions of no-color.org: a NO_COLOR variable
// (any value) or TERM=dumb disables colour; the caller adds --no-color and the
// TTY check on top.
func ColorEnabled(getenv func(string) string) bool {
	if _, set := lookupEnv(getenv, "NO_COLOR"); set {
		return false
	}
	return getenv("TERM") != "dumb"
}

func lookupEnv(getenv func(string) string, key string) (string, bool) {
	v := getenv(key)
	return v, v != ""
}
