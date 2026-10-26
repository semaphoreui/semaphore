// Package client implements `semaphore client`, the command-line client of the
// Semaphore UI REST API. It talks to a server over HTTP with an API token and
// never touches the database or the server configuration file.
package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/semaphoreui/semaphore/cli/cmd/client/config"
	"github.com/semaphoreui/semaphore/cli/cmd/client/output"
	"github.com/semaphoreui/semaphore/pkg/apiclient"
	"github.com/semaphoreui/semaphore/util"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// IO is where a command reads from and writes to. Tests pass buffers.
type IO struct {
	In     io.Reader
	Out    io.Writer
	Err    io.Writer
	Getenv func(string) string
	// Version of the client, sent as User-Agent and shown by `info`.
	Version string

	// forceInteractive makes the runtime behave as if stdin and stdout were
	// terminals. Tests use it to drive prompts through buffers.
	forceInteractive bool
}

// StdIO is the IO of a real process.
func StdIO() IO {
	return IO{In: os.Stdin, Out: os.Stdout, Err: os.Stderr, Getenv: os.Getenv, Version: util.Ver}
}

// flags are the persistent flags of `client`, bound once per command tree.
type flags struct {
	config.Flags
	fields    []string
	quiet     bool
	noHeaders bool
	noColor   bool
	yes       bool
}

// runtime is everything a command needs, resolved once in the persistent
// pre-run and carried in the command context. There is no package-level state.
type runtime struct {
	io       IO
	flags    *flags
	cfg      config.Effective
	cfgFile  *config.File
	cfgPath  string
	api      *apiclient.Client
	printer  *output.Printer
	prompter *Prompter
	// interactive is true when both stdin and stdout are terminals.
	interactive bool
}

type ctxKey struct{}

// annotationNewContext marks a command that may name a context which does not
// exist yet (auth login); for every other command an unknown --context is an
// error.
const annotationNewContext = "client.newContext"

// rt returns the runtime stored by the pre-run hook.
func rt(cmd *cobra.Command) *runtime {
	r, _ := cmd.Context().Value(ctxKey{}).(*runtime)
	if r == nil {
		panic("client runtime missing: command executed outside of NewCommand")
	}
	return r
}

// NewCommand builds the `client` command tree.
func NewCommand(stdio IO) *cobra.Command {
	f := &flags{}

	cmd := &cobra.Command{
		Use:   "client",
		Short: "Manage a Semaphore server over its REST API",
		Long: `Manage a Semaphore UI server from the terminal or from scripts over its REST API.

The client needs a server URL and an API token. Give them with flags, with
SEMAPHORE_CLIENT_* environment variables, or save them once as a named context
with 'semaphore client auth login' (stored in ~/.semaphore/client.yml with
mode 0600). Precedence: flag > environment > context > default.

Commands follow the form 'semaphore client <resource> <action> [id|name]'.
Results go to stdout as a table, JSON (-o json) or YAML (-o yaml); messages
and errors go to stderr.

Exit codes: 0 success, 1 client or API error, 2 usage error, 3 the awaited
task failed, 4 authentication failed, 5 not found, 6 the awaited task was
stopped, 124 wait timeout, 130 interrupted.`,
		Example: `  # Save a server and its token as the current context
  echo "$TOKEN" | semaphore client auth login --url https://semaphore.example.com --with-token

  # List templates of project 1 as JSON
  semaphore client template list -p 1 -o json

  # Run a template and follow its log; exit code tells how the task ended
  semaphore client template run deploy -p 1 --follow`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	pf := cmd.PersistentFlags()
	pf.StringVar(&f.URL, "url", "", "Server URL, e.g. https://semaphore.example.com ($"+config.EnvURL+")")
	pf.StringVar(&f.Context, "context", "", "Named context from the configuration file ($"+config.EnvContext+")")
	pf.StringVarP(&f.Project, "project", "p", "", "Project id or name for project-scoped commands ($"+config.EnvProject+")")
	pf.StringVarP(&f.Output, "output", "o", "", "Output format: table, json or yaml ($"+config.EnvOutput+", default table)")
	pf.StringSliceVar(&f.fields, "fields", nil, "Fields to show, comma-separated JSON field names (e.g. id,name,status)")
	pf.BoolVarP(&f.quiet, "quiet", "q", false, "Print only ids, one per line")
	pf.BoolVar(&f.noHeaders, "no-headers", false, "Omit the table header line")
	pf.BoolVar(&f.noColor, "no-color", false, "Disable colors (also $NO_COLOR)")
	pf.BoolVarP(&f.yes, "yes", "y", false, "Answer yes to confirmations")
	pf.BoolVar(&f.NoInput, "no-input", false, "Never prompt; fail instead ($"+config.EnvNoInput+", $CI)")
	pf.BoolVar(&f.Insecure, "insecure", false, "Skip TLS certificate verification ($"+config.EnvInsecure+")")
	pf.StringVar(&f.CACert, "ca-cert", "", "PEM file with the CA that signed the server certificate ($"+config.EnvCACert+")")
	pf.DurationVar(&f.Timeout, "timeout", 0, "HTTP request timeout, e.g. 30s ($"+config.EnvTimeout+", default 30s)")
	pf.BoolVar(&f.Debug, "debug", false, "Trace HTTP requests to stderr with secrets masked ($"+config.EnvDebug+")")

	// A wrong flag is a usage error (exit 2), not a generic failure.
	cmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		return usage(c.CommandPath(), "%s", err.Error())
	})

	cmd.PersistentPreRunE = func(c *cobra.Command, _ []string) error {
		r, err := newRuntime(c, stdio, f)
		if err != nil {
			return err
		}
		ctx, cancel := signal.NotifyContext(c.Context(), os.Interrupt, syscall.SIGTERM)
		c.SetContext(context.WithValue(ctx, ctxKey{}, r))
		// The cancel func is released when the process exits; cobra has no
		// post-run hook that runs on error, so we rely on that.
		_ = cancel
		return nil
	}

	cmd.AddCommand(newInfoCmd(), newAuthCmd(), newContextCmd())
	silence(cmd)
	return cmd
}

// silence stops cobra from printing usage and "Error:" lines on every command
// of the tree: cobra consults the executed command, so setting the flags on
// `client` alone is not enough. The root Execute prints the error once, and a
// UsageError carries its own help hint.
func silence(cmd *cobra.Command) {
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	for _, sub := range cmd.Commands() {
		silence(sub)
	}
}

// newRuntime resolves the configuration and wires the helpers. It does not
// talk to the server.
func newRuntime(cmd *cobra.Command, stdio IO, f *flags) (*runtime, error) {
	if stdio.Getenv == nil {
		stdio.Getenv = func(string) string { return "" }
	}

	cfgPath, err := config.Path(stdio.Getenv)
	if err != nil {
		return nil, err
	}
	cfgFile, warnings, err := config.Load(cfgPath)
	if err != nil {
		return nil, err
	}
	for _, w := range warnings {
		fmt.Fprintln(stdio.Err, "warning:", w)
	}

	cfg, err := config.Resolve(f.Flags, stdio.Getenv, cfgFile)
	if errors.Is(err, config.ErrUnknownContext) && f.Context != "" && cmd.Annotations[annotationNewContext] == "true" {
		// `auth login --context NAME` creates NAME: resolve without any
		// context so that the URL must come from --url or the environment.
		noCtx := f.Flags
		noCtx.Context = ""
		cfg, err = config.Resolve(noCtx, stdio.Getenv, &config.File{Contexts: cfgFile.Contexts, Defaults: cfgFile.Defaults})
	}
	if err != nil {
		return nil, err
	}

	interactive := stdio.forceInteractive || (output.IsTerminal(stdio.In) && output.IsTerminal(stdio.Out))
	outTTY := stdio.forceInteractive || output.IsTerminal(stdio.Out)

	r := &runtime{
		io:          stdio,
		flags:       f,
		cfg:         cfg,
		cfgFile:     cfgFile,
		cfgPath:     cfgPath,
		interactive: interactive,
		printer: output.New(stdio.Out, output.Options{
			Format:    cfg.Output,
			Fields:    f.fields,
			Quiet:     f.quiet,
			NoHeaders: f.noHeaders,
			TTY:       outTTY,
			Color:     !f.noColor && output.ColorEnabled(stdio.Getenv),
		}),
		prompter: &Prompter{
			in:          bufio.NewReader(stdio.In),
			out:         stdio.Err,
			interactive: interactive,
			yes:         f.yes,
			noInput:     cfg.NoInput,
			cmdPath:     cmd.CommandPath(),
		},
	}

	return r, nil
}

// client returns the API client, building it on first use so that commands
// which never talk to the server (context list, context set) do not fail on a
// missing URL or an unreadable CA file. A missing URL is a usage error.
func (r *runtime) client(cmd *cobra.Command) (*apiclient.Client, error) {
	if r.api != nil {
		return r.api, nil
	}
	if err := r.cfg.RequireURL(); err != nil {
		return nil, usage(cmd.CommandPath(), "%s", err.Error())
	}

	var debug io.Writer
	if r.cfg.Debug {
		debug = r.io.Err
	}
	api, err := apiclient.New(apiclient.Options{
		BaseURL:    r.cfg.URL,
		Token:      r.cfg.Token,
		CACertFile: r.cfg.CACert,
		Insecure:   r.cfg.Insecure,
		Timeout:    r.cfg.Timeout,
		UserAgent:  "semaphore-cli/" + r.io.Version,
		Debug:      debug,
	})
	if err != nil {
		return nil, usage(cmd.CommandPath(), "%s", err.Error())
	}
	if r.cfg.Insecure {
		r.warnf("warning: TLS certificate verification is disabled (--insecure)")
	} else if strings.HasPrefix(strings.ToLower(r.cfg.URL), "http://") && !isLoopback(r.cfg.URL) {
		r.warnf("warning: the server URL uses plain http; the token travels unencrypted")
	}
	r.api = api
	return r.api, nil
}

// requireToken fails early with a clear message instead of a bare 401.
func (r *runtime) requireToken(cmd *cobra.Command) error {
	if r.cfg.TokenError != nil {
		return r.cfg.TokenError
	}
	if r.cfg.Token == "" {
		return usage(cmd.CommandPath(), "no API token: set %s, %s or run 'semaphore client auth login'", config.EnvToken, config.EnvTokenFile)
	}
	return nil
}

// readLine reads one line from stdin through the shared buffered reader, so
// that a prompt and a piped token never fight over the same bytes.
func (r *runtime) readLine() (string, error) {
	return r.prompter.readLine()
}

// prompt writes a question to stderr and reads the answer. It refuses when the
// session is not interactive, so a script never hangs on a hidden question.
func (r *runtime) prompt(cmd *cobra.Command, question string) (string, error) {
	if r.cfg.NoInput || !r.interactive {
		return "", usage(cmd.CommandPath(), "%s: input required but the session is not interactive", strings.TrimSuffix(question, ": "))
	}
	fmt.Fprint(r.io.Err, question)
	return r.readLine()
}

// promptSecret is prompt without echo. On a real terminal the characters are
// hidden; on any other reader (tests) the line is read as-is.
func (r *runtime) promptSecret(cmd *cobra.Command, question string) (string, error) {
	if r.cfg.NoInput || !r.interactive {
		return "", usage(cmd.CommandPath(), "%s: input required but the session is not interactive", strings.TrimSuffix(question, ": "))
	}
	fmt.Fprint(r.io.Err, question)
	if f, ok := r.io.In.(*os.File); ok && output.IsTerminal(f) {
		b, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(r.io.Err)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(b)), nil
	}
	return r.readLine()
}

// saveConfig writes the configuration file with tight permissions.
func (r *runtime) saveConfig() error {
	return config.Save(r.cfgPath, r.cfgFile)
}

// warnf prints a message to stderr; data never goes there.
func (r *runtime) warnf(format string, args ...any) {
	fmt.Fprintf(r.io.Err, format+"\n", args...)
}

func isLoopback(rawURL string) bool {
	host := rawURL
	if i := strings.Index(host, "://"); i >= 0 {
		host = host[i+3:]
	}
	if i := strings.IndexAny(host, "/?#"); i >= 0 {
		host = host[:i]
	}
	if i := strings.LastIndex(host, ":"); i >= 0 && !strings.Contains(host[i:], "]") {
		host = host[:i]
	}
	host = strings.Trim(host, "[]")
	return host == "localhost" || host == "127.0.0.1" || host == "::1" || strings.HasPrefix(host, "127.")
}

// withTimeout derives a per-request context from the command context.
func withTimeout(cmd *cobra.Command, d time.Duration) (context.Context, context.CancelFunc) {
	if d <= 0 {
		return context.WithCancel(cmd.Context())
	}
	return context.WithTimeout(cmd.Context(), d)
}
