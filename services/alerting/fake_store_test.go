package alerting

import (
	"context"
	"sync"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/semaphoreui/semaphore/util"
)

// fakeStore is the minimal in-memory Store the service tests need.
type fakeStore struct {
	mu        sync.Mutex
	project   db.Project
	alerts    map[int]db.Alert
	nextID    int
	users     map[int]db.User
	members   []db.UserWithProjectRole
	admins    []db.User
	claims    map[string]bool
	templates map[int]db.Template
	tasks     map[int]db.Task
	schedules map[int]db.Schedule
	tplAlerts map[int][]int
	schAlerts map[int][]int
	keys      map[int]db.AccessKey
}

func newFakeStore(project db.Project) *fakeStore {
	return &fakeStore{
		project:   project,
		alerts:    make(map[int]db.Alert),
		users:     make(map[int]db.User),
		claims:    make(map[string]bool),
		templates: make(map[int]db.Template),
		tasks:     make(map[int]db.Task),
		schedules: make(map[int]db.Schedule),
		tplAlerts: make(map[int][]int),
		schAlerts: make(map[int][]int),
		keys:      make(map[int]db.AccessKey),
	}
}

func (s *fakeStore) GetAccessKey(projectID int, keyID int) (db.AccessKey, error) {
	k, ok := s.keys[keyID]
	if !ok || k.ProjectID == nil || *k.ProjectID != projectID {
		return db.AccessKey{}, db.ErrNotFound
	}
	return k, nil
}

func (s *fakeStore) addKey(typ db.AccessKeyType, secret string, login string) db.AccessKey {
	id := len(s.keys) + 1
	pid := s.project.ID
	key := db.AccessKey{ID: id, ProjectID: &pid, Type: typ, Name: "key"}
	switch typ {
	case db.AccessKeyString:
		key.String = secret
	case db.AccessKeyLoginPassword:
		key.LoginPassword = db.LoginPassword{Login: login, Password: secret}
	}
	s.keys[id] = key
	return key
}

func (s *fakeStore) addAlert(alert db.Alert) db.Alert {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	alert.ID = s.nextID
	alert.ProjectID = s.project.ID
	s.alerts[alert.ID] = alert
	return alert
}

func (s *fakeStore) GetAlert(projectID int, alertID int) (db.Alert, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.alerts[alertID]
	if !ok || a.ProjectID != projectID {
		return db.Alert{}, db.ErrNotFound
	}
	return a, nil
}

func (s *fakeStore) GetAlerts(projectID int, _ db.RetrieveQueryParams) ([]db.Alert, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]db.Alert, 0)
	for id := 1; id <= s.nextID; id++ {
		if a, ok := s.alerts[id]; ok && a.ProjectID == projectID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (s *fakeStore) GetDefaultAlertIDs(projectID int) ([]int, error) {
	alerts, _ := s.GetAlerts(projectID, db.RetrieveQueryParams{})
	out := make([]int, 0)
	for _, a := range alerts {
		if a.IsDefault && a.Enabled {
			out = append(out, a.ID)
		}
	}
	return out, nil
}

func (s *fakeStore) CreateAlert(alert db.Alert) (db.Alert, error) { return s.addAlert(alert), nil }
func (s *fakeStore) UpdateAlert(alert db.Alert) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alerts[alert.ID] = alert
	return nil
}
func (s *fakeStore) SetAlertActive(_ int, alertID int, active bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	a := s.alerts[alertID]
	a.Enabled = active
	s.alerts[alertID] = a
	return nil
}

func (s *fakeStore) DeleteAlert(_ int, alertID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.alerts, alertID)
	return nil
}
func (s *fakeStore) GetAlertRefs(int, int) (db.ObjectReferrers, error) {
	return db.ObjectReferrers{}, nil
}
func (s *fakeStore) GetTemplateAlerts(_ int, templateID int) ([]int, error) {
	return s.tplAlerts[templateID], nil
}
func (s *fakeStore) UpdateTemplateAlerts(_ int, templateID int, ids []int) error {
	s.tplAlerts[templateID] = ids
	return nil
}
func (s *fakeStore) GetScheduleAlerts(_ int, scheduleID int) ([]int, error) {
	return s.schAlerts[scheduleID], nil
}
func (s *fakeStore) UpdateScheduleAlerts(_ int, scheduleID int, ids []int) error {
	s.schAlerts[scheduleID] = ids
	return nil
}

func (s *fakeStore) ClaimAlertSend(taskID int, destination string, event db.AlertEvent) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := string(rune(taskID)) + destination + string(event)
	if s.claims[key] {
		return false, nil
	}
	s.claims[key] = true
	return true, nil
}

func (s *fakeStore) UnclaimAlertSend(taskID int, destination string, event db.AlertEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := string(rune(taskID)) + destination + string(event)
	delete(s.claims, key)
	return nil
}

func (s *fakeStore) GetUser(userID int) (db.User, error) {
	u, ok := s.users[userID]
	if !ok {
		return db.User{}, db.ErrNotFound
	}
	return u, nil
}

func (s *fakeStore) GetTask(_ int, taskID int) (db.Task, error) {
	t, ok := s.tasks[taskID]
	if !ok {
		return db.Task{}, db.ErrNotFound
	}
	return t, nil
}

func (s *fakeStore) GetTemplate(_ int, templateID int) (db.Template, error) {
	t, ok := s.templates[templateID]
	if !ok {
		return db.Template{}, db.ErrNotFound
	}
	return t, nil
}

func (s *fakeStore) GetSchedule(_ int, scheduleID int) (db.Schedule, error) {
	sch, ok := s.schedules[scheduleID]
	if !ok {
		return db.Schedule{}, db.ErrNotFound
	}
	return sch, nil
}

func (s *fakeStore) GetProject(projectID int) (db.Project, error) {
	if projectID != s.project.ID {
		return db.Project{}, db.ErrNotFound
	}
	return s.project, nil
}

func (s *fakeStore) GetProjectUsers(int, db.RetrieveQueryParams) ([]db.UserWithProjectRole, error) {
	return s.members, nil
}

func (s *fakeStore) GetAllAdmins() ([]db.User, error) { return s.admins, nil }

// fakeChannel records every send so tests can assert on routing decisions.
type fakeChannel struct {
	channelBase
	mu       sync.Mutex
	sent     []fakeSend
	instance *Destination
	sendErr  error
}

type fakeSend struct {
	Dest Destination
	Msg  Message
}

func newFakeChannel(typ db.AlertType, events db.AlertEvents) *fakeChannel {
	return &fakeChannel{
		channelBase: channelBase{
			typ:           typ,
			title:         string(typ),
			icon:          "mdi-test-tube",
			fields:        []Field{{Name: FieldURL}},
			defaultEvents: events,
			format:        BodyFormatText,
			templateFile:  "slack.tmpl",
		},
	}
}

func (c *fakeChannel) Validate(*util.ConfigType, Destination) error { return nil }

func (c *fakeChannel) InstanceDestination(*util.ConfigType, db.Project) (Destination, bool) {
	if c.instance == nil {
		return Destination{}, false
	}
	return *c.instance, true
}

func (c *fakeChannel) Send(_ context.Context, _ *util.ConfigType, dest Destination, msg Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sent = append(c.sent, fakeSend{Dest: dest, Msg: msg})
	return c.sendErr
}

func (c *fakeChannel) sends() []fakeSend {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]fakeSend(nil), c.sent...)
}

type recordingLogger struct {
	lines []string
}

func (l *recordingLogger) Logf(format string, args ...any) {
	l.lines = append(l.lines, format)
}

func testConfig() *util.ConfigType {
	return &util.ConfigType{WebHost: "https://semaphore.example"}
}

func newTestService(store *fakeStore, cfg *util.ConfigType, channels ...Channel) (Service, *Registry) {
	registry := &Registry{byType: make(map[db.AlertType]Channel)}
	for _, ch := range channels {
		registry.Register(ch)
	}
	return NewServiceWithConfig(store, registry, nil, func() *util.ConfigType { return cfg }), registry
}

func statusTask(id int, projectID int) db.Task {
	return db.Task{ID: id, ProjectID: projectID, TemplateID: 1, Status: task_logger.TaskSuccessStatus}
}
