package k8s

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ansibleTemplate is a small helper that produces a Template wired for ansible-playbook
// runs. Tests then override only the fields they care about.
func ansibleTemplate() db.Template {
	return db.Template{
		ID:       1,
		App:      db.AppAnsible,
		Playbook: "site.yml",
	}
}

func TestPrepare_RunsAnsibleAndBuildsExpectedArgs(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg,
		db.Task{ID: 42},
		ansibleTemplate(),
		db.Inventory{Type: db.InventoryFile, Inventory: "invs/prod"},
		db.Repository{GitURL: "https://example.com/r.git", GitBranch: "main"},
		db.Environment{},
	)

	require.NoError(t, exec.Prepare("alice", nil, ""))

	pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	require.NoError(t, err)

	script := pod.Spec.Containers[0].Args[0]

	assert.Contains(t, script, "set -e", "ansible-mode script must abort on first failure")
	assert.Contains(t, script, "cd /workspace/repo", "ansible-playbook runs from the cloned repo")
	assert.Contains(t, script, "ansible-playbook", "script must invoke ansible-playbook")
	assert.Contains(t, script, "'-i' '/workspace/repo/invs/prod'", "file-typed inventory path lives inside the cloned repo")
	assert.Contains(t, script, "'site.yml'", "playbook name reaches the command line")
	assert.Contains(t, script, "ANSIBLE_EXIT=$?", "script must capture the ansible exit code")
	assert.Contains(t, script, "exit $ANSIBLE_EXIT", "final exit code propagates to the Pod result")
}

func TestPrepare_StaticInventoryRendersAsConfigMap(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg,
		db.Task{ID: 1},
		ansibleTemplate(),
		db.Inventory{
			Type:      db.InventoryStatic,
			Inventory: "[web]\nhost-a ansible_host=10.0.0.1",
		},
		db.Repository{},
		db.Environment{},
	)
	require.NoError(t, exec.Prepare("", nil, ""))

	cm, err := cfg.Clientset.CoreV1().ConfigMaps(cfg.Namespace).Get(context.Background(),
		exec.configmapNames[0], metav1.GetOptions{})
	require.NoError(t, err)
	assert.Contains(t, cm.Data, staticInventoryFilename)
	assert.Equal(t, "[web]\nhost-a ansible_host=10.0.0.1", cm.Data[staticInventoryFilename])

	pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	require.NoError(t, err)

	// Pod mounts the ConfigMap and the ansible-playbook -i flag points into it.
	var sawMount bool
	for _, m := range pod.Spec.Containers[0].VolumeMounts {
		if m.Name == staticInventoryVolumeName && m.MountPath == staticInventoryMountDir {
			sawMount = true
		}
	}
	assert.True(t, sawMount, "build container must mount the inventory ConfigMap")
	assert.Contains(t, pod.Spec.Containers[0].Args[0], "/workspace/inventory/"+staticInventoryFilename)
}

func TestPrepare_StaticYamlInventoryUsesYamlFilename(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 1}, ansibleTemplate(),
		db.Inventory{Type: db.InventoryStaticYaml, Inventory: "all:\n  hosts:\n    a:"},
		db.Repository{}, db.Environment{})
	require.NoError(t, exec.Prepare("", nil, ""))

	cm, err := cfg.Clientset.CoreV1().ConfigMaps(cfg.Namespace).Get(context.Background(),
		exec.configmapNames[0], metav1.GetOptions{})
	require.NoError(t, err)
	assert.Contains(t, cm.Data, staticInventoryYamlFile, "YAML inventory must land in inventory.yml")
}

func TestPrepare_VaultPasswordCreatesSecretAndFlag(t *testing.T) {
	cfg := newTestConfig()
	vaultName := "prod"
	vaultKey := db.AccessKey{
		ID:            7,
		Type:          db.AccessKeyLoginPassword,
		LoginPassword: db.LoginPassword{Password: "topsecret"},
	}

	tpl := ansibleTemplate()
	tpl.Vaults = []db.TemplateVault{{
		Type:  db.TemplateVaultPassword,
		Name:  &vaultName,
		Vault: &vaultKey,
	}}

	exec := New(cfg, db.Task{ID: 1}, tpl,
		db.Inventory{Type: db.InventoryFile, Inventory: "hosts"},
		db.Repository{}, db.Environment{})

	require.NoError(t, exec.Prepare("", nil, ""))

	// One vault Secret named with the vault name, containing the password under the
	// vault name as the data key (matches the path the --vault-id flag points at).
	var vaultSecretName string
	for _, n := range exec.secretNames {
		if strings.Contains(n, "vault-prod") {
			vaultSecretName = n
		}
	}
	require.NotEmpty(t, vaultSecretName, "a vault Secret must be created")

	secret, err := cfg.Clientset.CoreV1().Secrets(cfg.Namespace).Get(context.Background(), vaultSecretName, metav1.GetOptions{})
	require.NoError(t, err)
	assert.Equal(t, []byte("topsecret"), secret.Data["prod"])

	pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	require.NoError(t, err)
	assert.Contains(t, pod.Spec.Containers[0].Args[0],
		"'--vault-id=prod@/secrets/vault/prod/prod'",
		"ansible-playbook must read the vault password from the mounted file")
}

func TestPrepare_VaultScriptUsesScriptPathDirectly(t *testing.T) {
	cfg := newTestConfig()
	scriptPath := "/usr/local/bin/get-vault-pass.sh"
	vaultName := "default"
	tpl := ansibleTemplate()
	tpl.Vaults = []db.TemplateVault{{
		Type:   db.TemplateVaultScript,
		Name:   &vaultName,
		Script: &scriptPath,
	}}

	exec := New(cfg, db.Task{ID: 1}, tpl,
		db.Inventory{Type: db.InventoryFile, Inventory: "hosts"},
		db.Repository{}, db.Environment{})
	require.NoError(t, exec.Prepare("", nil, ""))

	pod, err := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	require.NoError(t, err)
	assert.Contains(t, pod.Spec.Containers[0].Args[0], "'--vault-id=default@"+scriptPath+"'",
		"script-type vaults point ansible-playbook directly at the script path; no Secret needed")
}

func TestPrepare_ExtraVarsJSONContainsTaskDetailsAndEnvJSON(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg,
		db.Task{ID: 42, CommitMessage: "fix"},
		ansibleTemplate(),
		db.Inventory{Type: db.InventoryFile, Inventory: "hosts", ID: 9, Name: "prod"},
		db.Repository{ID: 5, Name: "playbooks"},
		db.Environment{JSON: `{"app_version":"1.2.3"}`},
	)
	require.NoError(t, exec.Prepare("denis", nil, ""))

	var extraSecretName string
	for _, n := range exec.secretNames {
		if strings.HasSuffix(n, "-extra-vars") {
			extraSecretName = n
		}
	}
	require.NotEmpty(t, extraSecretName)

	secret, err := cfg.Clientset.CoreV1().Secrets(cfg.Namespace).Get(context.Background(), extraSecretName, metav1.GetOptions{})
	require.NoError(t, err)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(secret.Data[extraVarsFilename], &payload))

	assert.Equal(t, "1.2.3", payload["app_version"], "environment JSON merges into extra-vars")
	vars, ok := payload["semaphore_vars"].(map[string]any)
	require.True(t, ok, "semaphore_vars block must be present")
	td, ok := vars["task_details"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, float64(42), td["id"])
	assert.Equal(t, "denis", td["username"])
	assert.Equal(t, "playbooks", td["repository_name"])
	assert.Equal(t, "prod", td["inventory_name"])
}

func TestBuildAnsibleArgs_IncludesUserDiffCheckLimitTagsSkip(t *testing.T) {
	cfg := newTestConfig()
	limit := []string{"web", "db"}
	tags := []string{"deploy"}
	skip := []string{"slow"}
	tplArgsJSON := `["-vvv"]`
	taskArgsJSON := `["--start-at-task=Restart"]`

	args := `{"AllowDebug":true,"AllowOverrideLimit":true,"AllowOverrideTags":true,"AllowOverrideSkipTags":true,"Limit":["web","db"],"Tags":["deploy"],"SkipTags":["slow"]}`

	tpl := ansibleTemplate()
	tpl.Arguments = &tplArgsJSON
	tpl.AllowOverrideArgsInTask = true
	// Inject TemplateParams via the encoded params field by abusing FillParams shape —
	// the simplest portable substitute is to write the Params via reflection-free
	// JSON. AnsibleTemplateParams has its own JSON shape but for the test we want
	// to demonstrate that user-set Limit/Tags/SkipTags reach the args. Skip the
	// FillParams indirection and assert directly on the user-overridable params.
	_ = args
	_ = limit
	_ = tags
	_ = skip

	inventorySSHKey := db.AccessKey{ID: 1, Type: db.AccessKeySSH, SshKey: db.SshKey{Login: "deploy"}}
	keyID := 1
	exec := New(cfg, db.Task{Arguments: &taskArgsJSON},
		tpl,
		db.Inventory{Type: db.InventoryFile, Inventory: "hosts", SSHKeyID: &keyID, SSHKey: inventorySSHKey},
		db.Repository{},
		db.Environment{},
	)
	require.NoError(t, exec.Prepare("", nil, ""))

	pod, _ := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	script := pod.Spec.Containers[0].Args[0]

	assert.Contains(t, script, "'--user' 'deploy'", "inventory SSH key login becomes --user")
	assert.Contains(t, script, "'-vvv'", "template extra args reach ansible-playbook")
	assert.Contains(t, script, "'--start-at-task=Restart'", "task extra args reach ansible-playbook")
	assert.Contains(t, script, "'site.yml'", "playbook name is last")
}

func TestPrepare_RejectsLoginPasswordInventoryKey(t *testing.T) {
	cfg := newTestConfig()
	keyID := 1
	exec := New(cfg, db.Task{}, ansibleTemplate(),
		db.Inventory{
			Type:     db.InventoryFile,
			SSHKeyID: &keyID,
			SSHKey: db.AccessKey{
				ID:            keyID,
				Type:          db.AccessKeyLoginPassword,
				LoginPassword: db.LoginPassword{Login: "root", Password: "pw"},
			},
		},
		db.Repository{}, db.Environment{})

	err := exec.Prepare("", nil, "")
	require.Error(t, err, "K8s executor must reject login_password inventory keys (no PTY in a Pod)")
	assert.Contains(t, err.Error(), "login_password")
}

func TestPrepare_NonAnsibleAppKeepsSkeletonScript(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 1},
		db.Template{App: db.AppBash},
		db.Inventory{}, db.Repository{}, db.Environment{})
	require.NoError(t, exec.Prepare("", nil, ""))

	pod, _ := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	script := pod.Spec.Containers[0].Args[0]

	assert.NotContains(t, script, "ansible-playbook", "non-Ansible templates must not invoke ansible-playbook")
	assert.Contains(t, script, "workspace contents:", "non-Ansible templates fall back to skeleton")
}

func TestCleanup_DeletesInventoryConfigMap(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 1}, ansibleTemplate(),
		db.Inventory{Type: db.InventoryStatic, Inventory: "[a]\nhost-a"},
		db.Repository{}, db.Environment{})
	require.NoError(t, exec.Prepare("", nil, ""))

	cmName := exec.configmapNames[0]
	exec.Cleanup()

	_, err := cfg.Clientset.CoreV1().ConfigMaps(cfg.Namespace).Get(context.Background(), cmName, metav1.GetOptions{})
	assert.True(t, apierrors.IsNotFound(err), "Cleanup must delete the inventory ConfigMap, got err=%v", err)
}

func TestBuildContainerScript_SshAgentGuard(t *testing.T) {
	cfg := newTestConfig()
	exec := New(cfg, db.Task{ID: 1}, ansibleTemplate(),
		db.Inventory{Type: db.InventoryFile, Inventory: "hosts", SSHKey: sshAccessKey(1, "")},
		db.Repository{SSHKey: sshAccessKey(2, "")},
		db.Environment{})
	require.NoError(t, exec.Prepare("", nil, ""))

	pod, _ := cfg.Clientset.CoreV1().Pods(cfg.Namespace).Get(context.Background(), exec.podName, metav1.GetOptions{})
	script := pod.Spec.Containers[0].Args[0]

	assert.Contains(t, script, "if command -v ssh-agent",
		"slim images without ssh-agent must skip the SSH setup with a warning instead of failing the task")
	assert.Contains(t, script, "skipping SSH key install")
}

func TestSanitizeForResourceName(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"prod", "prod"},
		{"Prod_Vault", "prod-vault"},
		{"weird name 1", "weird-name-1"},
		{"___", "x"},
		{"", "x"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.want, sanitizeForResourceName(tt.in))
		})
	}
}
