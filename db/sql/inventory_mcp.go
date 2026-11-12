package sql

import (
	"fmt"
	"strconv"
	"time"

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

// mcpServerWithTokenSelect reads servers and the keys of their tokens in one
// statement: two statements, even in one transaction, may see different
// commits (READ COMMITTED is the PostgreSQL default). It lists every column of
// access_key, each as key_<column>; mcpServerWithTokenRow maps them.
const mcpServerWithTokenSelect = "select m.id, m.inventory_id, m.name, m.url, m.enabled, " +
	"k.id as key_id, k.name as key_name, k.type as key_type, k.project_id as key_project_id, " +
	"k.secret as key_secret, k.plain as key_plain, k.storage_id as key_storage_id, " +
	"k.environment_id as key_environment_id, k.mcp_id as key_mcp_id, k.user_id as key_user_id, " +
	"k.task_id as key_task_id, k.expire_at as key_expire_at, k.owner as key_owner, " +
	"k.source_storage_id as key_source_storage_id, k.source_storage_key as key_source_storage_key, " +
	"k.source_storage_type as key_source_storage_type, k.synchronized as key_synchronized " +
	"from project__inventory_mcp m " +
	"join project__inventory i on i.id = m.inventory_id " +
	"left join access_key k on k.mcp_id = m.id and k.owner = ? and k.project_id = ? " +
	"where i.id = ? and i.project_id = ? " +
	"order by m.id"

// mcpServerWithTokenRow is a row of mcpServerWithTokenSelect. The key columns
// are all null for a server without a token.
type mcpServerWithTokenRow struct {
	ID          int    `db:"id"`
	InventoryID int    `db:"inventory_id"`
	Name        string `db:"name"`
	URL         string `db:"url"`
	Enabled     bool   `db:"enabled"`

	KeyID                *int                           `db:"key_id"`
	KeyName              *string                        `db:"key_name"`
	KeyType              *db.AccessKeyType              `db:"key_type"`
	KeyProjectID         *int                           `db:"key_project_id"`
	KeySecret            *string                        `db:"key_secret"`
	KeyPlain             *string                        `db:"key_plain"`
	KeyStorageID         *int                           `db:"key_storage_id"`
	KeyEnvironmentID     *int                           `db:"key_environment_id"`
	KeyMcpID             *int                           `db:"key_mcp_id"`
	KeyUserID            *int                           `db:"key_user_id"`
	KeyTaskID            *int                           `db:"key_task_id"`
	KeyExpireAt          *time.Time                     `db:"key_expire_at"`
	KeyOwner             *db.AccessKeyOwner             `db:"key_owner"`
	KeySourceStorageID   *int                           `db:"key_source_storage_id"`
	KeySourceStorageKey  *string                        `db:"key_source_storage_key"`
	KeySourceStorageType *db.AccessKeySourceStorageType `db:"key_source_storage_type"`
	KeySynchronized      *bool                          `db:"key_synchronized"`
}

func (r mcpServerWithTokenRow) server() db.McpServer {
	return db.McpServer{
		ID:          r.ID,
		InventoryID: r.InventoryID,
		Name:        r.Name,
		URL:         r.URL,
		Enabled:     r.Enabled,
		HasToken:    r.KeyID != nil,
	}
}

// token returns the key of the row, nil for a server without a token.
func (r mcpServerWithTokenRow) token() *db.AccessKey {
	if r.KeyID == nil {
		return nil
	}

	key := db.AccessKey{
		ID:                *r.KeyID,
		ProjectID:         r.KeyProjectID,
		Secret:            r.KeySecret,
		Plain:             r.KeyPlain,
		StorageID:         r.KeyStorageID,
		EnvironmentID:     r.KeyEnvironmentID,
		McpID:             r.KeyMcpID,
		UserID:            r.KeyUserID,
		TaskID:            r.KeyTaskID,
		ExpireAt:          r.KeyExpireAt,
		SourceStorageID:   r.KeySourceStorageID,
		SourceStorageKey:  r.KeySourceStorageKey,
		SourceStorageType: r.KeySourceStorageType,
	}

	if r.KeyName != nil {
		key.Name = *r.KeyName
	}
	if r.KeyType != nil {
		key.Type = *r.KeyType
	}
	if r.KeyOwner != nil {
		key.Owner = *r.KeyOwner
	}
	if r.KeySynchronized != nil {
		key.Synchronized = *r.KeySynchronized
	}

	return &key
}

func (d *SqlDb) GetMcpServersWithTokens(projectID int, inventoryID int) ([]db.McpServerWithToken, error) {
	rows := make([]mcpServerWithTokenRow, 0)

	if _, err := d.selectAll(&rows, mcpServerWithTokenSelect,
		db.AccessKeyMcp, projectID, inventoryID, projectID); err != nil {
		return nil, err
	}

	servers := make([]db.McpServerWithToken, 0, len(rows))
	for _, row := range rows {
		servers = append(servers, db.McpServerWithToken{Server: row.server(), Token: row.token()})
	}

	return servers, nil
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
		//
		// The number of affected rows does not tell whether the inventory
		// exists: MySQL counts changed rows, so an edit of the servers alone
		// reports 0 for a row that is there. setMcpServers looks the row up
		// itself, after this update holds it, and returns db.ErrNotFound.
		_, e := d.execTx(tx, inventoryUpdateQuery+" and type=?",
			append(inventoryUpdateArgs(inventory), db.InventoryMcp)...)
		if e != nil {
			return e
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
