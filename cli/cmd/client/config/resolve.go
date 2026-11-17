package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/semaphoreui/semaphore/pkg/apiclient"
)

// Source says where an effective value came from.
type Source string

const (
	SourceFlag    Source = "flag"
	SourceEnv     Source = "env"
	SourceContext Source = "context"
	SourceDefault Source = "default"
	SourceNone    Source = ""
)

// Flags are the values given on the command line. A zero value means the flag
// was not used; a bool flag can only switch a setting on.
type Flags struct {
	Context  string
	URL      string
	Project  string
	CACert   string
	Insecure bool
	Output   string
	Timeout  time.Duration
	NoInput  bool
	Debug    bool
}

// Effective is the resolved configuration of one invocation.
type Effective struct {
	// ContextName is the context in use, empty when the client runs on
	// environment and flags alone.
	ContextName string
	URL         string
	Token       string
	// TokenFile is the file Token was read from, when it came from one.
	TokenFile string
	// TokenError is set when a token file could not be read. It is deferred
	// rather than returned so that commands which need no token (context
	// list, context set) still work on a broken context.
	TokenError error
	Project    string
	CACert     string
	Insecure   bool
	Output     string
	Timeout    time.Duration
	NoInput    bool
	Debug      bool

	Sources Sources
}

// Sources records the origin of the values a user is likely to ask about.
type Sources struct {
	Context Source
	URL     Source
	Token   Source
	Project Source
	Output  Source
	// CACert and Insecure say whether the TLS settings came from the context.
	// Pointing that context at another server has to leave them behind.
	CACert   Source
	Insecure Source
}

// Errors a caller may want to tell apart.
var (
	// ErrNoURL means no flag, variable or context gave a server address.
	ErrNoURL = errors.New("server URL is not set")
	// ErrUnknownContext means the requested context is not in the file.
	ErrUnknownContext = errors.New("unknown context")
	// ErrTokenWithoutURL means SEMAPHORE_CLIENT_TOKEN is set while the URL
	// comes from the config file. The token would be sent to whichever server
	// the context names, which is how a CI token leaks to a foreign host.
	ErrTokenWithoutURL = errors.New(EnvToken + " is set but the server URL comes from the configuration file")
	// ErrTokenAndTokenFile means both variables are set and it is unclear which wins.
	ErrTokenAndTokenFile = errors.New(EnvToken + " and " + EnvTokenFile + " are both set; use one")
	// ErrTokenForOtherServer means the context holds a token saved for one
	// server while --url or SEMAPHORE_CLIENT_URL points at another. The token
	// is the credential of the first server and is not sent to the second.
	ErrTokenForOtherServer = errors.New("the saved token belongs to another server")
	// ErrBadOutput means the output format is not table, json or yaml.
	ErrBadOutput = errors.New("output format must be table, json or yaml")
)

// Resolve combines flags, environment and file into the effective settings.
// file may be nil. getenv is os.Getenv in production and a map lookup in tests.
func Resolve(flags Flags, getenv func(string) string, file *File) (Effective, error) {
	if file == nil {
		file = &File{}
	}
	var eff Effective

	// Context: flag > env > the one named "default". A context named on the
	// command line or in the environment must exist; the implicit default may
	// be absent, which is the normal state of an environment-only setup.
	var ctx Context
	named := true
	switch {
	case flags.Context != "":
		eff.ContextName, eff.Sources.Context = flags.Context, SourceFlag
	case getenv(EnvContext) != "":
		eff.ContextName, eff.Sources.Context = getenv(EnvContext), SourceEnv
	default:
		eff.ContextName, eff.Sources.Context, named = DefaultContextName, SourceDefault, false
	}
	var ok bool
	ctx, ok = file.Contexts[eff.ContextName]
	if !ok {
		if named {
			return eff, fmt.Errorf("%w %q; known contexts: %s", ErrUnknownContext, eff.ContextName, joinOrNone(file.ContextNames()))
		}
		eff.Sources.Context = SourceNone
	}

	// URL.
	switch {
	case flags.URL != "":
		eff.URL, eff.Sources.URL = flags.URL, SourceFlag
	case getenv(EnvURL) != "":
		eff.URL, eff.Sources.URL = getenv(EnvURL), SourceEnv
	case ctx.URL != "":
		eff.URL, eff.Sources.URL = ctx.URL, SourceContext
	}

	// Token. There is no flag on purpose: process arguments are visible to
	// every user of the machine and end up in shell history.
	envToken, envTokenFile := getenv(EnvToken), getenv(EnvTokenFile)
	if envToken != "" && envTokenFile != "" {
		return eff, ErrTokenAndTokenFile
	}
	switch {
	case envToken != "" || envTokenFile != "":
		if eff.Sources.URL == SourceContext {
			return eff, fmt.Errorf("%w (context %q); set %s as well or unset the token", ErrTokenWithoutURL, eff.ContextName, EnvURL)
		}
		eff.Sources.Token = SourceEnv
		if envToken != "" {
			eff.Token = strings.TrimSpace(envToken)
		} else {
			eff.TokenFile = expandHome(envTokenFile, getenv)
			eff.Token, eff.TokenError = readTokenFile(eff.TokenFile)
		}
	case ctx.Token != "" || ctx.TokenFile != "":
		// The context's token was issued by the context's server. When the
		// address is overridden it must stay behind, or `--url` turns into a
		// way to hand a production credential to any host that asks.
		if ctx.URL != "" && eff.Sources.URL != SourceContext && !sameServer(ctx.URL, eff.URL) {
			// Deferred, like an unreadable token file: a command that needs no
			// token (context list, auth login) still runs.
			eff.TokenError = fmt.Errorf("%w: context %q holds the token of %s, not of %s; use --context for that server or set %s",
				ErrTokenForOtherServer, eff.ContextName, ctx.URL, eff.URL, EnvToken)
			break
		}
		eff.Sources.Token = SourceContext
		if ctx.Token != "" {
			eff.Token = strings.TrimSpace(ctx.Token)
		} else {
			eff.TokenFile = expandHome(ctx.TokenFile, getenv)
			eff.Token, eff.TokenError = readTokenFile(eff.TokenFile)
		}
	}

	// Project.
	switch {
	case flags.Project != "":
		eff.Project, eff.Sources.Project = flags.Project, SourceFlag
	case getenv(EnvProject) != "":
		eff.Project, eff.Sources.Project = getenv(EnvProject), SourceEnv
	case ctx.Project != "":
		eff.Project, eff.Sources.Project = ctx.Project, SourceContext
	}

	// TLS.
	switch {
	case flags.CACert != "":
		eff.CACert, eff.Sources.CACert = expandHome(flags.CACert, getenv), SourceFlag
	case getenv(EnvCACert) != "":
		eff.CACert, eff.Sources.CACert = expandHome(getenv(EnvCACert), getenv), SourceEnv
	case ctx.CACert != "":
		eff.CACert, eff.Sources.CACert = expandHome(ctx.CACert, getenv), SourceContext
	}
	switch {
	case flags.Insecure:
		eff.Insecure, eff.Sources.Insecure = true, SourceFlag
	case truthy(getenv(EnvInsecure)):
		eff.Insecure, eff.Sources.Insecure = true, SourceEnv
	case ctx.Insecure:
		eff.Insecure, eff.Sources.Insecure = true, SourceContext
	}

	// Output.
	switch {
	case flags.Output != "":
		eff.Output, eff.Sources.Output = flags.Output, SourceFlag
	case getenv(EnvOutput) != "":
		eff.Output, eff.Sources.Output = getenv(EnvOutput), SourceEnv
	case file.Defaults.Output != "":
		eff.Output, eff.Sources.Output = file.Defaults.Output, SourceContext
	default:
		eff.Output, eff.Sources.Output = OutputTable, SourceDefault
	}
	eff.Output = strings.ToLower(eff.Output)
	switch eff.Output {
	case OutputTable, OutputJSON, OutputYAML:
	default:
		return eff, fmt.Errorf("%w, got %q", ErrBadOutput, eff.Output)
	}

	// Timeout.
	switch {
	case flags.Timeout > 0:
		eff.Timeout = flags.Timeout
	case getenv(EnvTimeout) != "":
		d, err := parseDuration(getenv(EnvTimeout))
		if err != nil {
			return eff, fmt.Errorf("%s: %w", EnvTimeout, err)
		}
		eff.Timeout = d
	case file.Defaults.Timeout != "":
		d, err := parseDuration(file.Defaults.Timeout)
		if err != nil {
			return eff, fmt.Errorf("defaults.timeout: %w", err)
		}
		eff.Timeout = d
	default:
		eff.Timeout = DefaultTimeout
	}

	eff.NoInput = flags.NoInput || truthy(getenv(EnvNoInput)) || truthy(getenv(EnvCI))
	eff.Debug = flags.Debug || truthy(getenv(EnvDebug))

	return eff, nil
}

// RequireURL is what a command that talks to the server calls first.
func (e Effective) RequireURL() error {
	if e.URL == "" {
		return fmt.Errorf("%w: pass --url, set %s or run `semaphore client auth login`", ErrNoURL, EnvURL)
	}
	return nil
}

// readTokenFile reads a token from a file, trimming the trailing newline an
// editor or `echo` leaves behind.
func readTokenFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read token file: %w", err)
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("token file %s is empty", path)
	}
	return token, nil
}

// parseDuration accepts Go durations ("30s", "2m") and bare seconds ("45").
func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if n, err := strconv.Atoi(s); err == nil {
		if n <= 0 {
			return 0, fmt.Errorf("timeout must be positive, got %d", n)
		}
		return time.Duration(n) * time.Second, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid timeout %q: use seconds or a duration like 30s", s)
	}
	if d <= 0 {
		return 0, fmt.Errorf("timeout must be positive, got %s", s)
	}
	return d, nil
}

// truthy reads an environment boolean: 1, true, yes, on (any case).
// sameServer compares two addresses the way the client sees them, so that a
// trailing slash or an "/api" suffix is not read as a different server.
func sameServer(a, b string) bool {
	return apiclient.NormalizeURL(a) == apiclient.NormalizeURL(b)
}

func truthy(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "on", "y":
		return true
	}
	return false
}

func joinOrNone(names []string) string {
	if len(names) == 0 {
		return "(none)"
	}
	return strings.Join(names, ", ")
}
