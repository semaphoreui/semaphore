package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/semaphoreui/semaphore/pkg/apiclient"
	"github.com/spf13/cobra"
)

// NotFoundError says no object of Kind matched the id or name the user gave.
// errors.Is(err, apiclient.ErrNotFound) holds, so it exits with 5 like a 404.
type NotFoundError struct {
	Kind string
	Name string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s %q not found", e.Kind, e.Name)
}

// Is makes a NotFoundError match apiclient.ErrNotFound.
func (e *NotFoundError) Is(target error) bool { return target == apiclient.ErrNotFound }

// resolveProject turns the configured project (flag, env or context) into an
// id, looking a name up in the user's projects. The result is cached for the
// invocation.
func (r *runtime) resolveProject(cmd *cobra.Command, api *apiclient.Client) (int, error) {
	if r.projectID != 0 {
		return r.projectID, nil
	}
	if r.cfg.Project == "" {
		return 0, usage(cmd.CommandPath(), "project is not set; pass --project/-p <id|name>, set SEMAPHORE_CLIENT_PROJECT or save it in the context with 'semaphore client context set <name> -p <id>'")
	}
	if id, err := strconv.Atoi(r.cfg.Project); err == nil && id > 0 {
		r.projectID = id
		return id, nil
	}

	ctx, cancel := withTimeout(cmd, r.cfg.Timeout)
	defer cancel()
	projects, err := api.ListProjects(ctx)
	if err != nil {
		return 0, err
	}
	var matches []int
	for _, p := range projects {
		if strings.EqualFold(p.Name, r.cfg.Project) {
			matches = append(matches, p.ID)
		}
	}
	switch len(matches) {
	case 0:
		return 0, &NotFoundError{Kind: "project", Name: r.cfg.Project}
	case 1:
		r.projectID = matches[0]
		return matches[0], nil
	default:
		return 0, &AmbiguousError{Kind: "project", Name: r.cfg.Project, Candidates: intsToStrings(matches)}
	}
}

// pickByIDOrName resolves arg to one item: a positive integer is an id, any
// other text is matched against match (case-insensitive by the caller). Zero
// matches is a NotFoundError, several are an AmbiguousError that names the
// candidates so the user can retry with an id.
func pickByIDOrName[T any](kind, arg string, items []T, id func(T) int, match func(T, string) bool, label func(T) string) (T, error) {
	var zero T
	if n, err := strconv.Atoi(arg); err == nil && n > 0 {
		for _, it := range items {
			if id(it) == n {
				return it, nil
			}
		}
		return zero, &NotFoundError{Kind: kind, Name: arg}
	}

	var found []T
	for _, it := range items {
		if match(it, arg) {
			found = append(found, it)
		}
	}
	switch len(found) {
	case 0:
		return zero, &NotFoundError{Kind: kind, Name: arg}
	case 1:
		return found[0], nil
	default:
		candidates := make([]string, 0, len(found))
		for _, it := range found {
			candidates = append(candidates, label(it))
		}
		return zero, &AmbiguousError{Kind: kind, Name: arg, Candidates: candidates}
	}
}

// parseID accepts only a positive integer; for commands that take ids alone.
func parseID(cmd *cobra.Command, kind, arg string) (int, error) {
	n, err := strconv.Atoi(arg)
	if err != nil || n <= 0 {
		return 0, usage(cmd.CommandPath(), "%s id must be a positive number, got %q", kind, arg)
	}
	return n, nil
}

func intsToStrings(ids []int) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, strconv.Itoa(id))
	}
	return out
}

// apiCtx is the per-request context with the configured timeout.
func (r *runtime) apiCtx(cmd *cobra.Command) (context.Context, context.CancelFunc) {
	return withTimeout(cmd, r.cfg.Timeout)
}
