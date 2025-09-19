package hooks

import "github.com/Digital-Data-Co/forge/db"

type Hook interface {
	End(store db.Store, projectID int, taskID int)
}
