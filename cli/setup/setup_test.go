package setup

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withPipedStdin(t *testing.T, input string) {
	t.Helper()

	r, w, err := os.Pipe()
	require.NoError(t, err)

	oldStdin := os.Stdin
	os.Stdin = r
	resetStdin()

	t.Cleanup(func() {
		os.Stdin = oldStdin
		resetStdin()
		_ = r.Close()
	})

	go func() {
		_, _ = io.WriteString(w, input)
		_ = w.Close()
	}()
}

func TestAskSecretValue_NonTTYUsesPlainInput(t *testing.T) {
	withPipedStdin(t, "s3cret with spaces\n")

	var got string
	askSecretValue("db Password", "default", &got)

	assert.Equal(t, "s3cret with spaces", got)
}

func TestAskSecretValue_EmptyInputKeepsDefault(t *testing.T) {
	withPipedStdin(t, "\n")

	var got string
	askSecretValue("Password for LDAP bind user", "pa55w0rd", &got)

	assert.Equal(t, "pa55w0rd", got)
}

func TestReadSecretLine_NonTTY(t *testing.T) {
	withPipedStdin(t, "admin-pass\n")

	assert.Equal(t, "admin-pass", ReadSecretLine(" > Password: "))
}

func TestSharedStdin_PreservesFollowingPrompts(t *testing.T) {
	// Mimics CI: secret prompt must not discard the next piped answer.
	withPipedStdin(t, "p455w0rd\nsemaphore\nadmin\n")

	var password, dbName string
	askSecretValue("db Password", "", &password)
	askValue("db Name", "default", &dbName)

	assert.Equal(t, "p455w0rd", password)
	assert.Equal(t, "semaphore", dbName)
	assert.Equal(t, "admin", ReadSecretLine(" > Password: "))
}
