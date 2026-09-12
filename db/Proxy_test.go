package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxy_Destination(t *testing.T) {
	user := "ansible-proxy"
	port := 2222

	tests := []struct {
		name     string
		proxy    Proxy
		expected string
	}{
		{"host only", Proxy{Host: "bastion.example.org"}, "bastion.example.org"},
		{"host and user", Proxy{Host: "bastion.example.org", User: &user}, "ansible-proxy@bastion.example.org"},
		{"host and port", Proxy{Host: "bastion.example.org", Port: &port}, "bastion.example.org:2222"},
		{"all parts", Proxy{Host: "bastion.example.org", User: &user, Port: &port}, "ansible-proxy@bastion.example.org:2222"},
		{"ipv6 host and port", Proxy{Host: "2001:db8::1", Port: &port}, "[2001:db8::1]:2222"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.proxy.Destination())
		})
	}
}

func TestProxy_Validate(t *testing.T) {
	badPort := 0
	okPort := 22
	injected := "user -oProxyCommand=evil"
	zeroKey := 0
	oneKey := 1

	tests := []struct {
		name    string
		proxy   Proxy
		wantErr bool
	}{
		{"valid", Proxy{Name: "bastion", Type: ProxySSH, Host: "bastion.example.org", Port: &okPort}, false},
		{"empty name", Proxy{Type: ProxySSH, Host: "bastion.example.org"}, true},
		{"empty host", Proxy{Name: "bastion", Type: ProxySSH}, true},
		{"unsupported type", Proxy{Name: "bastion", Type: "ftp", Host: "bastion.example.org"}, true},
		{"socks5", Proxy{Name: "socks", Type: ProxySOCKS5, Host: "proxy.example.org"}, false},
		{"http", Proxy{Name: "http", Type: ProxyHTTP, Host: "proxy.example.org"}, false},
		{"https", Proxy{Name: "https", Type: ProxyHTTPS, Host: "proxy.example.org"}, false},
		{"a socks proxy can not be chained", Proxy{Name: "socks", Type: ProxySOCKS5, Host: "proxy.example.org", RequiresProxyID: &oneKey}, true},
		{"port out of range", Proxy{Name: "bastion", Type: ProxySSH, Host: "bastion.example.org", Port: &badPort}, true},
		{"host with spaces", Proxy{Name: "bastion", Type: ProxySSH, Host: "bastion.example.org evil"}, true},
		{"host starting with dash", Proxy{Name: "bastion", Type: ProxySSH, Host: "-oProxyCommand=evil"}, true},
		{"user with injection", Proxy{Name: "bastion", Type: ProxySSH, Host: "bastion.example.org", User: &injected}, true},
		{"zero ssh key id", Proxy{Name: "bastion", Type: ProxySSH, Host: "bastion.example.org", SSHKeyID: &zeroKey}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.proxy.Validate()

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

// TestProxy_Validate_HostShellInjection covers the characters which reach the
// shell ssh runs a ProxyCommand with. The allowlist must reject the whole class,
// not just the separators: `#` comments out the rest of the command, and the
// globbing and expansion characters are equally shell syntax.
func TestProxy_Validate_HostShellInjection(t *testing.T) {
	hosts := []string{
		"bastion.example.org;id;",
		"bastion.example.org&id",
		"bastion.example.org|id",
		"bastion.example.org#",
		"bastion.example.org$(id)",
		"bastion.example.org`id`",
		"bastion.example.org${IFS}",
		"bastion.example.org>out",
		"bastion.example.org<in",
		"bastion.example.org(id)",
		"bastion.example.org*",
		"bastion.example.org?",
		"bastion.example.org[a]",
		"bastion.example.org{a,b}",
		"bastion.example.org~",
		"bastion.example.org!",
		"bastion.example.org evil",
		"bastion.example.org\nid",
		"bastion.example.org\tid",
		"bastion.example.org'id'",
		"bastion.example.org\"id\"",
		"bastion.example.org\\id",
		"-oProxyCommand=evil",
		"bastion.example.org-",
	}

	for _, host := range hosts {
		t.Run(host, func(t *testing.T) {
			p := Proxy{Name: "bastion", Type: ProxySSH, Host: host}
			assert.Error(t, p.Validate())
		})
	}
}

// TestProxy_Validate_UserShellInjection covers the same class on the user, which
// is interpolated into the very same command as "user@host".
func TestProxy_Validate_UserShellInjection(t *testing.T) {
	users := []string{
		"root;id;", "root&id", "root|id", "root#", "root$(id)", "root`id`",
		"root>out", "root(id)", "root*", "root~", "root!", "root evil",
		"root\nid", "root'id'", "root@other", "-oProxyCommand=evil",
	}

	for _, user := range users {
		t.Run(user, func(t *testing.T) {
			u := user
			p := Proxy{Name: "bastion", Type: ProxySSH, Host: "bastion.example.org", User: &u}
			assert.Error(t, p.Validate())
		})
	}
}

// Legitimate hosts and users must keep working.
func TestProxy_Validate_AcceptsRealHostsAndUsers(t *testing.T) {
	hosts := []string{
		"bastion", "bastion.example.org", "bastion-01.eu-west-1.example.org",
		"host_name", "192.168.1.10", "2001:db8::1", "::1", "127.0.0.1",
	}
	for _, host := range hosts {
		t.Run("host/"+host, func(t *testing.T) {
			p := Proxy{Name: "bastion", Type: ProxySSH, Host: host}
			assert.NoError(t, p.Validate())
		})
	}

	users := []string{"root", "ansible-proxy", "ansible_proxy", "first.last", "ec2-user", "u2"}
	for _, user := range users {
		t.Run("user/"+user, func(t *testing.T) {
			u := user
			p := Proxy{Name: "bastion", Type: ProxySSH, Host: "bastion.example.org", User: &u}
			assert.NoError(t, p.Validate())
		})
	}
}
