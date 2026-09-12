package tasks

import "github.com/semaphoreui/semaphore/db"

func ResolveAlerts(template db.Template, schedule *db.Schedule, defaultIDs []int) db.AlertSnapshot {
	onSuccess := db.ResolveAlertOn(template.AlertOnSuccess, template.SuppressSuccessAlerts)
	onError := db.ResolveAlertOn(template.AlertOnError, template.SuppressErrorAlerts)

	mode := template.AlertMode
	if mode == "" && len(template.AlertIDs) > 0 {
		mode = db.AlertModeIDs
	}

	var ids []int
	if mode == db.AlertModeIDs {
		ids = append([]int(nil), template.AlertIDs...)
	} else {
		ids = append([]int(nil), defaultIDs...)
	}

	if schedule != nil && schedule.AlertMode == db.AlertModeIDs {
		ids = append([]int(nil), schedule.AlertIDs...)
		if schedule.AlertOnSuccess != nil {
			onSuccess = *schedule.AlertOnSuccess
		}
		if schedule.AlertOnError != nil {
			onError = *schedule.AlertOnError
		}
	}

	if ids == nil {
		ids = []int{}
	}

	return db.AlertSnapshot{
		AlertIDs:  ids,
		OnSuccess: onSuccess,
		OnError:   onError,
	}
}

func SnapshotTaskAlerts(store db.Store, task *db.Task, template db.Template) error {
	var schedule *db.Schedule
	if task.ScheduleID != nil {
		s, err := store.GetSchedule(task.ProjectID, *task.ScheduleID)
		if err != nil {
			return err
		}
		schedule = &s
	}

	defaultIDs, err := store.GetDefaultAlertIDs(task.ProjectID)
	if err != nil {
		return err
	}

	snap := ResolveAlerts(template, schedule, defaultIDs)
	task.AlertSnapshot = &snap
	return nil
}
