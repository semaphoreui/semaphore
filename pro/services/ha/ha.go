package ha

import (
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

// Stubs – these are replaced by pro_impl via Go workspace.

func NewNodeRegistry() NodeRegistry                             { return nil }
func NewScheduleDeduplicator() schedules.ScheduleDeduplicator   { return nil }
func NewWSBroadcaster() sockets.Broadcaster                     { return nil }
func NewOrphanCleaner(_ db.Store) OrphanCleaner                 { return nil }
