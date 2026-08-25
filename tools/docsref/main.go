// Command docsref generates the configuration reference page of the
// documentation site directly from util.ConfigType.
//
// The reference used to be a hand-written table in the docs repository, which
// drifted from the code: defaults, environment variable names and whole keys
// were wrong. Generating it means the table cannot be stale while the build is
// green, and adding a config option updates the docs for free.
//
// Usage:
//
//	go run ./tools/docsref -out docs/docs/reference/configuration.md
//
// The struct tags supply the key name, the environment variable and the
// default; the doc comment above each field supplies the description. Grouping
// and edition badges come from groups.json next to this file, because neither
// can be derived from the Go source.
//
// descriptions.json is a bridge, not a second source of truth: it holds the
// prose that used to live in the hand-written documentation table for fields
// that have no doc comment yet. A doc comment always wins, and the command
// prints how many options still depend on the file, so the number can only be
// driven down.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// option is one leaf of the configuration tree: a key that a user can actually
// set in config.json.
type option struct {
	Key         string // dotted json path, e.g. "runner.executor.type"
	Env         string // SEMAPHORE_* variable, empty when the field has none
	Type        string // string, boolean, integer, array, object
	Default     string
	Values      []string // allowed values parsed from the rule tag
	Sensitive   bool
	Description string
	fromOverlay bool // description came from descriptions.json, not the Go source
}

// group is a section of the generated page. Keys are matched against Prefixes
// in order, so the first group that claims a key wins; anything unclaimed lands
// in the trailing "Other" group and is therefore visible rather than lost.
type group struct {
	Title    string   `json:"title"`
	Intro    string   `json:"intro,omitempty"`
	Edition  string   `json:"edition,omitempty"` // "pro" or "enterprise": badge for every key in the group
	Prefixes []string `json:"prefixes"`
}

type overlay struct {
	Groups   []group           `json:"groups"`
	Editions map[string]string `json:"editions"` // key prefix -> pro|enterprise
}

func main() {
	var (
		utilDir     = flag.String("util", "util", "directory of the util package")
		overlayPath = flag.String("groups", "tools/docsref/groups.json", "grouping and edition overlay")
		descPath    = flag.String("descriptions", "tools/docsref/descriptions.json", "descriptions for fields without a doc comment")
		out         = flag.String("out", "", "file to write (default: stdout)")
	)
	flag.Parse()

	structs, scalars, err := parsePackage(*utilDir)
	if err != nil {
		fail(err)
	}
	root, ok := structs["ConfigType"]
	if !ok {
		fail(fmt.Errorf("util.ConfigType not found in %s", *utilDir))
	}

	w := &walker{structs: structs, scalars: scalars}
	w.walk(root, "")

	ov, err := readOverlay(*overlayPath)
	if err != nil {
		fail(err)
	}

	descriptions, err := readDescriptions(*descPath)
	if err != nil {
		fail(err)
	}
	applyDescriptions(w.options, descriptions)

	page := render(w.options, ov)
	if *out == "" {
		fmt.Print(page)
		return
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		fail(err)
	}
	if err := os.WriteFile(*out, []byte(page), 0o644); err != nil {
		fail(err)
	}
	fmt.Fprintf(os.Stderr, "docsref: wrote %d options to %s\n", len(w.options), *out)
	reportDebt(w.options, descriptions)
}

// reportDebt keeps the two kinds of documentation debt visible on every run:
// options nobody has described anywhere, and descriptions.json entries that are
// no longer needed because the Go field has grown a doc comment (or has gone).
func reportDebt(opts []option, descriptions map[string]string) {
	used := map[string]bool{}
	var undocumented []string
	for _, o := range opts {
		if o.Description == "" {
			undocumented = append(undocumented, o.Key)
		}
		if _, ok := descriptions[o.Key]; ok && o.fromOverlay {
			used[o.Key] = true
		}
	}
	var stale []string
	for k := range descriptions {
		if !used[k] {
			stale = append(stale, k)
		}
	}
	sort.Strings(undocumented)
	sort.Strings(stale)

	if n := len(undocumented); n > 0 {
		fmt.Fprintf(os.Stderr, "docsref: %d option(s) have no description: %s\n", n, strings.Join(undocumented, ", "))
	}
	if n := len(stale); n > 0 {
		fmt.Fprintf(os.Stderr, "docsref: %d descriptions.json entr(ies) are unused and should be deleted: %s\n", n, strings.Join(stale, ", "))
	}
	if len(undocumented) == 0 && len(descriptions) == 0 {
		fmt.Fprintln(os.Stderr, "docsref: every option is documented in the Go source")
	}
}

func readDescriptions(path string) (map[string]string, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	m := map[string]string{}
	return m, json.Unmarshal(b, &m)
}

func applyDescriptions(opts []option, descriptions map[string]string) {
	for i := range opts {
		if opts[i].Description != "" {
			continue
		}
		if d, ok := descriptions[opts[i].Key]; ok {
			opts[i].Description = strings.ReplaceAll(d, "|", `\|`)
			opts[i].fromOverlay = true
		}
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "docsref:", err)
	os.Exit(1)
}

// --- parsing ---------------------------------------------------------------

// parsePackage returns every named struct type in the package, plus the set of
// named types that are really scalars (`type Foo string`) so that the walker
// does not try to descend into them.
func parsePackage(dir string) (map[string]*ast.StructType, map[string]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, err
	}

	fset := token.NewFileSet()
	structs := map[string]*ast.StructType{}
	scalars := map[string]string{}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		// Every .go file is parsed, build tags included: the platform-specific
		// files carry config fields that must show up in the reference too.
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ParseComments)
		if err != nil {
			return nil, nil, err
		}

		ast.Inspect(file, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			switch t := ts.Type.(type) {
			case *ast.StructType:
				structs[ts.Name.Name] = t
			case *ast.Ident:
				scalars[ts.Name.Name] = jsonType(t.Name)
			}
			return true
		})
	}
	return structs, scalars, nil
}

type walker struct {
	structs map[string]*ast.StructType
	scalars map[string]string
	options []option
	seen    []string // struct names on the current path, guards against cycles
}

func (w *walker) walk(st *ast.StructType, prefix string) {
	for _, field := range st.Fields.List {
		tag := structTag(field)
		name, ok := jsonName(tag)
		if !ok {
			continue
		}
		key := name
		if prefix != "" {
			key = prefix + "." + name
		}

		if nested, isStruct := w.structOf(field.Type); isStruct {
			if w.entered(nested.name) {
				continue
			}
			// A struct field that carries its own env tag can be set as one JSON
			// value, so the block itself is a settable option and gets a row of
			// its own in addition to its children.
			if env, sensitive := envVar(tag); env != "" {
				w.options = append(w.options, option{
					Key:         key,
					Env:         env,
					Type:        "object",
					Sensitive:   sensitive,
					Description: describe(field),
				})
			}
			w.seen = append(w.seen, nested.name)
			w.walk(nested.st, key)
			w.seen = w.seen[:len(w.seen)-1]
			continue
		}

		env, sensitive := envVar(tag)
		w.options = append(w.options, option{
			Key:         key,
			Env:         env,
			Type:        w.typeOf(field.Type),
			Default:     tagValue(tag, "default"),
			Values:      allowedValues(tagValue(tag, "rule")),
			Sensitive:   sensitive,
			Description: describe(field),
		})
	}
}

func (w *walker) entered(name string) bool {
	for _, s := range w.seen {
		if s == name {
			return true
		}
	}
	return false
}

type namedStruct struct {
	name string
	st   *ast.StructType
}

// structOf reports whether a field type is a struct we should descend into.
// Slices and maps of structs are leaves: a user sets them as one JSON value.
func (w *walker) structOf(expr ast.Expr) (namedStruct, bool) {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return w.structOf(t.X)
	case *ast.Ident:
		if st, ok := w.structs[t.Name]; ok {
			return namedStruct{t.Name, st}, true
		}
	case *ast.StructType:
		return namedStruct{"", t}, true
	}
	return namedStruct{}, false
}

func (w *walker) typeOf(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return w.typeOf(t.X)
	case *ast.ArrayType:
		return "array"
	case *ast.MapType:
		return "object"
	case *ast.SelectorExpr:
		return "object"
	case *ast.Ident:
		if s, ok := w.scalars[t.Name]; ok {
			return s
		}
		return jsonType(t.Name)
	}
	return "string"
}

func jsonType(goType string) string {
	switch goType {
	case "bool":
		return "boolean"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return "integer"
	case "float32", "float64":
		return "number"
	case "string":
		return "string"
	}
	return "object"
}

func structTag(f *ast.Field) string {
	if f.Tag == nil {
		return ""
	}
	tag, err := strconv.Unquote(f.Tag.Value)
	if err != nil {
		return ""
	}
	return tag
}

var tagRe = regexp.MustCompile(`(\w+):"([^"]*)"`)

func tagValue(tag, key string) string {
	for _, m := range tagRe.FindAllStringSubmatch(tag, -1) {
		if m[1] == key {
			return m[2]
		}
	}
	return ""
}

func jsonName(tag string) (string, bool) {
	v := tagValue(tag, "json")
	if v == "" || v == "-" {
		return "", false
	}
	name := strings.Split(v, ",")[0]
	if name == "" || name == "-" {
		return "", false
	}
	return name, true
}

func envVar(tag string) (string, bool) {
	v := tagValue(tag, "env")
	if v == "" {
		return "", false
	}
	parts := strings.Split(v, ",")
	sensitive := false
	for _, p := range parts[1:] {
		if strings.TrimSpace(p) == "sensitive" {
			sensitive = true
		}
	}
	return strings.TrimSpace(parts[0]), sensitive
}

// allowedValues turns the simple alternation regexes used in `rule` tags into a
// list. Anything more complicated is ignored rather than guessed at.
var simpleRule = regexp.MustCompile(`^\^\(?([\w|]+)\)?\??\$$`)

func allowedValues(rule string) []string {
	m := simpleRule.FindStringSubmatch(rule)
	if m == nil {
		return nil
	}
	var vals []string
	for _, v := range strings.Split(m[1], "|") {
		if v != "" {
			vals = append(vals, v)
		}
	}
	if len(vals) < 2 {
		return nil
	}
	return vals
}

// describe returns the first paragraph of the field's doc comment, with the Go
// field name stripped from the front so that sentences read naturally in the
// table ("Path to..." rather than "TmpPath is a path to...").
func describe(f *ast.Field) string {
	text := ""
	if f.Doc != nil {
		text = f.Doc.Text()
	} else if f.Comment != nil {
		text = f.Comment.Text()
	}
	if text == "" {
		return ""
	}
	if p, _, ok := strings.Cut(text, "\n\n"); ok {
		text = p
	}
	text = strings.Join(strings.Fields(text), " ")

	// Go doc comments conventionally open with the identifier ("TmpPath is a path
	// to..."), which reads wrong in a table keyed by the JSON name. Drop that
	// opening word, and the linking verb with it when there is one.
	if len(f.Names) > 0 {
		name := f.Names[0].Name
		if rest, ok := strings.CutPrefix(text, name+" "); ok {
			for _, verb := range []string{"is a ", "is an ", "is the ", "is ", "are "} {
				if trimmed, found := strings.CutPrefix(rest, verb); found {
					rest = trimmed
					break
				}
			}
			if rest != "" {
				text = strings.ToUpper(rest[:1]) + rest[1:]
			}
		}
	}
	return escapeMDX(strings.ReplaceAll(text, "|", `\|`))
}

// escapeMDX makes a doc comment safe to drop into a `.md` page, which Docusaurus
// parses as MDX. There, a bare `<` starts a JSX tag and a bare `{` starts an
// expression, so a comment mentioning `<id>` fails the whole build. Backslash
// escapes render as the literal character in MDX, but only outside code spans —
// inside backticks a backslash stays a backslash, and MDX does not parse JSX
// there anyway, so those spans are left alone.
func escapeMDX(text string) string {
	var b strings.Builder
	inCode := false
	for _, r := range text {
		switch {
		case r == '`':
			inCode = !inCode
		case !inCode && (r == '<' || r == '{'):
			b.WriteRune('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// --- rendering -------------------------------------------------------------

func readOverlay(path string) (overlay, error) {
	var ov overlay
	b, err := os.ReadFile(path)
	if err != nil {
		return ov, err
	}
	return ov, json.Unmarshal(b, &ov)
}

func render(opts []option, ov overlay) string {
	claimed := map[string]bool{}
	var b strings.Builder

	b.WriteString(`---
title: Configuration options
description: "Every option Semaphore accepts in config.json, with its environment variable, type, default, and meaning."
---

{/* Generated by tools/docsref in the semaphoreui/semaphore repository. Do not edit by hand. */}

# Configuration options

Every key below can be set in ` + "`config.json`" + ` or as the environment variable
next to it. Environment variables win over the file. Nested keys are written with
dots: ` + "`runner.executor.type`" + ` is ` + "`{\"runner\": {\"executor\": {\"type\": ...}}}`" + `.

This page is generated from the Semaphore source, so it always matches the
release it ships with. For how to supply these options, see
[Configuration](/admin-guide/configuration).

`)

	for _, g := range ov.Groups {
		var rows []option
		for _, o := range opts {
			if claimed[o.Key] {
				continue
			}
			if matchesAny(o.Key, g.Prefixes) {
				claimed[o.Key] = true
				rows = append(rows, o)
			}
		}
		if len(rows) == 0 {
			continue
		}
		writeSection(&b, g, rows, ov.Editions)
	}

	var rest []option
	for _, o := range opts {
		if !claimed[o.Key] {
			rest = append(rest, o)
		}
	}
	if len(rest) > 0 {
		sort.Slice(rest, func(i, j int) bool { return rest[i].Key < rest[j].Key })
		writeSection(&b, group{
			Title: "Other",
			Intro: "Options that have not been sorted into a section yet.",
		}, rest, ov.Editions)
	}

	return b.String()
}

func writeSection(b *strings.Builder, g group, rows []option, editions map[string]string) {
	fmt.Fprintf(b, "## %s {#%s}\n\n", g.Title, anchor(g.Title))
	if g.Intro != "" {
		fmt.Fprintf(b, "%s\n\n", g.Intro)
	}
	b.WriteString("| Option / Environment variable | Type / Default | Description |\n")
	b.WriteString("|---|---|---|\n")
	for _, o := range rows {
		badge := ""
		switch edition(o.Key, g.Edition, editions) {
		case "pro":
			badge = " <Pro />"
		case "enterprise":
			badge = " <Enterprise />"
		}

		optionAndEnv := "`" + o.Key + "`" + badge
		if o.Env != "" {
			optionAndEnv += "<br />`" + o.Env + "`"
		}
		typeAndDefault := o.Type
		if o.Default != "" {
			typeAndDefault += "<br />Default: `" + o.Default + "`"
		}
		desc := o.Description
		if len(o.Values) > 0 {
			quoted := make([]string, len(o.Values))
			for i, v := range o.Values {
				quoted[i] = "`" + v + "`"
			}
			desc = strings.TrimSpace(desc + " One of " + strings.Join(quoted, ", ") + ".")
		}
		if o.Sensitive {
			desc = strings.TrimSpace(desc + " Secret: keep it out of shell history and version control.")
		}
		if desc == "" {
			desc = "—"
		}
		fmt.Fprintf(b, "| %s | %s | %s |\n", optionAndEnv, typeAndDefault, desc)
	}
	b.WriteString("\n")
}

func edition(key, groupEdition string, editions map[string]string) string {
	best, bestLen := groupEdition, -1
	for prefix, ed := range editions {
		if matches(key, prefix) && len(prefix) > bestLen {
			best, bestLen = ed, len(prefix)
		}
	}
	return best
}

func matchesAny(key string, prefixes []string) bool {
	for _, p := range prefixes {
		if matches(key, p) {
			return true
		}
	}
	return false
}

// matches treats a prefix ending in "." as a subtree and anything else as an
// exact key, so that "log." claims log.tasks.format without "email" claiming
// "email_alert" by accident.
func matches(key, prefix string) bool {
	if strings.HasSuffix(prefix, ".") {
		return strings.HasPrefix(key, prefix)
	}
	if strings.HasSuffix(prefix, "*") {
		return strings.HasPrefix(key, strings.TrimSuffix(prefix, "*"))
	}
	return key == prefix
}

var nonAnchor = regexp.MustCompile(`[^a-z0-9]+`)

func anchor(title string) string {
	return strings.Trim(nonAnchor.ReplaceAllString(strings.ToLower(title), "-"), "-")
}
