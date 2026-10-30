package client

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/semaphoreui/semaphore/pkg/apiclient"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// applyFunc copies the flags a user passed into item. It is called after
// --from-file, so flags win over the file.
type applyFunc[T any] func(cmd *cobra.Command, r *runtime, api *apiclient.Client, projectID int, item *T) error

// mutable adds create, update and delete to a resource.
type mutable[T any] struct {
	res    resource[T]
	create func(ctx context.Context, api *apiclient.Client, projectID int, item T) (T, error)
	update func(ctx context.Context, api *apiclient.Client, projectID int, item T) error
	remove func(ctx context.Context, api *apiclient.Client, projectID, id int) error
	// bind registers the resource's flags on cmd and returns how to apply them.
	bind func(cmd *cobra.Command) applyFunc[T]
	// validate runs before create: required fields and the like.
	validate func(cmd *cobra.Command, item T) error
	name     func(T) string
	// createExample and updateExample are shown in --help.
	createExample string
	updateExample string
}

// commands returns create, update and delete; nil functions skip a verb.
func (m mutable[T]) commands() []*cobra.Command {
	var cmds []*cobra.Command
	if m.create != nil {
		cmds = append(cmds, m.createCmd())
	}
	if m.update != nil {
		cmds = append(cmds, m.updateCmd())
	}
	if m.remove != nil {
		cmds = append(cmds, m.deleteCmd())
	}
	return cmds
}

// setup is the preamble of every mutating command.
func (m mutable[T]) setup(cmd *cobra.Command) (*runtime, *apiclient.Client, int, error) {
	r := rt(cmd)
	api, err := r.client(cmd)
	if err != nil {
		return nil, nil, 0, err
	}
	if err := r.requireToken(cmd); err != nil {
		return nil, nil, 0, err
	}
	projectID, err := m.res.projectFor(cmd, r, api)
	if err != nil {
		return nil, nil, 0, err
	}
	return r, api, projectID, nil
}

func (m mutable[T]) createCmd() *cobra.Command {
	var fromFile string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a " + m.res.kind,
		Long: fmt.Sprintf(`Create a %s from flags, from a JSON or YAML file (--from-file, same field
names as the API), or both: flags override the file. The created object is
printed; with -q only its id.`, m.res.kind),
		Example: m.createExample,
		Args:    cobra.NoArgs,
	}
	apply := m.bind(cmd)
	cmd.Flags().StringVar(&fromFile, "from-file", "", "JSON or YAML file with the object (API field names)")
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		r, api, projectID, err := m.setup(cmd)
		if err != nil {
			return err
		}
		var item T
		if err := loadFromFile(cmd, fromFile, &item); err != nil {
			return err
		}
		if err := apply(cmd, r, api, projectID, &item); err != nil {
			return err
		}
		if m.validate != nil {
			if err := m.validate(cmd, item); err != nil {
				return err
			}
		}
		ctx, cancel := r.apiCtx(cmd)
		defer cancel()
		created, err := m.create(ctx, api, projectID, item)
		if err != nil {
			return err
		}
		r.warnf("%s %q created with id %d.", capitalize(m.res.kind), m.name(created), m.res.id(created))
		return r.printer.Print(m.res.viewOne(created), m.res.columns)
	}
	return cmd
}

func (m mutable[T]) updateCmd() *cobra.Command {
	var fromFile string
	cmd := &cobra.Command{
		Use:   "update <id|name>",
		Short: "Change a " + m.res.kind,
		Long: fmt.Sprintf(`Change fields of a %s. The object is read from the server first, the fields
from --from-file and then from the flags are applied, and the whole object is
written back — so nothing you did not mention is lost.`, m.res.kind),
		Example: m.updateExample,
		Args:    cobra.ExactArgs(1),
	}
	apply := m.bind(cmd)
	cmd.Flags().StringVar(&fromFile, "from-file", "", "JSON or YAML file with the fields to change (API field names)")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		r, api, projectID, err := m.setup(cmd)
		if err != nil {
			return err
		}
		item, err := m.res.resolve(cmd, r, api, projectID, args[0])
		if err != nil {
			return err
		}
		if err := loadFromFile(cmd, fromFile, &item); err != nil {
			return err
		}
		if err := apply(cmd, r, api, projectID, &item); err != nil {
			return err
		}
		ctx, cancel := r.apiCtx(cmd)
		defer cancel()
		if err := m.update(ctx, api, projectID, item); err != nil {
			return err
		}
		r.warnf("%s %q (id %d) updated.", capitalize(m.res.kind), m.name(item), m.res.id(item))
		return nil
	}
	return cmd
}

func (m mutable[T]) deleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <id|name>",
		Aliases: []string{"rm"},
		Short:   "Delete a " + m.res.kind,
		Long:    fmt.Sprintf("Delete a %s. Asks for confirmation in a terminal; pass --yes to skip it (required when stdin is not a terminal).", m.res.kind),
		Example: fmt.Sprintf("  semaphore client %[1]s delete %[2]q%[3]s -y", m.res.kind, m.res.exampleName(), m.res.projectFlag()),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, api, projectID, err := m.setup(cmd)
			if err != nil {
				return err
			}
			item, err := m.res.resolve(cmd, r, api, projectID, args[0])
			if err != nil {
				return err
			}
			id := m.res.id(item)
			if err := r.prompter.Confirm(cmd.Context(), fmt.Sprintf("Delete %s %q (id %d)?", m.res.kind, m.name(item), id)); err != nil {
				return err
			}
			ctx, cancel := r.apiCtx(cmd)
			defer cancel()
			if err := m.remove(ctx, api, projectID, id); err != nil {
				return err
			}
			r.warnf("%s %q (id %d) deleted.", capitalize(m.res.kind), m.name(item), id)
			return nil
		},
	}
}

// loadFromFile merges a JSON or YAML document into item. YAML goes through
// JSON so that the API field names (json tags) decide what lands where; keys
// absent from the file leave item untouched.
func loadFromFile(cmd *cobra.Command, path string, item any) error {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	var generic any
	if err := yaml.Unmarshal(data, &generic); err != nil {
		return usage(cmd.CommandPath(), "%s is neither valid JSON nor YAML: %v", path, err)
	}
	if _, ok := generic.(map[string]any); !ok {
		return usage(cmd.CommandPath(), "%s must contain one object (a mapping), not %T", path, generic)
	}
	asJSON, err := json.Marshal(generic)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(asJSON, item); err != nil {
		return usage(cmd.CommandPath(), "%s does not match the %s fields: %v", path, strings.TrimSuffix(cmd.Parent().Name(), "s"), err)
	}
	return nil
}

// textOrFile returns v itself when it looks like inline content (JSON object
// or array) and the content of the file v otherwise.
func textOrFile(v string) (string, error) {
	trimmed := strings.TrimSpace(v)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		return trimmed, nil
	}
	data, err := os.ReadFile(v)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", v, err)
	}
	return string(data), nil
}

// readSecretFile reads a secret from a file, trimming one trailing newline and
// nothing else (a private key keeps its internal newlines).
func readSecretFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read secret: %w", err)
	}
	return strings.TrimRight(string(data), "\r\n"), nil
}

// readSecretStdin reads the first line of stdin as a secret. It refuses an
// interactive terminal when the caller did not ask for a prompt, so a user
// does not end up typing a password into a hanging command by accident.
func (r *runtime) readSecretStdin(cmd *cobra.Command, what string) (string, error) {
	if r.interactive {
		return r.promptSecret(cmd, what+": ")
	}
	line, err := r.readLine(cmd.Context())
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(line) == "" {
		return "", usage(cmd.CommandPath(), "no %s on stdin", what)
	}
	return strings.TrimSpace(line), nil
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
