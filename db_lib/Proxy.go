package db_lib

import (
	"strconv"
	"strings"

	"github.com/semaphoreui/semaphore/db"
)

// ProxyChainKeys returns the keys of every proxy of the chain, so they can be
// held by a single agent. Each host of the chain rejects the keys which are not
// its own and ssh moves on to the next, which is why one agent is enough.
func ProxyChainKeys(proxy db.Proxy) (keys []db.AccessKey) {
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
