package ssh

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/random"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

// HostConfigInstallation is the ssh config and git rewrites generated from the
// credential mappings of a project, together with the agents holding the keys.
type HostConfigInstallation struct {
	ConfigFile string
	Agents     []*Agent

	// gitParams are the url.<replacement>.insteadOf rewrites of the URL
	// mappings, ready to be joined into GIT_CONFIG_PARAMETERS.
	gitParams []string
}

// InstallHostConfigs writes an ssh config binding each mapped host to the
// credential chosen for it, and returns the rewrites the URL mappings need.
//
// The key of a mapping is held by its own agent and bound with IdentityAgent
// rather than written out as an IdentityFile: a private key never reaches the
// filesystem, and one agent per mapping keeps the choice unambiguous. A single
// agent holding every key would let ssh offer the wrong one first, and sshd
// closes the connection after MaxAuthTries rejected keys.
//
// Returns a nil installation when the project has no mappings, so a project
// without them behaves exactly as before.
func InstallHostConfigs(
	projectID int,
	hostConfigs []db.HostConfig,
	logger task_logger.Logger,
) (installation *HostConfigInstallation, err error) {

	if len(hostConfigs) == 0 {
		return nil, nil
	}

	installation = &HostConfigInstallation{}

	defer func() {
		if err != nil {
			installation.Destroy()
			installation = nil
		}
	}()

	var blocks []string

	// ponytail: an agent per mapping, started whether or not the task reaches
	// that host. Fine for the handful of mappings a project has; install them
	// lazily, on the first connection to each host, if that ever changes.
	for _, hostConfig := range hostConfigs {
		// A login and password authenticate over https, so the URL carries them
		// and no ssh identity is involved: nothing to hold in an agent.
		if hostConfig.SSHKey.Type == db.AccessKeyLoginPassword {
			rewrite := credentialRewrite(hostConfig)
			if rewrite == "" {
				err = fmt.Errorf("host config %d has an invalid URL: %s", hostConfig.ID, hostConfig.Name)
				return
			}

			installation.gitParams = append(installation.gitParams, rewrite)
			continue
		}

		var agent Agent

		agent, err = StartSSHAgent(hostConfig.SSHKey, logger)
		if err != nil {
			return
		}

		installation.Agents = append(installation.Agents, &agent)

		switch hostConfig.Type {
		case db.HostConfigHost:
			var block string
			if block, err = hostBlock(hostConfig, agent.SocketFile); err != nil {
				return
			}

			blocks = append(blocks, block)
		case db.HostConfigURL:
			rewrite := urlRewrite(hostConfig)
			if rewrite == "" {
				// Only a row written around the API can get here; a mapping git
				// can not be told about must not silently look applied.
				err = fmt.Errorf("host config %d has an invalid URL: %s", hostConfig.ID, hostConfig.Name)
				return
			}

			var block string
			if block, err = urlBlock(hostConfig, agent.SocketFile); err != nil {
				return
			}

			blocks = append(blocks, block)
			installation.gitParams = append(installation.gitParams, rewrite)
		}
	}

	// Only an ssh mapping needs a generated configuration. A project whose
	// mappings all authenticate over https gets its rewrites and nothing else:
	// an empty file passed with -F takes away the configuration ssh would
	// otherwise read, rather than adding nothing to it.
	if len(blocks) == 0 {
		return
	}

	// Include comes last so a mapping wins over the same host in the
	// administrator's config: ssh keeps the first value it obtains for an
	// option. "Match all" is what takes the Include back out of the preceding
	// Host block, which would otherwise apply it only to that host.
	if adminConfig := util.Config.GetSshConfigPath(); adminConfig != "" {
		blocks = append(blocks, "Match all\nInclude "+adminConfig+"\n")
	}

	// The project directory only exists once something has been written to it.
	projectDir := util.Config.GetProjectTmpDir(projectID)
	if err = os.MkdirAll(projectDir, 0700); err != nil {
		return
	}

	installation.ConfigFile = path.Join(
		projectDir, fmt.Sprintf("ssh-config-%s.conf", random.String(10)))

	if err = os.WriteFile(installation.ConfigFile, []byte(strings.Join(blocks, "\n")), 0600); err != nil {
		return
	}

	// git runs as the configured process user, which could not otherwise read a
	// file the server wrote 0600. ChownDir reads Config.Process, which a config
	// built by hand — a test, a command — does not have.
	if util.Config.Process != nil {
		err = util.ChownDir(installation.ConfigFile)
	}

	return
}

// sshLoginRE matches a login safe to write into an ssh_config User directive.
// An allowlist, because the login comes from an access key which nothing else
// validates: a newline in it would inject further directives, such as a
// ProxyCommand, into the generated config.
var sshLoginRE = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9._-]*$`)

// hostBlock binds a host to the agent holding its credential.
//
// The login of the key is written out when it has one, but is not defaulted the
// way a URL mapping defaults it to "git": a host mapping also covers the hosts
// of an inventory, where a git login would be wrong.
func hostBlock(hostConfig db.HostConfig, agentSocket string) (string, error) {
	block := fmt.Sprintf("Host %s\n", hostConfig.Name)

	if login := hostConfig.SSHKey.SshKey.Login; login != "" {
		if !sshLoginRE.MatchString(login) {
			return "", fmt.Errorf(
				"the login of access key %d can not be used in an ssh configuration", hostConfig.SSHKeyID)
		}

		block += fmt.Sprintf("  User %s\n", login)
	}

	return block + fmt.Sprintf("  IdentityAgent %s\n", agentSocket), nil
}

// urlBlock binds the alias a URL mapping rewrites to. The alias carries the
// real host name, so several mappings on one host each keep their own
// credential.
func urlBlock(hostConfig db.HostConfig, agentSocket string) (string, error) {
	host := hostConfig.Name
	if u := parseMappingURL(host); u != nil {
		// Hostname() drops the port on purpose: it is the http port of the URL
		// being rewritten, and says nothing about where the host answers ssh.
		// ponytail: the alias therefore uses port 22; give the mapping its own
		// ssh port if a server on a non-standard one ever needs it.
		host = u.Hostname()
	}

	// The login of the key, as the git clients already default it.
	login := hostConfig.SSHKey.SshKey.Login
	if login == "" {
		login = "git"
	}

	if !sshLoginRE.MatchString(login) {
		return "", fmt.Errorf(
			"the login of access key %d can not be used in an ssh configuration", hostConfig.SSHKeyID)
	}

	return fmt.Sprintf("Host %s\n  HostName %s\n  User %s\n  IdentityAgent %s\n",
		hostConfig.SSHAlias(), host, login, agentSocket), nil
}

// urlRewrite renders the git rewrite of a URL mapping. git resolves several
// insteadOf rules by longest match, so a mapping of one repository already takes
// precedence over a mapping of the group containing it.
func urlRewrite(hostConfig db.HostConfig) string {
	u := parseMappingURL(hostConfig.Name)
	if u == nil {
		return ""
	}

	// No escaping: the URL is rejected if it holds an "=", because git splits an
	// entry at the first one and the key carries the path. Escaping it instead
	// makes the rewrite silently stop matching.
	replacement := hostConfig.SSHAlias() + ":" + strings.TrimPrefix(u.Path, "/")

	return sqQuote("url.git@" + replacement + ".insteadOf=" + hostConfig.Name)
}

// credentialRewrite renders the git rewrite of a URL mapping which authenticates
// with a login and a password, by putting them in the replacement URL.
//
// They travel in GIT_CONFIG_PARAMETERS rather than on a command line, so they
// stay out of the process list, and git only ever reports the pre-rewrite URL,
// so they stay out of the task log too.
func credentialRewrite(hostConfig db.HostConfig) string {
	u := parseMappingURL(hostConfig.Name)
	if u == nil {
		return ""
	}

	// Validation rejects this already; the rewrite is the last place the
	// credential can still be kept off a cleartext connection.
	if u.Scheme != "https" {
		return ""
	}

	login := hostConfig.SSHKey.LoginPassword

	withAuth := *u
	if login.Login == "" {
		// A token-only credential is the whole user part, as git expects.
		withAuth.User = url.User(login.Password)
	} else {
		withAuth.User = url.UserPassword(login.Login, login.Password)
	}

	// git splits a GIT_CONFIG_PARAMETERS entry at its first "=", and net/url
	// leaves "=" unescaped in userinfo, so a credential holding one would cut the
	// key short. git decodes the escape again when it authenticates. Only the
	// credential can hold one: the URL is rejected if it does.
	replacement := strings.ReplaceAll(withAuth.String(), "=", "%3D")

	return sqQuote("url." + replacement + ".insteadOf=" + hostConfig.Name)
}

// GitConfigParameters returns the value of GIT_CONFIG_PARAMETERS carrying every
// URL rewrite, or an empty string when no URL mapping is configured.
func (i *HostConfigInstallation) GitConfigParameters() string {
	if i == nil || len(i.gitParams) == 0 {
		return ""
	}

	return strings.Join(i.gitParams, " ")
}

// SSHConfigPath returns the generated config, or an empty string when nothing
// was generated.
func (i *HostConfigInstallation) SSHConfigPath() string {
	if i == nil {
		return ""
	}

	return i.ConfigFile
}

// Destroy closes every agent and removes the generated config.
func (i *HostConfigInstallation) Destroy() {
	if i == nil {
		return
	}

	for _, agent := range i.Agents {
		util.LogWarning(agent.Close())
	}
	i.Agents = nil

	if i.ConfigFile != "" {
		if err := os.Remove(i.ConfigFile); err != nil && !os.IsNotExist(err) {
			util.LogWarning(err)
		}
		i.ConfigFile = ""
	}
}

// parseMappingURL parses the URL of a mapping. It is already validated when it
// is stored, so a failure here means the row was written around the API.
func parseMappingURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return nil
	}
	return u
}

// sqQuote quotes an entry for GIT_CONFIG_PARAMETERS, which git reads as a
// shell-quoted list.
func sqQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
