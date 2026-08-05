package db

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type ProxyType string

const (
	ProxySSH ProxyType = "ssh"
)

// Proxy is a project scoped connection proxy. Currently only SSH jump hosts
// (ProxyJump) are supported; the type is stored so that SOCKS and HTTP proxies
// can be added without a schema change.
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
}

func (p *Proxy) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return &ValidationError{"proxy name can not be empty"}
	}

	if strings.TrimSpace(p.Host) == "" {
		return &ValidationError{"proxy host can not be empty"}
	}

	// The host ends up in an ssh command line, so refuse anything that could
	// add arguments or shell syntax.
	if strings.ContainsAny(p.Host, " \t\r\n\"'\\$`") || strings.HasPrefix(p.Host, "-") {
		return &ValidationError{"proxy host contains invalid characters"}
	}

	if p.Port != nil && (*p.Port < 1 || *p.Port > 65535) {
		return &ValidationError{"proxy port must be between 1 and 65535"}
	}

	if p.User != nil && strings.ContainsAny(*p.User, " \t\r\n\"'\\$`@") {
		return &ValidationError{"proxy user contains invalid characters"}
	}

	if p.SSHKeyID != nil && *p.SSHKeyID <= 0 {
		return &ValidationError{"proxy ssh key id must be a valid key"}
	}

	switch p.Type {
	case ProxySSH:
	default:
		return &ValidationError{"unsupported proxy type"}
	}

	return nil
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
