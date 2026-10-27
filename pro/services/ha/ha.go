package ha

import (
	"github.com/semaphoreui/semaphore/api/sockets"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
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
	Nodes() ([]pro_interfaces.NodeInfo, error)
	// RedisInfo returns Redis server / keyspace stats and a key-group breakdown.
	RedisInfo() (pro_interfaces.RedisInfo, error)
}

// Stubs – these are replaced by pro_impl via Go workspace.

// NewHARedis returns the shared HA Redis connection, nil in the open-source build.
func NewHARedis() pro_interfaces.HARedis { return nil }

func NewNodeRegistry(_ pro_interfaces.HARedis) NodeRegistry                           { return nil }
func NewScheduleDeduplicator(_ pro_interfaces.HARedis) schedules.ScheduleDeduplicator { return nil }
func NewWSBroadcaster(_ pro_interfaces.HARedis) sockets.Broadcaster                   { return nil }
func NewOrphanCleaner(_ pro_interfaces.HARedis, _ pro_interfaces.NodeRegistry, _ db.Store) OrphanCleaner {
	return nil
}
func NewClusterInspector(_ pro_interfaces.HARedis) ClusterInspector                  { return nil }
func NewWorkflowRunLocker(_ pro_interfaces.HARedis) pro_interfaces.WorkflowRunLocker { return nil }
func NewAuditExportLeaser(_ pro_interfaces.HARedis) pro_interfaces.AuditExportLeaser { return nil }
