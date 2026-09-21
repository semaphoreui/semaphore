package setup

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAskSecretValue_NonTTYUsesPlainInput(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)

	oldStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = oldStdin
		_ = r.Close()
	})

	go func() {
		_, _ = io.WriteString(w, "s3cret with spaces\n")
		_ = w.Close()
	}()

	var got string
	askSecretValue("db Password", "default", &got)

	assert.Equal(t, "s3cret with spaces", got)
}

func TestAskSecretValue_EmptyInputKeepsDefault(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)

	oldStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = oldStdin
		_ = r.Close()
	})

	go func() {
		_, _ = io.WriteString(w, "\n")
		_ = w.Close()
	}()

	var got string
	askSecretValue("Password for LDAP bind user", "pa55w0rd", &got)

	assert.Equal(t, "pa55w0rd", got)
}

func TestReadSecretLine_NonTTY(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)

	oldStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = oldStdin
		_ = r.Close()
	})

	go func() {
		_, _ = io.WriteString(w, "admin-pass\n")
		_ = w.Close()
	}()

	assert.Equal(t, "admin-pass", ReadSecretLine(" > Password: "))
}
