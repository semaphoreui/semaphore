package client

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/semaphoreui/semaphore/cli/cmd/client/output"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/apiclient"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// runFlags are the task overrides `template run` and `task rerun` accept.
type runFlags struct {
	wait            bool
	follow          bool
	waitTimeout     time.Duration
	stopOnInterrupt bool

	message   string
	branch    string
	arguments string
	inventory string
	vars      []string
	varsFile  string

	// Ansible
	check      bool
	diff       bool
	debugTask  bool
	debugLevel int
	limit      []string
	tags       []string
	skipTags   []string
	skipGalaxy bool
	// Terraform / OpenTofu
	plan        bool
	destroy     bool
	autoApprove bool
	upgrade     bool
	reconfigure bool
}

func (f *runFlags) bindWait(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&f.wait, "wait", false, "Wait until the task finishes; exit code follows its status (0 success, 3 error, 6 stopped)")
	cmd.Flags().BoolVar(&f.follow, "follow", false, "Print the task log while waiting (implies --wait)")
	cmd.Flags().DurationVar(&f.waitTimeout, "wait-timeout", 0, "Give up waiting after this long, exit 124; the task keeps running (default: no limit)")
	cmd.Flags().BoolVar(&f.stopOnInterrupt, "stop-on-interrupt", false, "On Ctrl+C ask the server to stop the task (by default only the client disconnects)")
}

func (f *runFlags) bindOverrides(cmd *cobra.Command) {
	fl := cmd.Flags()
	fl.StringVar(&f.message, "message", "", "Message shown in the task history")
	fl.StringVar(&f.branch, "branch", "", "Git branch to check out (if the template allows it)")
	fl.StringVar(&f.arguments, "arguments", "", "Extra CLI arguments as a JSON array, e.g. '[\"-vv\"]' (if the template allows it)")
	fl.StringVar(&f.inventory, "inventory", "", "Inventory id or name (if the template allows overriding it)")
	fl.StringArrayVar(&f.vars, "var", nil, "Extra variable NAME=VALUE; repeatable")
	fl.StringVar(&f.varsFile, "extra-vars-file", "", "JSON or YAML file with extra variables")

	fl.BoolVar(&f.check, "check", false, "Ansible check mode (dry run, no changes)")
	fl.BoolVar(&f.diff, "diff", false, "Ansible --diff")
	fl.BoolVar(&f.debugTask, "debug-task", false, "Ansible verbose output (-vvvv)")
	fl.IntVar(&f.debugLevel, "debug-level", 0, "Ansible verbosity level 1-4 (implies --debug-task)")
	fl.StringSliceVar(&f.limit, "limit", nil, "Ansible --limit hosts, comma-separated")
	fl.StringSliceVar(&f.tags, "tags", nil, "Ansible --tags, comma-separated")
	fl.StringSliceVar(&f.skipTags, "skip-tags", nil, "Ansible --skip-tags, comma-separated")
	fl.BoolVar(&f.skipGalaxy, "skip-galaxy", false, "Skip ansible-galaxy requirements install")

	fl.BoolVar(&f.plan, "plan", false, "Terraform/OpenTofu: plan only")
	fl.BoolVar(&f.destroy, "destroy", false, "Terraform/OpenTofu: destroy")
	fl.BoolVar(&f.autoApprove, "auto-approve", false, "Terraform/OpenTofu: apply without approval")
	fl.BoolVar(&f.upgrade, "upgrade", false, "Terraform/OpenTofu: init -upgrade")
	fl.BoolVar(&f.reconfigure, "reconfigure", false, "Terraform/OpenTofu: init -reconfigure")
}

// apply fills a task from the flags the user actually passed, so that
// template defaults stay in force for everything else.
func (f *runFlags) apply(cmd *cobra.Command, r *runtime, api *apiclient.Client, projectID int, task *db.Task) error {
	fl := cmd.Flags()
	if f.message != "" {
		task.Message = f.message
	}
	if fl.Changed("branch") {
		b := f.branch
		task.GitBranch = &b
	}
	if fl.Changed("arguments") {
		a := f.arguments
		task.Arguments = &a
	}
	if f.inventory != "" {
		inv, err := inventoryResource().resolve(cmd, r, api, projectID, f.inventory)
		if err != nil {
			return err
		}
		id := inv.ID
		task.InventoryID = &id
	}

	env, err := buildExtraVars(cmd, f.vars, f.varsFile)
	if err != nil {
		return err
	}
	if env != "" {
		task.Environment = env
	}

	params := map[string]any{}
	set := func(flag, key string, value any) {
		if fl.Changed(flag) {
			params[key] = value
		}
	}
	set("check", "dry_run", f.check)
	set("diff", "diff", f.diff)
	set("debug-task", "debug", f.debugTask)
	if fl.Changed("debug-level") {
		params["debug"] = true
		params["debug_level"] = f.debugLevel
	}
	set("limit", "limit", f.limit)
	set("tags", "tags", f.tags)
	set("skip-tags", "skip_tags", f.skipTags)
	set("skip-galaxy", "skip_galaxy_install", f.skipGalaxy)
	set("plan", "plan", f.plan)
	set("destroy", "destroy", f.destroy)
	set("auto-approve", "auto_approve", f.autoApprove)
	set("upgrade", "upgrade", f.upgrade)
	set("reconfigure", "reconfigure", f.reconfigure)
	if len(params) > 0 {
		if task.Params == nil {
			task.Params = db.MapStringAnyField{}
		}
		for k, v := range params {
			task.Params[k] = v
		}
	}
	return nil
}

// buildExtraVars merges --extra-vars-file and --var NAME=VALUE into the JSON
// the server stores as the task environment. Later --var wins over the file.
func buildExtraVars(cmd *cobra.Command, vars []string, file string) (string, error) {
	if len(vars) == 0 && file == "" {
		return "", nil
	}
	merged := map[string]any{}
	if file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("read extra vars: %w", err)
		}
		var fromFile map[string]any
		if err := yaml.Unmarshal(data, &fromFile); err != nil {
			return "", usage(cmd.CommandPath(), "%s is neither JSON nor YAML with a top-level mapping: %v", file, err)
		}
		for k, v := range fromFile {
			merged[k] = v
		}
	}
	for _, kv := range vars {
		name, value, ok := strings.Cut(kv, "=")
		if !ok || name == "" {
			return "", usage(cmd.CommandPath(), "--var expects NAME=VALUE, got %q", kv)
		}
		merged[name] = value
	}
	out, err := json.Marshal(merged)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// startAndMaybeWait creates the task and, depending on the flags, waits for
// it. Without --wait the id is printed and the command returns at once.
func (f *runFlags) startAndMaybeWait(cmd *cobra.Command, r *runtime, api *apiclient.Client, projectID int, task db.Task, what string) error {
	ctx, cancel := r.apiCtx(cmd)
	created, err := api.RunTask(ctx, projectID, task)
	cancel()
	if err != nil {
		return err
	}
	r.warnf("Task %d created (%s).", created.ID, what)

	wait := f.wait || f.follow
	if !wait {
		if r.printer.Options().Format == output.FormatTable {
			r.printer.Printf("%d\n", created.ID)
			return nil
		}
		return r.printer.Print(created, taskColumns())
	}

	final, err := r.waitTask(cmd, api, projectID, created.ID, waitOptions{
		follow: f.follow, timeout: f.waitTimeout, stopOnInterrupt: f.stopOnInterrupt,
	})
	if err != nil {
		return err
	}
	r.warnf("Task %d finished: %s.", final.ID, final.Status)
	if !f.follow {
		if err := r.printer.Print(final, taskColumns()); err != nil {
			return err
		}
	}
	return finishedError(final)
}

// newTemplateRunCmd is `template run <id|name>`.
func newTemplateRunCmd(tpl resource[db.Template]) *cobra.Command {
	f := &runFlags{}
	cmd := &cobra.Command{
		Use:   "run <id|name>",
		Short: "Start a task from a template",
		Long: `Start a task from a template. By default the task id is printed and the command
returns at once; --wait blocks until the task finishes and sets the exit code
from its status (0 success, 3 error, 6 stopped); --follow also streams the log.

Overrides (--limit, --tags, --check, --branch, --inventory, …) apply only when
the template allows them; otherwise the server rejects the task. Ctrl+C while
waiting only disconnects the client — the task goes on — unless
--stop-on-interrupt is given.`,
		Example: `  # Fire and forget: prints the task id
  semaphore client template run deploy -p 1

  # Run and follow the log; exit 0 only if the task succeeded
  semaphore client template run deploy -p 1 --follow

  # Check mode on two hosts with extra variables, give up after 30 minutes
  semaphore client template run deploy --check --limit web1,web2 --var version=1.4.2 --wait --wait-timeout 30m`,
		Args: cobra.ExactArgs(1),
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
			task := db.Task{TemplateID: t.ID}
			if err := f.apply(cmd, r, api, projectID, &task); err != nil {
				return err
			}
			return f.startAndMaybeWait(cmd, r, api, projectID, task, "template "+t.Name)
		},
	}
	f.bindWait(cmd)
	f.bindOverrides(cmd)
	return cmd
}

// newTaskCmd is the `task` tree: the lifecycle of tasks that exist already.
func newTaskCmd(tpl resource[db.Template]) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "task",
		Aliases: []string{"tasks"},
		Short:   "Inspect, follow, stop and rerun tasks",
		Long: `Work with the tasks of a project: list and inspect them, print or follow
their logs, wait for them, stop, confirm, reject or rerun them. New tasks are
started with 'template run'. Tasks are addressed by numeric id.

Waiting commands exit with 0 when the task ends in 'success', 3 on 'error',
6 on 'stopped' and 124 when --wait-timeout runs out (the task keeps running).`,
		Example: `  semaphore client task list -p 1 --status error
  semaphore client task logs 42 -p 1 --follow
  semaphore client task wait 42 -p 1 --wait-timeout 30m`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	cmd.AddCommand(
		newTaskListCmd(tpl), newTaskGetCmd(), newTaskLogsCmd(), newTaskWatchCmd(), newTaskWaitCmd(),
		newTaskStopCmd(), newTaskConfirmCmd(), newTaskRejectCmd(), newTaskRerunCmd(),
	)
	return cmd
}

// taskCommandSetup is the preamble every task command shares.
func taskCommandSetup(cmd *cobra.Command) (*runtime, *apiclient.Client, int, error) {
	r := rt(cmd)
	api, err := r.client(cmd)
	if err != nil {
		return nil, nil, 0, err
	}
	if err := r.requireToken(cmd); err != nil {
		return nil, nil, 0, err
	}
	projectID, err := r.resolveProject(cmd, api)
	if err != nil {
		return nil, nil, 0, err
	}
	return r, api, projectID, nil
}

func newTaskListCmd(tpl resource[db.Template]) *cobra.Command {
	var (
		template string
		status   string
		limit    int
		all      bool
	)
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List recent tasks of the project, newest first",
		Long: `List the recent tasks of the project, newest first. --template and --status
filter the list; --limit bounds the number of rows and --all lifts it.`,
		Example: `  semaphore client task list -p 1
  semaphore client task list -p 1 --template deploy --status success --limit 10
  semaphore client task list -p 1 -o json -q`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			r, api, projectID, err := taskCommandSetup(cmd)
			if err != nil {
				return err
			}
			templateID := 0
			if template != "" {
				t, err := tpl.resolve(cmd, r, api, projectID, template)
				if err != nil {
					return err
				}
				templateID = t.ID
			}
			if all {
				limit = 0
			}
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			tasks, err := api.ListAllTasks(ctx, projectID, templateID, limit)
			if err != nil {
				return err
			}
			if status != "" {
				filtered := make([]db.TaskWithTpl, 0, len(tasks))
				for _, t := range tasks {
					if strings.EqualFold(string(t.Status), status) {
						filtered = append(filtered, t)
					}
				}
				tasks = filtered
			}
			return r.printer.Print(tasks, taskColumns())
		},
	}
	cmd.Flags().StringVar(&template, "template", "", "Only tasks of this template (id or name)")
	cmd.Flags().StringVar(&status, "status", "", "Only tasks with this status (waiting, running, success, error, stopped, …)")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of tasks to fetch")
	cmd.Flags().BoolVar(&all, "all", false, "Fetch every task (ignores --limit)")
	return cmd
}

func newTaskGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "get <id>",
		Aliases: []string{"show"},
		Short:   "Show one task",
		Long:    `Show one task with all its fields: status, template, who started it, when it ran, its parameters.`,
		Example: `  semaphore client task get 42 -p 1 -o json`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, api, projectID, err := taskCommandSetup(cmd)
			if err != nil {
				return err
			}
			id, err := parseID(cmd, "task", args[0])
			if err != nil {
				return err
			}
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			task, err := api.GetTask(ctx, projectID, id)
			if err != nil {
				return err
			}
			return r.printer.Print(task, taskColumns())
		},
	}
}

func newTaskLogsCmd() *cobra.Command {
	var (
		follow bool
		raw    bool
	)
	cmd := &cobra.Command{
		Use:   "logs <id>",
		Short: "Print the log of a task",
		Long: `Print the log of a task. --raw streams the plain-text log the server keeps
(ANSI codes stripped); otherwise the stored lines are printed, or listed as
records with -o json. With --follow new lines are printed until the task
finishes and the exit code follows its status.`,
		Example: `  semaphore client task logs 42
  semaphore client task logs 42 -f
  semaphore client task logs 42 --raw > task-42.log`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, api, projectID, err := taskCommandSetup(cmd)
			if err != nil {
				return err
			}
			id, err := parseID(cmd, "task", args[0])
			if err != nil {
				return err
			}
			if follow {
				final, err := r.waitTask(cmd, api, projectID, id, waitOptions{follow: true})
				if err != nil {
					return err
				}
				return finishedError(final)
			}
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			if raw {
				return api.GetTaskRawOutput(ctx, projectID, id, r.io.Out)
			}
			lines, err := api.GetTaskOutput(ctx, projectID, id)
			if err != nil {
				return err
			}
			if r.printer.Options().Format != output.FormatTable {
				return r.printer.Print(lines, output.Columns("time", "output"))
			}
			for _, line := range lines {
				r.printLogLine(line.Output)
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Keep printing new lines until the task finishes")
	cmd.Flags().BoolVar(&raw, "raw", false, "Stream the plain-text log as the server stores it")
	return cmd
}

func newTaskWatchCmd() *cobra.Command {
	var (
		timeout    time.Duration
		exitStatus bool
	)
	cmd := &cobra.Command{
		Use:   "watch <id>",
		Short: "Show the status of a task until it finishes",
		Long: `Show the status of a task as it changes, until the task finishes. By default
the exit code is 0 whatever the outcome; with --exit-status it follows the task
(3 error, 6 stopped), like 'gh run watch --exit-status'.`,
		Example: `  semaphore client task watch 42 -p 1
  semaphore client task watch 42 -p 1 --exit-status --wait-timeout 1h`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, api, projectID, err := taskCommandSetup(cmd)
			if err != nil {
				return err
			}
			id, err := parseID(cmd, "task", args[0])
			if err != nil {
				return err
			}
			final, err := r.waitTask(cmd, api, projectID, id, waitOptions{
				timeout: timeout,
				onStatus: func(t db.TaskWithTpl) {
					r.printer.Printf("%s\ttask %d\t%s\n", time.Now().Format(time.TimeOnly), t.ID, t.Status)
				},
			})
			if err != nil {
				return err
			}
			if exitStatus {
				return finishedError(final)
			}
			return nil
		},
	}
	cmd.Flags().DurationVar(&timeout, "wait-timeout", 0, "Give up after this long, exit 124 (default: no limit)")
	cmd.Flags().BoolVar(&exitStatus, "exit-status", false, "Exit with 3 or 6 when the task failed or was stopped")
	return cmd
}

func newTaskWaitCmd() *cobra.Command {
	var timeout time.Duration
	cmd := &cobra.Command{
		Use:   "wait <id>",
		Short: "Wait silently until a task finishes; exit code follows its status",
		Long: `Wait until a task finishes without printing anything, for scripts. Exit code
0 on success, 3 on error, 6 when the task was stopped, 124 when --wait-timeout
runs out (the task keeps running on the server).`,
		Example: `  semaphore client task wait 42; case $? in 0) echo ok;; 3) echo failed;; 6) echo stopped;; 124) echo timeout;; esac`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, api, projectID, err := taskCommandSetup(cmd)
			if err != nil {
				return err
			}
			id, err := parseID(cmd, "task", args[0])
			if err != nil {
				return err
			}
			final, err := r.waitTask(cmd, api, projectID, id, waitOptions{timeout: timeout})
			if err != nil {
				return err
			}
			return finishedError(final)
		},
	}
	cmd.Flags().DurationVar(&timeout, "wait-timeout", 0, "Give up after this long, exit 124 (default: no limit)")
	return cmd
}

func newTaskStopCmd() *cobra.Command {
	var (
		force bool
		wait  bool
	)
	cmd := &cobra.Command{
		Use:   "stop <id>",
		Short: "Stop a running task",
		Long: `Ask the server to stop a task. Without --force the process gets a chance to
end gracefully; --force kills it at once. --wait blocks until the task has
actually stopped.`,
		Example: `  semaphore client task stop 42 -p 1
  semaphore client task stop 42 -p 1 --force --wait`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, api, projectID, err := taskCommandSetup(cmd)
			if err != nil {
				return err
			}
			id, err := parseID(cmd, "task", args[0])
			if err != nil {
				return err
			}
			ctx, cancel := r.apiCtx(cmd)
			err = api.StopTask(ctx, projectID, id, force)
			cancel()
			if err != nil {
				return err
			}
			r.warnf("Stop requested for task %d.", id)
			if !wait {
				return nil
			}
			final, err := r.waitTask(cmd, api, projectID, id, waitOptions{})
			if err != nil {
				return err
			}
			r.warnf("Task %d finished: %s.", final.ID, final.Status)
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Kill the task process immediately")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait until the task has stopped")
	return cmd
}

func newTaskConfirmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "confirm <id>",
		Short: "Let a task that waits for confirmation proceed",
		Long: `Approve a task whose template requires a confirmation before it runs (status
'waiting_confirmation'). The task is queued and runs as usual afterwards.`,
		Example: `  semaphore client task confirm 42 -p 1`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, api, projectID, err := taskCommandSetup(cmd)
			if err != nil {
				return err
			}
			id, err := parseID(cmd, "task", args[0])
			if err != nil {
				return err
			}
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			if err := api.ConfirmTask(ctx, projectID, id); err != nil {
				return err
			}
			r.warnf("Task %d confirmed.", id)
			return nil
		},
	}
}

func newTaskRejectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reject <id>",
		Short: "Refuse a task that waits for confirmation",
		Long: `Refuse a task whose template requires a confirmation before it runs (status
'waiting_confirmation'). The task does not run.`,
		Example: `  semaphore client task reject 42 -p 1`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, api, projectID, err := taskCommandSetup(cmd)
			if err != nil {
				return err
			}
			id, err := parseID(cmd, "task", args[0])
			if err != nil {
				return err
			}
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			if err := api.RejectTask(ctx, projectID, id); err != nil {
				return err
			}
			r.warnf("Task %d rejected.", id)
			return nil
		},
	}
}

func newTaskRerunCmd() *cobra.Command {
	f := &runFlags{}
	cmd := &cobra.Command{
		Use:   "rerun <id>",
		Short: "Start a new task with the parameters of an existing one",
		Long: `Start a new task with the template, inventory, branch, variables and
parameters of an existing task. Overrides given on the command line replace
the copied values. --wait and --follow work as for 'template run'.`,
		Example: `  semaphore client task rerun 42 -p 1
  semaphore client task rerun 42 -p 1 --branch hotfix --follow`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, api, projectID, err := taskCommandSetup(cmd)
			if err != nil {
				return err
			}
			id, err := parseID(cmd, "task", args[0])
			if err != nil {
				return err
			}
			ctx, cancel := r.apiCtx(cmd)
			src, err := api.GetTask(ctx, projectID, id)
			cancel()
			if err != nil {
				return err
			}
			task := db.Task{
				TemplateID:  src.TemplateID,
				InventoryID: src.InventoryID,
				GitBranch:   src.GitBranch,
				Arguments:   src.Arguments,
				Environment: src.Environment,
				Params:      src.Params,
				Message:     src.Message,
			}
			if err := f.apply(cmd, r, api, projectID, &task); err != nil {
				return err
			}
			return f.startAndMaybeWait(cmd, r, api, projectID, task, fmt.Sprintf("rerun of task %d", id))
		},
	}
	f.bindWait(cmd)
	f.bindOverrides(cmd)
	return cmd
}
