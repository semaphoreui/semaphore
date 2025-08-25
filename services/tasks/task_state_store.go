package tasks

import "sync"

// TaskRunnerHydrator constructs a TaskRunner for an existing task
// identified by taskID and projectID without starting it.
type TaskRunnerHydrator func(taskID int, projectID int) (*TaskRunner, error)

// TaskStateStore defines pluggable storage for task pool state
type TaskStateStore interface {
	// Start allows the store to initialize, restore its in-memory
	// pointers from the underlying backend and start background
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
