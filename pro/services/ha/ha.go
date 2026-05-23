package ha

import (
	"time"

	"github.com/semaphoreui/semaphore/api/sockets"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/schedules"
)

// NodeRegistry manages node heartbeats and cluster membership tracking
// in HA mode. In active-active setups every Semaphore instance registers
// itself and periodically refreshes a heartbeat so other nodes can detect
// liveness.
type NodeRegistry interface {
	Start() error
	Stop()
	NodeCount() int
	NodeID() string
}

// OrphanCleaner periodically detects tasks whose owning node has died and
// marks them as failed so they do not remain stuck in "running" forever.
type OrphanCleaner interface {
	Start()
	Stop()
}

// ClusterInspector is the read surface for the Cluster Dashboard. It exposes
// cluster membership and Redis keyspace stats. The Redis-backed implementation
// is supplied by pro_impl; the OSS stub returns nil.
type ClusterInspector interface {
	// Nodes returns current cluster membership with heartbeat info.
	Nodes() ([]NodeInfo, error)
	// RedisInfo returns Redis server / keyspace stats and a key-group breakdown.
	RedisInfo() (RedisInfo, error)
}

// NodeInfo describes a single node in the cluster.
type NodeInfo struct {
	NodeID        string    `json:"node_id"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	Alive         bool      `json:"alive"`
	IsSelf        bool      `json:"is_self"`
	StartedAt     time.Time `json:"started_at"`
	Version       string    `json:"version"`
}

// RedisInfo describes the Redis backend shared by the cluster.
type RedisInfo struct {
	Addr       string         `json:"addr"`
	Connected  bool           `json:"connected"`
	Version    string         `json:"version"`
	UsedMemory string         `json:"used_memory"`
	TotalKeys  int            `json:"total_keys"`
	KeyGroups  map[string]int `json:"key_groups"`
}

// Stubs – these are replaced by pro_impl via Go workspace.

func NewNodeRegistry() NodeRegistry                           { return nil }
func NewScheduleDeduplicator() schedules.ScheduleDeduplicator { return nil }
func NewWSBroadcaster() sockets.Broadcaster                   { return nil }
func NewOrphanCleaner(_ db.Store) OrphanCleaner               { return nil }
func NewClusterInspector() ClusterInspector                   { return nil }
