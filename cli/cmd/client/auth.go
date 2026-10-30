package client

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/semaphoreui/semaphore/cli/cmd/client/config"
	"github.com/semaphoreui/semaphore/cli/cmd/client/output"
	"github.com/semaphoreui/semaphore/pkg/apiclient"
	"github.com/spf13/cobra"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Log in to a server and manage the saved token",
		Long: `Log in to a Semaphore server and save the server URL and an API token as a
named context in ~/.semaphore/client.yml (mode 0600). Later commands use the
current context unless --url, --context or SEMAPHORE_CLIENT_* say otherwise.`,
		Example: `  semaphore client auth login --url https://semaphore.example.com
  echo "$TOKEN" | semaphore client auth login --url https://semaphore.example.com --with-token
  semaphore client auth status`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	cmd.AddCommand(newAuthLoginCmd(), newAuthStatusCmd(), newAuthTokenCmd(), newAuthLogoutCmd())
	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	var (
		withToken bool
		tokenName string
		ttl       time.Duration
		login     string
	)
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Save a server and an API token as a context",
		Long: `Save a server and an API token as a context and make it the current one.

Two ways to authenticate:

  * Paste a token: with --with-token, or whenever stdin is not a terminal, the
    first line of stdin is read as an existing API token and checked against
    the server. Create one in the web UI under Settings > API tokens.

  * Log in with a password: in an interactive terminal the command asks for the
    login and the password (and a one-time code when two-factor authentication
    is enabled), creates an API token named --name with lifetime --ttl, and
    closes the web session. The password is never stored.

The context is named by --context, or after the server host. The token is
written to the configuration file with mode 0600; use --ca-cert or --insecure
to describe a server with a private certificate and they are saved as well.`,
		Example: `  # Paste an existing token (CI, scripts)
  echo "$SEMAPHORE_TOKEN" | semaphore client auth login --url https://semaphore.example.com --with-token

  # Interactive login with a password, token valid for 30 days
  semaphore client auth login --url https://semaphore.example.com --ttl 720h

  # Save a second server under its own name
  semaphore client auth login --url https://staging.example.com --context staging`,
		Args:        cobra.NoArgs,
		Annotations: map[string]string{annotationNewContext: "true"},
		RunE: func(cmd *cobra.Command, _ []string) error {
			r := rt(cmd)
			api, err := r.client(cmd)
			if err != nil {
				return err
			}
			return runAuthLogin(cmd, r, api, withToken, tokenName, ttl, login)
		},
	}
	cmd.Flags().BoolVar(&withToken, "with-token", false, "Read an existing API token from the first line of stdin")
	cmd.Flags().StringVar(&tokenName, "name", "", "Name of the API token to create (default semaphore-cli@<hostname>)")
	cmd.Flags().DurationVar(&ttl, "ttl", 0, "Lifetime of the created token, e.g. 720h (default: never expires)")
	cmd.Flags().StringVar(&login, "login", "", "Login or e-mail for the password flow (asked when omitted)")
	return cmd
}

func runAuthLogin(cmd *cobra.Command, r *runtime, api *apiclient.Client, withToken bool, tokenName string, ttl time.Duration, login string) error {
	ctx, cancel := withTimeout(cmd, 0)
	defer cancel()

	// The token being replaced must not be used for the login itself.
	api.SetToken("")

	if err := api.Ping(ctx); err != nil {
		return err
	}

	// Flag first, then whatever Resolve settled on (the environment, or the
	// context named "default"). Logging in never renames or re-points a
	// context the user did not ask for.
	contextName := r.flags.Context
	if contextName == "" {
		contextName = r.cfg.ContextName
	}
	if contextName == "" {
		contextName = config.DefaultContextName
	}

	var token string
	switch {
	case withToken || !r.interactive || r.cfg.NoInput:
		line, err := r.readLine()
		if err != nil {
			return err
		}
		token = strings.TrimSpace(line)
		if token == "" {
			return usage(cmd.CommandPath(), "no token on stdin; pipe the API token or run in a terminal for the password flow")
		}
	default:
		created, err := loginWithPassword(cmd, r, api, tokenName, ttl, login)
		if err != nil {
			return err
		}
		token = created
	}

	api.SetToken(token)
	me, err := api.Me(ctx)
	if err != nil {
		if errors.Is(err, apiclient.ErrUnauthorized) {
			return fmt.Errorf("the server rejected the token: %w", err)
		}
		return err
	}

	saved, existed := r.cfgFile.Contexts[contextName]
	movedTo := existed && saved.URL != "" && saved.URL != api.BaseURL()
	if movedTo {
		// The context now points at another server: a default project saved
		// for the old one names an id that means something else here.
		r.warnf("Context %q pointed at %s; replacing it with %s.", contextName, saved.URL, api.BaseURL())
		if saved.Project != "" {
			r.warnf("Default project %q dropped: it belonged to the previous server.", saved.Project)
			saved.Project = ""
		}
	}
	saved.URL = api.BaseURL()
	saved.Token = token
	saved.TokenFile = ""
	saved.CACert = r.cfg.CACert
	saved.Insecure = r.cfg.Insecure
	r.cfgFile.SetContext(contextName, saved)
	if err := r.saveConfig(); err != nil {
		return err
	}

	r.warnf("Logged in to %s as %s.", api.BaseURL(), me.Username)
	r.warnf("Context %q saved to %s.", contextName, r.cfgPath)
	if contextName != config.DefaultContextName {
		r.warnf("Use it with --context %s, or set %s=%s.", contextName, config.EnvContext, contextName)
	}
	return nil
}

// loginWithPassword runs the interactive flow and returns a fresh API token.
func loginWithPassword(cmd *cobra.Command, r *runtime, api *apiclient.Client, tokenName string, ttl time.Duration, login string) (string, error) {
	ctx, cancel := withTimeout(cmd, 0)
	defer cancel()

	methods, err := api.LoginMethods(ctx)
	if err != nil {
		return "", err
	}
	if !methods.LoginWithPassword && !methods.LoginWithLdap {
		return "", fmt.Errorf("%s accepts no password login (single sign-on only); create an API token in the web UI at %s/tokens and run `auth login --with-token`", api.BaseURL(), api.BaseURL())
	}

	if login == "" {
		if login, err = r.prompt(cmd, "Login: "); err != nil {
			return "", err
		}
	}
	password, err := r.promptSecret(cmd, "Password: ")
	if err != nil {
		return "", err
	}
	if login == "" || password == "" {
		return "", usage(cmd.CommandPath(), "login and password are required")
	}

	if err := api.Login(ctx, login, password); err != nil {
		if errors.Is(err, apiclient.ErrUnauthorized) {
			return "", fmt.Errorf("login failed: wrong login or password: %w", err)
		}
		return "", err
	}

	// A session with a pending second factor answers 401 to everything.
	if _, err := api.Me(ctx); err != nil {
		if !errors.Is(err, apiclient.ErrUnauthorized) || methods.AuthMethods.Totp == nil {
			return "", err
		}
		code, err := r.promptSecret(cmd, "One-time code: ")
		if err != nil {
			return "", err
		}
		if err := api.Verify(ctx, code); err != nil {
			if errors.Is(err, apiclient.ErrUnauthorized) {
				return "", fmt.Errorf("two-factor verification failed: %w", err)
			}
			return "", err
		}
	}

	if tokenName == "" {
		host, _ := os.Hostname()
		if host == "" {
			host = "unknown-host"
		}
		tokenName = "semaphore-cli@" + host
	}
	var expiresAt *time.Time
	if ttl > 0 {
		t := time.Now().Add(ttl)
		expiresAt = &t
	}
	created, err := api.CreateToken(ctx, tokenName, expiresAt)
	if err != nil {
		return "", fmt.Errorf("create API token: %w", err)
	}

	if err := api.Logout(ctx); err != nil {
		r.warnf("warning: could not close the web session: %v", err)
	}
	return created.ID, nil
}

// statusResult is what `auth status` prints.
type statusResult struct {
	Context       string `json:"context,omitempty"`
	URL           string `json:"url"`
	User          string `json:"user,omitempty"`
	Name          string `json:"name,omitempty"`
	Admin         bool   `json:"admin"`
	TokenSource   string `json:"token_source"`
	TokenFile     string `json:"token_file,omitempty"`
	ConfigFile    string `json:"config_file"`
	ServerVersion string `json:"server_version,omitempty"`
}

func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show which server and token are in use and whether they work",
		Long: `Show the current context, the server URL, where the token comes from (flag,
environment, context file) and the user the token belongs to. Exit code 4 means
the server rejected the token.`,
		Example: `  semaphore client auth status
  semaphore client auth status --context staging -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			r := rt(cmd)
			api, err := r.client(cmd)
			if err != nil {
				return err
			}
			res := statusResult{
				Context:     r.cfg.ContextName,
				URL:         api.BaseURL(),
				TokenSource: tokenSourceLabel(r.cfg),
				TokenFile:   r.cfg.TokenFile,
				ConfigFile:  r.cfgPath,
			}
			if err := r.requireToken(cmd); err != nil {
				return err
			}

			ctx, cancel := withTimeout(cmd, r.cfg.Timeout)
			defer cancel()

			me, err := api.Me(ctx)
			if err != nil {
				if errors.Is(err, apiclient.ErrUnauthorized) {
					r.warnf("Server %s rejected the token (source: %s).", res.URL, res.TokenSource)
				}
				return err
			}
			res.User, res.Name, res.Admin = me.Username, me.Name, me.Admin
			if info, err := api.Info(ctx); err == nil {
				res.ServerVersion = info.Version
			}
			return r.printer.Print(res, output.Columns("context", "url", "user", "token_source"))
		},
	}
}

func tokenSourceLabel(cfg config.Effective) string {
	switch cfg.Sources.Token {
	case config.SourceEnv:
		if cfg.TokenFile != "" {
			return "env:" + config.EnvTokenFile
		}
		return "env:" + config.EnvToken
	case config.SourceContext:
		if cfg.TokenFile != "" {
			return "context:token_file"
		}
		return "context:token"
	}
	return "none"
}

func newAuthTokenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "token",
		Short: "Print the API token in use",
		Long: `Print the API token the client would send, for handing it to another tool:

  curl -H "Authorization: Bearer $(semaphore client auth token)" ...

The token is written to stdout only; a warning goes to stderr when stdout is a
terminal, because a token on screen tends to end up in screenshots and logs.`,
		Example: `  curl -H "Authorization: Bearer $(semaphore client auth token)" https://semaphore.example.com/api/user`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			r := rt(cmd)
			if err := r.requireToken(cmd); err != nil {
				return err
			}
			if output.IsTerminal(r.io.Out) {
				r.warnf("warning: printing the API token to the terminal")
			}
			fmt.Fprintln(r.io.Out, r.cfg.Token)
			return nil
		},
	}
}

func newAuthLogoutCmd() *cobra.Command {
	var revoke bool
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Forget the token of the current context",
		Long: `Remove the token from the current context in the configuration file. With
--revoke the token is also deleted on the server, so it stops working everywhere.
The context itself (URL, project, TLS settings) stays; delete it with
'semaphore client context delete'.`,
		Example: `  semaphore client auth logout
  semaphore client auth logout --context staging --revoke`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			r := rt(cmd)
			saved, ok := r.cfgFile.Contexts[r.cfg.ContextName]
			if !ok {
				// The implicit "default" may simply not be there yet; an
				// explicitly named context that is missing is an error.
				if r.cfg.Sources.Context == config.SourceNone {
					return usage(cmd.CommandPath(), "no context %q in %s; logout works on a context saved by 'auth login'", r.cfg.ContextName, r.cfgPath)
				}
				return fmt.Errorf("%w %q", config.ErrUnknownContext, r.cfg.ContextName)
			}
			if saved.Token == "" && saved.TokenFile == "" {
				r.warnf("Context %q holds no token; nothing to do.", r.cfg.ContextName)
				return nil
			}

			if revoke {
				if r.cfg.Sources.Token != config.SourceContext {
					return usage(cmd.CommandPath(), "--revoke uses the token of the context, but the token in use comes from %s; unset it first", tokenSourceLabel(r.cfg))
				}
				api, err := r.client(cmd)
				if err != nil {
					return err
				}
				ctx, cancel := withTimeout(cmd, r.cfg.Timeout)
				defer cancel()
				if err := api.DeleteToken(ctx, tokenPrefix(r.cfg.Token)); err != nil {
					return fmt.Errorf("revoke token on the server: %w", err)
				}
				r.warnf("Token revoked on %s.", api.BaseURL())
			}

			saved.Token = ""
			if saved.TokenFile != "" {
				r.warnf("Note: the token file %s is left in place; delete it yourself if the token is no longer needed.", saved.TokenFile)
				saved.TokenFile = ""
			}
			r.cfgFile.SetContext(r.cfg.ContextName, saved)
			if err := r.saveConfig(); err != nil {
				return err
			}
			r.warnf("Token removed from context %q in %s.", r.cfg.ContextName, r.cfgPath)
			return nil
		},
	}
	cmd.Flags().BoolVar(&revoke, "revoke", false, "Also delete the token on the server")
	return cmd
}

// tokenPrefix is the 8-character id the server lists and accepts for deletion,
// so that the full token never appears in a URL or an access log.
func tokenPrefix(token string) string {
	if len(token) > 8 {
		return token[:8]
	}
	return token
}

func hostOf(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Hostname() == "" {
		return "default"
	}
	return u.Hostname()
}
