package api

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
)

const auditPageSize = 50

// canReadAuditLog is the only place that decides who may read the audit log.
func canReadAuditLog(user *db.User) bool {
	return user.Admin
}

func auditLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A link on another site must not make an administrator's browser export the audit log.
		if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		user := helpers.GetFromContext(r, "user").(*db.User)
		if !canReadAuditLog(user) {
			helpers.RecordDenied(r, "audit_log", 0)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type auditPage struct {
	Events []audit.Envelope `json:"events"`
	Older  *int64           `json:"older"`
	Newer  *int64           `json:"newer"`
}

// getAuditEvents pages the whole audit log, a request with any other parameter goes to filtered.
func getAuditEvents(filtered http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		for name := range query {
			if name != "before" && name != "after" {
				filtered(w, r)
				return
			}
		}

		before, after, err := auditPageCursor(query)
		if err != nil {
			helpers.WriteErrorStatus(w, err.Error(), http.StatusBadRequest)
			return
		}
		page, err := helpers.Store(r).ListAuditEvents(r.Context(), before, after, auditPageSize)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		body := auditPage{Events: []audit.Envelope{}}
		for _, row := range page.Events {
			envelope, err := audit.EnvelopeFromRow(row)
			if err != nil {
				helpers.WriteError(w, err)
				return
			}
			body.Events = append(body.Events, envelope)
		}
		if page.Older > 0 {
			body.Older = &page.Older
		}
		if page.Newer > 0 {
			body.Newer = &page.Newer
		}
		helpers.WriteJSON(w, http.StatusOK, body)
	}
}

func auditPageCursor(query url.Values) (before, after int64, err error) {
	if value := query.Get("before"); value != "" {
		if before, err = strconv.ParseInt(value, 10, 64); err != nil || before < 1 {
			return 0, 0, errors.New("before must be a positive number")
		}
	}
	if value := query.Get("after"); value != "" {
		if after, err = strconv.ParseInt(value, 10, 64); err != nil || after < 1 {
			return 0, 0, errors.New("after must be a positive number")
		}
	}
	if before > 0 && after > 0 {
		return 0, 0, errors.New("before and after cannot be used together")
	}
	return before, after, nil
}
