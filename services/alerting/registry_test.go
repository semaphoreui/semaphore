package alerting

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistry_BuiltinChannels(t *testing.T) {
	registry := NewRegistry()

	var types []db.AlertType
	for _, ch := range registry.List() {
		types = append(types, ch.Type())
		assert.NotEmpty(t, ch.Title())
		assert.NotEmpty(t, ch.Icon())
		assert.NotEmpty(t, ch.DefaultBody())
		assert.NotEmpty(t, ch.DefaultEvents())
	}
	assert.ElementsMatch(t, []db.AlertType{
		TypeTelegram, TypeSlack, TypeEmail, TypeTeams, TypeRocketChat, TypeDingTalk, TypeGotify,
	}, types)

	_, err := registry.Get("pager")
	assert.ErrorContains(t, err, "unknown alert type")
}

func TestRegistry_DefaultEventsMatchLegacyBehaviour(t *testing.T) {
	registry := NewRegistry()

	email, err := registry.Get(TypeEmail)
	require.NoError(t, err)
	assert.Equal(t, db.AlertEvents{db.AlertEventError}, email.DefaultEvents(), "e-mail only reported failures")

	for _, typ := range []db.AlertType{TypeTelegram, TypeSlack, TypeTeams, TypeRocketChat, TypeDingTalk, TypeGotify} {
		ch, err := registry.Get(typ)
		require.NoError(t, err)
		assert.True(t, ch.DefaultEvents().Contains(db.AlertEventWaitingConfirmation), "%s must notify on waiting confirmation", typ)
		assert.True(t, ch.DefaultEvents().Contains(db.AlertEventSuccess))
		assert.True(t, ch.DefaultEvents().Contains(db.AlertEventError))
	}
}

func TestRegistry_Register_ReplacesSameType(t *testing.T) {
	registry := NewRegistry()
	count := len(registry.List())

	registry.Register(newFakeChannel(TypeSlack, chatEvents()))
	assert.Len(t, registry.List(), count)

	ch, err := registry.Get(TypeSlack)
	require.NoError(t, err)
	assert.Equal(t, "mdi-test-tube", ch.Icon())
}

func TestRegistry_Infos_ReflectServerConfig(t *testing.T) {
	registry := NewRegistry()
	chat := "project-chat"
	project := db.Project{ID: 1, AlertChat: &chat}

	cfg := &util.ConfigType{
		SlackAlert:    true,
		SlackUrl:      "https://hooks.slack.com/x",
		TelegramAlert: true,
		TelegramToken: "bot-token",
		EmailAlert:    false,
	}

	byType := make(map[db.AlertType]ChannelInfo)
	for _, info := range registry.Infos(cfg, project) {
		byType[info.Type] = info
	}

	assert.True(t, byType[TypeSlack].ServerConfigured)
	assert.True(t, byType[TypeSlack].ServerEnabled)
	assert.True(t, byType[TypeSlack].Ready)
	assert.Equal(t, []Field{{Name: FieldURL, Kind: FieldKindText, Required: true}}, byType[TypeSlack].Fields)

	assert.True(t, byType[TypeTelegram].ServerConfigured, "project chat plus server token is enough")
	assert.True(t, byType[TypeTelegram].ServerEnabled)
	assert.True(t, byType[TypeTelegram].Ready)

	assert.False(t, byType[TypeEmail].ServerConfigured)
	assert.False(t, byType[TypeEmail].ServerEnabled)
	assert.False(t, byType[TypeEmail].Ready)
	assert.Contains(t, byType[TypeEmail].ReadyError, "SMTP")

	assert.False(t, byType[TypeGotify].ServerConfigured)
	assert.False(t, byType[TypeGotify].Ready, "no server-wide gotify pair")
	require.NotNil(t, byType[TypeGotify].Secret)
	assert.Equal(t, db.AccessKeyString, byType[TypeGotify].Secret.KeyType)
	require.NotNil(t, byType[TypeEmail].Secret)
	assert.Equal(t, db.AccessKeyLoginPassword, byType[TypeEmail].Secret.KeyType)
	assert.True(t, byType[TypeEmail].Secret.Optional)
	assert.Nil(t, byType[TypeSlack].Secret)
	assert.Equal(t, BodyFormatHTML, byType[TypeEmail].BodyFormat)
}

func TestTelegram_InstanceDestination(t *testing.T) {
	ch := newTelegramChannel()
	cfg := &util.ConfigType{TelegramAlert: true, TelegramToken: "tok", TelegramChat: "global"}

	dest, ok := ch.InstanceDestination(cfg, db.Project{})
	require.True(t, ok)
	assert.Equal(t, "global", dest.ChatID)
	assert.True(t, dest.Trusted)

	chat := "project"
	dest, ok = ch.InstanceDestination(cfg, db.Project{AlertChat: &chat})
	require.True(t, ok)
	assert.Equal(t, "project", dest.ChatID)

	_, ok = ch.InstanceDestination(&util.ConfigType{TelegramAlert: true, TelegramToken: "tok"}, db.Project{})
	assert.False(t, ok, "no chat anywhere means nothing is sent")

	_, ok = ch.InstanceDestination(&util.ConfigType{TelegramToken: "tok", TelegramChat: "global"}, db.Project{})
	assert.False(t, ok, "telegram_alert=false disables the channel")
}

func TestChannelValidation(t *testing.T) {
	registry := NewRegistry()
	cfg := &util.ConfigType{
		TelegramToken: "server-bot",
		GotifyUrl:     "https://gotify.server",
		GotifyToken:   "server-app",
		EmailHost:     "smtp.server",
	}
	empty := &util.ConfigType{}
	token := &db.AccessKey{Type: db.AccessKeyString, String: "own"}
	creds := &db.AccessKey{Type: db.AccessKeyLoginPassword, LoginPassword: db.LoginPassword{Login: "u", Password: "p"}}

	tests := []struct {
		name    string
		typ     db.AlertType
		cfg     *util.ConfigType
		dest    Destination
		wantErr bool
	}{
		{"telegram needs chat", TypeTelegram, cfg, Destination{}, true},
		{"telegram server token", TypeTelegram, cfg, Destination{ChatID: "1"}, false},
		{"telegram no server token", TypeTelegram, empty, Destination{ChatID: "1"}, true},
		{"telegram own token", TypeTelegram, empty, Destination{ChatID: "1", Secret: token}, false},
		{"telegram empty key", TypeTelegram, empty, Destination{ChatID: "1", Secret: &db.AccessKey{Type: db.AccessKeyString}}, true},
		{"telegram bad thread", TypeTelegram, cfg, Destination{ChatID: "1", ThreadID: "abc"}, true},
		{"telegram thread ok", TypeTelegram, cfg, Destination{ChatID: "1", ThreadID: "7"}, false},
		{"slack needs url", TypeSlack, cfg, Destination{}, true},
		{"slack ok", TypeSlack, cfg, Destination{URL: "https://hooks.slack.com/x"}, false},
		{"slack loopback", TypeSlack, cfg, Destination{URL: "http://127.0.0.1/x"}, true},
		{"gotify server pair", TypeGotify, cfg, Destination{}, false},
		{"gotify server pair missing", TypeGotify, empty, Destination{}, true},
		{"gotify url without key", TypeGotify, cfg, Destination{URL: "https://gotify.example"}, true},
		{"gotify key without url", TypeGotify, cfg, Destination{Secret: token}, true},
		{"gotify own", TypeGotify, empty, Destination{URL: "https://gotify.example", Secret: token}, false},
		{"gotify own loopback", TypeGotify, empty, Destination{URL: "http://127.0.0.1", Secret: token}, true},
		{"email server", TypeEmail, cfg, Destination{Recipients: []string{"a@b.c"}}, false},
		{"email no server", TypeEmail, empty, Destination{Recipients: []string{"a@b.c"}}, true},
		{"email own server", TypeEmail, empty, Destination{Params: map[string]any{"smtp_host": "mail.example", "smtp_sender": "s@e.x"}}, false},
		{"email own server without sender", TypeEmail, empty, Destination{Params: map[string]any{"smtp_host": "mail.example"}}, true},
		{"email port without host", TypeEmail, cfg, Destination{Params: map[string]any{"smtp_port": "25"}}, true},
		{"email own creds on server host", TypeEmail, cfg, Destination{Secret: creds}, false},
		{"email bad address", TypeEmail, cfg, Destination{Recipients: []string{"nope"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch, err := registry.Get(tt.typ)
			require.NoError(t, err)
			err = ch.Validate(tt.cfg, tt.dest)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDestinationFromAlert_SplitsRecipients(t *testing.T) {
	recipients := "a@b.c, d@e.f;g@h.i\nj@k.l"
	dest := DestinationFromAlert(db.Alert{Recipients: &recipients})
	assert.Equal(t, []string{"a@b.c", "d@e.f", "g@h.i", "j@k.l"}, dest.Recipients)
	assert.False(t, dest.Trusted)
}
