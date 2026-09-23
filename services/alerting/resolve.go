package alerting

import (
	"github.com/semaphoreui/semaphore/db"
)

// Resolve decides, at task creation time, where a task reports:
//
//   - template in default mode: server channels (when the project allows
//     alerts, the pre-2.20.6 behaviour) plus the project alerts marked as
//     default;
//   - template in ids mode: only the alerts linked to the template;
//   - a schedule in ids mode replaces the template choice entirely.
//
// The template suppress_* flags are carried along so that a later edit does
// not change what an already queued task sends.
func Resolve(project db.Project, template db.Template, schedule *db.Schedule, defaultIDs []int) db.AlertSnapshot {
	snap := db.AlertSnapshot{
		OnSuccess: !template.SuppressSuccessAlerts,
		OnError:   !template.SuppressErrorAlerts,
	}

	mode := template.AlertMode
	if mode == "" {
		if len(template.AlertIDs) > 0 {
			mode = db.AlertModeIDs
		} else {
			mode = db.AlertModeDefault
		}
	}

	if mode == db.AlertModeIDs {
		snap.AlertIDs = uniqueIDs(template.AlertIDs)
	} else {
		snap.Instance = project.Alert
		snap.AlertIDs = uniqueIDs(defaultIDs)
	}

	if schedule != nil && schedule.AlertMode == db.AlertModeIDs {
		snap.Instance = false
		snap.AlertIDs = uniqueIDs(schedule.AlertIDs)
	}

	return snap
}

func uniqueIDs(ids []int) []int {
	out := make([]int, 0, len(ids))
	seen := make(map[int]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}
