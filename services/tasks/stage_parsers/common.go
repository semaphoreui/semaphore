package stage_parsers

import (
	"github.com/semaphoreui/semaphore/db"
	"strings"
)

type StageResultParser interface {
	IsStart(currentStage *db.TaskStage, output db.TaskOutput) bool
	IsEnd(currentStage *db.TaskStage, output db.TaskOutput) bool
	NeedParse() bool
	Parse(outputs []db.TaskOutput) (map[string]any, error)
}

func GetStageResultParser(app db.TemplateApp, stageType db.TaskStageType) StageResultParser {

	if stageType == db.TaskStageInit {
		return &InitStageParser{}
	}

	switch app {
	case db.AppAnsible:
		switch stageType {
		case db.TaskStageRunning:
			return &AnsibleRunningStageParser{}
		case db.TaskStagePrintResult:
			return &AnsibleResultStageParser{}
		}
	}

	return nil
}

func GetAllTaskStages(app db.TemplateApp) []db.TaskStageType {
	switch app {
	case db.AppAnsible:
		return []db.TaskStageType{
			db.TaskStageInit,
			db.TaskStageRunning,
			db.TaskStagePrintResult,
		}
	}

	return nil
}

type InitStageParser struct{}

func (p InitStageParser) NeedParse() bool {
	return false
}

func (p InitStageParser) IsStart(currentStage *db.TaskStage, output db.TaskOutput) bool {
	if currentStage != nil {
		return false
	}

	return strings.HasPrefix(output.Output, "Run TaskRunner with template: ")
}

func (p InitStageParser) IsEnd(currentStage *db.TaskStage, output db.TaskOutput) bool {
	return false
}

func (p InitStageParser) Parse(outputs []db.TaskOutput) (map[string]any, error) {
	return nil, nil
}
