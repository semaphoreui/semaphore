package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/semaphoreui/semaphore/cli/cmd/client/output"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/apiclient"
	"github.com/spf13/cobra"
)

// resource describes one kind of object well enough to generate its `list` and
// `get` commands and to resolve "<id|name>" arguments for the other verbs.
type resource[T any] struct {
	kind    string   // singular, used as the command name: "template"
	aliases []string // "templates", "tpl"
	short   string
	columns []output.Column
	// scope says where the objects live: in a project (default), globally, or
	// globally unless --project is given explicitly (runners).
	scope scope

	list  func(ctx context.Context, api *apiclient.Client, projectID int) ([]T, error)
	get   func(ctx context.Context, api *apiclient.Client, projectID, id int) (T, error) // nil: taken from list
	id    func(T) int
	match func(T, string) bool // name match for "<name>" arguments
	label func(T) string       // "id (name)" for ambiguity messages
	// view projects an item before printing, e.g. to drop secret values.
	view func(T) any
}

type scope int

const (
	scopeProject scope = iota
	scopeGlobal
	scopeProjectOptional
)

// newCmd builds "<kind>" with list and get plus any extra verbs.
func (res resource[T]) newCmd(extra ...*cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     res.kind,
		Aliases: res.aliases,
		Short:   res.short,
		Args:    cobra.NoArgs,
		RunE:    func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	cmd.AddCommand(res.listCmd(), res.getCmd())
	cmd.AddCommand(extra...)
	return cmd
}

// projectFor returns the project id a command works in, or 0 for a global
// resource.
func (res resource[T]) projectFor(cmd *cobra.Command, r *runtime, api *apiclient.Client) (int, error) {
	switch res.scope {
	case scopeGlobal:
		return 0, nil
	case scopeProjectOptional:
		if !cmd.Flags().Changed("project") {
			return 0, nil
		}
	}
	return r.resolveProject(cmd, api)
}

func (res resource[T]) listCmd() *cobra.Command {
	use := "list"
	short := "List " + plural(res.kind)
	if res.scope != scopeGlobal {
		short += " of the project"
	}
	return &cobra.Command{
		Use:     use,
		Aliases: []string{"ls"},
		Short:   short,
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
			projectID, err := res.projectFor(cmd, r, api)
			if err != nil {
				return err
			}
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			items, err := res.list(ctx, api, projectID)
			if err != nil {
				return err
			}
			return r.printer.Print(res.viewAll(items), res.columns)
		},
	}
}

func (res resource[T]) getCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "get <id|name>",
		Aliases: []string{"show"},
		Short:   "Show one " + res.kind,
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
			projectID, err := res.projectFor(cmd, r, api)
			if err != nil {
				return err
			}
			item, err := res.resolve(cmd, r, api, projectID, args[0])
			if err != nil {
				return err
			}
			return r.printer.Print(res.viewOne(item), res.columns)
		},
	}
}

// resolve finds one object by id or name. With a get function the object is
// re-read by id, because some lists omit fields (templates lose their vaults).
func (res resource[T]) resolve(cmd *cobra.Command, r *runtime, api *apiclient.Client, projectID int, arg string) (T, error) {
	var zero T
	ctx, cancel := r.apiCtx(cmd)
	defer cancel()

	id, isID := 0, false
	if n, err := strconv.Atoi(arg); err == nil && n > 0 {
		id, isID = n, true
	}
	if !isID {
		items, err := res.list(ctx, api, projectID)
		if err != nil {
			return zero, err
		}
		item, err := pickByIDOrName(res.kind, arg, items, res.id, res.match, res.label)
		if err != nil {
			return zero, err
		}
		if res.get == nil {
			return item, nil
		}
		id = res.id(item)
	}

	if res.get != nil {
		return res.get(ctx, api, projectID, id)
	}
	items, err := res.list(ctx, api, projectID)
	if err != nil {
		return zero, err
	}
	return pickByIDOrName(res.kind, strconv.Itoa(id), items, res.id, res.match, res.label)
}

func (res resource[T]) viewOne(item T) any {
	if res.view == nil {
		return item
	}
	return res.view(item)
}

func (res resource[T]) viewAll(items []T) any {
	if res.view == nil {
		return items
	}
	out := make([]any, 0, len(items))
	for _, it := range items {
		out = append(out, res.view(it))
	}
	return out
}

// Helpers shared by the resource definitions.

func nameMatch(name string) func(string) bool {
	return func(arg string) bool { return strings.EqualFold(name, arg) }
}

func idName(id int, name string) string { return fmt.Sprintf("%d (%s)", id, name) }

func plural(kind string) string {
	switch {
	case strings.HasSuffix(kind, "y"):
		return kind[:len(kind)-1] + "ies"
	default:
		return kind + "s"
	}
}

// Resource definitions. They are functions, not package variables: the
// project forbids globals, and a command tree is built once per process anyway.

func projectResource() resource[db.Project] {
	return resource[db.Project]{
		kind:    "project",
		aliases: []string{"projects"},
		short:   "Manage projects",
		columns: output.Columns("id", "name", "type", "created"),
		scope:   scopeGlobal,
		list: func(ctx context.Context, api *apiclient.Client, _ int) ([]db.Project, error) {
			return api.ListProjects(ctx)
		},
		get: func(ctx context.Context, api *apiclient.Client, _, id int) (db.Project, error) {
			return api.GetProject(ctx, id)
		},
		id:    func(p db.Project) int { return p.ID },
		match: func(p db.Project, arg string) bool { return nameMatch(p.Name)(arg) },
		label: func(p db.Project) string { return idName(p.ID, p.Name) },
	}
}

func templateResource() resource[db.Template] {
	return resource[db.Template]{
		kind:    "template",
		aliases: []string{"templates", "tpl"},
		short:   "Manage task templates",
		columns: output.Columns("id", "name", "app", "playbook", "inventory_id", "repository_id", "environment_id"),
		list: func(ctx context.Context, api *apiclient.Client, projectID int) ([]db.Template, error) {
			return api.ListTemplates(ctx, projectID)
		},
		// The list omits vaults and survey variables; get re-reads by id.
		get: func(ctx context.Context, api *apiclient.Client, projectID, id int) (db.Template, error) {
			return api.GetTemplate(ctx, projectID, id)
		},
		id:    func(t db.Template) int { return t.ID },
		match: func(t db.Template, arg string) bool { return nameMatch(t.Name)(arg) },
		label: func(t db.Template) string { return idName(t.ID, t.Name) },
	}
}

func inventoryResource() resource[db.Inventory] {
	return resource[db.Inventory]{
		kind:    "inventory",
		aliases: []string{"inventories", "inv"},
		short:   "Manage inventories",
		columns: output.Columns("id", "name", "type", "ssh_key_id", "repository_id"),
		list: func(ctx context.Context, api *apiclient.Client, projectID int) ([]db.Inventory, error) {
			return api.ListInventories(ctx, projectID)
		},
		get: func(ctx context.Context, api *apiclient.Client, projectID, id int) (db.Inventory, error) {
			return api.GetInventory(ctx, projectID, id)
		},
		id:    func(i db.Inventory) int { return i.ID },
		match: func(i db.Inventory, arg string) bool { return nameMatch(i.Name)(arg) },
		label: func(i db.Inventory) string { return idName(i.ID, i.Name) },
	}
}

// environmentView is an environment without secret values: only the names of
// its secrets and whether a password is set.
type environmentView struct {
	ID             int      `json:"id"`
	Name           string   `json:"name"`
	ProjectID      int      `json:"project_id"`
	JSON           string   `json:"json"`
	ENV            *string  `json:"env"`
	HasPassword    bool     `json:"has_password"`
	SecretNames    []string `json:"secret_names"`
	SecretStorage  *int     `json:"secret_storage_id,omitempty"`
	SyncEnabled    bool     `json:"sync_enabled"`
	SyncIntervalMn int      `json:"sync_interval"`
}

func environmentResource() resource[db.Environment] {
	return resource[db.Environment]{
		kind:    "environment",
		aliases: []string{"environments", "env"},
		short:   "Manage variable groups (environments)",
		columns: output.Columns("id", "name", "secret_names"),
		list: func(ctx context.Context, api *apiclient.Client, projectID int) ([]db.Environment, error) {
			return api.ListEnvironments(ctx, projectID)
		},
		get: func(ctx context.Context, api *apiclient.Client, projectID, id int) (db.Environment, error) {
			return api.GetEnvironment(ctx, projectID, id)
		},
		id:    func(e db.Environment) int { return e.ID },
		match: func(e db.Environment, arg string) bool { return nameMatch(e.Name)(arg) },
		label: func(e db.Environment) string { return idName(e.ID, e.Name) },
		view: func(e db.Environment) any {
			names := make([]string, 0, len(e.Secrets))
			for _, s := range e.Secrets {
				names = append(names, s.Name)
			}
			return environmentView{
				ID: e.ID, Name: e.Name, ProjectID: e.ProjectID, JSON: e.JSON, ENV: e.ENV,
				HasPassword: e.Password != nil && *e.Password != "", SecretNames: names,
				SecretStorage: e.SecretStorageID, SyncEnabled: e.SyncEnabled, SyncIntervalMn: e.SyncInterval,
			}
		},
	}
}

// keyView is an access key without its secret material.
type keyView struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	ProjectID *int   `json:"project_id"`
	Login     string `json:"login,omitempty"`
}

func keyResource() resource[db.AccessKey] {
	return resource[db.AccessKey]{
		kind:    "key",
		aliases: []string{"keys"},
		short:   "Manage access keys (SSH keys, passwords)",
		columns: output.Columns("id", "name", "type"),
		list: func(ctx context.Context, api *apiclient.Client, projectID int) ([]db.AccessKey, error) {
			return api.ListKeys(ctx, projectID)
		},
		get: func(ctx context.Context, api *apiclient.Client, projectID, id int) (db.AccessKey, error) {
			return api.GetKey(ctx, projectID, id)
		},
		id:    func(k db.AccessKey) int { return k.ID },
		match: func(k db.AccessKey, arg string) bool { return nameMatch(k.Name)(arg) },
		label: func(k db.AccessKey) string { return idName(k.ID, k.Name) },
		view: func(k db.AccessKey) any {
			return keyView{ID: k.ID, Name: k.Name, Type: string(k.Type), ProjectID: k.ProjectID, Login: k.LoginPassword.Login}
		},
	}
}

func repositoryResource() resource[db.Repository] {
	return resource[db.Repository]{
		kind:    "repository",
		aliases: []string{"repositories", "repo"},
		short:   "Manage repositories",
		columns: output.Columns("id", "name", "git_url", "git_branch", "ssh_key_id"),
		list: func(ctx context.Context, api *apiclient.Client, projectID int) ([]db.Repository, error) {
			return api.ListRepositories(ctx, projectID)
		},
		get: func(ctx context.Context, api *apiclient.Client, projectID, id int) (db.Repository, error) {
			return api.GetRepository(ctx, projectID, id)
		},
		id:    func(r db.Repository) int { return r.ID },
		match: func(r db.Repository, arg string) bool { return nameMatch(r.Name)(arg) },
		label: func(r db.Repository) string { return idName(r.ID, r.Name) },
	}
}

func scheduleResource() resource[db.ScheduleWithTpl] {
	return resource[db.ScheduleWithTpl]{
		kind:    "schedule",
		aliases: []string{"schedules"},
		short:   "Manage schedules (cron runs of templates)",
		columns: output.Columns("id", "name", "tpl_name", "cron_format", "active"),
		list: func(ctx context.Context, api *apiclient.Client, projectID int) ([]db.ScheduleWithTpl, error) {
			return api.ListSchedules(ctx, projectID)
		},
		// GET by id answers a bare schedule; the list carries the template name.
		get: func(ctx context.Context, api *apiclient.Client, projectID, id int) (db.ScheduleWithTpl, error) {
			s, err := api.GetSchedule(ctx, projectID, id)
			return db.ScheduleWithTpl{Schedule: s}, err
		},
		id:    func(s db.ScheduleWithTpl) int { return s.ID },
		match: func(s db.ScheduleWithTpl, arg string) bool { return nameMatch(s.Name)(arg) },
		label: func(s db.ScheduleWithTpl) string { return idName(s.ID, s.Name) },
	}
}

func viewResource() resource[db.View] {
	return resource[db.View]{
		kind:    "view",
		aliases: []string{"views"},
		short:   "Manage views (template groups on the project page)",
		columns: output.Columns("id", "title", "position"),
		list: func(ctx context.Context, api *apiclient.Client, projectID int) ([]db.View, error) {
			return api.ListViews(ctx, projectID)
		},
		get: func(ctx context.Context, api *apiclient.Client, projectID, id int) (db.View, error) {
			return api.GetView(ctx, projectID, id)
		},
		id:    func(v db.View) int { return v.ID },
		match: func(v db.View, arg string) bool { return nameMatch(v.Title)(arg) },
		label: func(v db.View) string { return idName(v.ID, v.Title) },
	}
}

func userResource() resource[db.User] {
	return resource[db.User]{
		kind:    "user",
		aliases: []string{"users"},
		short:   "Manage users of the server (administrators only)",
		columns: output.Columns("id", "username", "name", "email", "admin", "external"),
		scope:   scopeGlobal,
		list: func(ctx context.Context, api *apiclient.Client, _ int) ([]db.User, error) {
			return api.ListUsers(ctx)
		},
		get: func(ctx context.Context, api *apiclient.Client, _, id int) (db.User, error) {
			return api.GetUser(ctx, id)
		},
		id: func(u db.User) int { return u.ID },
		match: func(u db.User, arg string) bool {
			return strings.EqualFold(u.Username, arg) || strings.EqualFold(u.Email, arg)
		},
		label: func(u db.User) string { return idName(u.ID, u.Username) },
	}
}

func runnerResource() resource[db.Runner] {
	return resource[db.Runner]{
		kind:    "runner",
		aliases: []string{"runners"},
		short:   "Manage runners (global, or a project's with --project)",
		columns: output.Columns("id", "name", "tags", "active", "status", "touched"),
		scope:   scopeProjectOptional,
		list: func(ctx context.Context, api *apiclient.Client, projectID int) ([]db.Runner, error) {
			return api.ListRunners(ctx, projectID)
		},
		get: func(ctx context.Context, api *apiclient.Client, projectID, id int) (db.Runner, error) {
			return api.GetRunner(ctx, projectID, id)
		},
		id:    func(r db.Runner) int { return r.ID },
		match: func(r db.Runner, arg string) bool { return nameMatch(r.Name)(arg) },
		label: func(r db.Runner) string { return idName(r.ID, r.Name) },
	}
}

func memberResource() resource[apiclient.ProjectMember] {
	return resource[apiclient.ProjectMember]{
		kind:    "member",
		aliases: []string{"members"},
		short:   "Manage the users of a project and their roles",
		columns: output.Columns("id", "username", "name", "roles"),
		list: func(ctx context.Context, api *apiclient.Client, projectID int) ([]apiclient.ProjectMember, error) {
			return api.ListProjectMembers(ctx, projectID)
		},
		id:    func(m apiclient.ProjectMember) int { return m.ID },
		match: func(m apiclient.ProjectMember, arg string) bool { return strings.EqualFold(m.Username, arg) },
		label: func(m apiclient.ProjectMember) string { return idName(m.ID, m.Username) },
	}
}

// taskColumns is the default table of a task list.
func taskColumns() []output.Column {
	return output.Columns("id", "status", "tpl_alias", "user_name", "created", "start", "end")
}

// newTemplateTasksCmd is `template tasks <id|name>`: the recent tasks of one
// template, newest first.
func newTemplateTasksCmd(tpl resource[db.Template]) *cobra.Command {
	var limit int
	var all bool
	cmd := &cobra.Command{
		Use:   "tasks <id|name>",
		Short: "List recent tasks of a template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r := rt(cmd)
			api, err := r.client(cmd)
			if err != nil {
				return err
			}
			if err := r.requireToken(cmd); err != nil {
				return err
			}
			projectID, err := r.resolveProject(cmd, api)
			if err != nil {
				return err
			}
			t, err := tpl.resolve(cmd, r, api, projectID, args[0])
			if err != nil {
				return err
			}
			if all {
				limit = 0
			}
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			tasks, err := api.ListAllTasks(ctx, projectID, t.ID, limit)
			if err != nil {
				return err
			}
			return r.printer.Print(tasks, taskColumns())
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of tasks")
	cmd.Flags().BoolVar(&all, "all", false, "List every task (ignores --limit)")
	return cmd
}

// newViewTemplatesCmd is `view templates <id|name>`.
func newViewTemplatesCmd(views resource[db.View], tpl resource[db.Template]) *cobra.Command {
	return &cobra.Command{
		Use:   "templates <id|name>",
		Short: "List the templates shown in a view",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r := rt(cmd)
			api, err := r.client(cmd)
			if err != nil {
				return err
			}
			if err := r.requireToken(cmd); err != nil {
				return err
			}
			projectID, err := r.resolveProject(cmd, api)
			if err != nil {
				return err
			}
			v, err := views.resolve(cmd, r, api, projectID, args[0])
			if err != nil {
				return err
			}
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			templates, err := api.ListViewTemplates(ctx, projectID, v.ID)
			if err != nil {
				return err
			}
			return r.printer.Print(templates, tpl.columns)
		},
	}
}

// newProjectGetCmd replaces the generic get: without an argument it shows the
// configured project.
func newProjectGetCmd(res resource[db.Project]) *cobra.Command {
	return &cobra.Command{
		Use:     "get [id|name]",
		Aliases: []string{"show"},
		Short:   "Show one project (the configured one by default)",
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
				return usage(cmd.CommandPath(), "pass a project id or name, or set a default project with -p / SEMAPHORE_CLIENT_PROJECT / the context")
			}
			p, err := res.resolve(cmd, r, api, 0, arg)
			if err != nil {
				return err
			}
			return r.printer.Print(p, res.columns)
		},
	}
}

// resourceCommands assembles every resource command of the client.
func resourceCommands() []*cobra.Command {
	projects := projectResource()
	templates := templateResource()
	inventories := inventoryResource()
	environments := environmentResource()
	keys := keyResource()
	repositories := repositoryResource()
	schedules := scheduleResource()
	views := viewResource()
	users := userResource()
	runners := runnerResource()
	members := memberResource()

	projectCmd := &cobra.Command{
		Use:     projects.kind,
		Aliases: projects.aliases,
		Short:   projects.short,
		Args:    cobra.NoArgs,
		RunE:    func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	memberCmd := members.newCmd(newMemberAddCmd(members, users), newMemberUpdateCmd(members), newMemberRemoveCmd(members))
	projectCmd.AddCommand(projects.listCmd(), newProjectGetCmd(projects), newProjectCreateCmd(projects))
	projectCmd.AddCommand(projectMutable(projects).commands()...)
	projectCmd.AddCommand(newProjectDeleteCmd(projects), newProjectBackupCmd(projects), newProjectRestoreCmd(projects), memberCmd)

	templateCmd := templates.newCmd(newTemplateTasksCmd(templates), newTemplateRunCmd(templates), newTemplateStopAllCmd(templates))
	templateCmd.AddCommand(templateMutable(templates).commands()...)

	inventoryCmd := inventories.newCmd(inventoryMutable(inventories).commands()...)
	environmentCmd := environments.newCmd(environmentMutable(environments).commands()...)
	keyCmd := keys.newCmd(keyMutable(keys).commands()...)
	repositoryCmd := repositories.newCmd(repositoryMutable(repositories).commands()...)

	scheduleCmd := schedules.newCmd(newScheduleActiveCmd(schedules, true), newScheduleActiveCmd(schedules, false), newScheduleValidateCmd())
	scheduleCmd.AddCommand(scheduleMutable(schedules).commands()...)

	viewCmd := views.newCmd(newViewTemplatesCmd(views, templates))
	viewCmd.AddCommand(viewMutable(views).commands()...)

	userCmd := users.newCmd(newUserCreateCmd(users), newUserSetPasswordCmd(users))
	userCmd.AddCommand(userMutable(users).commands()...)

	runnerCmd := runners.newCmd(newRunnerCreateCmd(runners), newRunnerActiveCmd(runners, true), newRunnerActiveCmd(runners, false), newRunnerTokenCmd(runners), newRunnerClearCacheCmd(runners))
	runnerCmd.AddCommand(runnerMutable(runners).commands()...)

	return []*cobra.Command{
		projectCmd,
		templateCmd,
		newTaskCmd(templates),
		inventoryCmd,
		environmentCmd,
		keyCmd,
		repositoryCmd,
		scheduleCmd,
		viewCmd,
		userCmd,
		runnerCmd,
	}
}
