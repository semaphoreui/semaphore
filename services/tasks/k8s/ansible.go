package k8s

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/semaphoreui/semaphore/db"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Inside-container paths Ansible task assets are mounted at. Kept as constants because
// the same paths are referenced by both the Pod spec (mountPath) and the ansible-playbook
// CLI args (-i, --vault-id=...@FILE, --extra-vars @FILE) — drift between them is a
// silent class of bugs that's hard to catch otherwise.
const (
	vaultsMountRoot           = "/secrets/vault"
	staticInventoryMountDir   = "/workspace/inventory"
	staticInventoryFilename   = "inventory.cfg"
	staticInventoryYamlFile   = "inventory.yml"
	extraVarsMountDir         = "/workspace/extra-vars"
	extraVarsFilename         = "values.json"
	extraVarsVolumeName       = "extra-vars"
	staticInventoryVolumeName = "inventory"
)

// ansiblePrep captures everything we computed for an Ansible task: the K8s resources to
// create alongside the Pod, the Volume / VolumeMount entries to wire them into the
// build container, and the ansible-playbook CLI arguments the build script will invoke.
//
// It carries unwritten resources (Secrets, ConfigMaps) — Prepare turns it into actual
// API calls. Splitting "what to build" from "send to API" keeps the unit tests fast
// (they inspect the prep value directly) and lets a future dry-run mode print the
// plan without touching the cluster.
type ansiblePrep struct {
	// secrets / configMap are the cluster resources backing this task. Names are
	// generated up-front so the Pod spec can reference them.
	secrets   []*corev1.Secret
	configMap *corev1.ConfigMap

	volumes      []corev1.Volume
	volumeMounts []corev1.VolumeMount

	// playbookArgs is the full argv passed to ansible-playbook, in the order the
	// build script writes them. Last element is always the playbook path.
	playbookArgs []string
}

// buildAnsiblePrep computes the ansiblePrep for an Ansible task. Errors here surface
// to the caller before any K8s resources are created, so a half-finished plan never
// leaves orphans in the cluster.
//
// Returns nil-prep for non-Ansible templates so the caller can fall back to skeleton
// behaviour. Non-supported feature combinations (e.g. LoginPassword inventory key —
// no PTY in a Pod) return an explicit error rather than silently degrading.
func (e *Executor) buildAnsiblePrep(username string, incomingVersion *string) (*ansiblePrep, error) {
	if e.Template.App != db.AppAnsible {
		return nil, nil
	}

	if err := e.checkUnsupportedInventoryAuth(); err != nil {
		return nil, err
	}

	prep := &ansiblePrep{}

	repoPath := workspaceMountPath + "/" + workspaceRepoSubpath

	inventoryPath, err := e.addInventory(prep, repoPath)
	if err != nil {
		return nil, err
	}

	e.addVaults(prep)

	if err = e.addExtraVars(prep, username, incomingVersion); err != nil {
		return nil, err
	}

	prep.playbookArgs = e.buildAnsibleArgs(inventoryPath, prep)

	return prep, nil
}

// checkUnsupportedInventoryAuth rejects auth modes the K8s executor cannot yet
// handle. They all share the same root cause: ansible's --ask-pass / --ask-become-pass
// flags need an interactive prompt, but Pods don't have a controlling terminal. The
// LocalExecutor solves this by piping passwords through a PTY; doing the equivalent
// in K8s requires the attach-stream wiring scheduled for a later phase.
func (e *Executor) checkUnsupportedInventoryAuth() error {
	if e.Inventory.SSHKeyID != nil && e.Inventory.SSHKey.Type == db.AccessKeyLoginPassword {
		return fmt.Errorf("k8s executor: inventory login_password authentication is not supported (use an SSH key)")
	}
	if e.Inventory.BecomeKeyID != nil && e.Inventory.BecomeKey.Type == db.AccessKeyLoginPassword {
		return fmt.Errorf("k8s executor: inventory become login_password authentication is not supported")
	}
	return nil
}

// addInventory adds whatever K8s objects/mounts are needed to expose the task inventory
// inside the build container and returns the absolute path ansible-playbook should
// receive as `-i`. File-typed inventories live in the cloned repo and need no extra
// resources; static inventories are rendered as a ConfigMap.
func (e *Executor) addInventory(prep *ansiblePrep, repoPath string) (string, error) {
	switch e.Inventory.Type {
	case db.InventoryFile:
		if e.Inventory.GetFilename() == "" {
			return "", fmt.Errorf("k8s executor: inventory file path is empty")
		}
		return path.Join(repoPath, e.Inventory.GetFilename()), nil

	case db.InventoryStatic, db.InventoryStaticYaml:
		filename := staticInventoryFilename
		if e.Inventory.Type == db.InventoryStaticYaml {
			filename = staticInventoryYamlFile
		}

		cmName := fmt.Sprintf("%s-inventory", e.podName)
		prep.configMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cmName,
				Namespace: e.Config.Namespace,
				Labels: map[string]string{
					LabelRunner:  "semaphore",
					labelPodName: e.podName,
				},
			},
			Data: map[string]string{filename: e.Inventory.Inventory},
		}

		prep.volumes = append(prep.volumes, corev1.Volume{
			Name: staticInventoryVolumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: cmName},
				},
			},
		})
		prep.volumeMounts = append(prep.volumeMounts, corev1.VolumeMount{
			Name:      staticInventoryVolumeName,
			MountPath: staticInventoryMountDir,
			ReadOnly:  true,
		})
		return path.Join(staticInventoryMountDir, filename), nil

	default:
		// Terraform/Tofu/Terragrunt workspace inventories are not in Phase 5 scope.
		return "", fmt.Errorf("k8s executor: inventory type %q is not supported", e.Inventory.Type)
	}
}

// addVaults writes one Secret per password-typed Template.Vaults entry and mounts it at
// /secrets/vault/<vault_name>. Script-typed vaults need no Secret — the script path is
// referenced directly in the --vault-id flag.
func (e *Executor) addVaults(prep *ansiblePrep) {
	for _, vault := range e.Template.Vaults {
		name := vaultName(vault)
		if vault.Type != db.TemplateVaultPassword {
			continue
		}
		if vault.Vault == nil {
			continue
		}
		password := vaultPassword(*vault.Vault)
		if password == "" {
			continue
		}

		secretName := fmt.Sprintf("%s-vault-%s", e.podName, sanitizeForResourceName(name))
		prep.secrets = append(prep.secrets, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: e.Config.Namespace,
				Labels: map[string]string{
					LabelRunner:  "semaphore",
					labelPodName: e.podName,
				},
			},
			Type: corev1.SecretTypeOpaque,
			Data: map[string][]byte{name: []byte(password)},
		})

		volName := fmt.Sprintf("vault-%s", sanitizeForResourceName(name))
		mode := sshSecretFileMode
		prep.volumes = append(prep.volumes, corev1.Volume{
			Name: volName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  secretName,
					DefaultMode: &mode,
				},
			},
		})
		prep.volumeMounts = append(prep.volumeMounts, corev1.VolumeMount{
			Name:      volName,
			MountPath: path.Join(vaultsMountRoot, name),
			ReadOnly:  true,
			SubPath:   "", // mount whole secret directory; --vault-id consumes <dir>/<name>
		})
	}
}

// vaultName returns the canonical name for a Template.Vaults entry. Defaults to
// "default" when the user left it blank — mirrors LocalExecutor's installVaultKeyFiles.
func vaultName(v db.TemplateVault) string {
	if v.Name != nil && *v.Name != "" {
		return *v.Name
	}
	return "default"
}

// vaultPassword extracts the password material from an AccessKey of any vault-eligible
// shape. The Semaphore data model is forgiving here: vaults can be stored as ssh-typed,
// string-typed, or login_password-typed access keys depending on how the user uploaded
// them.
func vaultPassword(key db.AccessKey) string {
	switch key.Type {
	case db.AccessKeyString:
		return key.String
	case db.AccessKeyLoginPassword:
		return key.LoginPassword.Password
	case db.AccessKeySSH:
		// Vaults masquerading as ssh keys put the password in PrivateKey for legacy reasons.
		return key.SshKey.PrivateKey
	}
	return ""
}

// addExtraVars renders the `semaphore_vars` / EnvironmentSecretVar JSON into a Secret
// mounted at /workspace/extra-vars/values.json. Encoded as Secret (not ConfigMap)
// because env-secrets may contain credentials.
func (e *Executor) addExtraVars(prep *ansiblePrep, username string, incomingVersion *string) error {
	body, err := e.buildExtraVarsJSON(username, incomingVersion)
	if err != nil {
		return fmt.Errorf("k8s executor: build extra-vars JSON: %w", err)
	}

	secretName := fmt.Sprintf("%s-extra-vars", e.podName)
	prep.secrets = append(prep.secrets, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: e.Config.Namespace,
			Labels: map[string]string{
				LabelRunner:  "semaphore",
				labelPodName: e.podName,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{extraVarsFilename: body},
	})

	mode := sshSecretFileMode
	prep.volumes = append(prep.volumes, corev1.Volume{
		Name: extraVarsVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName:  secretName,
				DefaultMode: &mode,
			},
		},
	})
	prep.volumeMounts = append(prep.volumeMounts, corev1.VolumeMount{
		Name:      extraVarsVolumeName,
		MountPath: extraVarsMountDir,
		ReadOnly:  true,
	})
	return nil
}

// buildExtraVarsJSON mirrors LocalExecutor.getEnvironmentExtraVarsJSON: it merges
// Environment.JSON with any survey-secret JSON (not propagated through the runner
// protocol today, so always empty here) and inlines semaphore_vars.task_details.
func (e *Executor) buildExtraVarsJSON(username string, incomingVersion *string) ([]byte, error) {
	extraVars := make(map[string]any)
	if e.Environment.JSON != "" {
		if err := json.Unmarshal([]byte(e.Environment.JSON), &extraVars); err != nil {
			return nil, err
		}
	}

	extraVars["semaphore_vars"] = map[string]any{
		"task_details": e.taskDetails(username, incomingVersion),
	}

	return json.Marshal(extraVars)
}

// taskDetails replicates LocalExecutor.getTaskDetails verbatim so playbooks consuming
// semaphore_vars see the same shape regardless of which executor ran them.
func (e *Executor) taskDetails(username string, incomingVersion *string) map[string]any {
	details := map[string]any{
		"id":              e.Task.ID,
		"username":        username,
		"url":             e.Task.GetUrl(),
		"commit_hash":     e.Task.CommitHash,
		"commit_message":  e.Task.CommitMessage,
		"inventory_name":  e.Inventory.Name,
		"inventory_id":    e.Inventory.ID,
		"repository_name": e.Repository.Name,
		"repository_id":   e.Repository.ID,
	}
	if e.Task.Message != "" {
		details["message"] = e.Task.Message
	}
	if e.Template.Type != db.TemplateTask {
		details["type"] = e.Template.Type
		if incomingVersion != nil {
			details["incoming_version"] = incomingVersion
		}
		if e.Template.Type == db.TemplateBuild {
			details["target_version"] = e.Task.Version
		}
	}
	return details
}

// buildAnsibleArgs returns the argv ansible-playbook is invoked with. Order matches
// LocalExecutor.getPlaybookArgs so the two paths produce visually identical command
// lines (modulo --ask-pass which K8s never emits).
func (e *Executor) buildAnsibleArgs(inventoryPath string, prep *ansiblePrep) []string {
	args := []string{"-i", inventoryPath}

	// User from inventory SSH key (when it's an SSH key with a login set).
	if e.Inventory.SSHKeyID != nil && e.Inventory.SSHKey.Type == db.AccessKeySSH {
		if login := e.Inventory.SSHKey.SshKey.Login; login != "" {
			args = append(args, "--user", login)
		}
	}

	var tplParams db.AnsibleTemplateParams
	_ = e.Template.FillParams(&tplParams) // ignore: params are best-effort, defaults already populated
	var params db.AnsibleTaskParams
	_ = e.Task.ExtractParams(&params)

	if tplParams.AllowDebug && params.Debug {
		level := params.DebugLevel
		if level < 1 {
			level = 4
		}
		if level > 6 {
			level = 6
		}
		args = append(args, "-"+strings.Repeat("v", level))
	}
	if params.Diff {
		args = append(args, "--diff")
	}
	if params.DryRun {
		args = append(args, "--check")
	}

	for _, vault := range e.Template.Vaults {
		name := vaultName(vault)
		switch vault.Type {
		case db.TemplateVaultPassword:
			if vault.Vault != nil && vaultPassword(*vault.Vault) != "" {
				args = append(args, fmt.Sprintf("--vault-id=%s@%s", name, path.Join(vaultsMountRoot, name, name)))
			}
		case db.TemplateVaultScript:
			if vault.Script != nil && *vault.Script != "" {
				args = append(args, fmt.Sprintf("--vault-id=%s@%s", name, *vault.Script))
			}
		}
	}

	// Always pass extra-vars file: even when Environment.JSON is empty, semaphore_vars
	// must reach the playbook so users can rely on task_details.
	args = append(args, "--extra-vars", "@"+path.Join(extraVarsMountDir, extraVarsFilename))

	// EnvironmentSecretVar entries become individual --extra-vars name=value pairs.
	// These ARE secret values; passing on the command line is consistent with the
	// LocalExecutor behaviour, but a future hardening could move them into the JSON
	// file too.
	for _, secret := range e.Environment.Secrets {
		if secret.Type != db.EnvironmentSecretVar {
			continue
		}
		args = append(args, "--extra-vars", fmt.Sprintf("%s=%s", secret.Name, secret.Secret))
	}

	templateArgs := unmarshalArgsJSON(e.Template.Arguments)
	var taskArgs []string
	if e.Template.AllowOverrideArgsInTask {
		taskArgs = unmarshalArgsJSON(e.Task.Arguments)
	}

	if len(tplParams.Limit) > 0 || (tplParams.AllowOverrideLimit && params.Limit != nil) {
		limit := strings.Join(tplParams.Limit, ",")
		if tplParams.AllowOverrideLimit && params.Limit != nil {
			limit = strings.Join(params.Limit, ",")
		}
		if limit != "" {
			templateArgs = append(templateArgs, "--limit="+limit)
		}
	}
	if len(tplParams.Tags) > 0 || (tplParams.AllowOverrideTags && params.Tags != nil) {
		tags := strings.Join(tplParams.Tags, ",")
		if tplParams.AllowOverrideTags && params.Tags != nil {
			tags = strings.Join(params.Tags, ",")
		}
		if tags != "" {
			templateArgs = append(templateArgs, "--tags="+tags)
		}
	}
	if len(tplParams.SkipTags) > 0 || (tplParams.AllowOverrideSkipTags && params.SkipTags != nil) {
		skip := strings.Join(tplParams.SkipTags, ",")
		if tplParams.AllowOverrideSkipTags && params.SkipTags != nil {
			skip = strings.Join(params.SkipTags, ",")
		}
		if skip != "" {
			templateArgs = append(templateArgs, "--skip-tags="+skip)
		}
	}

	args = append(args, templateArgs...)
	args = append(args, taskArgs...)

	playbook := e.Task.Playbook
	if playbook == "" {
		playbook = e.Template.Playbook
	}
	args = append(args, playbook)

	return args
}

// unmarshalArgsJSON decodes the JSON-string CLI args fields on Template/Task. Errors
// are swallowed: an invalid value is treated as "no extra args" rather than failing
// the whole task — matches the most-forgiving interpretation used by LocalExecutor
// when callers don't surface the parse error to the user.
func unmarshalArgsJSON(raw *string) []string {
	if raw == nil || *raw == "" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(*raw), &out); err != nil {
		return nil
	}
	return out
}

// sanitizeForResourceName keeps the input DNS-1123 compatible — lowercase letters,
// digits, and dashes. K8s rejects underscores, dots, and uppercase in resource names,
// but vault names in Semaphore have no such constraint.
func sanitizeForResourceName(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-':
			b.WriteRune('-')
		default:
			b.WriteRune('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "x"
	}
	return out
}
