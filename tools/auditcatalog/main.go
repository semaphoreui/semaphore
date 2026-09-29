// Command auditcatalog generates the audit event reference from audit.Catalog.
package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/semaphoreui/semaphore/services/audit"
)

func main() {
	out := flag.String("out", "docs/docs/reference/audit-events.md", "output file")
	flag.Parse()
	if err := os.WriteFile(*out, []byte(render()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func render() string {
	var b strings.Builder
	b.WriteString("---\ntitle: Audit events\n")
	b.WriteString("description: Every security audit event Semaphore records through its API, with its type, outcomes, reasons, metadata fields and edition.\n---\n\n")
	b.WriteString("# Audit events\n\n")
	b.WriteString("<!-- Generated from the Semaphore source by tools/auditcatalog. Do not edit. -->\n\n")
	b.WriteString("See [Audit log](/admin-guide/audit-log) for the event schema.\n\n")
	b.WriteString("Events marked *Pro* are recorded only by Semaphore Pro.\n\n")
	b.WriteString("| event_code | action | type | outcomes | reasons | metadata | edition |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | --- |\n")
	for _, entry := range audit.Catalog() {
		outcomes := "success"
		switch {
		case entry.Type == audit.TypeDenied:
			outcomes = "failure"
		case len(entry.Reasons) > 0:
			outcomes = "success, failure"
		}
		// Analysts write SIEM rules against these values, so every allowed one is listed.
		reasons := make([]string, 0, len(entry.Reasons))
		for _, reason := range entry.Reasons {
			reasons = append(reasons, "`"+string(reason)+"`")
		}
		edition := ""
		if entry.Pro {
			edition = "Pro"
		}
		fmt.Fprintf(&b, "| `%s` | `%s` | `%s` | %s | %s | %s | %s |\n",
			entry.Kind.Code(), entry.Kind.Action(), entry.Type, outcomes, strings.Join(reasons, ", "), metadataFields(entry.Metadata), edition)
	}
	return b.String()
}

func metadataFields(metadata any) string {
	if metadata == nil {
		return ""
	}
	t := reflect.TypeOf(metadata)
	names := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		name, _, _ := strings.Cut(t.Field(i).Tag.Get("json"), ",")
		names = append(names, "`"+name+"`")
	}
	return strings.Join(names, ", ")
}
