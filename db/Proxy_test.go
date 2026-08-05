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

	tests := []struct {
		name    string
		proxy   Proxy
		wantErr bool
	}{
		{"valid", Proxy{Name: "bastion", Type: ProxySSH, Host: "bastion.example.org", Port: &okPort}, false},
		{"empty name", Proxy{Type: ProxySSH, Host: "bastion.example.org"}, true},
		{"empty host", Proxy{Name: "bastion", Type: ProxySSH}, true},
		{"unsupported type", Proxy{Name: "bastion", Type: "socks5", Host: "bastion.example.org"}, true},
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
