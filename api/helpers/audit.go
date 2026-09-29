package helpers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/semaphoreui/semaphore/services/audit"
)

func Audit(r *http.Request) audit.Recorder {
	if recorder, ok := GetOkFromContext(r, "audit"); ok {
		if typed, ok := recorder.(audit.Recorder); ok {
			return typed
		}
	}
	return audit.Nop{}
}

func RouteTemplate(r *http.Request) string {
	route := mux.CurrentRoute(r)
	if route == nil {
		return ""
	}
	template, err := route.GetPathTemplate()
	if err != nil {
		return ""
	}
	return template
}

func RecordDenied(r *http.Request, permission string, projectID int) {
	Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.AuthAuthorizationDeny,
		Outcome:   audit.OutcomeFailure,
		Reason:    audit.ReasonForbidden,
		Target:    &audit.Target{Type: audit.TargetRoute, ID: RouteTemplate(r)},
		ProjectID: projectID,
		Metadata:  audit.DenyMetadata{Method: r.Method, Permission: permission},
	})
}
