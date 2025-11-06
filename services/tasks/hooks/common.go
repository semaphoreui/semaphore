package hooks

import "github.com/semaphoreui/semaphore/db"

type Hook interface {
	End(store db.Store, projectID int, taskID int)
}
