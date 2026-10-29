package client

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/semaphoreui/semaphore/cli/cmd/client/output"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/apiclient"
	"github.com/spf13/cobra"
)

// resolveRef turns an "<id|name>" flag value into the id of a referenced
// object (inventory, repository, key, …).
func resolveRef[T any](cmd *cobra.Command, r *runtime, api *apiclient.Client, projectID int, res resource[T], value string) (int, error) {
	item, err := res.resolve(cmd, r, api, projectID, value)
	if err != nil {
		return 0, err
	}
	return res.id(item), nil
}

func requireFields(cmd *cobra.Command, missing ...string) error {
	var names []string
	for i := 0; i+1 < len(missing); i += 2 {
		if missing[i+1] == "" {
			names = append(names, missing[i])
		}
	}
	if len(names) > 0 {
		return usage(cmd.CommandPath(), "missing required: %s (flags or --from-file)", strings.Join(names, ", "))
	}
	return nil
}

// Templates.

func templateMutable(tpl resource[db.Template]) mutable[db.Template] {
	return mutable[db.Template]{
		res: tpl,
		create: func(ctx context.Context, api *apiclient.Client, p int, t db.Template) (db.Template, error) {
			return api.CreateTemplate(ctx, p, t)
		},
		update: func(ctx context.Context, api *apiclient.Client, p int, t db.Template) error {
			return api.UpdateTemplate(ctx, p, t)
		},
		remove: func(ctx context.Context, api *apiclient.Client, p, id int) error {
			return api.DeleteTemplate(ctx, p, id)
		},
		name: func(t db.Template) string { return t.Name },
		createExample: `  semaphore client template create -p 1 --name deploy --app ansible --playbook site.yml --repository infra --inventory prod --environment default
  semaphore client template create -p 1 --from-file template.yml`,
		updateExample: `  semaphore client template update deploy --description "Deploys the site"
  semaphore client template update 12 --inventory staging --branch release`,
		validate: func(cmd *cobra.Command, t db.Template) error {
			return requireFields(cmd, "--name", t.Name, "--playbook", t.Playbook, "--repository", nz(t.RepositoryID))
		},
		bind: func(cmd *cobra.Command) applyFunc[db.Template] {
			var name, app, playbook, repo, inv, env, description, arguments, branch, view, workDir string
			var allowArgs bool
			fl := cmd.Flags()
			fl.StringVar(&name, "name", "", "Template name")
			fl.StringVar(&app, "app", "", "Application: ansible, terraform, tofu, bash, python, powershell, pulumi, terragrunt")
			fl.StringVar(&playbook, "playbook", "", "Playbook or script path inside the repository")
			fl.StringVar(&repo, "repository", "", "Repository id or name")
			fl.StringVar(&inv, "inventory", "", "Inventory id or name")
			fl.StringVar(&env, "environment", "", "Environment (variable group) id or name")
			fl.StringVar(&description, "description", "", "Description")
			fl.StringVar(&arguments, "arguments", "", "Extra CLI arguments as a JSON array")
			fl.BoolVar(&allowArgs, "allow-override-args", false, "Let a task override the arguments (use =false to forbid)")
			fl.StringVar(&branch, "branch", "", "Git branch")
			fl.StringVar(&view, "view", "", "View id or title to show the template in")
			fl.StringVar(&workDir, "working-directory", "", "Working directory inside the repository")
			return func(cmd *cobra.Command, r *runtime, api *apiclient.Client, p int, t *db.Template) error {
				if fl.Changed("name") {
					t.Name = name
				}
				if fl.Changed("app") {
					t.App = db.TemplateApp(app)
				}
				if fl.Changed("playbook") {
					t.Playbook = playbook
				}
				if fl.Changed("description") {
					t.Description = strPtr(description)
				}
				if fl.Changed("arguments") {
					t.Arguments = strPtr(arguments)
				}
				if fl.Changed("allow-override-args") {
					t.AllowOverrideArgsInTask = allowArgs
				}
				if fl.Changed("branch") {
					t.GitBranch = strPtr(branch)
				}
				if fl.Changed("working-directory") {
					t.WorkingDirectory = strPtr(workDir)
				}
				if repo != "" {
					id, err := resolveRef(cmd, r, api, p, repositoryResource(), repo)
					if err != nil {
						return err
					}
					t.RepositoryID = id
				}
				if inv != "" {
					id, err := resolveRef(cmd, r, api, p, inventoryResource(), inv)
					if err != nil {
						return err
					}
					t.InventoryID = &id
				}
				if env != "" {
					id, err := resolveRef(cmd, r, api, p, environmentResource(), env)
					if err != nil {
						return err
					}
					t.EnvironmentID = id
					t.EnvironmentIDs = []int{id}
				}
				if view != "" {
					id, err := resolveRef(cmd, r, api, p, viewResource(), view)
					if err != nil {
						return err
					}
					t.ViewID = &id
				}
				return nil
			}
		},
	}
}

func newTemplateStopAllCmd(tpl resource[db.Template]) *cobra.Command {
	return &cobra.Command{
		Use:   "stop-all <id|name>",
		Short: "Stop every running task of a template",
		Long: `Ask the server to stop every running or queued task of one template. Asks
for confirmation in a terminal; pass --yes to skip it.`,
		Example: `  semaphore client template stop-all deploy -p 1 -y`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, api, projectID, err := taskCommandSetup(cmd)
			if err != nil {
				return err
			}
			t, err := tpl.resolve(cmd, r, api, projectID, args[0])
			if err != nil {
				return err
			}
			if err := r.prompter.Confirm(fmt.Sprintf("Stop all running tasks of template %q?", t.Name)); err != nil {
				return err
			}
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			if err := api.StopAllTemplateTasks(ctx, projectID, t.ID); err != nil {
				return err
			}
			r.warnf("Stop requested for all tasks of template %q.", t.Name)
			return nil
		},
	}
}

// Inventories.

func inventoryMutable(res resource[db.Inventory]) mutable[db.Inventory] {
	return mutable[db.Inventory]{
		res: res,
		create: func(ctx context.Context, api *apiclient.Client, p int, i db.Inventory) (db.Inventory, error) {
			return api.CreateInventory(ctx, p, i)
		},
		update: func(ctx context.Context, api *apiclient.Client, p int, i db.Inventory) error {
			return api.UpdateInventory(ctx, p, i)
		},
		remove: func(ctx context.Context, api *apiclient.Client, p, id int) error {
			return api.DeleteInventory(ctx, p, id)
		},
		name: func(i db.Inventory) string { return i.Name },
		createExample: `  semaphore client inventory create -p 1 --name prod --type static --inventory-file hosts.ini --ssh-key deploy
  semaphore client inventory create -p 1 --name repo-hosts --type file --inventory inventories/prod.yml --repository infra`,
		updateExample: `  semaphore client inventory update prod --inventory-file hosts.ini`,
		validate: func(cmd *cobra.Command, i db.Inventory) error {
			return requireFields(cmd, "--name", i.Name, "--type", string(i.Type))
		},
		bind: func(cmd *cobra.Command) applyFunc[db.Inventory] {
			var name, typ, inventory, inventoryFile, sshKey, becomeKey, repo string
			fl := cmd.Flags()
			fl.StringVar(&name, "name", "", "Inventory name")
			fl.StringVar(&typ, "type", "", "Type: static, static-yaml, file, terraform-workspace, tofu-workspace, terragrunt-workspace")
			fl.StringVar(&inventory, "inventory", "", "Inventory content (static types) or path inside the repository (file type)")
			fl.StringVar(&inventoryFile, "inventory-file", "", "Local file whose content becomes the inventory")
			fl.StringVar(&sshKey, "ssh-key", "", "Access key id or name for SSH")
			fl.StringVar(&becomeKey, "become-key", "", "Access key id or name for become (sudo)")
			fl.StringVar(&repo, "repository", "", "Repository id or name (file type)")
			return func(cmd *cobra.Command, r *runtime, api *apiclient.Client, p int, i *db.Inventory) error {
				if fl.Changed("name") {
					i.Name = name
				}
				if fl.Changed("type") {
					i.Type = db.InventoryType(typ)
				}
				if fl.Changed("inventory") {
					i.Inventory = inventory
				}
				if inventoryFile != "" {
					data, err := os.ReadFile(inventoryFile)
					if err != nil {
						return fmt.Errorf("read inventory: %w", err)
					}
					i.Inventory = string(data)
				}
				if sshKey != "" {
					id, err := resolveRef(cmd, r, api, p, keyResource(), sshKey)
					if err != nil {
						return err
					}
					i.SSHKeyID = &id
				}
				if becomeKey != "" {
					id, err := resolveRef(cmd, r, api, p, keyResource(), becomeKey)
					if err != nil {
						return err
					}
					i.BecomeKeyID = &id
				}
				if repo != "" {
					id, err := resolveRef(cmd, r, api, p, repositoryResource(), repo)
					if err != nil {
						return err
					}
					i.RepositoryID = &id
				}
				return nil
			}
		},
	}
}

// Environments.

func environmentMutable(res resource[db.Environment]) mutable[db.Environment] {
	return mutable[db.Environment]{
		res: res,
		create: func(ctx context.Context, api *apiclient.Client, p int, e db.Environment) (db.Environment, error) {
			return api.CreateEnvironment(ctx, p, e)
		},
		update: func(ctx context.Context, api *apiclient.Client, p int, e db.Environment) error {
			return api.UpdateEnvironment(ctx, p, e)
		},
		remove: func(ctx context.Context, api *apiclient.Client, p, id int) error {
			return api.DeleteEnvironment(ctx, p, id)
		},
		name: func(e db.Environment) string { return e.Name },
		createExample: `  semaphore client environment create -p 1 --name prod --json vars.json --secrets-file secrets.yml
  semaphore client environment create -p 1 --name empty --json '{}'`,
		updateExample: `  semaphore client environment update prod --env '{"ANSIBLE_FORCE_COLOR":"1"}'
  semaphore client environment update prod --secrets-file secrets.yml`,
		validate: func(cmd *cobra.Command, e db.Environment) error {
			if e.JSON == "" {
				e.JSON = "{}"
			}
			return requireFields(cmd, "--name", e.Name)
		},
		bind: func(cmd *cobra.Command) applyFunc[db.Environment] {
			var name, jsonVars, envVars, secretsFile string
			var secrets, envSecrets, deleteSecrets []string
			fl := cmd.Flags()
			fl.StringArrayVar(&deleteSecrets, "delete-secret", nil, "Name of a secret to delete (update only); repeatable")
			fl.StringVar(&name, "name", "", "Environment name")
			fl.StringVar(&jsonVars, "json", "", "Extra variables as a JSON object, inline or a file path")
			fl.StringVar(&envVars, "env", "", "Environment variables as a JSON object, inline or a file path")
			fl.StringArrayVar(&secrets, "secret", nil, "Secret extra variable NAME=VALUE (visible in `ps`; prefer --secrets-file); repeatable")
			fl.StringArrayVar(&envSecrets, "secret-env", nil, "Secret environment variable NAME=VALUE (visible in `ps`); repeatable")
			fl.StringVar(&secretsFile, "secrets-file", "", "JSON or YAML file: {vars: {NAME: VALUE}, env: {NAME: VALUE}} or a flat {NAME: VALUE} map of secret variables")
			return func(cmd *cobra.Command, r *runtime, api *apiclient.Client, _ int, e *db.Environment) error {
				if fl.Changed("name") {
					e.Name = name
				}
				if jsonVars != "" {
					v, err := textOrFile(jsonVars)
					if err != nil {
						return err
					}
					e.JSON = v
				}
				if envVars != "" {
					v, err := textOrFile(envVars)
					if err != nil {
						return err
					}
					e.ENV = strPtr(v)
				}
				if e.JSON == "" {
					e.JSON = "{}"
				}
				// The server applies secrets as operations: existing ones come
				// back from GET without an operation and must not be re-sent,
				// or the update is rejected. Deletions need the secret's id.
				existing := e.Secrets
				e.Secrets = nil
				for _, name := range deleteSecrets {
					found := false
					for _, sec := range existing {
						if sec.Name == name {
							e.Secrets = append(e.Secrets, db.EnvironmentSecret{ID: sec.ID, Type: sec.Type, Name: sec.Name, Operation: db.EnvironmentSecretDelete})
							found = true
						}
					}
					if !found {
						return &NotFoundError{Kind: "secret", Name: name}
					}
				}
				if len(secrets)+len(envSecrets) > 0 {
					r.warnf("warning: secrets given on the command line are visible to other users of this machine; prefer --secrets-file")
				}
				add := func(typ db.EnvironmentSecretType, kv string) error {
					k, v, ok := strings.Cut(kv, "=")
					if !ok || k == "" {
						return usage(cmd.CommandPath(), "expected NAME=VALUE, got %q", kv)
					}
					e.Secrets = append(e.Secrets, db.EnvironmentSecret{Type: typ, Name: k, Secret: v, Operation: db.EnvironmentSecretCreate})
					return nil
				}
				for _, kv := range secrets {
					if err := add(db.EnvironmentSecretVar, kv); err != nil {
						return err
					}
				}
				for _, kv := range envSecrets {
					if err := add(db.EnvironmentSecretEnv, kv); err != nil {
						return err
					}
				}
				if secretsFile != "" {
					var doc map[string]any
					if err := loadFromFile(cmd, secretsFile, &doc); err != nil {
						return err
					}
					vars, env := splitSecretsDoc(doc)
					for k, v := range vars {
						e.Secrets = append(e.Secrets, db.EnvironmentSecret{Type: db.EnvironmentSecretVar, Name: k, Secret: fmt.Sprint(v), Operation: db.EnvironmentSecretCreate})
					}
					for k, v := range env {
						e.Secrets = append(e.Secrets, db.EnvironmentSecret{Type: db.EnvironmentSecretEnv, Name: k, Secret: fmt.Sprint(v), Operation: db.EnvironmentSecretCreate})
					}
				}
				return nil
			}
		},
	}
}

// splitSecretsDoc accepts {vars: {...}, env: {...}} or a flat map of vars.
func splitSecretsDoc(doc map[string]any) (vars, env map[string]any) {
	v, hasVars := doc["vars"].(map[string]any)
	e, hasEnv := doc["env"].(map[string]any)
	if hasVars || hasEnv {
		return v, e
	}
	return doc, nil
}

// Access keys. Secret material never comes from an argument.

func keyMutable(res resource[db.AccessKey]) mutable[db.AccessKey] {
	return mutable[db.AccessKey]{
		res: res,
		create: func(ctx context.Context, api *apiclient.Client, p int, k db.AccessKey) (db.AccessKey, error) {
			return api.CreateKey(ctx, p, k)
		},
		update: func(ctx context.Context, api *apiclient.Client, p int, k db.AccessKey) error {
			return api.UpdateKey(ctx, p, k)
		},
		remove: func(ctx context.Context, api *apiclient.Client, p, id int) error { return api.DeleteKey(ctx, p, id) },
		name:   func(k db.AccessKey) string { return k.Name },
		createExample: `  semaphore client key create -p 1 --name deploy --type ssh --login git --ssh-key-file ~/.ssh/id_ed25519
  semaphore client key create -p 1 --name vault --type login_password --login admin --password-file ./pw.txt
  echo "$SECRET" | semaphore client key create -p 1 --name api --type string --secret-stdin`,
		updateExample: `  semaphore client key update deploy --ssh-key-file ~/.ssh/new_key
  semaphore client key update 4 --name "deploy (old)"`,
		validate: func(cmd *cobra.Command, k db.AccessKey) error {
			return requireFields(cmd, "--name", k.Name, "--type", string(k.Type))
		},
		bind: func(cmd *cobra.Command) applyFunc[db.AccessKey] {
			var name, typ, login, sshKeyFile, passphraseFile, passwordFile, secretFile string
			var passwordStdin, secretStdin bool
			fl := cmd.Flags()
			fl.StringVar(&name, "name", "", "Key name")
			fl.StringVar(&typ, "type", "", "Type: ssh, login_password, string, none")
			fl.StringVar(&login, "login", "", "Login (ssh, login_password)")
			fl.StringVar(&sshKeyFile, "ssh-key-file", "", "File with the private key (ssh)")
			fl.StringVar(&passphraseFile, "passphrase-file", "", "File with the key passphrase (ssh)")
			fl.StringVar(&passwordFile, "password-file", "", "File with the password (login_password)")
			fl.BoolVar(&passwordStdin, "password-stdin", false, "Read the password from stdin (login_password)")
			fl.StringVar(&secretFile, "secret-file", "", "File with the secret (string)")
			fl.BoolVar(&secretStdin, "secret-stdin", false, "Read the secret from stdin (string)")
			return func(cmd *cobra.Command, r *runtime, _ *apiclient.Client, _ int, k *db.AccessKey) error {
				if fl.Changed("name") {
					k.Name = name
				}
				if fl.Changed("type") {
					k.Type = db.AccessKeyType(typ)
				}
				if fl.Changed("login") {
					k.SshKey.Login = login
					k.LoginPassword.Login = login
				}
				secretGiven := false
				if sshKeyFile != "" {
					v, err := readSecretFile(sshKeyFile)
					if err != nil {
						return err
					}
					k.SshKey.PrivateKey, secretGiven = v, true
				}
				if passphraseFile != "" {
					v, err := readSecretFile(passphraseFile)
					if err != nil {
						return err
					}
					k.SshKey.Passphrase, secretGiven = v, true
				}
				if passwordFile != "" && passwordStdin {
					return usage(cmd.CommandPath(), "use either --password-file or --password-stdin")
				}
				if passwordFile != "" {
					v, err := readSecretFile(passwordFile)
					if err != nil {
						return err
					}
					k.LoginPassword.Password, secretGiven = v, true
				}
				if passwordStdin {
					v, err := r.readSecretStdin(cmd, "password")
					if err != nil {
						return err
					}
					k.LoginPassword.Password, secretGiven = v, true
				}
				if secretFile != "" {
					v, err := readSecretFile(secretFile)
					if err != nil {
						return err
					}
					k.String, secretGiven = v, true
				}
				if secretStdin {
					v, err := r.readSecretStdin(cmd, "secret")
					if err != nil {
						return err
					}
					k.String, secretGiven = v, true
				}
				// On update the server keeps the stored secret unless told to
				// replace it; a login or type change alone must not wipe it.
				k.OverrideSecret = secretGiven || k.ID == 0
				return nil
			}
		},
	}
}

// Repositories.

func repositoryMutable(res resource[db.Repository]) mutable[db.Repository] {
	return mutable[db.Repository]{
		res: res,
		create: func(ctx context.Context, api *apiclient.Client, p int, repo db.Repository) (db.Repository, error) {
			return api.CreateRepository(ctx, p, repo)
		},
		update: func(ctx context.Context, api *apiclient.Client, p int, repo db.Repository) error {
			return api.UpdateRepository(ctx, p, repo)
		},
		remove: func(ctx context.Context, api *apiclient.Client, p, id int) error {
			return api.DeleteRepository(ctx, p, id)
		},
		name:          func(repo db.Repository) string { return repo.Name },
		createExample: `  semaphore client repository create -p 1 --name infra --url git@github.com:org/infra.git --branch main --ssh-key deploy`,
		updateExample: `  semaphore client repository update infra --branch release`,
		validate: func(cmd *cobra.Command, repo db.Repository) error {
			return requireFields(cmd, "--name", repo.Name, "--url", repo.GitURL, "--branch", repo.GitBranch, "--ssh-key", nz(repo.SSHKeyID))
		},
		bind: func(cmd *cobra.Command) applyFunc[db.Repository] {
			var name, gitURL, branch, sshKey string
			fl := cmd.Flags()
			fl.StringVar(&name, "name", "", "Repository name")
			fl.StringVar(&gitURL, "url", "", "Git URL or local path")
			fl.StringVar(&branch, "branch", "", "Branch")
			fl.StringVar(&sshKey, "ssh-key", "", "Access key id or name used to fetch (type none for public repositories)")
			return func(cmd *cobra.Command, r *runtime, api *apiclient.Client, p int, repo *db.Repository) error {
				if fl.Changed("name") {
					repo.Name = name
				}
				if fl.Changed("url") {
					repo.GitURL = gitURL
				}
				if fl.Changed("branch") {
					repo.GitBranch = branch
				}
				if sshKey != "" {
					id, err := resolveRef(cmd, r, api, p, keyResource(), sshKey)
					if err != nil {
						return err
					}
					repo.SSHKeyID = id
				}
				return nil
			}
		},
	}
}

// Schedules.

func scheduleMutable(res resource[db.ScheduleWithTpl]) mutable[db.ScheduleWithTpl] {
	return mutable[db.ScheduleWithTpl]{
		res: res,
		create: func(ctx context.Context, api *apiclient.Client, p int, s db.ScheduleWithTpl) (db.ScheduleWithTpl, error) {
			created, err := api.CreateSchedule(ctx, p, s.Schedule)
			return db.ScheduleWithTpl{Schedule: created}, err
		},
		update: func(ctx context.Context, api *apiclient.Client, p int, s db.ScheduleWithTpl) error {
			return api.UpdateSchedule(ctx, p, s.Schedule)
		},
		remove: func(ctx context.Context, api *apiclient.Client, p, id int) error {
			return api.DeleteSchedule(ctx, p, id)
		},
		name: func(s db.ScheduleWithTpl) string { return s.Name },
		createExample: `  semaphore client schedule create -p 1 --template deploy --cron "0 3 * * *" --name nightly
  semaphore client schedule create -p 1 --template 12 --cron "@hourly" --inactive`,
		updateExample: `  semaphore client schedule update nightly --cron "0 4 * * *"`,
		validate: func(cmd *cobra.Command, s db.ScheduleWithTpl) error {
			return requireFields(cmd, "--template", nz(s.TemplateID), "--cron", s.CronFormat)
		},
		bind: func(cmd *cobra.Command) applyFunc[db.ScheduleWithTpl] {
			var name, template, cron string
			var inactive bool
			fl := cmd.Flags()
			fl.StringVar(&name, "name", "", "Schedule name")
			fl.StringVar(&template, "template", "", "Template id or name to run")
			fl.StringVar(&cron, "cron", "", "Cron expression, e.g. \"0 3 * * *\" or @daily")
			fl.BoolVar(&inactive, "inactive", false, "Create the schedule disabled")
			return func(cmd *cobra.Command, r *runtime, api *apiclient.Client, p int, s *db.ScheduleWithTpl) error {
				if s.ID == 0 && !fl.Changed("inactive") {
					s.Active = true
				}
				if fl.Changed("inactive") {
					s.Active = !inactive
				}
				if fl.Changed("name") {
					s.Name = name
				}
				if fl.Changed("cron") {
					s.CronFormat = cron
				}
				if template != "" {
					id, err := resolveRef(cmd, r, api, p, templateResource(), template)
					if err != nil {
						return err
					}
					s.TemplateID = id
				}
				return nil
			}
		},
	}
}

func newScheduleActiveCmd(res resource[db.ScheduleWithTpl], active bool) *cobra.Command {
	verb := "enable"
	if !active {
		verb = "disable"
	}
	return &cobra.Command{
		Use:   verb + " <id|name>",
		Short: capitalize(verb) + " a schedule",
		Long: fmt.Sprintf(`%s a schedule without changing anything else about it. A disabled
schedule keeps its cron expression but does not start tasks.`, capitalize(verb)),
		Example: fmt.Sprintf(`  semaphore client schedule %s nightly -p 1`, verb),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, api, projectID, err := taskCommandSetup(cmd)
			if err != nil {
				return err
			}
			s, err := res.resolve(cmd, r, api, projectID, args[0])
			if err != nil {
				return err
			}
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			if err := api.SetScheduleActive(ctx, projectID, s.ID, active); err != nil {
				return err
			}
			r.warnf("Schedule %q (id %d) %sd.", s.Name, s.ID, verb)
			return nil
		},
	}
}

func newScheduleValidateCmd() *cobra.Command {
	var cron string
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Check a cron expression with the server",
		Long: `Ask the server whether a cron expression is valid, before putting it into a
schedule. Exit code 0 when it is, 1 with the server's message when it is not.`,
		Example: `  semaphore client schedule validate --cron "0 3 * * *" -p 1`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			r, api, projectID, err := taskCommandSetup(cmd)
			if err != nil {
				return err
			}
			if cron == "" {
				return usage(cmd.CommandPath(), "--cron is required")
			}
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			if err := api.ValidateCron(ctx, projectID, cron); err != nil {
				return err
			}
			r.warnf("Cron expression %q is valid.", cron)
			return nil
		},
	}
	cmd.Flags().StringVar(&cron, "cron", "", "Cron expression to check")
	return cmd
}

// Views.

func viewMutable(res resource[db.View]) mutable[db.View] {
	return mutable[db.View]{
		res: res,
		create: func(ctx context.Context, api *apiclient.Client, p int, v db.View) (db.View, error) {
			return api.CreateView(ctx, p, v)
		},
		update: func(ctx context.Context, api *apiclient.Client, p int, v db.View) error {
			return api.UpdateView(ctx, p, v)
		},
		remove:        func(ctx context.Context, api *apiclient.Client, p, id int) error { return api.DeleteView(ctx, p, id) },
		name:          func(v db.View) string { return v.Title },
		createExample: `  semaphore client view create -p 1 --title Ops --position 2`,
		updateExample: `  semaphore client view update Ops --title Operations`,
		validate:      func(cmd *cobra.Command, v db.View) error { return requireFields(cmd, "--title", v.Title) },
		bind: func(cmd *cobra.Command) applyFunc[db.View] {
			var title string
			var position int
			fl := cmd.Flags()
			fl.StringVar(&title, "title", "", "View title")
			fl.IntVar(&position, "position", 0, "Position among the views")
			return func(_ *cobra.Command, _ *runtime, _ *apiclient.Client, _ int, v *db.View) error {
				if fl.Changed("title") {
					v.Title = title
				}
				if fl.Changed("position") {
					v.Position = position
				}
				return nil
			}
		},
	}
}

// Users (global).

func userMutable(res resource[db.User]) mutable[db.User] {
	return mutable[db.User]{
		res: res,
		// create is custom (password), see newUserCreateCmd.
		update: func(ctx context.Context, api *apiclient.Client, _ int, u db.User) error {
			return api.UpdateUser(ctx, u)
		},
		remove: func(ctx context.Context, api *apiclient.Client, _, id int) error { return api.DeleteUser(ctx, id) },
		name:   func(u db.User) string { return u.Username },
		updateExample: `  semaphore client user update bob --email bob@example.com
  semaphore client user update bob --admin=false`,
		bind: userFlags,
	}
}

// userFlags binds the user fields shared by create and update.
func userFlags(cmd *cobra.Command) applyFunc[db.User] {
	var login, name, email string
	var admin, external, alert bool
	fl := cmd.Flags()
	fl.StringVar(&login, "login", "", "Login (username)")
	fl.StringVar(&name, "name", "", "Full name")
	fl.StringVar(&email, "email", "", "E-mail")
	fl.BoolVar(&admin, "admin", false, "Administrator (use --admin=false to revoke)")
	fl.BoolVar(&external, "external", false, "External user (LDAP/OIDC), no local password")
	fl.BoolVar(&alert, "alert", false, "Send alerts to this user")
	return func(_ *cobra.Command, _ *runtime, _ *apiclient.Client, _ int, u *db.User) error {
		if fl.Changed("login") {
			u.Username = login
		}
		if fl.Changed("name") {
			u.Name = name
		}
		if fl.Changed("email") {
			u.Email = email
		}
		if fl.Changed("admin") {
			u.Admin = admin
		}
		if fl.Changed("external") {
			u.External = external
		}
		if fl.Changed("alert") {
			u.Alert = alert
		}
		return nil
	}
}

func newUserCreateCmd(res resource[db.User]) *cobra.Command {
	var fromFile string
	var passwordStdin bool
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a user",
		Long: `Create a user. The password is read from stdin with --password-stdin and is
never accepted as an argument; an --external user has no local password.`,
		Example: `  echo "$PASSWORD" | semaphore client user create --login bob --name "Bob" --email bob@example.com --password-stdin
  semaphore client user create --login alice --name Alice --email alice@example.com --external`,
		Args: cobra.NoArgs,
	}
	apply := userFlags(cmd)
	cmd.Flags().StringVar(&fromFile, "from-file", "", "JSON or YAML file with the user fields")
	cmd.Flags().BoolVar(&passwordStdin, "password-stdin", false, "Read the password from stdin")
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		r := rt(cmd)
		api, err := r.client(cmd)
		if err != nil {
			return err
		}
		if err := r.requireToken(cmd); err != nil {
			return err
		}
		var u db.User
		if err := loadFromFile(cmd, fromFile, &u); err != nil {
			return err
		}
		if err := apply(cmd, r, api, 0, &u); err != nil {
			return err
		}
		if err := requireFields(cmd, "--login", u.Username, "--name", u.Name, "--email", u.Email); err != nil {
			return err
		}
		password := ""
		if passwordStdin {
			if password, err = r.readSecretStdin(cmd, "password"); err != nil {
				return err
			}
		} else if !u.External {
			return usage(cmd.CommandPath(), "a local user needs --password-stdin (or pass --external)")
		}
		ctx, cancel := r.apiCtx(cmd)
		defer cancel()
		created, err := api.CreateUser(ctx, u, password)
		if err != nil {
			return err
		}
		r.warnf("User %q created with id %d.", created.Username, created.ID)
		return r.printer.Print(created, res.columns)
	}
	return cmd
}

func newUserSetPasswordCmd(res resource[db.User]) *cobra.Command {
	var passwordStdin bool
	cmd := &cobra.Command{
		Use:   "set-password <id|login>",
		Short: "Set a user's password (read from stdin)",
		Long: `Set a new password for a user. The password is read from stdin with
--password-stdin (or prompted for in a terminal) and is never accepted as an
argument, so it does not show up in the shell history or in 'ps'.`,
		Example: `  echo "$NEW_PASSWORD" | semaphore client user set-password bob --password-stdin`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r := rt(cmd)
			api, err := r.client(cmd)
			if err != nil {
				return err
			}
			if err := r.requireToken(cmd); err != nil {
				return err
			}
			if !passwordStdin {
				return usage(cmd.CommandPath(), "--password-stdin is required; the password is never taken from an argument")
			}
			u, err := res.resolve(cmd, r, api, 0, args[0])
			if err != nil {
				return err
			}
			password, err := r.readSecretStdin(cmd, "password")
			if err != nil {
				return err
			}
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			if err := api.SetUserPassword(ctx, u.ID, password, ""); err != nil {
				return err
			}
			r.warnf("Password of user %q (id %d) changed.", u.Username, u.ID)
			return nil
		},
	}
	cmd.Flags().BoolVar(&passwordStdin, "password-stdin", false, "Read the new password from stdin")
	return cmd
}

// Projects.

func projectMutable(res resource[db.Project]) mutable[db.Project] {
	return mutable[db.Project]{
		res: res,
		// create and delete are custom: --demo, and --confirm=<name>.
		update: func(ctx context.Context, api *apiclient.Client, _ int, p db.Project) error {
			return api.UpdateProject(ctx, p)
		},
		name: func(p db.Project) string { return p.Name },
		updateExample: `  semaphore client project update Homelab --max-parallel-tasks 4
  semaphore client project update 3 --alert --alert-chat ops`,
		bind: projectFlags,
	}
}

func projectFlags(cmd *cobra.Command) applyFunc[db.Project] {
	var name, typ, alertChat string
	var maxParallel int
	var alert bool
	fl := cmd.Flags()
	fl.StringVar(&name, "name", "", "Project name")
	fl.StringVar(&typ, "type", "", "Project type (empty for a regular project)")
	fl.IntVar(&maxParallel, "max-parallel-tasks", 0, "Maximum tasks running at once (0 = unlimited)")
	fl.BoolVar(&alert, "alert", false, "Send alerts about tasks (use --alert=false to stop)")
	fl.StringVar(&alertChat, "alert-chat", "", "Alert chat id")
	return func(_ *cobra.Command, _ *runtime, _ *apiclient.Client, _ int, p *db.Project) error {
		if fl.Changed("name") {
			p.Name = name
		}
		if fl.Changed("type") {
			p.Type = typ
		}
		if fl.Changed("max-parallel-tasks") {
			p.MaxParallelTasks = maxParallel
		}
		if fl.Changed("alert") {
			p.Alert = alert
		}
		if fl.Changed("alert-chat") {
			p.AlertChat = strPtr(alertChat)
		}
		return nil
	}
}

func newProjectCreateCmd(res resource[db.Project]) *cobra.Command {
	var fromFile string
	var demo bool
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project",
		Long: `Create a project from flags or from a JSON/YAML file (--from-file, API field
names); flags override the file. --demo also fills it with the demo
template, inventory and repository. The created project is printed.`,
		Example: `  semaphore client project create --name Homelab
  semaphore client project create --name Playground --demo`,
		Args: cobra.NoArgs,
	}
	apply := projectFlags(cmd)
	cmd.Flags().StringVar(&fromFile, "from-file", "", "JSON or YAML file with the project fields")
	cmd.Flags().BoolVar(&demo, "demo", false, "Fill the project with the demo resources")
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		r := rt(cmd)
		api, err := r.client(cmd)
		if err != nil {
			return err
		}
		if err := r.requireToken(cmd); err != nil {
			return err
		}
		var p db.Project
		if err := loadFromFile(cmd, fromFile, &p); err != nil {
			return err
		}
		if err := apply(cmd, r, api, 0, &p); err != nil {
			return err
		}
		if err := requireFields(cmd, "--name", p.Name); err != nil {
			return err
		}
		ctx, cancel := r.apiCtx(cmd)
		defer cancel()
		created, err := api.CreateProject(ctx, p, demo)
		if err != nil {
			return err
		}
		r.warnf("Project %q created with id %d.", created.Name, created.ID)
		return r.printer.Print(created, res.columns)
	}
	return cmd
}

func newProjectDeleteCmd(res resource[db.Project]) *cobra.Command {
	var confirm string
	cmd := &cobra.Command{
		Use:     "delete <id|name>",
		Aliases: []string{"rm"},
		Short:   "Delete a project with everything in it",
		Long: `Delete a project and all its templates, tasks, keys and other resources. This
cannot be undone, so the project name has to be typed back (or passed with
--confirm=<name>); --yes alone is not enough.`,
		Example: `  semaphore client project delete Playground --confirm=Playground`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r := rt(cmd)
			api, err := r.client(cmd)
			if err != nil {
				return err
			}
			if err := r.requireToken(cmd); err != nil {
				return err
			}
			p, err := res.resolve(cmd, r, api, 0, args[0])
			if err != nil {
				return err
			}
			if err := r.prompter.ConfirmName("project", p.Name, confirm); err != nil {
				return err
			}
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			if err := api.DeleteProject(ctx, p.ID); err != nil {
				return err
			}
			r.warnf("Project %q (id %d) deleted.", p.Name, p.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&confirm, "confirm", "", "The project name, to confirm without a prompt")
	return cmd
}

func newProjectBackupCmd(res resource[db.Project]) *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "backup [id|name]",
		Short: "Download a project backup (JSON)",
		Long: `Download the backup of a project — its templates, inventories, keys,
repositories, environments, views and schedules — as the JSON document the
'Restore' page of the UI accepts. Without --file the document goes to stdout;
with --file it is written with mode 0600, because it contains secrets.`,
		Example: `  semaphore client project backup Homelab --file homelab.json`,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r := rt(cmd)
			api, err := r.client(cmd)
			if err != nil {
				return err
			}
			if err := r.requireToken(cmd); err != nil {
				return err
			}
			arg := r.cfg.Project
			if len(args) == 1 {
				arg = args[0]
			}
			if arg == "" {
				return usage(cmd.CommandPath(), "pass a project id or name, or set a default project")
			}
			p, err := res.resolve(cmd, r, api, 0, arg)
			if err != nil {
				return err
			}
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			if file == "" {
				return api.BackupProject(ctx, p.ID, r.io.Out)
			}
			f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
			if err != nil {
				return err
			}
			defer f.Close()
			if err := api.BackupProject(ctx, p.ID, f); err != nil {
				return err
			}
			r.warnf("Backup of project %q written to %s.", p.Name, file)
			return nil
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "Write the backup to this file (mode 0600) instead of stdout")
	return cmd
}

func newProjectRestoreCmd(res resource[db.Project]) *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Create a project from a backup file",
		Long: `Create a new project from a backup file produced by 'project backup' or by
the UI. The new project is printed.`,
		Example: `  semaphore client project restore --file homelab.json`,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			r := rt(cmd)
			api, err := r.client(cmd)
			if err != nil {
				return err
			}
			if err := r.requireToken(cmd); err != nil {
				return err
			}
			if file == "" {
				return usage(cmd.CommandPath(), "--file is required")
			}
			f, err := os.Open(file)
			if err != nil {
				return err
			}
			defer f.Close()
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			created, err := api.RestoreProject(ctx, f)
			if err != nil {
				return err
			}
			r.warnf("Project %q restored with id %d.", created.Name, created.ID)
			return r.printer.Print(created, res.columns)
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "Backup file produced by 'project backup'")
	return cmd
}

// Project members.

func newMemberAddCmd(members resource[apiclient.ProjectMember], users resource[db.User]) *cobra.Command {
	var role string
	cmd := &cobra.Command{
		Use:   "add <user id|login|email>",
		Short: "Add a user to the project with a role",
		Long: `Add an existing user of the server to the project with a role: owner,
manager, task_runner or guest.`,
		Example: `  semaphore client project member add bob --role manager -p 1`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, api, projectID, err := taskCommandSetup(cmd)
			if err != nil {
				return err
			}
			if role == "" {
				return usage(cmd.CommandPath(), "--role is required (owner, manager, task_runner, guest or a custom role slug)")
			}
			u, err := users.resolve(cmd, r, api, 0, args[0])
			if err != nil {
				return err
			}
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			if err := api.AddProjectMember(ctx, projectID, u.ID, role); err != nil {
				return err
			}
			r.warnf("User %q added to project %d as %s.", u.Username, projectID, role)
			return nil
		},
	}
	cmd.Flags().StringVar(&role, "role", "", "Role slug: owner, manager, task_runner, guest or a custom role")
	return cmd
}

func newMemberUpdateCmd(members resource[apiclient.ProjectMember]) *cobra.Command {
	var role string
	cmd := &cobra.Command{
		Use:     "update <user id|login>",
		Short:   "Change the role of a project member",
		Long:    `Give a project member another role: owner, manager, task_runner or guest.`,
		Example: `  semaphore client project member update bob --role guest -p 1`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, api, projectID, err := taskCommandSetup(cmd)
			if err != nil {
				return err
			}
			if role == "" {
				return usage(cmd.CommandPath(), "--role is required")
			}
			m, err := members.resolve(cmd, r, api, projectID, args[0])
			if err != nil {
				return err
			}
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			if err := api.UpdateProjectMember(ctx, projectID, m.ID, role); err != nil {
				return err
			}
			r.warnf("Role of %q in project %d is now %s.", m.Username, projectID, role)
			return nil
		},
	}
	cmd.Flags().StringVar(&role, "role", "", "New role slug")
	return cmd
}

func newMemberRemoveCmd(members resource[apiclient.ProjectMember]) *cobra.Command {
	return &cobra.Command{
		Use:     "remove <user id|login>",
		Aliases: []string{"rm", "delete"},
		Short:   "Remove a user from the project",
		Long: `Remove a user from the project. The user keeps the account on the server.
Asks for confirmation in a terminal; pass --yes to skip it.`,
		Example: `  semaphore client project member remove bob -p 1 -y`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, api, projectID, err := taskCommandSetup(cmd)
			if err != nil {
				return err
			}
			m, err := members.resolve(cmd, r, api, projectID, args[0])
			if err != nil {
				return err
			}
			if err := r.prompter.Confirm(fmt.Sprintf("Remove %q from project %d?", m.Username, projectID)); err != nil {
				return err
			}
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			if err := api.RemoveProjectMember(ctx, projectID, m.ID); err != nil {
				return err
			}
			r.warnf("User %q removed from project %d.", m.Username, projectID)
			return nil
		},
	}
}

// Runners.

func runnerMutable(res resource[db.Runner]) mutable[db.Runner] {
	return mutable[db.Runner]{
		res: res,
		// create is custom: it prints the registration token once.
		update: func(ctx context.Context, api *apiclient.Client, p int, rn db.Runner) error {
			return api.UpdateRunner(ctx, p, rn)
		},
		remove:        func(ctx context.Context, api *apiclient.Client, p, id int) error { return api.DeleteRunner(ctx, p, id) },
		name:          func(rn db.Runner) string { return rn.Name },
		updateExample: `  semaphore client runner update builder-1 --tags docker,linux --max-parallel-tasks 4`,
		bind:          runnerFlags,
	}
}

func runnerFlags(cmd *cobra.Command) applyFunc[db.Runner] {
	var name, webhook string
	var tags []string
	var maxParallel int
	fl := cmd.Flags()
	fl.StringVar(&name, "name", "", "Runner name")
	fl.StringSliceVar(&tags, "tags", nil, "Tags, comma-separated")
	fl.IntVar(&maxParallel, "max-parallel-tasks", 0, "Maximum tasks at once")
	fl.StringVar(&webhook, "webhook", "", "Webhook URL called when a task is assigned")
	return func(_ *cobra.Command, _ *runtime, _ *apiclient.Client, _ int, rn *db.Runner) error {
		if fl.Changed("name") {
			rn.Name = name
		}
		if fl.Changed("tags") {
			rn.Tags = tags
		}
		if fl.Changed("max-parallel-tasks") {
			rn.MaxParallelTasks = maxParallel
		}
		if fl.Changed("webhook") {
			rn.Webhook = webhook
		}
		return nil
	}
}

func newRunnerCreateCmd(res resource[db.Runner]) *cobra.Command {
	var fromFile string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Register a runner record and print its registration token",
		Long: `Create a runner record (global, or a project's with --project) and print the
registration token the runner process needs. The token is shown only once;
re-issue it later with 'runner registration-token'.`,
		Example: `  semaphore client runner create --name builder-1 --tags docker,linux
  semaphore client runner create --name proj-runner -p 1 -q   # prints only the token`,
		Args: cobra.NoArgs,
	}
	apply := runnerFlags(cmd)
	cmd.Flags().StringVar(&fromFile, "from-file", "", "JSON or YAML file with the runner fields")
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		r := rt(cmd)
		api, err := r.client(cmd)
		if err != nil {
			return err
		}
		if err := r.requireToken(cmd); err != nil {
			return err
		}
		projectID, err := res.projectFor(cmd, r, api)
		if err != nil {
			return err
		}
		var rn db.Runner
		if err := loadFromFile(cmd, fromFile, &rn); err != nil {
			return err
		}
		if err := apply(cmd, r, api, projectID, &rn); err != nil {
			return err
		}
		if err := requireFields(cmd, "--name", rn.Name); err != nil {
			return err
		}
		ctx, cancel := r.apiCtx(cmd)
		defer cancel()
		created, err := api.CreateRunner(ctx, projectID, rn)
		if err != nil {
			return err
		}
		r.warnf("Runner %q created with id %d. The registration token below is shown once.", created.Name, created.ID)
		if r.printer.Options().Quiet || r.printer.Options().Format == output.FormatTable {
			r.printer.Printf("%s\n", created.Token)
			return nil
		}
		return r.printer.Print(struct {
			ID                int    `json:"id"`
			Name              string `json:"name"`
			RegistrationToken string `json:"registration_token"`
		}{created.ID, created.Name, created.Token}, nil)
	}
	return cmd
}

func newRunnerActiveCmd(res resource[db.Runner], active bool) *cobra.Command {
	verb := "enable"
	if !active {
		verb = "disable"
	}
	return &cobra.Command{
		Use:   verb + " <id|name>",
		Short: capitalize(verb) + " a runner",
		Long: fmt.Sprintf(`%s a runner. A disabled runner stays registered but is not given
tasks until it is enabled again.`, capitalize(verb)),
		Example: fmt.Sprintf(`  semaphore client runner %s worker-1`, verb),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, api, projectID, rn, err := runnerSetup(cmd, res, args[0])
			if err != nil {
				return err
			}
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			if err := api.SetRunnerActive(ctx, projectID, rn.ID, active); err != nil {
				return err
			}
			r.warnf("Runner %q (id %d) %sd.", rn.Name, rn.ID, verb)
			return nil
		},
	}
}

func newRunnerTokenCmd(res resource[db.Runner]) *cobra.Command {
	return &cobra.Command{
		Use:   "registration-token <id|name>",
		Short: "Re-issue the registration token of a runner (printed once)",
		Long: `Issue a new registration token for a runner and print it. The previous
token stops working; the server keeps only a hash, so the token cannot be shown
again later.`,
		Example: `  semaphore client runner registration-token worker-1 > runner.token`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, api, projectID, rn, err := runnerSetup(cmd, res, args[0])
			if err != nil {
				return err
			}
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			tok, err := api.RegenerateRunnerToken(ctx, projectID, rn.ID)
			if err != nil {
				return err
			}
			r.warnf("New registration token for runner %q (id %d); the previous one is void.", rn.Name, rn.ID)
			if r.printer.Options().Format == output.FormatTable || r.printer.Options().Quiet {
				r.printer.Printf("%s\n", tok.RegistrationToken)
				return nil
			}
			return r.printer.Print(tok, nil)
		},
	}
}

func newRunnerClearCacheCmd(res resource[db.Runner]) *cobra.Command {
	return &cobra.Command{
		Use:   "clear-cache <id|name>",
		Short: "Ask a runner to drop its repository cache",
		Long: `Ask a runner to delete its local clones of the repositories, so the next
task fetches them afresh.`,
		Example: `  semaphore client runner clear-cache worker-1`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, api, projectID, rn, err := runnerSetup(cmd, res, args[0])
			if err != nil {
				return err
			}
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			if err := api.ClearRunnerCache(ctx, projectID, rn.ID); err != nil {
				return err
			}
			r.warnf("Cache clearing requested for runner %q (id %d).", rn.Name, rn.ID)
			return nil
		},
	}
}

func runnerSetup(cmd *cobra.Command, res resource[db.Runner], arg string) (*runtime, *apiclient.Client, int, db.Runner, error) {
	r := rt(cmd)
	api, err := r.client(cmd)
	if err != nil {
		return nil, nil, 0, db.Runner{}, err
	}
	if err := r.requireToken(cmd); err != nil {
		return nil, nil, 0, db.Runner{}, err
	}
	projectID, err := res.projectFor(cmd, r, api)
	if err != nil {
		return nil, nil, 0, db.Runner{}, err
	}
	rn, err := res.resolve(cmd, r, api, projectID, arg)
	return r, api, projectID, rn, err
}

// nz renders an int for requireFields: zero means "not given".
func nz(n int) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprint(n)
}
