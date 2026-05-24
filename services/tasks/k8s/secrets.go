package k8s

import (
	"fmt"
	"strings"

	"github.com/semaphoreui/semaphore/db"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Files inside each SSH-key Secret. Match the spec (section 6.1): the private key is
// always named "id_rsa" regardless of algorithm — the OpenSSH client doesn't care
// about filename, only file contents. The passphrase, when present, lives alongside.
const (
	sshSecretFilePrivateKey = "id_rsa"
	sshSecretFilePassphrase = "passphrase"

	// sshSecretsMountRoot is where SSH-key Secrets are mounted in the build container.
	// Each key lives in a subdirectory named after its AccessKey ID so the build script
	// can address them deterministically: /secrets/ssh/<id>/id_rsa.
	sshSecretsMountRoot = "/secrets/ssh"

	// sshSecretFileMode is the default mode the kubelet applies to mounted Secret
	// files. 0400 prevents the build container from accidentally exposing keys to
	// child processes that drop their UID. Encoded as octal int32 for K8s API.
	sshSecretFileMode int32 = 0o400
)

// sshKeyInstallation captures one SSH AccessKey we plan to expose inside the Pod. The
// origin field (repository / inventory / inventory-repository) is preserved purely so
// the build container's logs name where each key came from, which helps users debug
// permission-denied errors against multiple targets.
type sshKeyInstallation struct {
	key    db.AccessKey
	origin string
}

// collectSSHKeys walks the task's data and returns the unique set of SSH access keys
// that need to land in the Pod as K8s Secrets. Repeated AccessKey IDs are de-duplicated:
// if the same key is used for both the playbook repository and the inventory user,
// one Secret is sufficient. Non-SSH and empty keys are filtered out.
func (e *Executor) collectSSHKeys() []sshKeyInstallation {
	seen := make(map[int]bool)
	var out []sshKeyInstallation

	add := func(key db.AccessKey, origin string) {
		if key.Type != db.AccessKeySSH {
			return
		}
		if key.SshKey.PrivateKey == "" {
			return
		}
		if seen[key.ID] {
			return
		}
		seen[key.ID] = true
		out = append(out, sshKeyInstallation{key: key, origin: origin})
	}

	add(e.Repository.SSHKey, "repository")
	add(e.Inventory.SSHKey, "inventory")
	if e.Inventory.Repository != nil {
		add(e.Inventory.Repository.SSHKey, "inventory_repository")
	}

	return out
}

// sshSecretName produces the K8s Secret name for an installed key. Format binds the
// Secret to its owning Pod so orphan cleanup (Phase 5+) can list-by-prefix and so
// concurrent tasks on the same runner never collide.
func sshSecretName(podName string, keyID int) string {
	return fmt.Sprintf("%s-ssh-%d", podName, keyID)
}

// sshVolumeName names the Pod volume backed by an SSH-key Secret. Must be DNS-1123
// (lowercase, no underscores) — K8s rejects everything else.
func sshVolumeName(keyID int) string {
	return fmt.Sprintf("ssh-key-%d", keyID)
}

// sshMountPath returns the directory inside the build container where the Secret for
// this key is mounted (e.g. /secrets/ssh/42). The private key file is always
// "<path>/id_rsa".
func sshMountPath(keyID int) string {
	return fmt.Sprintf("%s/%d", sshSecretsMountRoot, keyID)
}

// buildSSHSecret renders the K8s Secret object that holds one private key (and
// optionally its passphrase). Labels mirror those on the owning Pod so that the
// Phase-5 orphan-cleanup pass — which sweeps stale resources after a runner
// restart — can list them via the same selector used for Pods.
func buildSSHSecret(podName, namespace string, install sshKeyInstallation) *corev1.Secret {
	data := map[string][]byte{
		sshSecretFilePrivateKey: normalizePrivateKey(install.key.SshKey.PrivateKey),
	}
	if pass := install.key.SshKey.Passphrase; pass != "" {
		data[sshSecretFilePassphrase] = []byte(pass)
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sshSecretName(podName, install.key.ID),
			Namespace: namespace,
			Labels: map[string]string{
				LabelRunner: "semaphore",
				// Pod-name label is what cleanup pivots on: when the Pod goes away,
				// every Secret carrying its name is fair game to delete.
				labelPodName: podName,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: data,
	}
}

// normalizePrivateKey guarantees the PEM payload ends with a newline. OpenSSH refuses
// to load a key whose final line lacks a terminator, and the form Semaphore stores
// in the DB has historically been newline-stripped for some import paths.
func normalizePrivateKey(pem string) []byte {
	if !strings.HasSuffix(pem, "\n") {
		pem += "\n"
	}
	return []byte(pem)
}

// sshAgentSetupScript renders the shell snippet that the build container runs before
// the user's command. It boots an ssh-agent, then for each installed key runs
// ssh-add — using SSH_ASKPASS for passphrase-protected keys, since ssh-add does not
// reliably accept passphrases on stdin across OpenSSH versions.
//
// Returns an empty string when there are no keys to install; the caller checks for
// this and skips the snippet entirely so the build script stays portable to images
// that don't ship ssh-agent (e.g. plain alpine, used by the skeleton).
func sshAgentSetupScript(installs []sshKeyInstallation) string {
	if len(installs) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("# Phase 4: bring up ssh-agent and load installed SSH keys.\n")
	b.WriteString("eval \"$(ssh-agent -s)\"\n")
	for _, inst := range installs {
		mount := sshMountPath(inst.key.ID)
		fmt.Fprintf(&b, "echo 'k8s: loading SSH key %d (%s)'\n", inst.key.ID, inst.origin)

		if inst.key.SshKey.Passphrase == "" {
			fmt.Fprintf(&b, "ssh-add %s/%s\n", mount, sshSecretFilePrivateKey)
			continue
		}

		// SSH_ASKPASS pattern: write a tiny helper that prints the passphrase, then
		// invoke ssh-add via `setsid -w` so OpenSSH detaches from the controlling
		// terminal and consults the helper. SSH_ASKPASS_REQUIRE=force is required on
		// OpenSSH ≥8.4 to bypass the "no DISPLAY" guard.
		askpass := fmt.Sprintf("/tmp/askpass-%d.sh", inst.key.ID)
		fmt.Fprintf(&b, "cat > %s <<'__SEMASKPASS__'\n", askpass)
		b.WriteString("#!/bin/sh\n")
		fmt.Fprintf(&b, "cat %s/%s\n", mount, sshSecretFilePassphrase)
		b.WriteString("__SEMASKPASS__\n")
		fmt.Fprintf(&b, "chmod 700 %s\n", askpass)
		fmt.Fprintf(&b, "DISPLAY=:0 SSH_ASKPASS=%s SSH_ASKPASS_REQUIRE=force setsid -w ssh-add %s/%s </dev/null\n",
			askpass, mount, sshSecretFilePrivateKey)
		fmt.Fprintf(&b, "rm -f %s\n", askpass)
	}
	return b.String()
}

// sshSecretVolumes builds the Pod-level Volume entries for each installed key. The
// Secret name on each Volume must match what buildSSHSecret produced.
func sshSecretVolumes(podName string, installs []sshKeyInstallation) []corev1.Volume {
	if len(installs) == 0 {
		return nil
	}
	out := make([]corev1.Volume, 0, len(installs))
	mode := sshSecretFileMode
	for _, inst := range installs {
		out = append(out, corev1.Volume{
			Name: sshVolumeName(inst.key.ID),
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  sshSecretName(podName, inst.key.ID),
					DefaultMode: &mode,
				},
			},
		})
	}
	return out
}

// sshSecretVolumeMounts builds the build-container VolumeMount entries for each key.
func sshSecretVolumeMounts(installs []sshKeyInstallation) []corev1.VolumeMount {
	if len(installs) == 0 {
		return nil
	}
	out := make([]corev1.VolumeMount, 0, len(installs))
	for _, inst := range installs {
		out = append(out, corev1.VolumeMount{
			Name:      sshVolumeName(inst.key.ID),
			MountPath: sshMountPath(inst.key.ID),
			ReadOnly:  true,
		})
	}
	return out
}

// labelPodName is set on every per-task Secret so Cleanup can sweep them after the
// Pod is gone. Lives here (not in executor.go) because Pod and Secret resources both
// reference it and the constant is otherwise nowhere natural to put.
const labelPodName = "semaphoreui.com/pod-name"
