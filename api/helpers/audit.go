package helpers

import (
	"net/http"

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
