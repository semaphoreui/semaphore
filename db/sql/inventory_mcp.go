package sql

import (
	"fmt"
	"strconv"

	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

// mcpServerSelect reads servers together with whether each has a token. The
// join on the inventory scopes every query by project.
const mcpServerSelect = "select m.id, m.inventory_id, m.name, m.url, m.enabled, " +
	"(k.id is not null) as has_token " +
	"from project__inventory_mcp m " +
	"join project__inventory i on i.id = m.inventory_id " +
	"left join access_key k on k.mcp_id = m.id and k.owner = ? " +
	"where i.id = ? and i.project_id = ? " +
	"order by m.id"

func (d *SqlDb) GetMcpServers(projectID int, inventoryID int) ([]db.McpServer, error) {
	servers := make([]db.McpServer, 0)

	_, err := d.selectAll(&servers, mcpServerSelect, db.AccessKeyMcp, inventoryID, projectID)

	return servers, err
}

func (d *SqlDb) GetMcpServerTokens(projectID int, inventoryID int) ([]db.AccessKey, error) {
	keys := make([]db.AccessKey, 0)

	_, err := d.selectAll(&keys,
		"select k.* from access_key k "+
			"join project__inventory_mcp m on m.id = k.mcp_id "+
			"join project__inventory i on i.id = m.inventory_id "+
			"where i.id = ? and i.project_id = ? and k.project_id = ? and k.owner = ? "+
			"order by k.mcp_id",
		inventoryID, projectID, projectID, db.AccessKeyMcp)

	return keys, err
}

func (d *SqlDb) SetMcpServers(projectID int, inventoryID int, changes []db.McpServerChange) (servers []db.McpServer, err error) {
	if err = validateMcpServerChanges(changes); err != nil {
		return
	}

	err = d.mcpTx(func(tx *gorp.Transaction) error {
		if e := d.setMcpServers(tx, projectID, inventoryID, changes); e != nil {
			return e
		}
		var e error
		servers, e = d.selectMcpServersTx(tx, projectID, inventoryID)
		return e
	})

	return
}

func (d *SqlDb) CreateMcpInventory(inventory db.Inventory, changes []db.McpServerChange) (newInventory db.Inventory, servers []db.McpServer, err error) {
	if inventory.Type != db.InventoryMcp {
		err = common_errors.NewValidationError("inventory type must be mcp")
		return
	}

	if err = validateMcpServerChanges(changes); err != nil {
		return
	}

	err = d.mcpTx(func(tx *gorp.Transaction) error {
		id, e := d.connection.InsertTx(tx, "id", inventoryInsertQuery, inventoryInsertArgs(inventory)...)
		if e != nil {
			return e
		}

		newInventory = inventory
		newInventory.ID = id

		if e = d.setMcpServers(tx, inventory.ProjectID, id, changes); e != nil {
			return e
		}

		servers, e = d.selectMcpServersTx(tx, inventory.ProjectID, id)
		return e
	})

	if err != nil {
		// The row was rolled back; do not hand out its ID.
		return db.Inventory{}, nil, err
	}

	return
}

func (d *SqlDb) UpdateMcpInventory(inventory db.Inventory, changes []db.McpServerChange) (servers []db.McpServer, err error) {
	if inventory.Type != db.InventoryMcp {
		err = common_errors.NewValidationError("inventory type must be mcp")
		return
	}

	if err = validateMcpServerChanges(changes); err != nil {
		return
	}

	err = d.mcpTx(func(tx *gorp.Transaction) error {
		// The type check keeps an inventory of another type from being turned
		// into mcp here: its own data would stay behind, unseen.
		res, e := d.execTx(tx, inventoryUpdateQuery+" and type=?",
			append(inventoryUpdateArgs(inventory), db.InventoryMcp)...)
		if e != nil {
			return e
		}

		if n, e := res.RowsAffected(); e != nil {
			return e
		} else if n == 0 {
			return db.ErrNotFound
		}

		if e = d.setMcpServers(tx, inventory.ProjectID, inventory.ID, changes); e != nil {
			return e
		}

		servers, e = d.selectMcpServersTx(tx, inventory.ProjectID, inventory.ID)
		return e
	})

	return
}

// validateMcpServerChanges checks the whole list before anything is written.
func validateMcpServerChanges(changes []db.McpServerChange) error {
	list := make([]db.McpServer, len(changes))
	for i, c := range changes {
		list[i] = c.Server

		if c.Token != nil {
			if err := c.Token.Validate(); err != nil {
				return err
			}
		}
	}

	return db.ValidateMcpServers(list)
}

// mcpTx runs fn in a transaction, committing only when fn succeeds.
func (d *SqlDb) mcpTx(fn func(tx *gorp.Transaction) error) (err error) {
	tx, err := d.Sql().Begin()
	if err != nil {
		return
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = fn(tx); err != nil {
		return
	}

	err = tx.Commit()
	return
}

func (d *SqlDb) selectMcpServersTx(tx *gorp.Transaction, projectID int, inventoryID int) ([]db.McpServer, error) {
	servers := make([]db.McpServer, 0)
	_, err := tx.Select(&servers, d.PrepareQuery(mcpServerSelect), db.AccessKeyMcp, inventoryID, projectID)
	return servers, err
}

func (d *SqlDb) setMcpServers(tx *gorp.Transaction, projectID int, inventoryID int, changes []db.McpServerChange) error {
	count, err := tx.SelectInt(
		d.PrepareQuery("select count(*) from project__inventory where id = ? and project_id = ? and type = ?"),
		inventoryID, projectID, db.InventoryMcp)
	if err != nil {
		return err
	}
	if count == 0 {
		return db.ErrNotFound
	}

	var existing []db.McpServer
	if _, err = tx.Select(&existing, d.PrepareQuery(mcpServerSelect), db.AccessKeyMcp, inventoryID, projectID); err != nil {
		return err
	}

	existingByID := make(map[int]db.McpServer, len(existing))
	for _, s := range existing {
		existingByID[s.ID] = s
	}

	kept := make(map[int]struct{}, len(changes))
	for _, c := range changes {
		if c.Server.ID == 0 {
			continue
		}
		if _, ok := existingByID[c.Server.ID]; !ok {
			return common_errors.NewValidationError(fmt.Sprintf("mcp server %d does not belong to the inventory", c.Server.ID))
		}
		if _, ok := kept[c.Server.ID]; ok {
			return common_errors.NewValidationError(fmt.Sprintf("mcp server %d is listed twice", c.Server.ID))
		}
		kept[c.Server.ID] = struct{}{}
	}

	// A server missing from the list goes together with its token. The key is
	// deleted explicitly rather than left to the cascade, which SQLite applies
	// only with foreign_keys on.
	for _, s := range existing {
		if _, ok := kept[s.ID]; ok {
			continue
		}
		if err = d.deleteMcpToken(tx, projectID, s.ID); err != nil {
			return err
		}
		if _, err = d.execTx(tx, "delete from project__inventory_mcp where id = ?", s.ID); err != nil {
			return err
		}
	}

	// Renamed servers first get a name no valid server can have ('~' is not
	// allowed), so that two servers can swap names within one request
	// without hitting the unique (inventory_id, name) index.
	for _, c := range changes {
		if c.Server.ID == 0 || existingByID[c.Server.ID].Name == c.Server.Name {
			continue
		}
		if _, err = d.execTx(tx, "update project__inventory_mcp set name = ? where id = ?",
			"~"+strconv.Itoa(c.Server.ID), c.Server.ID); err != nil {
			return err
		}
	}

	for _, c := range changes {
		serverID := c.Server.ID
		hasToken := false

		if serverID == 0 {
			serverID, err = d.connection.InsertTx(tx, "id",
				"insert into project__inventory_mcp (inventory_id, name, url, enabled) values (?, ?, ?, ?)",
				inventoryID, c.Server.Name, c.Server.URL, c.Server.Enabled)
			if err != nil {
				return err
			}
		} else {
			hasToken = existingByID[serverID].HasToken

			if _, err = d.execTx(tx, "update project__inventory_mcp set name = ?, url = ?, enabled = ? where id = ?",
				c.Server.Name, c.Server.URL, c.Server.Enabled, serverID); err != nil {
				return err
			}

			// The key carries the server name to be recognisable in the audit log.
			if _, err = d.execTx(tx, "update access_key set name = ? where mcp_id = ? and owner = ? and project_id = ?",
				c.Server.Name, serverID, db.AccessKeyMcp, projectID); err != nil {
				return err
			}
		}

		if c.Token == nil {
			continue
		}

		if err = d.applyMcpTokenChange(tx, projectID, serverID, c.Server.Name, hasToken, *c.Token); err != nil {
			return err
		}
	}

	return nil
}

func (d *SqlDb) applyMcpTokenChange(
	tx *gorp.Transaction,
	projectID int,
	serverID int,
	serverName string,
	hasToken bool,
	change db.McpTokenChange,
) error {
	keyType := change.Key.Type
	if keyType == "" {
		keyType = db.AccessKeyString
	}

	switch change.Operation {
	case db.McpTokenCreate:
		if hasToken {
			return common_errors.NewValidationError("mcp server " + serverName + " already has a token, use update")
		}

		// Owner, project and server come from here, never from the caller's key.
		_, err := d.connection.InsertTx(tx, "id",
			"insert into access_key (name, type, project_id, secret, owner, mcp_id) values (?, ?, ?, ?, ?, ?)",
			serverName, keyType, projectID, change.Key.Secret, db.AccessKeyMcp, serverID)
		return err

	case db.McpTokenUpdate:
		if !hasToken {
			return common_errors.NewValidationError("mcp server " + serverName + " has no token, use create")
		}

		_, err := d.execTx(tx, "update access_key set type = ?, secret = ? where mcp_id = ? and owner = ? and project_id = ?",
			keyType, change.Key.Secret, serverID, db.AccessKeyMcp, projectID)
		return err

	case db.McpTokenDelete:
		return d.deleteMcpToken(tx, projectID, serverID)

	default:
		return common_errors.NewValidationError("invalid mcp token operation: " + string(change.Operation))
	}
}

func (d *SqlDb) deleteMcpToken(tx *gorp.Transaction, projectID int, serverID int) error {
	_, err := d.execTx(tx, "delete from access_key where mcp_id = ? and owner = ? and project_id = ?",
		serverID, db.AccessKeyMcp, projectID)
	return err
}
