package client

import (
	"fmt"
	"sort"

	"github.com/semaphoreui/semaphore/cli/cmd/client/config"
	"github.com/semaphoreui/semaphore/cli/cmd/client/output"
	"github.com/semaphoreui/semaphore/pkg/apiclient"
	"github.com/spf13/cobra"
)

// contextRow is a context as shown to the user: never the token itself, only
// where it is kept.
type contextRow struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	Project   string `json:"project,omitempty"`
	Token     string `json:"token"`
	TokenFile string `json:"token_file,omitempty"`
	CACert    string `json:"ca_cert,omitempty"`
	Insecure  bool   `json:"insecure_skip_tls_verify"`
}

func toContextRow(name string, c config.Context) contextRow {
	tok := "none"
	switch {
	case c.TokenFile != "":
		tok = "file"
	case c.Token != "":
		tok = "inline"
	}
	return contextRow{Name: name, URL: c.URL, Project: c.Project, Token: tok, TokenFile: c.TokenFile, CACert: c.CACert, Insecure: c.Insecure}
}

var contextColumns = output.Columns("name", "url", "project", "token")

func newContextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "context",
		Aliases: []string{"contexts", "ctx"},
		Short:   "Manage saved servers (contexts)",
		Long: `Manage the contexts saved in ~/.semaphore/client.yml. A context is a server
URL with its token, an optional default project and TLS settings.

The context named "default" is the one commands use when nothing says
otherwise; --context or SEMAPHORE_CLIENT_CONTEXT picks another one per call.
There is no "current context" to set: a file that remembered a choice made
'auth login' against another server overwrite whichever context was selected.`,
		Example: `  semaphore client context list
  semaphore client context set prod --url https://semaphore.example.com --project 3
  semaphore client project list --context prod`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	cmd.AddCommand(newContextListCmd(), newContextGetCmd(), newContextSetCmd(), newContextDeleteCmd())
	return cmd
}

func newContextListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List saved contexts",
		Long: `List the contexts of the configuration file: name, server URL, default
project and where its token comes from (inline, file or none). Token values are
never shown. The context named "default" is the one used when --context and
SEMAPHORE_CLIENT_CONTEXT are absent. Works without a server connection.`,
		Example: `  semaphore client context list`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			r := rt(cmd)
			rows := make([]contextRow, 0, len(r.cfgFile.Contexts))
			for _, name := range r.cfgFile.ContextNames() {
				rows = append(rows, toContextRow(name, r.cfgFile.Contexts[name]))
			}
			return r.printer.Print(rows, contextColumns)
		},
	}
}

func newContextGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "get [NAME]",
		Aliases: []string{"show"},
		Short:   "Show one context (\"default\" when no name is given)",
		Long:    `Show one saved context without its token. Without a name the context "default" is shown.`,
		Example: `  semaphore client context get
  semaphore client context get staging -o yaml`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r := rt(cmd)
			name := config.DefaultContextName
			if len(args) == 1 {
				name = args[0]
			}
			c, ok := r.cfgFile.Contexts[name]
			if !ok {
				return fmt.Errorf("%w %q; known contexts: %s", config.ErrUnknownContext, name, joinNames(r.cfgFile))
			}
			return r.printer.Print(toContextRow(name, c), contextColumns)
		},
	}
}

func newContextSetCmd() *cobra.Command {
	var (
		tokenFile  string
		noInsecure bool
		clearProj  bool
	)
	cmd := &cobra.Command{
		Use:   "set NAME",
		Short: "Create or change a context without logging in",
		Long: `Create or change a context field by field. The global flags --url, --project,
--ca-cert and --insecure give the new values; only the flags you pass change.
The token comes from --token-file (a file holding the token) or from
'semaphore client auth login'; there is no flag that takes the token itself.

Aiming an existing context at another server drops what belonged to the old
one — its token, its default project and its TLS settings — so that a
production credential is never handed to a stranger. Values passed in the same
command are applied afterwards and survive.`,
		Example: `  semaphore client context set staging --url https://staging.example.com --token-file ~/.semaphore/staging.token
  semaphore client context set prod -p 3
  semaphore client context set lab --insecure`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r := rt(cmd)
			name := args[0]
			c, existed := r.cfgFile.Contexts[name]
			changed := false

			if cmd.Flags().Changed("url") {
				// Stored in the canonical form the client talks to, like the
				// one `auth login` saves, so the file says what is really used.
				url := apiclient.NormalizeURL(r.flags.URL)
				r.repointContext(name, &c, url)
				c.URL, changed = url, true
			}
			if cmd.Flags().Changed("project") {
				c.Project, changed = r.flags.Project, true
			}
			if clearProj {
				c.Project, changed = "", true
			}
			if cmd.Flags().Changed("ca-cert") {
				c.CACert, changed = r.flags.CACert, true
			}
			if cmd.Flags().Changed("insecure") {
				c.Insecure, changed = true, true
			}
			if noInsecure {
				c.Insecure, changed = false, true
			}
			if cmd.Flags().Changed("token-file") {
				c.TokenFile, c.Token, changed = tokenFile, "", true
			}

			if !existed && c.URL == "" {
				return usage(cmd.CommandPath(), "a new context needs --url")
			}
			if !changed && existed {
				return usage(cmd.CommandPath(), "nothing to change; pass at least one of --url, --project, --ca-cert, --insecure, --no-insecure, --token-file")
			}

			r.cfgFile.SetContext(name, c)
			if err := r.saveConfig(); err != nil {
				return err
			}
			if existed {
				r.warnf("Context %q updated.", name)
			} else {
				r.warnf("Context %q created.", name)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&tokenFile, "token-file", "", "File holding the API token (replaces an inline token)")
	cmd.Flags().BoolVar(&noInsecure, "no-insecure", false, "Turn TLS verification back on")
	cmd.Flags().BoolVar(&clearProj, "no-project", false, "Remove the default project")
	return cmd
}

func newContextDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete NAME",
		Aliases: []string{"rm"},
		Short:   "Delete a context",
		Long: `Remove a context from the configuration file. The token stored in it is
forgotten but not revoked on the server; use 'auth logout --revoke' for that.
Asks for confirmation in a terminal; pass --yes to skip it.`,
		Example: `  semaphore client context delete staging -y`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r := rt(cmd)
			name := args[0]
			if _, ok := r.cfgFile.Contexts[name]; !ok {
				return fmt.Errorf("%w %q; known contexts: %s", config.ErrUnknownContext, name, joinNames(r.cfgFile))
			}
			if err := r.prompter.Confirm(cmd.Context(), fmt.Sprintf("Delete context %q?", name)); err != nil {
				return err
			}
			delete(r.cfgFile.Contexts, name)
			if err := r.saveConfig(); err != nil {
				return err
			}
			r.warnf("Context %q deleted.", name)
			return nil
		},
	}
}

// sameServer says whether two spellings address the same Semaphore, so that a
// cosmetic change (a trailing slash, an "/api" suffix) is not mistaken for a
// move to another server.
func sameServer(a, b string) bool {
	return apiclient.NormalizeURL(a) == apiclient.NormalizeURL(b)
}

// repointContext drops what belonged to the server a context used to address.
// A token, a default project and the TLS settings are all per-server: carried
// over, the token is handed to a stranger, the project id names something else
// there, and "insecure" silently stops verifying a server nobody said that
// about. Anything the same command passes explicitly is applied afterwards.
func (r *runtime) repointContext(name string, c *config.Context, newURL string) bool {
	if c.URL == "" || sameServer(c.URL, newURL) {
		return false
	}
	r.warnf("Context %q pointed at %s; replacing it with %s.", name, c.URL, newURL)
	if c.Token != "" || c.TokenFile != "" {
		r.warnf("Its token is dropped: it belonged to the previous server.")
		c.Token, c.TokenFile = "", ""
	}
	if c.Project != "" {
		r.warnf("Default project %q dropped: it belonged to the previous server.", c.Project)
		c.Project = ""
	}
	if c.CACert != "" || c.Insecure {
		r.warnf("TLS settings dropped: they belonged to the previous server.")
		c.CACert, c.Insecure = "", false
	}
	return true
}

func joinNames(f *config.File) string {
	names := f.ContextNames()
	sort.Strings(names)
	if len(names) == 0 {
		return "(none)"
	}
	out := ""
	for i, n := range names {
		if i > 0 {
			out += ", "
		}
		out += n
	}
	return out
}
