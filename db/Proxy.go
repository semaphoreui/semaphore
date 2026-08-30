package db

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

type ProxyType string

const (
	ProxySSH    ProxyType = "ssh"
	ProxyHTTP   ProxyType = "http"
	ProxyHTTPS  ProxyType = "https"
	ProxySOCKS5 ProxyType = "socks5"
)

// IsSSH reports whether the proxy is an SSH jump host, which is reached by
// opening an ssh connection to it, rather than a proxy protocol carrying the
// connection of another program.
func (t ProxyType) IsSSH() bool {
	return t == ProxySSH
}

// Scheme returns the URL scheme of a proxy protocol, as git and the connector
// expect it. It is empty for SSH jump hosts, which are not a URL.
func (t ProxyType) Scheme() string {
	switch t {
	case ProxyHTTP, ProxyHTTPS, ProxySOCKS5:
		return string(t)
	default:
		return ""
	}
}

// Proxy is a project scoped connection proxy: either an SSH jump host or a
// SOCKS5/HTTP(S) proxy carrying the connection.
type Proxy struct {
	ID        int       `db:"id" json:"id" backup:"-"`
	ProjectID int       `db:"project_id" json:"project_id" backup:"-"`
	Name      string    `db:"name" json:"name" binding:"required"`
	Type      ProxyType `db:"type" json:"type"`
	Host      string    `db:"host" json:"host" binding:"required"`
	Port      *int      `db:"port" json:"port,omitempty"`
	User      *string   `db:"user" json:"user,omitempty"`

	// SSHKeyID is the key used to authenticate against the proxy itself, which
	// is usually not the key used for the target host.
	SSHKeyID *int      `db:"ssh_key_id" json:"ssh_key_id,omitempty" backup:"-"`
	SSHKey   AccessKey `db:"-" json:"-" backup:"-"`

	// RequiresProxyID is another proxy of the same project which must be passed
	// through to reach this one, for chained proxy situations.
	RequiresProxyID *int   `db:"requires_proxy_id" json:"requires_proxy_id,omitempty" backup:"-"`
	RequiresProxy   *Proxy `db:"-" json:"-" backup:"-"`
}

// MaxProxyChainLength bounds how many proxies may be chained. Each hop is one
// more key offered to every host of the chain, and sshd closes a connection
// after MaxAuthTries (6 by default) rejected keys.
const MaxProxyChainLength = 5

// Chain returns the proxies to pass through to reach a host, in the order they
// are connected to: the proxy which requires no other comes first, p comes last.
func (p Proxy) Chain() []Proxy {
	chain := []Proxy{p}

	for required := p.RequiresProxy; required != nil; required = required.RequiresProxy {
		chain = append([]Proxy{*required}, chain...)

		if len(chain) >= MaxProxyChainLength {
			break
		}
	}

	return chain
}

func (p *Proxy) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return common_errors.NewValidationError("proxy name can not be empty")
	}

	if strings.TrimSpace(p.Host) == "" {
		return common_errors.NewValidationError("proxy host can not be empty")
	}

	// The host ends up in an ssh command line, so refuse anything that could
	// add arguments or shell syntax.
	if strings.ContainsAny(p.Host, " \t\r\n\"'\\$`") || strings.HasPrefix(p.Host, "-") {
		return common_errors.NewValidationError("proxy host contains invalid characters")
	}

	if p.Port != nil && (*p.Port < 1 || *p.Port > 65535) {
		return common_errors.NewValidationError("proxy port must be between 1 and 65535")
	}

	if p.User != nil && strings.ContainsAny(*p.User, " \t\r\n\"'\\$`@") {
		return common_errors.NewValidationError("proxy user contains invalid characters")
	}

	if p.SSHKeyID != nil && *p.SSHKeyID <= 0 {
		return common_errors.NewValidationError("proxy ssh key id must be a valid key")
	}

	switch p.Type {
	case ProxySSH:
	case ProxyHTTP, ProxyHTTPS, ProxySOCKS5:
		if p.RequiresProxyID != nil {
			return common_errors.NewValidationError("only an ssh proxy can be chained")
		}
	default:
		return common_errors.NewValidationError("unsupported proxy type")
	}

	return nil
}

// URL renders a non-SSH proxy as the URL git and the connector take, for example
// "socks5://proxy.example.org:1080".
//
// Credentials are deliberately left out: the URL ends up on a command line,
// which every process on the host can read. They are passed as environment
// variables instead.
func (p Proxy) URL() string {
	scheme := p.Type.Scheme()
	if scheme == "" {
		return ""
	}

	host := p.Host
	if p.Port != nil {
		host = net.JoinHostPort(host, strconv.Itoa(*p.Port))
	}

	return fmt.Sprintf("%s://%s", scheme, host)
}

// SSHDestination renders the proxy as "user@host", without the port, for use
// with ssh commands which take the port as a separate argument.
func (p Proxy) SSHDestination() string {
	if p.User != nil && *p.User != "" {
		return fmt.Sprintf("%s@%s", *p.User, p.Host)
	}
	return p.Host
}

// Destination renders the proxy as an OpenSSH ProxyJump destination,
// for example "ansible-proxy@bastion.example.org:2222".
func (p Proxy) Destination() string {
	dst := p.Host

	if p.Port != nil {
		dst = net.JoinHostPort(dst, strconv.Itoa(*p.Port))
	}

	if p.User != nil && *p.User != "" {
		dst = fmt.Sprintf("%s@%s", *p.User, dst)
	}

	return dst
}
