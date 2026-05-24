package k8s

import (
	"context"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func sshAccessKey(id int, passphrase string) db.AccessKey {
	return db.AccessKey{
		ID:   id,
		Type: db.AccessKeySSH,
		SshKey: db.SshKey{
			PrivateKey: "-----BEGIN OPENSSH PRIVATE KEY-----\nFAKE\n-----END OPENSSH PRIVATE KEY-----",
			Passphrase: passphrase,
		},
	}
}

func TestCollectSSHKeys_Deduplicates(t *testing.T) {
	shared := sshAccessKey(7, "")
	exec := New(newTestConfig(),
		db.Task{},
		db.Template{},
		db.Inventory{
			SSHKey:     shared, // same key used both as inventory user and template repo
			Repository: &db.Repository{SSHKey: sshAccessKey(99, "")},
		},
		db.Repository{SSHKey: shared},
		db.Environment{},
	)

	installs := exec.collectSSHKeys()

	require.Len(t, installs, 2, "duplicate IDs across slots collapse to one installation")
	ids := []int{installs[0].key.ID, installs[1].key.ID}
	assert.Contains(t, ids, 7)
	assert.Contains(t, ids, 99)
}

func TestCollectSSHKeys_SkipsNonSSHAndEmpty(t *testing.T) {
	exec := New(newTestConfig(),
		db.Task{},
		db.Template{},
		db.Inventory{
			// LoginPassword is not an SSH key, and an SSH-typed key with no private
			// key bytes is just an empty placeholder.
			SSHKey: db.AccessKey{ID: 1, Type: db.AccessKeyLoginPassword},
		},
		db.Repository{SSHKey: db.AccessKey{ID: 2, Type: db.AccessKeySSH, SshKey: db.SshKey{PrivateKey: ""}}},
		db.Environment{},
	)

	assert.Empty(t, exec.collectSSHKeys(), "non-SSH and empty keys must be filtered")
}

func TestPrepare_CreatesSSHSecretAndMountsIt(t *testing.T) {
	cfg := newTestConfig()
	key := sshAccessKey(42, "")
	exec := New(cfg,
		db.Task{ID: 5},
		db.Template{},
		db.Inventory{SSHKey: key},
		db.Repository{GitURL: "https://example.com/repo.git", GitBranch: "main"},
		db.Environment{},
	)

	require.NoError(t, exec.Prepare("", nil, ""))

	// Secret created with normalized PEM + correct labels.
	secret, err := cfg.Clientset.CoreV1().Secrets(cfg.Namespace).Get(context.Background(),
		sshSecretName(exec.podName, key.ID), metav1.GetOptions{})
	require.NoError(t, err)

	assert.Equal(t, "semaphore", secret.Labels[LabelRunner])
	assert.Equal(t, exec.podName, secret.Labels[labelPodName])

	require.Contains(t, secret.Data, sshSecretFilePrivateKey)
	assert.True(t, strings.HasSuffix(string(secret.Data[sshSecretFilePrivateKey]), "\n"),
		"PEM must end with newline so OpenSSH accepts it")
	assert.NotContains(t, secret.Data, sshSecretFilePassphrase,
		"passphrase-less key must not produce a passphrase file")

	// Pod mounts the Secret in the build container at the documented path.
	pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	require.NoError(t, err)

	var sawSecretVolume bool
	for _, v := range pod.Spec.Volumes {
		if v.Name == sshVolumeName(key.ID) {
			require.NotNil(t, v.Secret, "ssh volume must be Secret-backed")
			assert.Equal(t, sshSecretName(exec.podName, key.ID), v.Secret.SecretName)
			require.NotNil(t, v.Secret.DefaultMode)
			assert.Equal(t, int32(0o400), *v.Secret.DefaultMode, "private key files must be mode 0400")
			sawSecretVolume = true
		}
	}
	assert.True(t, sawSecretVolume, "Pod must declare a volume for each installed SSH key")

	var sawMount bool
	for _, m := range pod.Spec.Containers[0].VolumeMounts {
		if m.Name == sshVolumeName(key.ID) {
			assert.Equal(t, sshMountPath(key.ID), m.MountPath)
			assert.True(t, m.ReadOnly, "ssh mount is read-only")
			sawMount = true
		}
	}
	assert.True(t, sawMount, "build container must mount each SSH-key volume")
}

func TestPrepare_IncludesPassphraseSecretAndAskPassFlow(t *testing.T) {
	cfg := newTestConfig()
	key := sshAccessKey(11, "s3cret")
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{SSHKey: key},
		db.Repository{}, db.Environment{})

	require.NoError(t, exec.Prepare("", nil, ""))

	secret, err := cfg.Clientset.CoreV1().Secrets(cfg.Namespace).Get(context.Background(),
		sshSecretName(exec.podName, key.ID), metav1.GetOptions{})
	require.NoError(t, err)

	assert.Equal(t, []byte("s3cret"), secret.Data[sshSecretFilePassphrase])

	pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	require.NoError(t, err)

	script := pod.Spec.Containers[0].Args[0]
	assert.Contains(t, script, "SSH_ASKPASS",
		"passphrase-protected keys must use the SSH_ASKPASS helper flow")
	assert.Contains(t, script, sshMountPath(key.ID)+"/"+sshSecretFilePassphrase,
		"helper script must cat the mounted passphrase file")
}

func TestPrepare_SetsAnsibleHostKeyCheckingEnv(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{SSHKey: sshAccessKey(1, "")},
		db.Repository{}, db.Environment{})
	require.NoError(t, exec.Prepare("", nil, ""))

	pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	require.NoError(t, err)

	var found bool
	for _, env := range pod.Spec.Containers[0].Env {
		if env.Name == "ANSIBLE_HOST_KEY_CHECKING" {
			assert.Equal(t, "False", env.Value)
			found = true
		}
	}
	assert.True(t, found, "ANSIBLE_HOST_KEY_CHECKING=False must be set when SSH keys are installed")
}

func TestPrepare_NoSSHKeys_NoEnvNoSecret(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 1}, db.Template{}, db.Inventory{}, db.Repository{}, db.Environment{})
	require.NoError(t, exec.Prepare("", nil, ""))

	pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	require.NoError(t, err)

	// No SSH-keyed volumes, no env var, no secrets list.
	for _, v := range pod.Spec.Volumes {
		assert.Nil(t, v.Secret, "no SSH keys means no Secret-backed volumes")
	}
	assert.Empty(t, pod.Spec.Containers[0].Env, "env vars are added only when SSH keys present")
	assert.Empty(t, exec.secretNames)
}

func TestCleanup_DeletesAllSSHSecrets(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 1}, db.Template{},
		db.Inventory{SSHKey: sshAccessKey(7, "")},
		db.Repository{SSHKey: sshAccessKey(8, "phrase")},
		db.Environment{})

	require.NoError(t, exec.Prepare("", nil, ""))
	require.Len(t, exec.secretNames, 2)

	exec.Cleanup()

	for _, name := range []string{sshSecretName(exec.podName, 7), sshSecretName(exec.podName, 8)} {
		_, err := cfg.Clientset.CoreV1().Secrets(cfg.Namespace).Get(context.Background(), name, metav1.GetOptions{})
		assert.True(t, apierrors.IsNotFound(err), "Cleanup must delete secret %s, got err=%v", name, err)
	}
}

func TestSSHAgentSetupScript_LoadsEachKey(t *testing.T) {
	installs := []sshKeyInstallation{
		{key: sshAccessKey(1, ""), origin: "repository"},
		{key: sshAccessKey(2, "pw"), origin: "inventory"},
	}

	script := sshAgentSetupScript(installs)

	assert.Contains(t, script, "eval \"$(ssh-agent -s)\"")
	assert.Contains(t, script, "/secrets/ssh/1/id_rsa")
	assert.Contains(t, script, "/secrets/ssh/2/id_rsa")
	assert.Contains(t, script, "SSH_ASKPASS", "passphrase-protected key must use SSH_ASKPASS")
	// No passphrase on key 1, so the askpass helper must NOT be invoked for it.
	assert.Contains(t, script, "ssh-add /secrets/ssh/1/id_rsa")
}

func TestSSHAgentSetupScript_EmptyInstallsReturnsEmpty(t *testing.T) {
	assert.Empty(t, sshAgentSetupScript(nil),
		"no installs → empty snippet so plain images without ssh-agent still work")
}

func TestSSHAgentSetupScript_NoHeredocSoIndentingIsSafe(t *testing.T) {
	// Regression: when nested inside `if command -v ssh-agent…; then … fi`, the
	// snippet is indented two spaces. A heredoc terminator pushed off column 0
	// would never match — busybox sh would swallow the rest of the script and
	// fail with "Syntax error: end of file unexpected (expecting fi)".
	installs := []sshKeyInstallation{
		{key: sshAccessKey(11, "s3cret"), origin: "repository"},
	}
	script := sshAgentSetupScript(installs)

	assert.NotContains(t, script, "<<", "ssh-agent snippet must be heredoc-free so it can be safely indented")
}
