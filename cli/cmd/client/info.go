package client

import (
	"strings"

	"github.com/semaphoreui/semaphore/cli/cmd/client/output"
	"github.com/spf13/cobra"
)

// infoResult is what `info` prints: the server as it describes itself plus the
// client version, so one record answers "what am I talking to".
type infoResult struct {
	URL               string `json:"url"`
	Context           string `json:"context,omitempty"`
	ServerVersion     string `json:"server_version"`
	ClientVersion     string `json:"client_version"`
	WebHost           string `json:"web_host,omitempty"`
	UseRemoteRunner   bool   `json:"use_remote_runner"`
	LoginWithPassword bool   `json:"login_with_password"`
	TotpEnabled       bool   `json:"totp_enabled"`
	SubscriptionState string `json:"subscription_state,omitempty"`
	GitClient         string `json:"git_client,omitempty"`
	ScheduleTimezone  string `json:"schedule_timezone,omitempty"`
	BoltdbUsed        bool   `json:"boltdb_used"`
	ServerAndClientOK bool   `json:"versions_compatible"`
}

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show the server version and settings",
		Long: `Show the version and public settings of the server the client is pointed at,
next to the client version. Use it to check that the URL and the token work.`,
		Example: `  semaphore client info
  semaphore client info --url https://semaphore.example.com -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			r := rt(cmd)
			api, err := r.client(cmd)
			if err != nil {
				return err
			}
			if err := r.requireToken(cmd); err != nil {
				return err
			}

			ctx, cancel := withTimeout(cmd, r.cfg.Timeout)
			defer cancel()

			info, err := api.Info(ctx)
			if err != nil {
				return err
			}

			res := infoResult{
				URL:               api.BaseURL(),
				Context:           r.cfg.ContextName,
				ServerVersion:     info.Version,
				ClientVersion:     r.io.Version,
				WebHost:           info.WebHost,
				UseRemoteRunner:   info.UseRemoteRunner,
				LoginWithPassword: info.LoginWithPassword,
				TotpEnabled:       info.AuthMethods.Totp != nil,
				SubscriptionState: info.SubscriptionState,
				GitClient:         info.GitClient,
				ScheduleTimezone:  info.ScheduleTimezone,
				BoltdbUsed:        info.BoltdbUsed,
				ServerAndClientOK: sameMinor(info.Version, r.io.Version),
			}
			if !res.ServerAndClientOK {
				r.warnf("warning: client %s and server %s differ in major or minor version; some commands may not match the server", r.io.Version, info.Version)
			}
			return r.printer.Print(res, output.Columns("url", "server_version", "client_version"))
		},
	}
}

// sameMinor compares "major.minor" of two versions such as "v2.20.3" and
// "2.20.1-abc". Unknown or development versions ("development", "") compare
// as compatible, since there is nothing to warn about.
func sameMinor(a, b string) bool {
	ma, oka := majorMinor(a)
	mb, okb := majorMinor(b)
	if !oka || !okb {
		return true
	}
	return ma == mb
}

func majorMinor(v string) (string, bool) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", false
	}
	for _, p := range parts[:2] {
		for _, ch := range p {
			if ch < '0' || ch > '9' {
				return "", false
			}
		}
	}
	return parts[0] + "." + parts[1], true
}
