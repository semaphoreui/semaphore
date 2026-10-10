package ansible_vault

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const password = "s3cr3t-pa55word {{ 1+1 }}"

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
	}{
		{"yaml document", "token: !unsafe \"s3cr3t {{ 1+1 }}\"\ndb_password: !unsafe \"p@ss\"\n"},
		{"empty", ""},
		{"one block exactly", strings.Repeat("a", 16)},
		{"two blocks exactly", strings.Repeat("b", 32)},
		{"one byte", "x"},
		{"binary", string([]byte{0, 1, 2, 255, 254, '\n', '\n'})},
		{"long", strings.Repeat("secret line with unicode ключ\n", 200)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vaulttext, err := Encrypt([]byte(tt.plaintext), password)
			require.NoError(t, err)

			got, err := Decrypt(vaulttext, password)
			require.NoError(t, err)
			assert.Equal(t, []byte(tt.plaintext), got)
		})
	}
}

func TestEncrypt_Envelope(t *testing.T) {
	vaulttext, err := Encrypt([]byte(strings.Repeat("payload\n", 50)), password)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSuffix(string(vaulttext), "\n"), "\n")
	require.Greater(t, len(lines), 2)
	assert.Equal(t, "$ANSIBLE_VAULT;1.1;AES256", lines[0], "header without a vault id")
	assert.True(t, bytes.HasSuffix(vaulttext, []byte("\n")), "newline terminated")

	for i, line := range lines[1:] {
		assert.LessOrEqual(t, len(line), 80, "line %d", i+1)
		assert.Regexp(t, `^[0-9a-f]+$`, line, "line %d is hex", i+1)
	}
	for _, line := range lines[1 : len(lines)-1] {
		assert.Len(t, line, 80, "every line but the last is full")
	}
}

func TestEncrypt_SaltIsRandom(t *testing.T) {
	a, err := Encrypt([]byte("same"), password)
	require.NoError(t, err)
	b, err := Encrypt([]byte("same"), password)
	require.NoError(t, err)
	assert.NotEqual(t, a, b, "a fresh salt every time")
}

func TestDecrypt_WrongPassword(t *testing.T) {
	vaulttext, err := Encrypt([]byte("payload"), password)
	require.NoError(t, err)

	_, err = Decrypt(vaulttext, "not the password")
	assert.ErrorIs(t, err, ErrAuth)
}

func TestDecrypt_Tampered(t *testing.T) {
	vaulttext, err := Encrypt([]byte("payload that spans more than one block"), password)
	require.NoError(t, err)

	// Change the last hex digit of the ciphertext inside the double-hex body:
	// the ciphertext changes, the HMAC does not, and the body stays valid hex.
	headerEnd := bytes.IndexByte(vaulttext, '\n')
	body := bytes.ReplaceAll(vaulttext[headerEnd+1:], []byte("\n"), nil)
	inner, err := hexDecode(body)
	require.NoError(t, err)
	last := len(inner) - 1
	if inner[last] == '0' {
		inner[last] = '1'
	} else {
		inner[last] = '0'
	}
	tampered := append(append([]byte{}, vaulttext[:headerEnd+1]...), hexEncode(inner)...)

	_, err = Decrypt(tampered, password)
	assert.ErrorIs(t, err, ErrAuth)
}

func TestDecrypt_BadEnvelope(t *testing.T) {
	good, err := Encrypt([]byte("payload"), password)
	require.NoError(t, err)
	body := string(good[strings.Index(string(good), "\n")+1:])

	tests := []struct {
		name string
		text string
	}{
		{"empty", ""},
		{"header only", "$ANSIBLE_VAULT;1.1;AES256\n"},
		{"not a vault", "token: value\n"},
		{"other cipher", "$ANSIBLE_VAULT;1.1;AES\n" + body},
		{"other version", "$ANSIBLE_VAULT;2.0;AES256\n" + body},
		{"not hex", "$ANSIBLE_VAULT;1.1;AES256\nzz" + body[2:]},
		{"truncated body", "$ANSIBLE_VAULT;1.1;AES256\n" + body[:40] + "\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decrypt([]byte(tt.text), password)
			assert.ErrorIs(t, err, ErrFormat)
		})
	}
}

func TestDecrypt_Version12WithVaultID(t *testing.T) {
	good, err := Encrypt([]byte("payload"), password)
	require.NoError(t, err)
	v12 := strings.Replace(string(good), "$ANSIBLE_VAULT;1.1;AES256", "$ANSIBLE_VAULT;1.2;AES256;prod", 1)

	got, err := Decrypt([]byte(v12), password)
	require.NoError(t, err)
	assert.Equal(t, "payload", string(got))
}

func TestPKCS7(t *testing.T) {
	for n := 0; n <= 33; n++ {
		data := bytes.Repeat([]byte{7}, n)
		padded := pkcs7Pad(data, 16)
		assert.Equal(t, 0, len(padded)%16)
		assert.Greater(t, len(padded), n, "padding is never empty")
		got, err := pkcs7Unpad(padded, 16)
		require.NoError(t, err)
		assert.Equal(t, data, got)
	}

	_, err := pkcs7Unpad([]byte{}, 16)
	assert.ErrorIs(t, err, ErrFormat)
	_, err = pkcs7Unpad(bytes.Repeat([]byte{17}, 16), 16)
	assert.ErrorIs(t, err, ErrFormat, "pad byte above the block size")
	_, err = pkcs7Unpad(append(bytes.Repeat([]byte{1}, 14), 2, 3), 16)
	assert.ErrorIs(t, err, ErrFormat, "inconsistent pad bytes")
}

// The interop tests run the real ansible-vault when it is installed, and are
// skipped otherwise. They are what proves the format, the unit tests above
// only prove the round trip.
func ansibleVault(t *testing.T) string {
	t.Helper()
	bin, err := exec.LookPath("ansible-vault")
	if err != nil {
		t.Skip("ansible-vault is not installed")
	}
	return bin
}

func runAnsibleVault(t *testing.T, bin string, args ...string) []byte {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), "ANSIBLE_NOCOLOR=1", "ANSIBLE_LOCAL_TEMP="+t.TempDir())
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	require.NoError(t, err, "ansible-vault %s: %s", strings.Join(args, " "), stderr.String())
	return out
}

func TestInterop_AnsibleDecryptsGo(t *testing.T) {
	bin := ansibleVault(t)
	dir := t.TempDir()

	plaintext := "token: !unsafe \"s3cr3t {{ 1+1 }}\"\ndb_password: !unsafe \"p@ss\"\n"
	vaulttext, err := Encrypt([]byte(plaintext), password)
	require.NoError(t, err)

	vaultFile := filepath.Join(dir, "secret_vars.yml")
	passFile := filepath.Join(dir, "password")
	require.NoError(t, os.WriteFile(vaultFile, vaulttext, 0o600))
	require.NoError(t, os.WriteFile(passFile, []byte(password+"\n"), 0o600))

	out := runAnsibleVault(t, bin, "decrypt", "--vault-password-file", passFile, "--output", "-", vaultFile)
	assert.Equal(t, plaintext, string(out))
}

func TestInterop_GoDecryptsAnsible(t *testing.T) {
	bin := ansibleVault(t)
	dir := t.TempDir()

	plaintext := "vaulted_var: from ansible-vault\n"
	plainFile := filepath.Join(dir, "plain.yml")
	passFile := filepath.Join(dir, "password")
	require.NoError(t, os.WriteFile(plainFile, []byte(plaintext), 0o600))
	require.NoError(t, os.WriteFile(passFile, []byte(password+"\n"), 0o600))

	vaulttext := runAnsibleVault(t, bin, "encrypt", "--vault-password-file", passFile, "--output", "-", plainFile)
	require.True(t, bytes.HasPrefix(vaulttext, []byte("$ANSIBLE_VAULT;1.1;AES256\n")), "%s", vaulttext)

	got, err := Decrypt(vaulttext, password)
	require.NoError(t, err)
	assert.Equal(t, plaintext, string(got))

	// With a vault id ansible-vault writes a 1.2 header.
	withID := runAnsibleVault(t, bin, "encrypt", "--vault-id", "prod@"+passFile, "--output", "-", plainFile)
	require.True(t, bytes.HasPrefix(withID, []byte("$ANSIBLE_VAULT;1.2;AES256;prod\n")), "%s", withID)
	got, err = Decrypt(withID, password)
	require.NoError(t, err)
	assert.Equal(t, plaintext, string(got))
}
