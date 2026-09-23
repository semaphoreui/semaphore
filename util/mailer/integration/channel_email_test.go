//go:build integration

package integration

import (
	"context"
	"runtime"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/alerting"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func emailChannel(t *testing.T) alerting.Channel {
	t.Helper()
	ch, err := alerting.NewRegistry().Get(alerting.TypeEmail)
	require.NoError(t, err)
	return ch
}

// serverConfig points the server-wide SMTP settings at the container.
func serverConfig(m *mailpit) *util.ConfigType {
	return &util.ConfigType{
		EmailHost:   m.host,
		EmailPort:   m.smtpPort,
		EmailSender: "semaphore@example.com",
	}
}

func message() alerting.Message {
	return alerting.Message{Subject: "Task #1 failed", Body: "<p>failed</p>"}
}

// requireContainerIP skips scenarios that connect to the container address
// directly: the bridge network is routable from the test process on Linux
// only.
func requireContainerIP(t *testing.T, m *mailpit) {
	t.Helper()
	if runtime.GOOS != "linux" || m.ip == "" {
		t.Skipf("container IP %q is not routable on %s", m.ip, runtime.GOOS)
	}
}

// TestEmailChannel_ServerSMTP: the destination comes from config.json.
func TestEmailChannel_ServerSMTP(t *testing.T) {
	m := startMailpit(t, mailpitConfig{})
	ch := emailChannel(t)

	t.Run("E1 one message per recipient", func(t *testing.T) {
		m.clear()
		dest := alerting.Destination{
			Trusted:    true,
			Recipients: []string{"a@example.com", "b@example.com", "c@example.com"},
		}
		require.NoError(t, ch.Send(context.Background(), serverConfig(m), dest, message()))

		msgs := m.waitForMessages(3)
		var to []string
		for _, msg := range msgs {
			assert.Equal(t, "Task #1 failed", msg.Subject)
			assert.Equal(t, "semaphore@example.com", msg.From.Address)
			require.Len(t, msg.To, 1)
			to = append(to, msg.To[0].Address)
		}
		assert.ElementsMatch(t, []string{"a@example.com", "b@example.com", "c@example.com"}, to)
	})

	t.Run("E3 project alert may not point its own smtp host at localhost", func(t *testing.T) {
		m.clear()
		dest := alerting.Destination{
			Recipients: []string{"a@example.com"},
			Params: map[string]any{
				"smtp_host":   m.host,
				"smtp_port":   m.smtpPort,
				"smtp_sender": "alert@example.com",
			},
		}
		err := ch.Send(context.Background(), serverConfig(m), dest, message())
		assert.ErrorContains(t, err, "smtp_host is not allowed")
		m.assertNoMessages()
	})

	t.Run("E2 project alert with its own smtp host on the private network", func(t *testing.T) {
		requireContainerIP(t, m)
		m.clear()
		dest := alerting.Destination{
			Recipients: []string{"a@example.com"},
			Params: map[string]any{
				"smtp_host":   m.ip,
				"smtp_port":   "1025",
				"smtp_sender": "alert@example.com",
			},
		}
		require.NoError(t, ch.Send(context.Background(), &util.ConfigType{}, dest, message()))
		msg := m.waitForOne()
		assert.Equal(t, "alert@example.com", msg.From.Address)
	})
}

// TestEmailChannel_OwnCredentials (E4): the alert brings a login_password
// key, the host stays the server one.
func TestEmailChannel_OwnCredentials(t *testing.T) {
	m := startMailpit(t, mailpitConfig{auth: "alertuser:alertpass", allowInsecureAuth: true})
	ch := emailChannel(t)

	cfg := serverConfig(m)
	cfg.EmailSecure = true
	cfg.EmailUsername, cfg.EmailPassword = "serveruser", "serverpass"

	t.Run("key credentials win over the server ones", func(t *testing.T) {
		m.clear()
		dest := alerting.Destination{
			Recipients: []string{"a@example.com"},
			Secret: &db.AccessKey{
				Type:          db.AccessKeyLoginPassword,
				LoginPassword: db.LoginPassword{Login: "alertuser", Password: "alertpass"},
			},
		}
		require.NoError(t, ch.Send(context.Background(), cfg, dest, message()))
		assert.Equal(t, "alertuser", m.waitForOne().Username)
	})

	t.Run("server credentials are rejected by this server", func(t *testing.T) {
		m.clear()
		dest := alerting.Destination{Trusted: true, Recipients: []string{"a@example.com"}}
		err := ch.Send(context.Background(), cfg, dest, message())
		assert.ErrorContains(t, err, "535")
		m.assertNoMessages()
	})
}

// TestEmailChannel_RejectedRecipient (E5): one address is refused by the
// server; the others are still delivered and the error names the failure.
func TestEmailChannel_RejectedRecipient(t *testing.T) {
	m := startMailpit(t, mailpitConfig{allowedRecipients: `@example\.com$`})
	ch := emailChannel(t)

	m.clear()
	dest := alerting.Destination{
		Trusted:    true,
		Recipients: []string{"a@example.com", "bad@other.org", "b@example.com"},
	}
	err := ch.Send(context.Background(), serverConfig(m), dest, message())
	require.Error(t, err)

	msgs := m.waitForMessages(2)
	var to []string
	for _, msg := range msgs {
		to = append(to, msg.To[0].Address)
	}
	assert.ElementsMatch(t, []string{"a@example.com", "b@example.com"}, to)
}
