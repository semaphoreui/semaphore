// Package artifacts implements AWX-style cross-template variable passing for
// Semaphore workflows. An upstream task may write a JSON object to the file
// pointed to by the SEMAPHORE_ARTIFACTS_FILE environment variable; that object
// is persisted on the task row and then merged into the extra-vars / env of
// every downstream task in the same WorkflowRun.
package artifacts

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/semaphoreui/semaphore/db"
)

// MaxSize is the maximum size in bytes accepted for an artifacts file.
// Files larger than this are rejected and an error is returned to the caller
// so it can log a warning to the task output.
const MaxSize = 256 * 1024

// keyPattern enforces a conservative key whitelist so artifact names can be
// safely projected as Ansible variables and shell environment variables.
var keyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// reservedKeys lists top-level keys that must not be overridden by artifacts.
var reservedKeys = map[string]struct{}{
	"semaphore_vars":               {},
	"semaphore_workflow_artifacts": {},
	"task_details":                 {},
	"incoming_version":             {},
}

// EnvVarPrefix is the prefix used when exporting scalar artifact values as
// shell environment variables.
const EnvVarPrefix = "SEMAPHORE_WF_"

// ErrInvalidArtifacts indicates the artifacts file could not be parsed or did
// not satisfy the schema (must be a JSON object with whitelisted keys).
var ErrInvalidArtifacts = errors.New("invalid artifacts payload")

// ErrTooLarge indicates the artifacts file exceeded MaxSize.
var ErrTooLarge = errors.New("artifacts file is too large")

// LoadFile reads, validates and returns the JSON object stored at path.
// Returns (nil, nil, nil) if the file does not exist (the common no-artifacts case).
func LoadFile(path string) (map[string]any, []byte, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, nil, err
	}
	if stat.Size() > MaxSize {
		return nil, nil, fmt.Errorf("%w: %d bytes (limit %d)", ErrTooLarge, stat.Size(), MaxSize)
	}

	raw, err := io.ReadAll(io.LimitReader(f, MaxSize+1))
	if err != nil {
		return nil, nil, err
	}
	if int64(len(raw)) > MaxSize {
		return nil, nil, fmt.Errorf("%w: limit %d", ErrTooLarge, MaxSize)
	}

	return Parse(raw)
}

// Parse validates raw JSON bytes as an artifacts object and returns the
// canonical representation alongside a re-serialized JSON blob suitable for
// persistence. Empty/whitespace-only inputs produce (nil, nil, nil).
func Parse(raw []byte) (map[string]any, []byte, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return nil, nil, nil
	}

	var decoded any
	if err := json.Unmarshal([]byte(trimmed), &decoded); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrInvalidArtifacts, err)
	}

	obj, ok := decoded.(map[string]any)
	if !ok {
		return nil, nil, fmt.Errorf("%w: top-level value must be a JSON object", ErrInvalidArtifacts)
	}

	cleaned := make(map[string]any, len(obj))
	for k, v := range obj {
		if !keyPattern.MatchString(k) {
			return nil, nil, fmt.Errorf("%w: invalid key %q", ErrInvalidArtifacts, k)
		}
		if _, reserved := reservedKeys[k]; reserved {
			return nil, nil, fmt.Errorf("%w: key %q is reserved", ErrInvalidArtifacts, k)
		}
		cleaned[k] = v
	}

	if len(cleaned) == 0 {
		return nil, nil, nil
	}

	out, err := json.Marshal(cleaned)
	if err != nil {
		return nil, nil, err
	}
	return cleaned, out, nil
}

// Merge merges b on top of a. Both inputs may be nil. Later values win, mirroring
// AWX's set_stats merge semantics.
func Merge(a, b map[string]any) map[string]any {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}
	out := make(map[string]any, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

// CollectFromTasks merges artifacts from a slice of tasks ordered by creation
// time. The optional excludeTaskID is skipped (used to avoid feeding a task its
// own artifacts back into itself).
func CollectFromTasks(tasks []db.Task, excludeTaskID *int) map[string]any {
	// Defensive: ensure deterministic ordering by ID even if the caller did
	// not sort. Lower IDs come first; later (newer) tasks override.
	sorted := make([]db.Task, len(tasks))
	copy(sorted, tasks)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID })

	var merged map[string]any
	for _, task := range sorted {
		if excludeTaskID != nil && task.ID == *excludeTaskID {
			continue
		}
		if task.Artifacts == nil || *task.Artifacts == "" {
			continue
		}
		obj, _, err := Parse([]byte(*task.Artifacts))
		if err != nil || obj == nil {
			continue
		}
		merged = Merge(merged, obj)
	}
	return merged
}

// ToShellEnv converts an artifacts map to shell env entries (KEY=value) using
// the SEMAPHORE_WF_ prefix. Only scalar values are exported; nested objects
// and arrays are skipped (Ansible playbooks can still consume them as extra
// vars). Keys are uppercased and sanitized.
func ToShellEnv(m map[string]any) []string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]string, 0, len(keys))
	for _, k := range keys {
		envKey := EnvVarPrefix + sanitizeEnvKey(k)
		if envKey == EnvVarPrefix {
			continue
		}
		val, ok := scalarString(m[k])
		if !ok {
			continue
		}
		out = append(out, fmt.Sprintf("%s=%s", envKey, val))
	}
	return out
}

func sanitizeEnvKey(k string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(k) {
		switch {
		case r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}

func scalarString(v any) (string, bool) {
	switch t := v.(type) {
	case string:
		return t, true
	case bool:
		return strconv.FormatBool(t), true
	case float64:
		// json.Unmarshal decodes all numbers as float64 by default. Render
		// integers without a decimal point so set_stats counters look right.
		if t >= math.MinInt64 && t <= math.MaxInt64 && t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10), true
		}
		return strconv.FormatFloat(t, 'f', -1, 64), true
	case json.Number:
		return t.String(), true
	case nil:
		return "", false
	default:
		return "", false
	}
}
