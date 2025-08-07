package tasks

import (
	"context"
	"strconv"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
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
}

func NewRedisTaskStateStore() *RedisTaskStateStore {
	keyPrefix := "tasks:"

	p := keyPrefix
	if p != "" && !strings.HasSuffix(p, ":") {
		p += ":"
	}

	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
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

// Queue operations
func (s *RedisTaskStateStore) Enqueue(task *TaskRunner) {
	s.mu.Lock()
	s.byID[task.Task.ID] = task
	s.mu.Unlock()
	ctx := context.Background()
	if err := s.client.RPush(ctx, s.key("queue"), strconv.Itoa(task.Task.ID)).Err(); err != nil {
		log.WithError(err).Error("redis enqueue failed")
	}
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
}

func (s *RedisTaskStateStore) DeleteRunning(taskID int) {
	ctx := context.Background()
	if err := s.client.SRem(ctx, s.key("running"), taskID).Err(); err != nil {
		log.WithError(err).Error("redis delete running failed")
	}
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
}

func (s *RedisTaskStateStore) RemoveActive(projectID int, taskID int) {
	ctx := context.Background()
	if err := s.client.SRem(ctx, s.key("active", strconv.Itoa(projectID)), taskID).Err(); err != nil {
		log.WithError(err).Error("redis remove active failed")
	}
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
}
