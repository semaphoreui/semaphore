package ssh

import (
	"fmt"
	"net/url"
	"os"
	"path"
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
		var agent Agent

		agent, err = StartSSHAgent(hostConfig.SSHKey, logger)
		if err != nil {
			return
		}

		installation.Agents = append(installation.Agents, &agent)

		switch hostConfig.Type {
		case db.HostConfigHost:
			blocks = append(blocks, hostBlock(hostConfig.Name, agent.SocketFile))
		case db.HostConfigURL:
			rewrite := urlRewrite(hostConfig)
			if rewrite == "" {
				// Only a row written around the API can get here; a mapping git
				// can not be told about must not silently look applied.
				err = fmt.Errorf("host config %d has an invalid URL: %s", hostConfig.ID, hostConfig.Name)
				return
			}

			blocks = append(blocks, urlBlock(hostConfig, agent.SocketFile))
			installation.gitParams = append(installation.gitParams, rewrite)
		}
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

	err = os.WriteFile(installation.ConfigFile, []byte(strings.Join(blocks, "\n")), 0600)

	return
}

// hostBlock binds a host to the agent holding its credential.
func hostBlock(host string, agentSocket string) string {
	return fmt.Sprintf("Host %s\n  IdentityAgent %s\n", host, agentSocket)
}

// urlBlock binds the alias a URL mapping rewrites to. The alias carries the
// real host name, so several mappings on one host each keep their own
// credential.
func urlBlock(hostConfig db.HostConfig, agentSocket string) string {
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

	return fmt.Sprintf("Host %s\n  HostName %s\n  User %s\n  IdentityAgent %s\n",
		hostConfig.SSHAlias(), host, login, agentSocket)
}

// urlRewrite renders the git rewrite of a URL mapping. git resolves several
// insteadOf rules by longest match, so a mapping of one repository already takes
// precedence over a mapping of the group containing it.
func urlRewrite(hostConfig db.HostConfig) string {
	u := parseMappingURL(hostConfig.Name)
	if u == nil {
		return ""
	}

	replacement := hostConfig.SSHAlias() + ":" + strings.TrimPrefix(u.Path, "/")

	// git splits a GIT_CONFIG_PARAMETERS entry at its first "=", so an "=" in
	// the key would cut it short. git decodes the escape again when it matches.
	replacement = strings.ReplaceAll(replacement, "=", "%3D")
	matched := strings.ReplaceAll(hostConfig.Name, "=", "%3D")

	return sqQuote("url.git@" + replacement + ".insteadOf=" + matched)
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
