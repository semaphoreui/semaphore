package bolt

import (
	"fmt"
	"github.com/semaphoreui/semaphore/pkg/conv"
)

type migration_2_14_7 struct {
	migration
}

func (d migration_2_14_7) Apply() (err error) {
	projectIDs, err := d.getProjectIDs()

	if err != nil {
		return
	}

	for _, projectID := range projectIDs {
		projectSchedules, err2 := d.getObjects(projectID, "schedule")
		if err2 != nil {
			return err2
		}

		for scheduleID, schedule := range projectSchedules {
			tplID, ok := conv.ConvertFloatToIntIfPossible(schedule["template_id"])
			if !ok {
				return fmt.Errorf("schedule template id %s is not a valid integer", schedule["template_id"])
			}

			tpl, err3 := d.getObject(projectID, "template", string(intObjectID(int(tplID)).ToBytes()))
			if err3 != nil {
				return err3
			}

			if tpl == nil {
				err3 = d.deleteObject(projectID, "schedule", scheduleID)
			}

			if err3 != nil {
				return err3
			}
		}
	}

	return
}
