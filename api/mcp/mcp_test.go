package mcp

import (
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/semaphoreui/semaphore/db"
	mcpservice "github.com/semaphoreui/semaphore/services/mcp"
)

type mockProjectStore struct {
	projects []db.Project
}

func (m *mockProjectStore) GetAllProjects() ([]db.Project, error) {
	return m.projects, nil
}

type mockTaskPool struct{}

func (m *mockTaskPool) AddTask(task db.Task, userID *int, username string, projectID int, needAlias bool) (db.Task, error) {
	task.ID = 1
	return task, nil
}

func setupServer(t *testing.T) (*websocket.Conn, func()) {
	store := &mockProjectStore{projects: []db.Project{{ID: 1, Name: "demo"}}}
	pool := &mockTaskPool{}
	srv := mcpservice.NewServer(store, pool)
	r := mux.NewRouter()
	Route(r, srv)
	ts := httptest.NewServer(r)
	wsURL := "ws" + ts.URL[len("http"):len(ts.URL)] + "/mcp/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	cleanup := func() { conn.Close(); ts.Close() }
	return conn, cleanup
}

func TestHandshakeAndListProjects(t *testing.T) {
	conn, cleanup := setupServer(t)
	defer cleanup()

	if err := conn.WriteJSON(map[string]string{"command": "handshake"}); err != nil {
		t.Fatalf("write handshake: %v", err)
	}
	var resp map[string]interface{}
	if err := conn.ReadJSON(&resp); err != nil {
		t.Fatalf("read handshake: %v", err)
	}
	if resp["status"] != "ok" {
		t.Fatalf("handshake failed: %v", resp)
	}

	if err := conn.WriteJSON(map[string]string{"command": "list_projects"}); err != nil {
		t.Fatalf("write list: %v", err)
	}
	var list struct {
		Projects []db.Project `json:"projects"`
	}
	if err := conn.ReadJSON(&list); err != nil {
		t.Fatalf("read list: %v", err)
	}
	if len(list.Projects) != 1 || list.Projects[0].Name != "demo" {
		t.Fatalf("unexpected projects: %+v", list.Projects)
	}
}

func TestTriggerTask(t *testing.T) {
	conn, cleanup := setupServer(t)
	defer cleanup()
	_ = conn.WriteJSON(map[string]string{"command": "handshake"})
	_ = conn.ReadJSON(&map[string]interface{}{})

	if err := conn.WriteJSON(map[string]interface{}{"command": "trigger_task", "project_id": 1, "template_id": 2}); err != nil {
		t.Fatalf("write trigger: %v", err)
	}
	var resp struct {
		TaskID int `json:"task_id"`
	}
	if err := conn.ReadJSON(&resp); err != nil {
		t.Fatalf("read trigger: %v", err)
	}
	if resp.TaskID != 1 {
		t.Fatalf("unexpected task id: %d", resp.TaskID)
	}
}

func TestUnknownCommand(t *testing.T) {
	conn, cleanup := setupServer(t)
	defer cleanup()
	_ = conn.WriteJSON(map[string]string{"command": "handshake"})
	_ = conn.ReadJSON(&map[string]interface{}{})

	if err := conn.WriteJSON(map[string]string{"command": "bad"}); err != nil {
		t.Fatalf("write bad: %v", err)
	}
	var resp map[string]interface{}
	if err := conn.ReadJSON(&resp); err != nil {
		t.Fatalf("read bad: %v", err)
	}
	if resp["error"] != "unknown_command" {
		t.Fatalf("unexpected response: %v", resp)
	}
}
