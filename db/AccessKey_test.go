package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccessKeyType_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		typ   AccessKeyType
		valid bool
	}{
		{"ssh", AccessKeySSH, true},
		{"login_password", AccessKeyLoginPassword, true},
		{"none", AccessKeyNone, true},
		{"string", AccessKeyString, true},
		{"empty", AccessKeyType(""), false},
		{"unknown", AccessKeyType("password"), false},
		{"wrong case", AccessKeyType("SSH"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.typ.IsValid())
		})
	}
}
