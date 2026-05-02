package projects

import "github.com/semaphoreui/semaphore/db"

// mergeTemplateEnvironmentIDs preserves variable-group associations when the client
// omits environment_ids in JSON (Unmarshal leaves the slice nil). Without this,
// SqlDb.UpdateTemplate would clear project__template_environment and drop all links.
// Explicit [] means clear; non-nil slice replaces; deprecated environment_id still applies.
func mergeTemplateEnvironmentIDs(updated *db.Template, previous db.Template) {
	if updated.EnvironmentIDs != nil {
		return
	}
	if updated.EnvironmentID > 0 {
		updated.EnvironmentIDs = []int{updated.EnvironmentID}
		return
	}
	updated.EnvironmentIDs = previous.EnvironmentIDs
}
