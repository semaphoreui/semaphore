package db_lib

import (
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/semaphoreui/semaphore/db"
)

// connectorCommand renders the ProxyCommand which carries an ssh connection
// through a SOCKS5 or HTTP proxy. OpenSSH speaks neither protocol, and rather
// than depend on netcat or socat being installed next to Semaphore, the
// Semaphore binary connects on its behalf.
func connectorCommand(proxy db.Proxy) string {
	return strings.Join([]string{
		semaphoreExecutable(),
		"proxy-connect",
		"--proxy", proxy.URL(),
		"%h", "%p",
	}, " ")
}

// semaphoreExecutable is the path ssh runs the connector with. It falls back to
// the command name so a PATH installation still works if the path can not be
// resolved.
func semaphoreExecutable() string {
	if exe, err := os.Executable(); err == nil {
		return exe
	}

	return "semaphore"
}

// ProxyEnv returns the environment a proxy needs, currently the credentials the
// connector authenticates with, which are kept off the command line.
func ProxyEnv(proxy *db.Proxy) (env []string) {
	if proxy == nil || proxy.Type.IsSSH() || proxy.SSHKeyID == nil {
		return
	}

	login := proxy.SSHKey.LoginPassword

	if login.Login == "" && login.Password == "" {
		return
	}

	return []string{
		"SEMAPHORE_PROXY_USER=" + login.Login,
		"SEMAPHORE_PROXY_PASSWORD=" + login.Password,
	}
}

// ProxyURLWithCredentials renders the proxy URL including its credentials, for
// programs which take the whole thing, such as git. It is only ever put in an
// environment variable, never on a command line.
func ProxyURLWithCredentials(proxy db.Proxy) string {
	raw := proxy.URL()
	if raw == "" || proxy.SSHKeyID == nil {
		return raw
	}

	login := proxy.SSHKey.LoginPassword
	if login.Login == "" && login.Password == "" {
		return raw
	}

	scheme, host, found := strings.Cut(raw, "://")
	if !found {
		return raw
	}

	return scheme + "://" + url.UserPassword(login.Login, login.Password).String() + "@" + host
}

// ProxyChainKeys returns the keys of every proxy of the chain, so they can be
// held by a single agent. Each host of the chain rejects the keys which are not
// its own and ssh moves on to the next, which is why one agent is enough.
func ProxyChainKeys(proxy db.Proxy) (keys []db.AccessKey) {
	if !proxy.Type.IsSSH() {
		return
	}

	for _, hop := range proxy.Chain() {
		if hop.SSHKeyID != nil {
			keys = append(keys, hop.SSHKey)
		}
	}

	return
}

// ProxyCommandOption renders the chain of a proxy as an ssh ProxyCommand which
// opens a tunnel to the target host.
//
// ProxyCommand instead of ProxyJump: the proxies and the target host use
// different keys, held in different agents, and IdentityAgent applies to the
// outer connection only. A nested ssh is spawned for the jump with the agent of
// the proxy keys attached to it.
//
// The hops of a chain are nested the same way rather than passed as one
// ProxyJump list, because ssh takes the configuration of a ProxyJump host from
// ssh_config and ignores the options given on the command line, so the hops
// would lose the host key policy of the connection.
//
// agentSocket may be empty when no proxy of the chain has a key.
func ProxyCommandOption(proxy db.Proxy, agentSocket string) string {
	if !proxy.Type.IsSSH() {
		return "ProxyCommand=" + connectorCommand(proxy)
	}

	chain := proxy.Chain()

	// Built from the first hop outwards: each hop opens a tunnel to the next,
	// and the last one tunnels to the target host of the connection.
	command := ""

	for i, hop := range chain {
		target := "%h:%p"
		if i+1 < len(chain) {
			next := chain[i+1]
			target = next.Host + ":" + strconv.Itoa(sshPort(next.Port))
		}

		command = hopCommand(hop, agentSocket, target, command)
	}

	return "ProxyCommand=" + command
}

// hopCommand renders one ssh hop. proxyCommand is the command reaching this hop,
// empty when it is reachable directly.
func hopCommand(hop db.Proxy, agentSocket string, target string, proxyCommand string) string {
	cmd := []string{"ssh"}

	if agentSocket != "" {
		cmd = append(cmd, "-o", "IdentityAgent="+agentSocket)
	}

	if proxyCommand != "" {
		cmd = append(cmd, "-o", "ProxyCommand="+shellQuote(proxyCommand))
	}

	cmd = append(cmd, "-o", "StrictHostKeyChecking=no", "-W", target)

	if hop.Port != nil {
		cmd = append(cmd, "-p", strconv.Itoa(*hop.Port))
	}

	return strings.Join(append(cmd, hop.SSHDestination()), " ")
}

func sshPort(port *int) int {
	if port == nil {
		return 22
	}
	return *port
}

// shellQuote quotes a command for the shell ssh runs a ProxyCommand with. Nested
// hops quote each other, so an already quoted command must survive being quoted
// again.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
