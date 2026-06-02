package artifacts

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/semaphoreui/semaphore/db"
)

func ptr[T any](v T) *T { return &v }

func TestParseValidObject(t *testing.T) {
	obj, raw, err := Parse([]byte(`{"foo": "bar", "count": 3}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj["foo"] != "bar" {
		t.Fatalf("expected foo=bar got %v", obj["foo"])
	}
	if !strings.Contains(string(raw), `"foo"`) {
		t.Fatalf("expected re-serialized JSON to contain key, got %q", raw)
	}
}

func TestParseRejectsNonObject(t *testing.T) {
	_, _, err := Parse([]byte(`[1,2,3]`))
	if !errors.Is(err, ErrInvalidArtifacts) {
		t.Fatalf("expected ErrInvalidArtifacts, got %v", err)
	}
}

func TestParseRejectsBadKey(t *testing.T) {
	_, _, err := Parse([]byte(`{"bad-key": 1}`))
	if !errors.Is(err, ErrInvalidArtifacts) {
		t.Fatalf("expected ErrInvalidArtifacts for bad key, got %v", err)
	}
}

func TestParseRejectsReservedKey(t *testing.T) {
	_, _, err := Parse([]byte(`{"semaphore_vars": {}}`))
	if !errors.Is(err, ErrInvalidArtifacts) {
		t.Fatalf("expected ErrInvalidArtifacts for reserved key, got %v", err)
	}
}

func TestParseEmptyReturnsNil(t *testing.T) {
	obj, raw, err := Parse([]byte(`   `))
	if err != nil || obj != nil || raw != nil {
		t.Fatalf("expected (nil,nil,nil) for empty input, got obj=%v raw=%v err=%v", obj, raw, err)
	}
}

func TestMergePrecedence(t *testing.T) {
	a := map[string]any{"x": 1, "y": 2}
	b := map[string]any{"y": 99, "z": 3}
	got := Merge(a, b)
	want := map[string]any{"x": 1, "y": 99, "z": 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("merge precedence wrong: got %v want %v", got, want)
	}
}

func TestCollectFromTasksOrdering(t *testing.T) {
	tasks := []db.Task{
		{ID: 2, Artifacts: ptr(`{"k":"second"}`)},
		{ID: 1, Artifacts: ptr(`{"k":"first","only_in_first":1}`)},
		{ID: 3, Artifacts: ptr(`{"k":"third"}`)},
	}
	merged := CollectFromTasks(tasks, nil)
	if merged["k"] != "third" {
		t.Fatalf("expected newest task to win, got %v", merged["k"])
	}
	if merged["only_in_first"] != float64(1) {
		t.Fatalf("expected non-overlapping keys to be preserved, got %v", merged)
	}
}

func TestCollectFromTasksExcludesCurrent(t *testing.T) {
	tasks := []db.Task{
		{ID: 1, Artifacts: ptr(`{"k":"first"}`)},
		{ID: 2, Artifacts: ptr(`{"k":"second"}`)},
	}
	merged := CollectFromTasks(tasks, ptr(2))
	if merged["k"] != "first" {
		t.Fatalf("expected current task to be excluded, got %v", merged)
	}
}

func TestCollectSkipsInvalidArtifacts(t *testing.T) {
	tasks := []db.Task{
		{ID: 1, Artifacts: ptr(`not json`)},
		{ID: 2, Artifacts: ptr(`{"k":"ok"}`)},
	}
	merged := CollectFromTasks(tasks, nil)
	if merged["k"] != "ok" {
		t.Fatalf("expected invalid artifacts to be skipped, got %v", merged)
	}
}

func TestToShellEnvScalarsOnly(t *testing.T) {
	m := map[string]any{
		"hello":   "world",
		"count":   float64(42),
		"ratio":   1.5,
		"flag":    true,
		"nested":  map[string]any{"x": 1},
		"arr":     []any{1, 2},
		"weird@k": "ignored", // would be filtered by Parse, but test sanitizer
	}
	got := ToShellEnv(m)
	sort.Strings(got)

	mustContain := []string{
		"SEMAPHORE_WF_HELLO=world",
		"SEMAPHORE_WF_COUNT=42",
		"SEMAPHORE_WF_RATIO=1.5",
		"SEMAPHORE_WF_FLAG=true",
	}
	for _, want := range mustContain {
		found := false
		for _, e := range got {
			if e == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing %q in env output: %v", want, got)
		}
	}

	for _, e := range got {
		if strings.HasPrefix(e, "SEMAPHORE_WF_NESTED=") || strings.HasPrefix(e, "SEMAPHORE_WF_ARR=") {
			t.Fatalf("nested/array values should not be exported: %q", e)
		}
	}
}

func TestLoadFileMissingReturnsNil(t *testing.T) {
	obj, raw, err := LoadFile(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil || obj != nil || raw != nil {
		t.Fatalf("expected (nil,nil,nil) for missing file, got obj=%v raw=%v err=%v", obj, raw, err)
	}
}

func TestLoadFileTooLarge(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "artifacts.json")
	big := strings.Repeat("a", MaxSize+10)
	if err := os.WriteFile(path, []byte(`{"k":"`+big+`"}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, _, err := LoadFile(path)
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("expected ErrTooLarge, got %v", err)
	}
}

func TestLoadFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "artifacts.json")
	if err := os.WriteFile(path, []byte(`{"version":"1.2.3","ready":true}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	obj, raw, err := LoadFile(path)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if obj["version"] != "1.2.3" || obj["ready"] != true {
		t.Fatalf("unexpected obj: %v", obj)
	}
	if len(raw) == 0 {
		t.Fatal("expected re-serialized blob")
	}
}
