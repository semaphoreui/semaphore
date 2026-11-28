package projects

import (
	"net/http"
	"strconv"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
)

// GetIntegrationCalls returns a page of the project's integration call history,
// newest first. The history is project-wide on purpose: an alias or an
// integration can be deleted and its past calls must stay visible. The optional
// `alias` and `integration_id` query parameters filter the list, `count` and
// `before` page it, and X-Has-Next reports whether older calls exist.
func GetIntegrationCalls(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)
	query := r.URL.Query()

	// The same keyset cursor the task history uses: `count` capped, `before` cursor.
	params, pageSize := parseTasksPageParams(query, helpers.QueryParams(r.URL))

	var integrationID *int
	if str := query.Get("integration_id"); str != "" {
		id, err := strconv.Atoi(str)
		if err != nil {
			helpers.WriteErrorStatus(w, "Invalid integration_id", http.StatusBadRequest)
			return
		}
		integrationID = &id
	}

	calls, err := helpers.Store(r).GetIntegrationCalls(project.ID, params, query.Get("alias"), integrationID, false)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	hasNext := len(calls) > pageSize
	if hasNext {
		calls = calls[:pageSize]
	}

	w.Header().Set("X-Has-Next", strconv.FormatBool(hasNext))
	helpers.WriteJSON(w, http.StatusOK, calls)
}

// GetIntegrationCall returns one call of the project's integration history,
// with the request body the list leaves out.
func GetIntegrationCall(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	callID, ok := helpers.GetIntParamOrAbort("call_id", w, r)
	if !ok {
		return
	}

	call, err := helpers.Store(r).GetIntegrationCall(project.ID, callID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, call)
}
