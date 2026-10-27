package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/semaphoreui/semaphore/cli/cmd/client/config"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// taskServer fakes project 7 with one template and a task whose status walks
// through `statuses`, one step per GET, while its log grows one line per GET.
type taskServer struct {
	mu       sync.Mutex
	statuses []task_logger.TaskStatus
	polls    int
	lines    []string
	created  []db.Task
	stops    []bool
	confirms int
	rejects  int
	// failPolls makes the first n GET task calls answer 502.
	failPolls int
	nextID    int
}

func newTaskServer(statuses ...task_logger.TaskStatus) *taskServer {
	return &taskServer{statuses: statuses, nextID: 42, lines: []string{"PLAY [all]", "TASK [ping]", "ok: [web1]", "PLAY RECAP"}}
}

func (s *taskServer) status() task_logger.TaskStatus {
	i := s.polls
	if i >= len(s.statuses) {
		i = len(s.statuses) - 1
	}
	return s.statuses[i]
}

func (s *taskServer) handler(t *testing.T) http.Handler {
	str := func(v string) *string { return &v }
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		if r.Header.Get("Authorization") != "Bearer tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		enc := func(v any) { _ = json.NewEncoder(w).Encode(v) }
		p := r.URL.Path
		switch {
		case p == "/api/project/7/templates":
			enc([]db.Template{{ID: 3, Name: "Deploy", App: "ansible"}})
		case p == "/api/project/7/templates/3":
			enc(db.Template{ID: 3, Name: "Deploy", App: "ansible"})
		case p == "/api/project/7/inventory":
			enc([]db.Inventory{{ID: 9, Name: "staging"}})
		case p == "/api/project/7/inventory/9":
			enc(db.Inventory{ID: 9, Name: "staging"})
		case p == "/api/project/7/tasks" && r.Method == http.MethodPost:
			var task db.Task
			require.NoError(t, json.NewDecoder(r.Body).Decode(&task))
			s.created = append(s.created, task)
			task.ID = s.nextID
			task.Status = task_logger.TaskWaitingStatus
			w.WriteHeader(http.StatusCreated)
			enc(task)
		case p == "/api/project/7/tasks/last":
			enc([]db.TaskWithTpl{
				{Task: db.Task{ID: 42, TemplateID: 3, Status: "success"}, TemplateAlias: "Deploy", UserName: str("admin")},
				{Task: db.Task{ID: 41, TemplateID: 3, Status: "error"}, TemplateAlias: "Deploy"},
			})
		case p == "/api/project/7/templates/3/tasks/last":
			enc([]db.TaskWithTpl{{Task: db.Task{ID: 42, TemplateID: 3, Status: "success"}, TemplateAlias: "Deploy"}})
		case p == "/api/project/7/tasks/42":
			if s.failPolls > 0 {
				s.failPolls--
				w.WriteHeader(http.StatusBadGateway)
				return
			}
			st := s.status()
			s.polls++
			inv := 9
			enc(db.TaskWithTpl{Task: db.Task{ID: 42, TemplateID: 3, Status: st, InventoryID: &inv, Environment: `{"a":"1"}`, Params: db.MapStringAnyField{"dry_run": true}, Message: "orig"}, TemplateAlias: "Deploy"})
		case p == "/api/project/7/tasks/42/output":
			n := min(s.polls, len(s.lines))
			out := make([]db.TaskOutput, 0, n)
			for i, line := range s.lines[:n] {
				out = append(out, db.TaskOutput{ID: i + 1, TaskID: 42, Output: line})
			}
			enc(out)
		case p == "/api/project/7/tasks/42/raw_output":
			_, _ = w.Write([]byte(strings.Join(s.lines, "\n") + "\n"))
		case p == "/api/project/7/tasks/42/stop":
			var body map[string]bool
			_ = json.NewDecoder(r.Body).Decode(&body)
			s.stops = append(s.stops, body["force"])
			w.WriteHeader(http.StatusNoContent)
		case p == "/api/project/7/tasks/42/confirm":
			s.confirms++
			w.WriteHeader(http.StatusNoContent)
		case p == "/api/project/7/tasks/42/reject":
			s.rejects++
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

func taskEnv(srv *httptest.Server) map[string]string {
	return map[string]string{config.EnvURL: srv.URL, config.EnvToken: "tok", config.EnvProject: "7"}
}

func TestTemplateRun_FireAndForget(t *testing.T) {
	s := newTaskServer(task_logger.TaskWaitingStatus)
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()

	out, errOut, err := run(t, taskEnv(srv), "", "template", "run", "Deploy")
	require.NoError(t, err)
	assert.Equal(t, "42\n", out)
	assert.Contains(t, errOut, "Task 42 created (template Deploy).")
	require.Len(t, s.created, 1)
	assert.Equal(t, 3, s.created[0].TemplateID)
	assert.Nil(t, s.created[0].Params, "no override flags → no params, template defaults apply")
	assert.Empty(t, s.created[0].Environment)

	out, _, err = run(t, taskEnv(srv), "", "template", "run", "3", "-o", "json")
	require.NoError(t, err)
	var task map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &task))
	assert.EqualValues(t, 42, task["id"])
}

func TestTemplateRun_Overrides(t *testing.T) {
	s := newTaskServer(task_logger.TaskWaitingStatus)
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()

	varsFile := filepath.Join(t.TempDir(), "vars.yml")
	require.NoError(t, os.WriteFile(varsFile, []byte("region: eu\nversion: 1.0\n"), 0o600))

	_, _, err := run(t, taskEnv(srv), "", "template", "run", "Deploy",
		"--check", "--diff", "--limit", "web1,web2", "--tags", "deploy", "--skip-tags", "slow", "--debug-level", "3",
		"--inventory", "staging", "--branch", "release", "--message", "hotfix", "--arguments", `["-vv"]`,
		"--var", "version=1.4.2", "--var", "flag=on", "--extra-vars-file", varsFile)
	require.NoError(t, err)
	require.Len(t, s.created, 1)
	c := s.created[0]
	assert.Equal(t, 9, *c.InventoryID, "inventory resolved by name")
	assert.Equal(t, "release", *c.GitBranch)
	assert.Equal(t, "hotfix", c.Message)
	assert.Equal(t, `["-vv"]`, *c.Arguments)
	assert.Equal(t, map[string]any{
		"dry_run": true, "diff": true, "debug": true, "debug_level": float64(3),
		"limit": []any{"web1", "web2"}, "tags": []any{"deploy"}, "skip_tags": []any{"slow"},
	}, map[string]any(c.Params))

	var env map[string]any
	require.NoError(t, json.Unmarshal([]byte(c.Environment), &env))
	assert.Equal(t, map[string]any{"region": "eu", "version": "1.4.2", "flag": "on"}, env, "--var wins over the file")

	// Terraform flags and explicit false values travel too.
	_, _, err = run(t, taskEnv(srv), "", "template", "run", "Deploy", "--plan", "--destroy=false", "--auto-approve")
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"plan": true, "destroy": false, "auto_approve": true}, map[string]any(s.created[1].Params))

	// Bad inputs are usage errors.
	_, _, err = run(t, taskEnv(srv), "", "template", "run", "Deploy", "--var", "novalue")
	var u *UsageError
	require.ErrorAs(t, err, &u)
	_, _, err = run(t, taskEnv(srv), "", "template", "run", "Deploy", "--extra-vars-file", filepath.Join(t.TempDir(), "missing.yml"))
	assert.ErrorContains(t, err, "read extra vars")
	_, _, err = run(t, taskEnv(srv), "", "template", "run", "Deploy", "--inventory", "nope")
	assert.Equal(t, ExitNotFound, ExitCode(err))
}

func TestTemplateRun_WaitExitCodes(t *testing.T) {
	tests := []struct {
		name     string
		statuses []task_logger.TaskStatus
		wantCode int
	}{
		{"success", []task_logger.TaskStatus{task_logger.TaskWaitingStatus, task_logger.TaskRunningStatus, task_logger.TaskSuccessStatus}, ExitOK},
		{"error", []task_logger.TaskStatus{task_logger.TaskRunningStatus, task_logger.TaskFailStatus}, ExitTaskFailed},
		{"stopped", []task_logger.TaskStatus{task_logger.TaskRunningStatus, task_logger.TaskStoppingStatus, task_logger.TaskStoppedStatus}, ExitTaskStopped},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTaskServer(tt.statuses...)
			srv := httptest.NewServer(s.handler(t))
			defer srv.Close()

			out, errOut, err := run(t, taskEnv(srv), "", "template", "run", "Deploy", "--wait")
			assert.Equal(t, tt.wantCode, ExitCode(err))
			assert.Contains(t, errOut, "Task 42 finished: "+string(tt.statuses[len(tt.statuses)-1]))
			assert.Contains(t, out, "STATUS\t"+string(tt.statuses[len(tt.statuses)-1])+"\n", "the final task is printed as a record")
			assert.GreaterOrEqual(t, s.polls, len(tt.statuses))
		})
	}
}

func TestTemplateRun_FollowPrintsEachLineOnce(t *testing.T) {
	s := newTaskServer(task_logger.TaskRunningStatus, task_logger.TaskRunningStatus, task_logger.TaskRunningStatus, task_logger.TaskSuccessStatus)
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()

	out, _, err := run(t, taskEnv(srv), "", "template", "run", "Deploy", "--follow")
	require.NoError(t, err)
	assert.Equal(t, "PLAY [all]\nTASK [ping]\nok: [web1]\nPLAY RECAP\n", out)
}

func TestTemplateRun_WaitTimeout(t *testing.T) {
	s := newTaskServer(task_logger.TaskRunningStatus)
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()

	_, errOut, err := run(t, taskEnv(srv), "", "template", "run", "Deploy", "--wait", "--wait-timeout", "40ms")
	var to *WaitTimeoutError
	require.ErrorAs(t, err, &to)
	assert.Equal(t, 42, to.TaskID)
	assert.Equal(t, ExitTimeout, ExitCode(err))
	assert.Empty(t, s.stops, "a timeout never stops the task")
	assert.NotContains(t, errOut, "Stop requested")
}

func TestWait_TransientErrorsAreRetried(t *testing.T) {
	s := newTaskServer(task_logger.TaskRunningStatus, task_logger.TaskSuccessStatus)
	s.failPolls = 3
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()

	_, errOut, err := run(t, taskEnv(srv), "", "task", "wait", "42")
	require.NoError(t, err)
	assert.Contains(t, errOut, "retry 1/5")
	assert.Contains(t, errOut, "retry 3/5")

	s = newTaskServer(task_logger.TaskRunningStatus)
	s.failPolls = 100
	srv2 := httptest.NewServer(s.handler(t))
	defer srv2.Close()
	_, _, err = run(t, taskEnv(srv2), "", "task", "wait", "42")
	assert.ErrorContains(t, err, "lost the server")
	assert.Equal(t, ExitError, ExitCode(err))
}

func TestWait_ConfirmationHint(t *testing.T) {
	s := newTaskServer(task_logger.TaskWaitingConfirmation, task_logger.TaskWaitingConfirmation, task_logger.TaskConfirmed, task_logger.TaskRunningStatus, task_logger.TaskSuccessStatus)
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()

	_, errOut, err := run(t, taskEnv(srv), "", "task", "wait", "42")
	require.NoError(t, err)
	assert.Equal(t, 1, strings.Count(errOut, "task confirm 42"), "the hint is printed once")
}

func TestTaskWatchAndWait(t *testing.T) {
	s := newTaskServer(task_logger.TaskWaitingStatus, task_logger.TaskRunningStatus, task_logger.TaskRunningStatus, task_logger.TaskFailStatus)
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()

	out, _, err := run(t, taskEnv(srv), "", "task", "watch", "42")
	require.NoError(t, err, "watch exits 0 by default")
	assert.Equal(t, 3, strings.Count(out, "task 42"), "one line per status change")
	assert.Contains(t, out, "\ttask 42\twaiting\n")
	assert.Contains(t, out, "\ttask 42\trunning\n")
	assert.Contains(t, out, "\ttask 42\terror\n")

	s.polls = 0
	_, _, err = run(t, taskEnv(srv), "", "task", "watch", "42", "--exit-status")
	assert.Equal(t, ExitTaskFailed, ExitCode(err))

	s.polls = 0
	out, _, err = run(t, taskEnv(srv), "", "task", "wait", "42")
	assert.Equal(t, ExitTaskFailed, ExitCode(err))
	assert.Empty(t, out, "wait is silent")

	_, _, err = run(t, taskEnv(srv), "", "task", "wait", "abc")
	assert.Equal(t, ExitUsage, ExitCode(err))
}

func TestTaskListGetLogs(t *testing.T) {
	s := newTaskServer(task_logger.TaskRunningStatus, task_logger.TaskSuccessStatus)
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	env := taskEnv(srv)

	out, _, err := run(t, env, "", "task", "list", "--fields", "id,status,tpl_alias,user_name")
	require.NoError(t, err)
	assert.Equal(t, "ID\tSTATUS\tTPL ALIAS\tUSER NAME\n42\tsuccess\tDeploy\tadmin\n41\terror\tDeploy\t\n", out)

	out, _, err = run(t, env, "", "task", "list", "--status", "error", "-q")
	require.NoError(t, err)
	assert.Equal(t, "41\n", out)

	out, _, err = run(t, env, "", "tasks", "ls", "--template", "Deploy", "-q")
	require.NoError(t, err)
	assert.Equal(t, "42\n", out)

	out, _, err = run(t, env, "", "task", "get", "42", "--fields", "id,status,tpl_alias")
	require.NoError(t, err)
	assert.Equal(t, "ID\t42\nSTATUS\trunning\nTPL ALIAS\tDeploy\n", out)

	// logs: stored lines, raw stream, JSON records.
	s.polls = len(s.lines)
	out, _, err = run(t, env, "", "task", "logs", "42")
	require.NoError(t, err)
	assert.Equal(t, "PLAY [all]\nTASK [ping]\nok: [web1]\nPLAY RECAP\n", out)

	out, _, err = run(t, env, "", "task", "logs", "42", "--raw")
	require.NoError(t, err)
	assert.Equal(t, "PLAY [all]\nTASK [ping]\nok: [web1]\nPLAY RECAP\n", out)

	out, _, err = run(t, env, "", "task", "logs", "42", "-o", "json", "--fields", "output")
	require.NoError(t, err)
	var recs []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &recs))
	assert.Len(t, recs, 4)

	// logs -f follows and sets the exit code.
	s.polls = 0
	s.statuses = []task_logger.TaskStatus{task_logger.TaskRunningStatus, task_logger.TaskFailStatus}
	out, _, err = run(t, env, "", "task", "logs", "42", "-f")
	assert.Equal(t, ExitTaskFailed, ExitCode(err))
	assert.True(t, strings.HasPrefix(out, "PLAY [all]\n"), out)
}

func TestLogLinesStripANSIOutsideTerminal(t *testing.T) {
	s := newTaskServer(task_logger.TaskSuccessStatus)
	s.lines = []string{"\x1b[0;32mok: [web1]\x1b[0m"}
	s.polls = 1
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()

	out, _, err := run(t, taskEnv(srv), "", "task", "logs", "42")
	require.NoError(t, err)
	assert.Equal(t, "ok: [web1]\n", out)
}

func TestTaskStopConfirmReject(t *testing.T) {
	s := newTaskServer(task_logger.TaskStoppingStatus, task_logger.TaskStoppedStatus)
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()
	env := taskEnv(srv)

	_, errOut, err := run(t, env, "", "task", "stop", "42")
	require.NoError(t, err)
	assert.Equal(t, []bool{false}, s.stops)
	assert.Contains(t, errOut, "Stop requested for task 42")

	_, errOut, err = run(t, env, "", "task", "stop", "42", "--force", "--wait")
	require.NoError(t, err, "stop --wait exits 0 even though the task ends as stopped")
	assert.Equal(t, []bool{false, true}, s.stops)
	assert.Contains(t, errOut, "Task 42 finished: stopped")

	_, _, err = run(t, env, "", "task", "confirm", "42")
	require.NoError(t, err)
	_, _, err = run(t, env, "", "task", "reject", "42")
	require.NoError(t, err)
	assert.Equal(t, 1, s.confirms)
	assert.Equal(t, 1, s.rejects)

	_, _, err = run(t, env, "", "task", "stop", "999")
	assert.Equal(t, ExitNotFound, ExitCode(err))
}

func TestTaskRerun(t *testing.T) {
	s := newTaskServer(task_logger.TaskRunningStatus, task_logger.TaskSuccessStatus)
	srv := httptest.NewServer(s.handler(t))
	defer srv.Close()

	out, errOut, err := run(t, taskEnv(srv), "", "task", "rerun", "42", "--message", "again", "--limit", "web2")
	require.NoError(t, err)
	assert.Equal(t, "42\n", out)
	assert.Contains(t, errOut, "rerun of task 42")
	require.Len(t, s.created, 1)
	c := s.created[0]
	assert.Equal(t, 3, c.TemplateID)
	assert.Equal(t, 9, *c.InventoryID, "copied from the source task")
	assert.Equal(t, `{"a":"1"}`, c.Environment)
	assert.Equal(t, "again", c.Message, "override replaces the copied message")
	assert.Equal(t, map[string]any{"dry_run": true, "limit": []any{"web2"}}, map[string]any(c.Params), "copied params plus the override")
}

func TestBuildExtraVars(t *testing.T) {
	cmd := newTaskCmd(templateResource())
	env, err := buildExtraVars(cmd, nil, "")
	require.NoError(t, err)
	assert.Empty(t, env)

	env, err = buildExtraVars(cmd, []string{"a=1", "b=x=y"}, "")
	require.NoError(t, err)
	assert.JSONEq(t, `{"a":"1","b":"x=y"}`, env)

	f := filepath.Join(t.TempDir(), "v.json")
	require.NoError(t, os.WriteFile(f, []byte(`{"n": 5, "list": [1,2]}`), 0o600))
	env, err = buildExtraVars(cmd, []string{"n=7"}, f)
	require.NoError(t, err)
	assert.JSONEq(t, `{"n":"7","list":[1,2]}`, env)

	require.NoError(t, os.WriteFile(f, []byte("- just\n- a list\n"), 0o600))
	_, err = buildExtraVars(cmd, nil, f)
	assert.ErrorContains(t, err, "top-level mapping")
}
