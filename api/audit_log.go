package api

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	log "github.com/sirupsen/logrus"
)

const auditPageSize = 50

// canReadAuditLog is the only place that decides who may read the audit log.
func canReadAuditLog(user *db.User) bool {
	return user.Admin
}

func auditLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only this site's pages or a typed address may make a browser read the audit log.
		if site := r.Header.Get("Sec-Fetch-Site"); site != "" && site != "same-origin" && site != "none" {
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
			log.WithError(err).WithField("context", "audit_log").Error("Failed to read audit events")
			helpers.WriteErrorStatus(w, "Failed to read audit events", http.StatusInternalServerError)
			return
		}

		body := auditPage{Events: []audit.Envelope{}}
		for _, row := range page.Events {
			envelope, err := audit.EnvelopeFromRow(row)
			if err != nil {
				log.WithError(err).WithField("context", "audit_log").Error("Failed to read audit events")
				helpers.WriteErrorStatus(w, "Failed to read audit events", http.StatusInternalServerError)
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
	for _, name := range []string{"before", "after"} {
		if len(query[name]) > 1 {
			return 0, 0, fmt.Errorf("%s must be given once", name)
		}
	}
	if value := query.Get("before"); value != "" {
		if before, err = strconv.ParseInt(value, 10, 64); err != nil || before < 1 {
			return 0, 0, errors.New("before must be a positive number")
		}
	}
	if value := query.Get("after"); value != "" {
		// The store reads from after+1, which the largest int64 would overflow.
		if after, err = strconv.ParseInt(value, 10, 64); err != nil || after < 1 || after == math.MaxInt64 {
			return 0, 0, errors.New("after must be a number from 1 to 9223372036854775806")
		}
	}
	if before > 0 && after > 0 {
		return 0, 0, errors.New("before and after cannot be used together")
	}
	return before, after, nil
}
