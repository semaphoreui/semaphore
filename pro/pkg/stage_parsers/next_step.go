package stage_parsers

import (
	"github.com/Digital-Data-Co/forge/db"
	"github.com/Digital-Data-Co/forge/pro_interfaces"
)

func MoveToNextStage(
	store db.Store,
	ansibleTaskRepo db.AnsibleTaskRepository,
	logWriter pro_interfaces.LogWriteService,
	app db.TemplateApp,
	projectID int,
	currentState any,
	currentStage *db.TaskStage,
	currentOutput *db.TaskOutput,
	newOutput db.TaskOutput,
) (newStage *db.TaskStage, newState any, err error) {
	return
}
