package alerting

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

const (
	TypeSlack      db.AlertType = "slack"
	TypeTeams      db.AlertType = "teams"
	TypeRocketChat db.AlertType = "rocketchat"
	TypeDingTalk   db.AlertType = "dingtalk"
)

func newSlackChannel() Channel {
	return newWebhookChannel(
		TypeSlack,
		"Slack",
		"mdi-slack",
		"slack.tmpl",
		[]int{200},
		map[task_logger.TaskStatus]string{
			task_logger.TaskSuccessStatus:       "good",
			task_logger.TaskFailStatus:          "danger",
			task_logger.TaskRunningStatus:       "#333CFF",
			task_logger.TaskWaitingStatus:       "#FFFC33",
			task_logger.TaskWaitingConfirmation: "#FFFC33",
			task_logger.TaskStoppingStatus:      "#BEBEBE",
			task_logger.TaskStoppedStatus:       "#5B5B5B",
		},
		func(cfg *util.ConfigType) bool { return cfg.SlackAlert },
		func(cfg *util.ConfigType) string { return cfg.SlackUrl },
	)
}

func newTeamsChannel() Channel {
	return newWebhookChannel(
		TypeTeams,
		"Microsoft Teams",
		"mdi-microsoft-teams",
		"microsoft-teams.tmpl",
		[]int{200, 202},
		nil,
		func(cfg *util.ConfigType) bool { return cfg.MicrosoftTeamsAlert },
		func(cfg *util.ConfigType) string { return cfg.MicrosoftTeamsUrl },
	)
}

func newRocketChatChannel() Channel {
	return newWebhookChannel(
		TypeRocketChat,
		"Rocket.Chat",
		"mdi-rocket-launch-outline",
		"rocketchat.tmpl",
		[]int{200},
		map[task_logger.TaskStatus]string{
			task_logger.TaskSuccessStatus:       "#00EE00",
			task_logger.TaskFailStatus:          "#EE0000",
			task_logger.TaskRunningStatus:       "#333CFF",
			task_logger.TaskWaitingStatus:       "#FFFC33",
			task_logger.TaskWaitingConfirmation: "#FFFC33",
			task_logger.TaskStoppingStatus:      "#BEBEBE",
			task_logger.TaskStoppedStatus:       "#5B5B5B",
		},
		func(cfg *util.ConfigType) bool { return cfg.RocketChatAlert },
		func(cfg *util.ConfigType) string { return cfg.RocketChatUrl },
	)
}

func newDingTalkChannel() Channel {
	return newWebhookChannel(
		TypeDingTalk,
		"DingTalk",
		"mdi-message-text-outline",
		"dingtalk.tmpl",
		[]int{200},
		nil,
		func(cfg *util.ConfigType) bool { return cfg.DingTalkAlert },
		func(cfg *util.ConfigType) string { return cfg.DingTalkUrl },
	)
}
