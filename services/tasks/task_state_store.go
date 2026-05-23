package tasks

import "sync"

// TaskStateInspector is optionally implemented by a TaskStateStore to
// expose its records for the Cluster Dashboard. Callers must type-assert,
// so stores that do not implement it are unaffected.
type TaskStateInspector interface {
	// Snapshot returns a serializable view of all task records.
	Snapshot() TaskStateSnapshot
	// ClearTasks removes task records from the backend. scope selects
	// which record groups to clear.
	ClearTasks(scope ClearScope) (ClearResult, error)
}

// TaskStateSnapshot is a serializable view of all task pool records.
type TaskStateSnapshot struct {
	Queue        []TaskRecord         `json:"queue"`
	Running      []TaskRecord         `json:"running"`
	ActiveByProj map[int][]TaskRecord `json:"active_by_project"`
	Aliases      map[string]int       `json:"aliases"` // alias -> taskID
	Claims       []int                `json:"claims"`  // claimed taskIDs
}

// TaskRecord is a single task entry in a TaskStateSnapshot.
type TaskRecord struct {
	TaskID     int    `json:"task_id"`
	ProjectID  int    `json:"project_id"`
	TemplateID int    `json:"template_id"`
	Status     string `json:"status"`
	RunnerID   int    `json:"runner_id"`
	Username   string `json:"username"`
	Alias      string `json:"alias"`
	NodeID     string `json:"node_id"` // owning node, "" if unknown
}

// ClearScope selects which record groups ClearTasks removes.
type ClearScope struct {
	Queue         bool `json:"queue"`
	Running       bool `json:"running"`
	Active        bool `json:"active"`
	Aliases       bool `json:"aliases"`
	Claims        bool `json:"claims"`
	RuntimeFields bool `json:"runtime_fields"`
}

// ClearResult reports what ClearTasks removed.
type ClearResult struct {
	DeletedKeys int            `json:"deleted_keys"`
	PerGroup    map[string]int `json:"per_group"`
}

// taskToRecord builds a TaskRecord from a TaskRunner.
func taskToRecord(t *TaskRunner) TaskRecord {
	if t == nil {
		return TaskRecord{}
	}
	rec := TaskRecord{
		TaskID:     t.Task.ID,
		ProjectID:  t.Task.ProjectID,
		TemplateID: t.Task.TemplateID,
		Status:     string(t.Task.Status),
		Username:   t.Username,
		Alias:      t.Alias,
	}
	if t.Task.RunnerID != nil {
		rec.RunnerID = *t.Task.RunnerID
	}
	return rec
}

// TaskRunnerHydrator constructs a TaskRunner for an existing task
// identified by taskID and projectID without starting it.
type TaskRunnerHydrator func(taskID int, projectID int) (*TaskRunner, error)

// TaskStateStore defines pluggable storage for task pool state
type TaskStateStore interface {
	// Start allows the store to initialize, restore its in-memory
	// pointers from the underlying backend, and start background
	// sync listeners (e.g., Redis Pub/Sub). Implementations may no-op.
	Start(hydrator TaskRunnerHydrator) error

	// Queue operations
	Enqueue(task *TaskRunner)
	DequeueAt(index int) error
	QueueRange() []*TaskRunner
	QueueGet(index int) *TaskRunner
	QueueLen() int

	// Running tasks map operations
	SetRunning(task *TaskRunner)
	DeleteRunning(taskID int)
	RunningRange() []*TaskRunner
	RunningCount() int

	// Active-by-project operations
	AddActive(projectID int, task *TaskRunner)
	RemoveActive(projectID int, taskID int)
	GetActive(projectID int) []*TaskRunner
	ActiveCount(projectID int) int

	// Aliases operations
	SetAlias(alias string, task *TaskRunner)
	GetByAlias(alias string) *TaskRunner
	DeleteAlias(alias string)

	// Distributed claim to ensure single runner starts a task
	TryClaim(taskID int) bool
	DeleteClaim(taskID int)

	// UpdateRuntimeFields persists transient fields of TaskRunner so
	// they can be restored after restart in HA mode.
	UpdateRuntimeFields(task *TaskRunner)

	// LoadRuntimeFields fills runtime fields (RunnerID, Username, IncomingVersion, Alias)
	// from the backend into the provided task. No-op if not supported.
	LoadRuntimeFields(task *TaskRunner)
}

// MemoryTaskStateStore is an in-memory implementation of TaskStateStore
type MemoryTaskStateStore struct {
	mu         sync.RWMutex
	queue      []*TaskRunner
	running    map[int]*TaskRunner
	activeProj map[int]map[int]*TaskRunner // projectID -> taskID -> task
	aliases    map[string]*TaskRunner
}

func NewMemoryTaskStateStore() *MemoryTaskStateStore {
	return &MemoryTaskStateStore{
		queue:      make([]*TaskRunner, 0),
		running:    make(map[int]*TaskRunner),
		activeProj: make(map[int]map[int]*TaskRunner),
		aliases:    make(map[string]*TaskRunner),
	}
}

// Start is a no-op for the in-memory store
func (s *MemoryTaskStateStore) Start(_ TaskRunnerHydrator) error { return nil }

// Claims always succeed in memory single-process mode
func (s *MemoryTaskStateStore) TryClaim(_ int) bool               { return true }
func (s *MemoryTaskStateStore) DeleteClaim(_ int)                 {}
func (s *MemoryTaskStateStore) UpdateRuntimeFields(_ *TaskRunner) {}
func (s *MemoryTaskStateStore) LoadRuntimeFields(_ *TaskRunner)   {}

// Queue
func (s *MemoryTaskStateStore) Enqueue(task *TaskRunner) {
	s.mu.Lock()
	s.queue = append(s.queue, task)
	s.mu.Unlock()
}

func (s *MemoryTaskStateStore) DequeueAt(index int) error {
	s.mu.Lock()
	if index < 0 || index >= len(s.queue) {
		s.mu.Unlock()
		return nil
	}
	s.queue = append(s.queue[:index], s.queue[index+1:]...)
	s.mu.Unlock()
	return nil
}

func (s *MemoryTaskStateStore) QueueRange() []*TaskRunner {
	s.mu.RLock()
	out := make([]*TaskRunner, len(s.queue))
	copy(out, s.queue)
	s.mu.RUnlock()
	return out
}

func (s *MemoryTaskStateStore) QueueGet(index int) *TaskRunner {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if index < 0 || index >= len(s.queue) {
		return nil
	}
	return s.queue[index]
}

func (s *MemoryTaskStateStore) QueueLen() int {
	s.mu.RLock()
	l := len(s.queue)
	s.mu.RUnlock()
	return l
}

// Running
func (s *MemoryTaskStateStore) SetRunning(task *TaskRunner) {
	s.mu.Lock()
	s.running[task.Task.ID] = task
	s.mu.Unlock()
}

func (s *MemoryTaskStateStore) DeleteRunning(taskID int) {
	s.mu.Lock()
	delete(s.running, taskID)
	s.mu.Unlock()
}

func (s *MemoryTaskStateStore) RunningRange() []*TaskRunner {
	s.mu.RLock()
	res := make([]*TaskRunner, 0, len(s.running))
	for _, t := range s.running {
		res = append(res, t)
	}
	s.mu.RUnlock()
	return res
}

func (s *MemoryTaskStateStore) RunningCount() int {
	s.mu.RLock()
	l := len(s.running)
	s.mu.RUnlock()
	return l
}

// Active by project
func (s *MemoryTaskStateStore) AddActive(projectID int, task *TaskRunner) {
	s.mu.Lock()
	m, ok := s.activeProj[projectID]
	if !ok {
		m = make(map[int]*TaskRunner)
		s.activeProj[projectID] = m
	}
	m[task.Task.ID] = task
	s.mu.Unlock()
}

func (s *MemoryTaskStateStore) RemoveActive(projectID int, taskID int) {
	s.mu.Lock()
	if s.activeProj[projectID] != nil {
		delete(s.activeProj[projectID], taskID)
		if len(s.activeProj[projectID]) == 0 {
			delete(s.activeProj, projectID)
		}
	}
	s.mu.Unlock()
}

func (s *MemoryTaskStateStore) GetActive(projectID int) []*TaskRunner {
	s.mu.RLock()
	res := make([]*TaskRunner, 0)
	if s.activeProj[projectID] != nil {
		for _, t := range s.activeProj[projectID] {
			res = append(res, t)
		}
	}
	s.mu.RUnlock()
	return res
}

func (s *MemoryTaskStateStore) ActiveCount(projectID int) int {
	s.mu.RLock()
	l := 0
	if s.activeProj[projectID] != nil {
		l = len(s.activeProj[projectID])
	}
	s.mu.RUnlock()
	return l
}

// Aliases
func (s *MemoryTaskStateStore) SetAlias(alias string, task *TaskRunner) {
	s.mu.Lock()
	s.aliases[alias] = task
	s.mu.Unlock()
}

func (s *MemoryTaskStateStore) GetByAlias(alias string) *TaskRunner {
	s.mu.RLock()
	t := s.aliases[alias]
	s.mu.RUnlock()
	return t
}

func (s *MemoryTaskStateStore) DeleteAlias(alias string) {
	s.mu.Lock()
	delete(s.aliases, alias)
	s.mu.Unlock()
}

// Snapshot returns a serializable view of all in-memory task records.
// The memory store does not track claims, so Claims is always empty.
func (s *MemoryTaskStateStore) Snapshot() TaskStateSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snap := TaskStateSnapshot{
		Queue:        make([]TaskRecord, 0, len(s.queue)),
		Running:      make([]TaskRecord, 0, len(s.running)),
		ActiveByProj: make(map[int][]TaskRecord),
		Aliases:      make(map[string]int),
		Claims:       []int{},
	}

	for _, t := range s.queue {
		snap.Queue = append(snap.Queue, taskToRecord(t))
	}
	for _, t := range s.running {
		snap.Running = append(snap.Running, taskToRecord(t))
	}
	for projectID, tasks := range s.activeProj {
		recs := make([]TaskRecord, 0, len(tasks))
		for _, t := range tasks {
			recs = append(recs, taskToRecord(t))
		}
		snap.ActiveByProj[projectID] = recs
	}
	for alias, t := range s.aliases {
		if t != nil {
			snap.Aliases[alias] = t.Task.ID
		}
	}

	return snap
}

// ClearTasks removes the selected record groups from the in-memory store.
func (s *MemoryTaskStateStore) ClearTasks(scope ClearScope) (ClearResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	res := ClearResult{PerGroup: map[string]int{}}

	if scope.Queue {
		n := len(s.queue)
		s.queue = make([]*TaskRunner, 0)
		res.PerGroup["queue"] = n
		res.DeletedKeys += n
	}
	if scope.Running {
		n := len(s.running)
		s.running = make(map[int]*TaskRunner)
		res.PerGroup["running"] = n
		res.DeletedKeys += n
	}
	if scope.Active {
		n := 0
		for _, tasks := range s.activeProj {
			n += len(tasks)
		}
		s.activeProj = make(map[int]map[int]*TaskRunner)
		res.PerGroup["active"] = n
		res.DeletedKeys += n
	}
	if scope.Aliases {
		n := len(s.aliases)
		s.aliases = make(map[string]*TaskRunner)
		res.PerGroup["aliases"] = n
		res.DeletedKeys += n
	}
	// The memory store has no separate claims or runtime fields.
	if scope.Claims {
		res.PerGroup["claims"] = 0
	}
	if scope.RuntimeFields {
		res.PerGroup["runtime_fields"] = 0
	}

	return res, nil
}
