package tasks

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"strconv"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

// RedisTaskStateStore is a Redis-backed implementation of TaskStateStore.
// Notes:
//   - It stores only task identifiers in Redis and keeps an in-process pointer cache
//     to resolve TaskRunner instances. This is sufficient for single-process
//     deployments and basic multi-process visibility. For true cross-process
//     pointer resolution, a separate hydration mechanism would be required.
type RedisTaskStateStore struct {
	client    *redis.Client
	keyPrefix string

	mu      sync.RWMutex
	byID    map[int]*TaskRunner
	byAlias map[string]*TaskRunner

	// pub/sub
	pubsub       *redis.PubSub
	cancelListen context.CancelFunc
}

func NewRedisTaskStateStore() *RedisTaskStateStore {
	keyPrefix := "tasks:"

	p := keyPrefix
	if p != "" && !strings.HasSuffix(p, ":") {
		p += ":"
	}

	var redisTLS *tls.Config
	var addr string
	var dbNum int
	var pass string
	var user string
	var skipVerify bool
	var enableTLS bool

	if util.Config.HA != nil && util.Config.HA.Redis != nil {
		addr = util.Config.HA.Redis.Addr
		dbNum = util.Config.HA.Redis.DB
		pass = util.Config.HA.Redis.Pass
		user = util.Config.HA.Redis.User
		enableTLS = util.Config.HA.Redis.TLS
		skipVerify = util.Config.HA.Redis.TLSSkipVerify
	}
	if enableTLS {
		redisTLS = &tls.Config{InsecureSkipVerify: skipVerify}
	}

	if addr == "" {
		addr = "127.0.0.1:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr:      addr,
		DB:        dbNum,
		Password:  pass,
		Username:  user,
		TLSConfig: redisTLS,
	})

	return &RedisTaskStateStore{
		client:    client,
		keyPrefix: p,
		byID:      make(map[int]*TaskRunner),
		byAlias:   make(map[string]*TaskRunner),
	}
}

func (s *RedisTaskStateStore) key(parts ...string) string {
	return s.keyPrefix + strings.Join(parts, ":")
}

// redis message envelope
type redisEvent struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type taskRef struct {
	TaskID    int    `json:"task_id"`
	ProjectID int    `json:"project_id"`
	Alias     string `json:"alias,omitempty"`
	RunnerID  int    `json:"runner_id,omitempty"`
	Username  string `json:"username,omitempty"`
	Incoming  string `json:"incoming_version,omitempty"`
}

func (s *RedisTaskStateStore) publish(ctx context.Context, ev redisEvent) {
	b, _ := json.Marshal(ev)
	if err := s.client.Publish(ctx, s.key("events"), string(b)).Err(); err != nil {
		log.WithError(err).Error("redis publish failed")
	}
}

// Start restores state from Redis and begins listening to Pub/Sub events
func (s *RedisTaskStateStore) Start(hydrator TaskRunnerHydrator) error {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelListen = cancel

	// Restore queued tasks
	ids, err := s.client.LRange(ctx, s.key("queue"), 0, -1).Result()
	if err != nil {
		log.WithError(err).Error("redis restore queue failed")
	} else {
		for _, idStr := range ids {
			id, convErr := strconv.Atoi(idStr)
			if convErr != nil {
				continue
			}
			// We need project id; store it next to task id in a hash
			projStr, herr := s.client.HGet(ctx, s.key("task_project"), idStr).Result()
			if herr != nil {
				continue
			}
			projID, _ := strconv.Atoi(projStr)
			if hydrator != nil {
				if tr, hErr := hydrator(id, projID); hErr == nil && tr != nil {
					s.mu.Lock()
					s.byID[id] = tr
					if tr.Alias != "" {
						s.byAlias[tr.Alias] = tr
					}
					s.mu.Unlock()
				}
			}
		}
	}

	// Restore active tasks by project
	// Find all keys tasks:active:*
	var cursor uint64
	for {
		keys, cur, err := s.client.Scan(ctx, cursor, s.key("active", "*"), 100).Result()
		if err != nil {
			log.WithError(err).Error("redis scan active keys failed")
			break
		}
		for _, k := range keys {
			// extract project id from key suffix
			parts := strings.Split(k, ":")
			if len(parts) == 0 {
				continue
			}
			projStr := parts[len(parts)-1]
			projectID, _ := strconv.Atoi(projStr)
			ids, gerr := s.client.SMembers(ctx, k).Result()
			if gerr != nil {
				continue
			}
			for _, idStr := range ids {
				id, convErr := strconv.Atoi(idStr)
				if convErr != nil {
					continue
				}
				if hydrator != nil {
					if tr, hErr := hydrator(id, projectID); hErr == nil && tr != nil {
						s.mu.Lock()
						s.byID[id] = tr
						if tr.Alias != "" {
							s.byAlias[tr.Alias] = tr
						}
						s.mu.Unlock()
					}
				}
			}
		}
		cursor = cur
		if cursor == 0 {
			break
		}
	}

	// Restore running tasks set
	runIDs, err := s.client.SMembers(ctx, s.key("running")).Result()
	if err != nil {
		log.WithError(err).Error("redis restore running failed")
	} else {
		for _, idStr := range runIDs {
			id, convErr := strconv.Atoi(idStr)
			if convErr != nil {
				continue
			}
			projStr, herr := s.client.HGet(ctx, s.key("task_project"), idStr).Result()
			if herr != nil {
				continue
			}
			projID, _ := strconv.Atoi(projStr)
			if hydrator != nil {
				if tr, hErr := hydrator(id, projID); hErr == nil && tr != nil {
					s.mu.Lock()
					s.byID[id] = tr
					if tr.Alias != "" {
						s.byAlias[tr.Alias] = tr
					}
					s.mu.Unlock()
				}
			}
		}
	}

	// Restore aliases pointers from Redis hash where value is task id
	aliasMap, err := s.client.HGetAll(ctx, s.key("aliases")).Result()
	if err == nil {
		for alias, idStr := range aliasMap {
			id, convErr := strconv.Atoi(idStr)
			if convErr != nil {
				continue
			}
			projStr, herr := s.client.HGet(ctx, s.key("task_project"), idStr).Result()
			if herr != nil {
				continue
			}
			projID, _ := strconv.Atoi(projStr)
			if hydrator != nil {
				if tr, hErr := hydrator(id, projID); hErr == nil && tr != nil {
					s.mu.Lock()
					s.byID[id] = tr
					s.byAlias[alias] = tr
					s.mu.Unlock()
				}
			}
		}
	}

	// Start Pub/Sub listener
	s.pubsub = s.client.Subscribe(ctx, s.key("events"))
	go func() {
		for {
			msg, rerr := s.pubsub.ReceiveMessage(ctx)
			if rerr != nil {
				return
			}
			var ev redisEvent
			if err := json.Unmarshal([]byte(msg.Payload), &ev); err != nil {
				continue
			}
			switch ev.Type {
			case "enqueue":
				var ref taskRef
				if json.Unmarshal(ev.Data, &ref) == nil && hydrator != nil {
					if tr, hErr := hydrator(ref.TaskID, ref.ProjectID); hErr == nil && tr != nil {
						s.mu.Lock()
						s.byID[ref.TaskID] = tr
						if ref.Alias != "" {
							s.byAlias[ref.Alias] = tr
						}
						// restore runtime fields
						if ref.RunnerID != 0 {
							tr.RunnerID = ref.RunnerID
						}
						if ref.Username != "" {
							tr.Username = ref.Username
						}
						if ref.Incoming != "" {
							v := ref.Incoming
							tr.IncomingVersion = &v
						}
						s.mu.Unlock()
					}
				}
			case "dequeue":
				var ref taskRef
				if json.Unmarshal(ev.Data, &ref) == nil {
					s.mu.Lock()
					delete(s.byID, ref.TaskID)
					s.mu.Unlock()
				}
			case "set_running":
				var ref taskRef
				if json.Unmarshal(ev.Data, &ref) == nil && hydrator != nil {
					if tr, hErr := hydrator(ref.TaskID, ref.ProjectID); hErr == nil && tr != nil {
						s.mu.Lock()
						s.byID[ref.TaskID] = tr
						if ref.Alias != "" {
							s.byAlias[ref.Alias] = tr
						}
						if ref.RunnerID != 0 {
							tr.RunnerID = ref.RunnerID
						}
						if ref.Username != "" {
							tr.Username = ref.Username
						}
						if ref.Incoming != "" {
							v := ref.Incoming
							tr.IncomingVersion = &v
						}
						s.mu.Unlock()
					}
				}
			case "delete_running":
				var ref taskRef
				if json.Unmarshal(ev.Data, &ref) == nil {
					s.mu.Lock()
					delete(s.byID, ref.TaskID)
					s.mu.Unlock()
				}
			case "active_add":
				var ref taskRef
				if json.Unmarshal(ev.Data, &ref) == nil && hydrator != nil {
					if tr, hErr := hydrator(ref.TaskID, ref.ProjectID); hErr == nil && tr != nil {
						s.mu.Lock()
						s.byID[ref.TaskID] = tr
						if ref.RunnerID != 0 {
							tr.RunnerID = ref.RunnerID
						}
						if ref.Username != "" {
							tr.Username = ref.Username
						}
						if ref.Incoming != "" {
							v := ref.Incoming
							tr.IncomingVersion = &v
						}
						s.mu.Unlock()
					}
				}
			case "active_remove":
				var ref taskRef
				if json.Unmarshal(ev.Data, &ref) == nil {
					s.mu.Lock()
					delete(s.byID, ref.TaskID)
					s.mu.Unlock()
				}
			case "alias_set":
				var ref taskRef
				if json.Unmarshal(ev.Data, &ref) == nil && hydrator != nil {
					if tr, hErr := hydrator(ref.TaskID, ref.ProjectID); hErr == nil && tr != nil {
						s.mu.Lock()
						s.byID[ref.TaskID] = tr
						if ref.Alias != "" {
							s.byAlias[ref.Alias] = tr
						}
						if ref.RunnerID != 0 {
							tr.RunnerID = ref.RunnerID
						}
						if ref.Username != "" {
							tr.Username = ref.Username
						}
						if ref.Incoming != "" {
							v := ref.Incoming
							tr.IncomingVersion = &v
						}
						s.mu.Unlock()
					}
				}
			case "alias_delete":
				var ref taskRef
				if json.Unmarshal(ev.Data, &ref) == nil {
					s.mu.Lock()
					delete(s.byAlias, ref.Alias)
					s.mu.Unlock()
				}
			}
		}
	}()

	return nil
}

// Queue operations
func (s *RedisTaskStateStore) Enqueue(task *TaskRunner) {
	s.mu.Lock()
	s.byID[task.Task.ID] = task
	s.mu.Unlock()
	ctx := context.Background()
	if err := s.client.RPush(ctx, s.key("queue"), strconv.Itoa(task.Task.ID)).Err(); err != nil {
		log.WithError(err).Error("redis enqueue failed")
	}
	// store project for hydrator
	_ = s.client.HSet(ctx, s.key("task_project"), strconv.Itoa(task.Task.ID), strconv.Itoa(task.Task.ProjectID)).Err()
	// notify others with runtime fields
	s.publish(ctx, redisEvent{Type: "enqueue", Data: mustJSON(taskRef{TaskID: task.Task.ID, ProjectID: task.Task.ProjectID, Alias: task.Alias, RunnerID: task.RunnerID, Username: task.Username, Incoming: derefStr(task.IncomingVersion)})})
}

func (s *RedisTaskStateStore) DequeueAt(index int) error {
	ctx := context.Background()
	idStr, err := s.client.LIndex(ctx, s.key("queue"), int64(index)).Result()
	if err != nil {
		return nil
	}
	if err := s.client.LRem(ctx, s.key("queue"), 1, idStr).Err(); err != nil {
		log.WithError(err).Error("redis dequeue failed")
	}
	if id, convErr := strconv.Atoi(idStr); convErr == nil {
		s.publish(ctx, redisEvent{Type: "dequeue", Data: mustJSON(taskRef{TaskID: id})})
	}
	return nil
}

func (s *RedisTaskStateStore) QueueRange() []*TaskRunner {
	ctx := context.Background()
	ids, err := s.client.LRange(ctx, s.key("queue"), 0, -1).Result()
	if err != nil {
		log.WithError(err).Error("redis queue range failed")
		return nil
	}
	s.mu.RLock()
	res := make([]*TaskRunner, 0, len(ids))
	for _, idStr := range ids {
		id, convErr := strconv.Atoi(idStr)
		if convErr != nil {
			continue
		}
		if t := s.byID[id]; t != nil {
			res = append(res, t)
		}
	}
	s.mu.RUnlock()
	return res
}

func (s *RedisTaskStateStore) QueueGet(index int) *TaskRunner {
	ctx := context.Background()
	idStr, err := s.client.LIndex(ctx, s.key("queue"), int64(index)).Result()
	if err != nil {
		return nil
	}
	id, convErr := strconv.Atoi(idStr)
	if convErr != nil {
		return nil
	}
	s.mu.RLock()
	t := s.byID[id]
	s.mu.RUnlock()
	return t
}

func (s *RedisTaskStateStore) QueueLen() int {
	ctx := context.Background()
	n, err := s.client.LLen(ctx, s.key("queue")).Result()
	if err != nil {
		log.WithError(err).Error("redis queue len failed")
		return 0
	}
	return int(n)
}

// Running operations
func (s *RedisTaskStateStore) SetRunning(task *TaskRunner) {
	s.mu.Lock()
	s.byID[task.Task.ID] = task
	s.mu.Unlock()
	ctx := context.Background()
	if err := s.client.SAdd(ctx, s.key("running"), task.Task.ID).Err(); err != nil {
		log.WithError(err).Error("redis set running failed")
	}
	s.publish(ctx, redisEvent{Type: "set_running", Data: mustJSON(taskRef{TaskID: task.Task.ID, ProjectID: task.Task.ProjectID, Alias: task.Alias, RunnerID: task.RunnerID, Username: task.Username, Incoming: derefStr(task.IncomingVersion)})})
}

func (s *RedisTaskStateStore) DeleteRunning(taskID int) {
	ctx := context.Background()
	if err := s.client.SRem(ctx, s.key("running"), taskID).Err(); err != nil {
		log.WithError(err).Error("redis delete running failed")
	}
	s.publish(ctx, redisEvent{Type: "delete_running", Data: mustJSON(taskRef{TaskID: taskID})})
}

func (s *RedisTaskStateStore) RunningRange() []*TaskRunner {
	ctx := context.Background()
	ids, err := s.client.SMembers(ctx, s.key("running")).Result()
	if err != nil {
		log.WithError(err).Error("redis running range failed")
		return nil
	}
	s.mu.RLock()
	res := make([]*TaskRunner, 0, len(ids))
	for _, idStr := range ids {
		id, convErr := strconv.Atoi(idStr)
		if convErr != nil {
			continue
		}
		if t := s.byID[id]; t != nil {
			res = append(res, t)
		}
	}
	s.mu.RUnlock()
	return res
}

func (s *RedisTaskStateStore) RunningCount() int {
	ctx := context.Background()
	n, err := s.client.SCard(ctx, s.key("running")).Result()
	if err != nil {
		log.WithError(err).Error("redis running count failed")
		return 0
	}
	return int(n)
}

// Active-by-project operations
func (s *RedisTaskStateStore) AddActive(projectID int, task *TaskRunner) {
	s.mu.Lock()
	s.byID[task.Task.ID] = task
	s.mu.Unlock()
	ctx := context.Background()
	if err := s.client.SAdd(ctx, s.key("active", strconv.Itoa(projectID)), task.Task.ID).Err(); err != nil {
		log.WithError(err).Error("redis add active failed")
	}
	_ = s.client.HSet(ctx, s.key("task_project"), strconv.Itoa(task.Task.ID), strconv.Itoa(projectID)).Err()
	s.publish(ctx, redisEvent{Type: "active_add", Data: mustJSON(taskRef{TaskID: task.Task.ID, ProjectID: projectID, RunnerID: task.RunnerID, Username: task.Username, Incoming: derefStr(task.IncomingVersion)})})
}

func (s *RedisTaskStateStore) RemoveActive(projectID int, taskID int) {
	ctx := context.Background()
	if err := s.client.SRem(ctx, s.key("active", strconv.Itoa(projectID)), taskID).Err(); err != nil {
		log.WithError(err).Error("redis remove active failed")
	}
	s.publish(ctx, redisEvent{Type: "active_remove", Data: mustJSON(taskRef{TaskID: taskID, ProjectID: projectID})})
}

func (s *RedisTaskStateStore) GetActive(projectID int) []*TaskRunner {
	ctx := context.Background()
	ids, err := s.client.SMembers(ctx, s.key("active", strconv.Itoa(projectID))).Result()
	if err != nil {
		log.WithError(err).Error("redis get active failed")
		return nil
	}
	s.mu.RLock()
	res := make([]*TaskRunner, 0, len(ids))
	for _, idStr := range ids {
		id, convErr := strconv.Atoi(idStr)
		if convErr != nil {
			continue
		}
		if t := s.byID[id]; t != nil {
			res = append(res, t)
		}
	}
	s.mu.RUnlock()
	return res
}

func (s *RedisTaskStateStore) ActiveCount(projectID int) int {
	ctx := context.Background()
	n, err := s.client.SCard(ctx, s.key("active", strconv.Itoa(projectID))).Result()
	if err != nil {
		log.WithError(err).Error("redis active count failed")
		return 0
	}
	return int(n)
}

// Alias operations
func (s *RedisTaskStateStore) SetAlias(alias string, task *TaskRunner) {
	s.mu.Lock()
	s.byAlias[alias] = task
	s.byID[task.Task.ID] = task
	s.mu.Unlock()
	ctx := context.Background()
	if err := s.client.HSet(ctx, s.key("aliases"), alias, task.Task.ID).Err(); err != nil {
		log.WithError(err).Error("redis set alias failed")
	}
	s.publish(ctx, redisEvent{Type: "alias_set", Data: mustJSON(taskRef{TaskID: task.Task.ID, ProjectID: task.Task.ProjectID, Alias: alias, RunnerID: task.RunnerID, Username: task.Username, Incoming: derefStr(task.IncomingVersion)})})
}

func (s *RedisTaskStateStore) GetByAlias(alias string) *TaskRunner {
	s.mu.RLock()
	if t := s.byAlias[alias]; t != nil {
		s.mu.RUnlock()
		return t
	}
	s.mu.RUnlock()
	ctx := context.Background()
	idStr, err := s.client.HGet(ctx, s.key("aliases"), alias).Result()
	if err != nil {
		return nil
	}
	id, convErr := strconv.Atoi(idStr)
	if convErr != nil {
		return nil
	}
	s.mu.RLock()
	t := s.byID[id]
	s.mu.RUnlock()
	return t
}

func (s *RedisTaskStateStore) DeleteAlias(alias string) {
	ctx := context.Background()
	if err := s.client.HDel(ctx, s.key("aliases"), alias).Err(); err != nil {
		log.WithError(err).Error("redis delete alias failed")
	}
	s.mu.Lock()
	delete(s.byAlias, alias)
	s.mu.Unlock()
	s.publish(ctx, redisEvent{Type: "alias_delete", Data: mustJSON(taskRef{Alias: alias})})
}

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// UpdateRuntimeFields persists transient TaskRunner fields in Redis
func (s *RedisTaskStateStore) UpdateRuntimeFields(task *TaskRunner) {
	ctx := context.Background()
	// We store them in a hash keyed by task id for easy update
	fields := map[string]interface{}{
		"runner_id":        strconv.Itoa(task.RunnerID),
		"username":         task.Username,
		"incoming_version": derefStr(task.IncomingVersion),
		"alias":            task.Alias,
		"project_id":       strconv.Itoa(task.Task.ProjectID),
	}
	if err := s.client.HSet(ctx, s.key("runtime", strconv.Itoa(task.Task.ID)), fields).Err(); err != nil {
		log.WithError(err).Error("redis update runtime failed")
	}
}

// LoadRuntimeFields restores transient fields from Redis for a task
func (s *RedisTaskStateStore) LoadRuntimeFields(task *TaskRunner) {
	ctx := context.Background()
	m, err := s.client.HGetAll(ctx, s.key("runtime", strconv.Itoa(task.Task.ID))).Result()
	if err != nil || len(m) == 0 {
		return
	}
	if v := m["runner_id"]; v != "" {
		if id, err := strconv.Atoi(v); err == nil {
			task.RunnerID = id
		}
	}
	if v := m["username"]; v != "" {
		task.Username = v
	}
	if v := m["incoming_version"]; v != "" {
		task.IncomingVersion = &v
	}
	if v := m["alias"]; v != "" {
		task.Alias = v
	}
}

// TryClaim atomically tries to claim a task for execution using SET NX
func (s *RedisTaskStateStore) TryClaim(taskID int) bool {
	ctx := context.Background()
	key := s.key("claim", strconv.Itoa(taskID))
	ok, err := s.client.SetNX(ctx, key, "1", 0).Result()
	if err != nil {
		log.WithError(err).Error("redis try claim failed")
		return false
	}
	return ok
}

// DeleteClaim releases the execution claim for a task
func (s *RedisTaskStateStore) DeleteClaim(taskID int) {
	ctx := context.Background()
	if err := s.client.Del(ctx, s.key("claim", strconv.Itoa(taskID))).Err(); err != nil {
		log.WithError(err).Error("redis delete claim failed")
	}
}
