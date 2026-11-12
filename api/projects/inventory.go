package projects

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/services/audit"
	"github.com/semaphoreui/semaphore/services/server"

	"os"
	"path/filepath"
	"strings"
)

// InventoryMiddleware ensures an inventory exists and loads it to the context
func InventoryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		project := helpers.GetFromContext(r, "project").(db.Project)
		inventoryID, ok := helpers.GetIntParamOrAbort("inventory_id", w, r)
		if !ok {
			return
		}

		inventory, err := helpers.Store(r).GetInventory(project.ID, inventoryID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		r = helpers.SetContextValue(r, "inventory", inventory)
		next.ServeHTTP(w, r)
	})
}

func GetInventoryRefs(w http.ResponseWriter, r *http.Request) {
	inventory := helpers.GetFromContext(r, "inventory").(db.Inventory)
	refs, err := helpers.Store(r).GetInventoryRefs(inventory.ProjectID, inventory.ID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, refs)
}

// InventoryController serves the inventory API. It needs the encryption
// service because an mcp inventory carries the tokens of its servers.
type InventoryController struct {
	encryptionService server.AccessKeyEncryptionService
}

func NewInventoryController(encryptionService server.AccessKeyEncryptionService) *InventoryController {
	return &InventoryController{encryptionService: encryptionService}
}

// mcpTokenRequest is a change of the token of one MCP server. The secret is
// write-only: it is encrypted on arrival and never returned.
type mcpTokenRequest struct {
	Operation db.McpTokenOperation `json:"operation"`
	Secret    string               `json:"secret"`
}

type mcpServerRequest struct {
	// ID of an existing server of the inventory; empty for a new one.
	ID      int    `json:"id"`
	Name    string `json:"name"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
	// Token is nil when the token stays as it is.
	Token *mcpTokenRequest `json:"token"`
}

// inventoryRequest is the body of POST and PUT. McpServers shadows the
// read-only db.Inventory.McpServers. On PUT, a nil list keeps the servers as
// they are; an empty list removes them all.
type inventoryRequest struct {
	db.Inventory
	McpServers []mcpServerRequest `json:"mcp_servers"`
}

// GetInventory returns an inventory from the database
func (c *InventoryController) GetInventory(w http.ResponseWriter, r *http.Request) {
	if inventory := helpers.GetFromContext(r, "inventory"); inventory != nil {
		inv := inventory.(db.Inventory)

		if inv.Type == db.InventoryMcp {
			servers, err := helpers.Store(r).GetMcpServers(inv.ProjectID, inv.ID)
			if err != nil {
				helpers.WriteError(w, err)
				return
			}
			inv.McpServers = servers
		}

		helpers.WriteJSON(w, http.StatusOK, inv)
		return
	}

	project := helpers.GetFromContext(r, "project").(db.Project)

	params := helpers.QueryParamsWithOwner(r.URL, db.InventoryProps)

	app := r.URL.Query().Get("app")

	var types []db.InventoryType

	if app != "" {
		types = db.TemplateApp(app).InventoryTypes()
	}

	inventories, err := helpers.Store(r).GetInventories(project.ID, params, types)

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, inventories)
}

// validateMcpInventoryFields rejects the fields that make no sense for an mcp
// inventory: it has no inventory text, hosts or repository.
func validateMcpInventoryFields(inventory db.Inventory) error {
	if inventory.Inventory != "" {
		return common_errors.NewValidationError("mcp inventory must not have the inventory field, use mcp_servers")
	}
	if inventory.SSHKeyID != nil || inventory.BecomeKeyID != nil {
		return common_errors.NewValidationError("mcp inventory must not have ssh_key_id or become_key_id")
	}
	if inventory.RepositoryID != nil {
		return common_errors.NewValidationError("mcp inventory must not have repository_id")
	}
	return nil
}

// mcpServerChanges turns the request into store changes. Every token secret is
// encrypted here, before the store sees it; the plain value goes no further.
// The returned metadata counts the token changes for the audit log.
func (c *InventoryController) mcpServerChanges(projectID int, servers []mcpServerRequest) ([]db.McpServerChange, audit.InventoryMetadata, error) {
	meta := audit.InventoryMetadata{Type: string(db.InventoryMcp)}
	changes := make([]db.McpServerChange, 0, len(servers))

	for _, s := range servers {
		change := db.McpServerChange{
			Server: db.McpServer{ID: s.ID, Name: s.Name, URL: s.URL, Enabled: s.Enabled},
		}

		if s.Token != nil {
			token := db.McpTokenChange{Operation: s.Token.Operation}

			switch s.Token.Operation {
			case db.McpTokenCreate, db.McpTokenUpdate:
				if s.Token.Secret == "" {
					return nil, meta, common_errors.NewValidationError("mcp server " + s.Name + ": token secret can not be empty")
				}

				key := db.AccessKey{
					Name:      s.Name,
					Type:      db.AccessKeyString,
					String:    s.Token.Secret,
					ProjectID: &projectID,
					Owner:     db.AccessKeyMcp,
				}
				if err := c.encryptionService.SerializeSecret(&key); err != nil {
					return nil, meta, err
				}

				token.Key = db.AccessKey{Type: key.Type, Secret: key.Secret}

				if s.Token.Operation == db.McpTokenCreate {
					meta.TokensCreated++
				} else {
					meta.TokensUpdated++
				}
			case db.McpTokenDelete:
				meta.TokensDeleted++
			default:
				return nil, meta, common_errors.NewValidationError("mcp server " + s.Name + ": invalid token operation: " + string(s.Token.Operation))
			}

			change.Token = &token
		}

		changes = append(changes, change)
	}

	return changes, meta, nil
}

// AddInventory creates an inventory in the database
func (c *InventoryController) AddInventory(w http.ResponseWriter, r *http.Request) {
	project := helpers.GetFromContext(r, "project").(db.Project)

	var req inventoryRequest

	if !helpers.Bind(w, r, &req) {
		return
	}

	inventory := req.Inventory
	inventory.McpServers = nil

	if inventory.ProjectID != project.ID {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Project ID in body and URL must be the same",
		})
		return
	}

	switch inventory.Type {
	case db.InventoryStatic,
		db.InventoryStaticYaml,
		db.InventoryFile,
		db.InventoryTofuWorkspace,
		db.InventoryTerraformWorkspace,
		db.InventoryMcp:
		break
	default:
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Not supported inventory type",
		})
		return
	}

	if inventory.Type != db.InventoryMcp && len(req.McpServers) > 0 {
		helpers.WriteErrorStatus(w, "mcp_servers is allowed only for an mcp inventory", http.StatusBadRequest)
		return
	}

	if inventory.Type == db.InventoryMcp {
		if err := validateMcpInventoryFields(inventory); err != nil {
			helpers.WriteError(w, err)
			return
		}
	}

	err := db.ValidateInventory(helpers.Store(r), &inventory)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	var newInventory db.Inventory
	meta := audit.InventoryMetadata{Type: string(inventory.Type)}

	if inventory.Type == db.InventoryMcp {
		var changes []db.McpServerChange
		changes, meta, err = c.mcpServerChanges(project.ID, req.McpServers)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		newInventory, newInventory.McpServers, err = helpers.Store(r).CreateMcpInventory(inventory, changes)
		meta.McpServers = len(newInventory.McpServers)
	} else {
		newInventory, err = helpers.Store(r).CreateInventory(inventory)
	}

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogCreate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   project.ID,
		ObjectType:  db.EventInventory,
		ObjectID:    newInventory.ID,
		Description: fmt.Sprintf("Inventory %s created", inventory.Name),
	})

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceInventoryCreate,
		Target:    audit.ResourceTarget(audit.TargetInventory, newInventory.ID, newInventory.Name),
		ProjectID: project.ID,
		Metadata:  meta,
	})

	helpers.WriteJSON(w, http.StatusCreated, newInventory)
}

// IsValidInventoryPath tests a path to ensure it is below the cwd
func IsValidInventoryPath(path string) bool {

	currentPath, err := os.Getwd()
	if err != nil {
		return false
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	relPath, err := filepath.Rel(currentPath, absPath)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(relPath, "..")
}

// UpdateInventory writes updated values to an existing inventory item in the database
func (c *InventoryController) UpdateInventory(w http.ResponseWriter, r *http.Request) {
	oldInventory := helpers.GetFromContext(r, "inventory").(db.Inventory)

	var req inventoryRequest

	if !helpers.Bind(w, r, &req) {
		return
	}

	inventory := req.Inventory
	inventory.McpServers = nil

	if inventory.ID != oldInventory.ID {
		helpers.WriteErrorStatus(w,
			"Inventory ID in body and URL must be the same",
			http.StatusBadRequest)
		return
	}

	if inventory.ProjectID != oldInventory.ProjectID {
		helpers.WriteErrorStatus(w,
			"project ID in body and URL must be the same",
			http.StatusBadRequest)
		return
	}

	// An mcp inventory keeps its servers and tokens in other tables, and other
	// types are used by templates that can not run with MCP servers.
	if (inventory.Type == db.InventoryMcp) != (oldInventory.Type == db.InventoryMcp) {
		helpers.WriteErrorStatus(w,
			"inventory type can not be changed to or from mcp",
			http.StatusBadRequest)
		return
	}

	switch inventory.Type {
	case db.InventoryTerraformWorkspace, db.InventoryTofuWorkspace:
	case db.InventoryStatic, db.InventoryStaticYaml:
	case db.InventoryMcp:
		if err := validateMcpInventoryFields(inventory); err != nil {
			helpers.WriteError(w, err)
			return
		}
	case db.InventoryFile:
		if !IsValidInventoryPath(inventory.Inventory) {
			helpers.WriteErrorStatus(w, "Invalid inventory file pathname. Must be: path/to/inventory.", http.StatusBadRequest)
			return
		}
	default:
		helpers.WriteErrorStatus(w,
			"unknown inventory type: "+string(inventory.Type),
			http.StatusBadRequest)
		return
	}

	if inventory.Type != db.InventoryMcp && len(req.McpServers) > 0 {
		helpers.WriteErrorStatus(w, "mcp_servers is allowed only for an mcp inventory", http.StatusBadRequest)
		return
	}

	if err := db.ValidateInventory(helpers.Store(r), &inventory); err != nil {
		helpers.WriteError(w, err)
		return
	}

	meta := audit.InventoryMetadata{Type: string(inventory.Type)}

	if inventory.Type == db.InventoryMcp && req.McpServers != nil {
		changes, mcpMeta, err := c.mcpServerChanges(inventory.ProjectID, req.McpServers)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		servers, err := helpers.Store(r).UpdateMcpInventory(inventory, changes)
		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		meta = mcpMeta
		meta.McpServers = len(servers)
	} else if err := helpers.Store(r).UpdateInventory(inventory); err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogUpdate, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   oldInventory.ProjectID,
		ObjectType:  db.EventInventory,
		ObjectID:    oldInventory.ID,
		Description: fmt.Sprintf("Inventory %s updated", inventory.Name),
	})

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceInventoryUpdate,
		Target:    audit.ResourceTarget(audit.TargetInventory, oldInventory.ID, inventory.Name),
		ProjectID: oldInventory.ProjectID,
		Metadata:  meta,
	})

	w.WriteHeader(http.StatusNoContent)
}

// RemoveInventory deletes an inventory from the database
func RemoveInventory(w http.ResponseWriter, r *http.Request) {
	inventory := helpers.GetFromContext(r, "inventory").(db.Inventory)
	err := helpers.Store(r).DeleteInventory(inventory.ProjectID, inventory.ID)
	if errors.Is(err, db.ErrInvalidOperation) {
		helpers.WriteJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Inventory is in use by one or more templates",
			"inUse": true,
		})
		return
	}

	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	helpers.EventLog(r, helpers.EventLogDelete, helpers.EventLogItem{
		UserID:      helpers.UserFromContext(r).ID,
		ProjectID:   inventory.ProjectID,
		ObjectType:  db.EventInventory,
		ObjectID:    inventory.ID,
		Description: fmt.Sprintf("Inventory %s deleted", inventory.Name),
	})

	helpers.Audit(r).Record(r.Context(), audit.Event{
		Kind:      audit.ResourceInventoryDelete,
		Target:    audit.ResourceTarget(audit.TargetInventory, inventory.ID, inventory.Name),
		ProjectID: inventory.ProjectID,
	})

	w.WriteHeader(http.StatusNoContent)
}
