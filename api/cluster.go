package api

import (
	"encoding/json"
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	taskServices "github.com/semaphoreui/semaphore/services/tasks"
	"github.com/semaphoreui/semaphore/util"
	log "github.com/sirupsen/logrus"
)

// clusterInspectorFromContext returns the ClusterInspector injected by the
// router middleware, or nil if HA is disabled / the overlay is absent.
func clusterInspectorFromContext(r *http.Request) pro_interfaces.ClusterInspector {
	ci, _ := helpers.GetFromContext(r, "cluster_inspector").(pro_interfaces.ClusterInspector)
	return ci
}

// taskStateInspectorFromContext returns the TaskStateInspector for the running
// task pool, or nil if the store does not support introspection.
func taskStateInspectorFromContext(r *http.Request) taskServices.TaskStateInspector {
	pool, ok := helpers.GetFromContext(r, "task_pool").(*taskServices.TaskPool)
	if !ok || pool == nil {
		return nil
	}
	inspector, _ := pool.StateStore().(taskServices.TaskStateInspector)
	return inspector
}

// getClusterStatus reports HA mode, cluster membership and Redis stats.
// It never fails: with HA disabled or no inspector it returns ha_enabled:false.
func getClusterStatus(w http.ResponseWriter, r *http.Request) {
	body := map[string]any{
		"ha_enabled": util.HAEnabled(),
	}
	if util.Config.HA != nil {
		body["node_id"] = util.Config.HA.NodeID
	}

	ci := clusterInspectorFromContext(r)

	if ci == nil {
		helpers.WriteErrorStatus(w, "cluster inspection is unavailable (HA mode disabled or overlay missing)", http.StatusServiceUnavailable)
		return
	}

	if nodes, err := ci.Nodes(); err != nil {
		log.WithError(err).Error("cluster: failed to list nodes")
	} else {
		body["nodes"] = nodes
	}

	if redisInfo, err := ci.RedisInfo(); err != nil {
		log.WithError(err).Error("cluster: failed to read redis info")
	} else {
		body["redis"] = redisInfo
	}

	helpers.WriteJSON(w, http.StatusOK, body)
}

// getClusterTasks returns a snapshot of the task pool records (queue, running,
// active, aliases, claims). Works in both HA and non-HA mode.
func getClusterTasks(w http.ResponseWriter, r *http.Request) {
	inspector := taskStateInspectorFromContext(r)
	if inspector == nil {
		helpers.WriteJSON(w, http.StatusOK, taskServices.TaskStateSnapshot{
			Queue:        []taskServices.TaskRecord{},
			Running:      []taskServices.TaskRecord{},
			ActiveByProj: map[int][]taskServices.TaskRecord{},
			Aliases:      map[string]int{},
			Claims:       []int{},
		})
		return
	}

	helpers.WriteJSON(w, http.StatusOK, inspector.Snapshot())
}

// clearClusterTasksRequest is the body of DELETE /api/cluster/tasks.
type clearClusterTasksRequest struct {
	Scope taskServices.ClearScope `json:"scope"`
}

// clearClusterTasks removes the selected task record groups from the backend.
// This is a maintenance action for recovering from a stuck cluster state.
func clearClusterTasks(w http.ResponseWriter, r *http.Request) {
	var req clearClusterTasksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helpers.WriteErrorStatus(w, "an explicit scope is required", http.StatusBadRequest)
		return
	}

	scope := req.Scope
	if !scope.Queue &&
		!scope.Running &&
		!scope.Active &&
		!scope.Aliases &&
		!scope.Claims &&
		!scope.RuntimeFields {
		helpers.WriteErrorStatus(w, "no record groups selected", http.StatusBadRequest)
		return
	}

	inspector := taskStateInspectorFromContext(r)
	if inspector == nil {
		helpers.WriteErrorStatus(w, "task state store does not support clearing", http.StatusNotImplemented)
		return
	}

	result, err := inspector.ClearTasks(scope)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	user := helpers.UserFromContext(r)
	log.WithFields(log.Fields{
		"context":      "cluster",
		"user":         user.Username,
		"scope":        scope,
		"deleted_keys": result.DeletedKeys,
		"per_group":    result.PerGroup,
	}).Info("cluster tasks cleared from backend")

	helpers.WriteJSON(w, http.StatusOK, result)
}
