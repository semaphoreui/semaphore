package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// taskListServer serves /api/project/1/tasks/last with keyset pagination over
// task ids total..1, the way api/projects/tasks.go does.
func taskListServer(t *testing.T, total int) (*httptest.Server, *[]string) {
	t.Helper()
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.URL.RequestURI())
		require.Equal(t, "/api/project/1/tasks/last", r.URL.Path)

		count, _ := strconv.Atoi(r.URL.Query().Get("count"))
		if count <= 0 || count > MaxTasksPage {
			count = MaxTasksPage
		}
		start := total
		if b := r.URL.Query().Get("before"); b != "" {
			before, _ := strconv.Atoi(b)
			start = before - 1
		}

		var page []db.TaskWithTpl
		for id := start; id >= 1 && len(page) < count; id-- {
			page = append(page, db.TaskWithTpl{Task: db.Task{ID: id}})
		}
		if len(page) > 0 && page[len(page)-1].ID > 1 {
			w.Header().Set("X-Has-Next", "true")
		}
		_ = json.NewEncoder(w).Encode(page)
	}))
	return srv, &calls
}

func TestListTasks_PageAndCursor(t *testing.T) {
	srv, calls := taskListServer(t, 5)
	defer srv.Close()
	c := newTestClient(t, srv, nil)

	page, err := c.ListTasks(context.Background(), 1, 0, 2, 0)
	require.NoError(t, err)
	assert.Len(t, page.Tasks, 2)
	assert.True(t, page.HasNext)
	assert.Equal(t, 4, page.Before())
	assert.Equal(t, "/api/project/1/tasks/last?count=2", (*calls)[0])

	page, err = c.ListTasks(context.Background(), 1, 0, 2, page.Before())
	require.NoError(t, err)
	assert.Equal(t, []int{3, 2}, ids(page.Tasks))
	assert.Equal(t, "/api/project/1/tasks/last?before=4&count=2", (*calls)[1])
}

func TestListTasks_TemplateScope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/project/1/templates/9/tasks/last", r.URL.Path)
		_, _ = w.Write([]byte("[]"))
	}))
	defer srv.Close()
	c := newTestClient(t, srv, nil)

	page, err := c.ListTasks(context.Background(), 1, 9, 0, 0)
	require.NoError(t, err)
	assert.Empty(t, page.Tasks)
	assert.False(t, page.HasNext)
	assert.Equal(t, 0, page.Before())
}

func TestListAllTasks(t *testing.T) {
	tests := []struct {
		name      string
		total     int
		limit     int
		wantIDs   []int
		wantCalls int
	}{
		{"everything over several pages", 450, 0, seq(450, 1), 3},
		{"limit inside first page", 450, 10, seq(450, 441), 1},
		{"limit across pages", 450, 250, seq(450, 201), 2},
		{"empty history", 0, 0, nil, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, calls := taskListServer(t, tt.total)
			defer srv.Close()
			c := newTestClient(t, srv, nil)

			tasks, err := c.ListAllTasks(context.Background(), 1, 0, tt.limit)
			require.NoError(t, err)
			assert.Equal(t, tt.wantIDs, ids(tasks))
			assert.Len(t, *calls, tt.wantCalls)
		})
	}
}

func TestTaskActions(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotMethod = r.URL.Path, r.Method
		gotBody = nil
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	c := newTestClient(t, srv, nil)
	ctx := context.Background()

	require.NoError(t, c.StopTask(ctx, 1, 5, true))
	assert.Equal(t, "POST /api/project/1/tasks/5/stop", gotMethod+" "+gotPath)
	assert.Equal(t, map[string]any{"force": true}, gotBody)

	require.NoError(t, c.ConfirmTask(ctx, 1, 5))
	assert.Equal(t, "POST /api/project/1/tasks/5/confirm", gotMethod+" "+gotPath)

	require.NoError(t, c.RejectTask(ctx, 1, 5))
	assert.Equal(t, "POST /api/project/1/tasks/5/reject", gotMethod+" "+gotPath)

	require.NoError(t, c.DeleteTask(ctx, 1, 5))
	assert.Equal(t, "DELETE /api/project/1/tasks/5", gotMethod+" "+gotPath)
}

func TestTaskOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/project/1/tasks/5/output":
			_ = json.NewEncoder(w).Encode([]db.TaskOutput{{TaskID: 5, Output: "PLAY [all]"}})
		case "/api/project/1/tasks/5/raw_output":
			_, _ = w.Write([]byte("line 1\nline 2\n"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	c := newTestClient(t, srv, nil)

	lines, err := c.GetTaskOutput(context.Background(), 1, 5)
	require.NoError(t, err)
	require.Len(t, lines, 1)
	assert.Equal(t, "PLAY [all]", lines[0].Output)

	var raw bytes.Buffer
	require.NoError(t, c.GetTaskRawOutput(context.Background(), 1, 5, &raw))
	assert.Equal(t, "line 1\nline 2\n", raw.String())

	assert.ErrorIs(t, c.GetTaskRawOutput(context.Background(), 1, 6, &raw), ErrNotFound)
}

func ids(tasks []db.TaskWithTpl) []int {
	if len(tasks) == 0 {
		return nil
	}
	out := make([]int, 0, len(tasks))
	for _, tsk := range tasks {
		out = append(out, tsk.ID)
	}
	return out
}

// seq returns from, from-1, …, to (inclusive, descending).
func seq(from, to int) []int {
	var out []int
	for i := from; i >= to; i-- {
		out = append(out, i)
	}
	return out
}
