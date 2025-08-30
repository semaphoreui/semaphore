package mcp

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/semaphoreui/semaphore/db"
)

// ProjectService defines the methods required to list projects.
type ProjectService interface {
	GetAllProjects() ([]db.Project, error)
}

// TaskService defines the methods required to trigger tasks.
type TaskService interface {
	AddTask(task db.Task, userID *int, username string, projectID int, needAlias bool) (db.Task, error)
}

// Server implements a minimal Model Context Protocol server.
type Server struct {
	projects ProjectService
	tasks    TaskService
}

// NewServer creates a new MCP server.
func NewServer(projects ProjectService, tasks TaskService) *Server {
	return &Server{projects: projects, tasks: tasks}
}

// request represents a client command.
type request struct {
	Command    string `json:"command"`
	ProjectID  int    `json:"project_id,omitempty"`
	TemplateID int    `json:"template_id,omitempty"`
}

// response is sent back to the client.
type response struct {
	Command  string       `json:"command,omitempty"`
	Status   string       `json:"status,omitempty"`
	Error    string       `json:"error,omitempty"`
	Projects []db.Project `json:"projects,omitempty"`
	TaskID   int          `json:"task_id,omitempty"`
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

// ServeWS upgrades the connection to WebSocket and handles MCP commands.
func (s *Server) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	handshaked := false
	for {
		var req request
		if err := conn.ReadJSON(&req); err != nil {
			return
		}

		if !handshaked && req.Command != "handshake" {
			_ = conn.WriteJSON(response{Error: "handshake_required"})
			continue
		}

		switch req.Command {
		case "handshake":
			handshaked = true
			_ = conn.WriteJSON(response{Command: "handshake", Status: "ok"})
		case "list_projects":
			projects, err := s.projects.GetAllProjects()
			if err != nil {
				_ = conn.WriteJSON(response{Command: "list_projects", Error: err.Error()})
				continue
			}
			_ = conn.WriteJSON(response{Command: "list_projects", Projects: projects})
		case "trigger_task":
			task, err := s.tasks.AddTask(db.Task{TemplateID: req.TemplateID}, nil, "", req.ProjectID, false)
			if err != nil {
				_ = conn.WriteJSON(response{Command: "trigger_task", Error: err.Error()})
				continue
			}
			_ = conn.WriteJSON(response{Command: "trigger_task", TaskID: task.ID, Status: "queued"})
		default:
			_ = conn.WriteJSON(response{Error: "unknown_command"})
		}
	}
}
