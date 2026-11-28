package projects

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/server"
	log "github.com/sirupsen/logrus"
)

type InventorySourceController struct {
	service *server.InventorySourceService
}

func NewInventorySourceController(service *server.InventorySourceService) *InventorySourceController {
	return &InventorySourceController{service: service}
}

// inventorySourceRequest holds the only fields a request may set, the sync writes the rest
type inventorySourceRequest struct {
	Name         string                 `json:"name"`
	Type         db.InventorySourceType `json:"type"`
	RepositoryID *int                   `json:"repository_id"`
	KeyID        *int                   `json:"key_id"`
	Params       db.MapStringAnyField   `json:"params"`
}

// settings returns the params of the request, the path of a git source without the spaces around it
func (r inventorySourceRequest) settings() db.MapStringAnyField {
	if path, ok := r.Params["path"].(string); ok {
		r.Params["path"] = strings.TrimSpace(path)
	}
	return r.Params
}

// recordInventorySource writes one event log entry and one audit record of a source change
func recordInventorySource(r *http.Request, action helpers.EventLogType, verb string, kind audit.Kind, inventory db.Inventory, source db.InventorySource, repository string) {
	helpers.EventLog(r, action, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   inventory.ProjectID,
		ObjectType:  db.EventInventorySource,
		ObjectID:    source.ID,
		Description: fmt.Sprintf("Source %s %s in inventory %s", source.Name, verb, inventory.Name),
	})

	metadata := audit.InventorySourceMetadata{
		InventoryID:   inventory.ID,
		InventoryName: inventory.Name,
		Type:          string(source.Type),
		Repository:    repository,
		Params:        source.Params,
	}
	if kind == audit.ResourceInventorySourceDelete {
		metadata.Hosts, metadata.Groups = source.HostCount, source.GroupCount
	}

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      kind,
		Target:    audit.ResourceTarget(audit.TargetInventorySource, source.ID, source.Name),
		ProjectID: inventory.ProjectID,
		Metadata:  metadata,
	})
}

// validSource answers 400 when the source is invalid or its repository is not in the project, and returns the repository name
func (c *InventorySourceController) validSource(w http.ResponseWriter, r *http.Request, source db.InventorySource) (string, bool) {
	if err := c.service.Validate(source); err != nil {
		helpers.WriteErrorStatus(w, err.Error(), http.StatusBadRequest)
		return "", false
	}

	if source.RepositoryID == nil {
		return "", true
	}

	repo, err := helpers.Store(r).GetRepository(source.ProjectID, *source.RepositoryID)
	if errors.Is(err, db.ErrNotFound) {
		helpers.WriteErrorStatus(w, "repository is not in the project", http.StatusBadRequest)
		return "", false
	}
	if err != nil {
		helpers.WriteError(w, err)
		return "", false
	}
	return repo.Name, true
}

// GetInventorySources returns the sources of a managed inventory
func (c *InventorySourceController) GetInventorySources(w http.ResponseWriter, r *http.Request) {
	inventory, ok := managedInventory(w, r)
	if !ok {
		return
	}

	sources, err := helpers.Store(r).GetInventorySources(inventory.ProjectID, inventory.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	for i, source := range sources {
		sources[i] = c.service.Present(source)
	}

	helpers.WriteJSON(w, http.StatusOK, sources)
}

// AddInventorySource creates a source in a managed inventory
func (c *InventorySourceController) AddInventorySource(w http.ResponseWriter, r *http.Request) {
	inventory, ok := managedInventory(w, r)
	if !ok {
		return
	}

	var body inventorySourceRequest
	if !helpers.Bind(w, r, &body) {
		return
	}

	source := db.InventorySource{
		ProjectID:    inventory.ProjectID,
		InventoryID:  inventory.ID,
		Name:         body.Name,
		Type:         body.Type,
		RepositoryID: body.RepositoryID,
		KeyID:        body.KeyID,
		Params:       body.settings(),
	}

	repository, ok := c.validSource(w, r, source)
	if !ok {
		return
	}

	newSource, err := helpers.Store(r).CreateInventorySource(source)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	recordInventorySource(r, helpers.EventLogCreate, "created", audit.ResourceInventorySourceCreate, inventory, newSource, repository)

	helpers.WriteJSON(w, http.StatusCreated, c.service.Present(newSource))
}

// GetInventorySource returns a source of a managed inventory
func (c *InventorySourceController) GetInventorySource(w http.ResponseWriter, r *http.Request) {
	inventory, ok := managedInventory(w, r)
	if !ok {
		return
	}

	sourceID, ok := helpers.GetIntParamOrAbort("inventory_source_id", w, r)
	if !ok {
		return
	}

	source, err := helpers.Store(r).GetInventorySource(inventory.ProjectID, inventory.ID, sourceID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, c.service.Present(source))
}

// UpdateInventorySource changes the name, the repository and the path of a source, its type stays
func (c *InventorySourceController) UpdateInventorySource(w http.ResponseWriter, r *http.Request) {
	inventory, ok := managedInventory(w, r)
	if !ok {
		return
	}

	sourceID, ok := helpers.GetIntParamOrAbort("inventory_source_id", w, r)
	if !ok {
		return
	}

	source, err := helpers.Store(r).GetInventorySource(inventory.ProjectID, inventory.ID, sourceID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	var body inventorySourceRequest
	if !helpers.Bind(w, r, &body) {
		return
	}

	source.Name = body.Name
	source.RepositoryID = body.RepositoryID
	source.KeyID = body.KeyID
	source.Params = body.settings()

	repository, ok := c.validSource(w, r, source)
	if !ok {
		return
	}

	if err := helpers.Store(r).UpdateInventorySource(source); err != nil {
		helpers.WriteError(w, err)
		return
	}

	recordInventorySource(r, helpers.EventLogUpdate, "updated", audit.ResourceInventorySourceUpdate, inventory, source, repository)

	w.WriteHeader(http.StatusNoContent)
}

// RemoveInventorySource deletes a source of a managed inventory with its hosts and groups
func (c *InventorySourceController) RemoveInventorySource(w http.ResponseWriter, r *http.Request) {
	inventory, ok := managedInventory(w, r)
	if !ok {
		return
	}

	sourceID, ok := helpers.GetIntParamOrAbort("inventory_source_id", w, r)
	if !ok {
		return
	}

	source, err := helpers.Store(r).GetInventorySource(inventory.ProjectID, inventory.ID, sourceID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	// A failed read only leaves the repository name out of the record
	var repository string
	if source.RepositoryID != nil {
		repo, err := helpers.Store(r).GetRepository(source.ProjectID, *source.RepositoryID)
		if err != nil {
			log.WithError(err).Error("Failed to read the repository of an inventory source")
		}
		repository = repo.Name
	}

	if err := helpers.Store(r).DeleteInventorySource(inventory.ProjectID, inventory.ID, sourceID); err != nil {
		helpers.WriteError(w, err)
		return
	}

	recordInventorySource(r, helpers.EventLogDelete, "deleted", audit.ResourceInventorySourceDelete, inventory, source, repository)

	w.WriteHeader(http.StatusNoContent)
}

// SyncInventorySource starts a sync of a source in the background
func (c *InventorySourceController) SyncInventorySource(w http.ResponseWriter, r *http.Request) {
	inventory, ok := managedInventory(w, r)
	if !ok {
		return
	}

	sourceID, ok := helpers.GetIntParamOrAbort("inventory_source_id", w, r)
	if !ok {
		return
	}

	source, err := helpers.Store(r).GetInventorySource(inventory.ProjectID, inventory.ID, sourceID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	// The sync outlives the request, so done uses only what is taken here
	recorder := helpers.Audit(r)
	ctx := context.WithoutCancel(r.Context())
	done := func(result db.InventorySyncResult, err error) {
		if err != nil || !result.Changed() {
			return
		}
		recorder.Record(ctx, audit.Event{
			Kind:      audit.ResourceInventorySourceSync,
			Target:    audit.ResourceTarget(audit.TargetInventorySource, source.ID, source.Name),
			ProjectID: inventory.ProjectID,
			Metadata: audit.InventorySourceSyncMetadata{
				InventoryID:      inventory.ID,
				InventoryName:    inventory.Name,
				HostsAdded:       result.HostsAdded,
				HostsChanged:     result.HostsChanged,
				HostsRemoved:     result.HostsRemoved,
				GroupsAdded:      result.GroupsAdded,
				GroupsChanged:    result.GroupsChanged,
				GroupsRemoved:    result.GroupsRemoved,
				Adopted:          result.Adopted,
				Skipped:          result.SkippedCount,
				VariablesChanged: result.VariablesChanged,
			},
		})
	}

	started, err := c.service.Sync(source, done)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	if !started {
		helpers.WriteErrorStatus(w, "source is already syncing", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
