package sql

import (
	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
)

type migration_2_20_6 struct {
	db *SqlDb
}

type migrationProjectAlert struct {
	ID    int    `db:"id"`
	Alert bool   `db:"alert"`
	Chat  string `db:"alert_chat"`
}

type migrationTemplateAlert struct {
	ID                    int  `db:"id"`
	ProjectID             int  `db:"project_id"`
	SuppressSuccessAlerts bool `db:"suppress_success_alerts"`
	SuppressErrorAlerts   bool `db:"suppress_error_alerts"`
}

func (m migration_2_20_6) insertAlert(
	tx *gorp.Transaction,
	projectID int,
	name string,
	alertType db.AlertType,
	chatID string,
	url string,
	token string,
	isDefault bool,
) (int64, error) {
	var chat, webhook, tok *string
	if chatID != "" {
		chat = &chatID
	}
	if url != "" {
		webhook = &url
	}
	if token != "" {
		tok = &token
	}

	insertQuery := "insert into project__alert " +
		"(project_id, name, `type`, enabled, is_default, chat_id, url, token) values (?, ?, ?, ?, ?, ?, ?, ?)"

	switch m.db.Sql().Dialect.(type) {
	case gorp.PostgresDialect:
		return tx.SelectInt(
			m.db.PrepareQuery(insertQuery+" returning id"),
			projectID, name, string(alertType), true, isDefault, chat, webhook, tok,
		)
	default:
		res, err := tx.Exec(
			m.db.PrepareQuery(insertQuery),
			projectID, name, string(alertType), true, isDefault, chat, webhook, tok,
		)
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	}
}

func (m migration_2_20_6) PostApply(tx *gorp.Transaction) error {
	if util.Config == nil {
		return nil
	}

	var projects []migrationProjectAlert
	_, err := tx.Select(
		&projects,
		m.db.PrepareQuery("select id, alert, coalesce(alert_chat, '') as alert_chat from project"),
	)
	if err != nil {
		return err
	}

	cfg := util.Config

	for _, p := range projects {
		seeds := []seed{
			{cfg.EmailAlert, "Email", db.AlertTypeEmail, "", "", ""},
			{cfg.TelegramAlert, "Telegram", db.AlertTypeTelegram, firstNonEmpty(p.Chat, cfg.TelegramChat), "", ""},
			{cfg.SlackAlert, "Slack", db.AlertTypeSlack, "", cfg.SlackUrl, ""},
			{cfg.MicrosoftTeamsAlert, "Microsoft Teams", db.AlertTypeTeams, "", cfg.MicrosoftTeamsUrl, ""},
			{cfg.RocketChatAlert, "Rocket.Chat", db.AlertTypeRocketChat, "", cfg.RocketChatUrl, ""},
			{cfg.DingTalkAlert, "DingTalk", db.AlertTypeDingTalk, "", cfg.DingTalkUrl, ""},
			{cfg.GotifyAlert, "Gotify", db.AlertTypeGotify, "", "", ""},
		}

		var created []int64
		for _, s := range seeds {
			if !seedHasDestination(s) {
				continue
			}
			id, err2 := m.insertAlert(tx, p.ID, s.name, s.typ, s.chat, s.url, s.token, p.Alert)
			if err2 != nil {
				return err2
			}
			created = append(created, id)
		}

		var templates []migrationTemplateAlert
		_, err = tx.Select(
			&templates,
			m.db.PrepareQuery(
				"select id, project_id, suppress_success_alerts, suppress_error_alerts "+
					"from project__template where project_id=?",
			),
			p.ID,
		)
		if err != nil {
			return err
		}

		for _, tpl := range templates {
			mode := db.AlertModeDefault
			if !p.Alert || len(created) == 0 || (tpl.SuppressSuccessAlerts && tpl.SuppressErrorAlerts) {
				mode = db.AlertModeIDs
			}
			_, err = tx.Exec(
				m.db.PrepareQuery("update project__template set alert_mode=? where id=? and project_id=?"),
				mode, tpl.ID, p.ID,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type seed struct {
	enabled bool
	name    string
	typ     db.AlertType
	chat    string
	url     string
	token   string
}

func seedHasDestination(s seed) bool {
	if !s.enabled {
		return false
	}
	switch s.typ {
	case db.AlertTypeTelegram:
		return s.chat != ""
	case db.AlertTypeSlack, db.AlertTypeTeams, db.AlertTypeRocketChat, db.AlertTypeDingTalk:
		return s.url != "" && db.ValidateAlertURL(s.url) == nil
	default:
		return true
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
