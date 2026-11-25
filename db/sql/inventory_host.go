package sql

import (
	"database/sql"
	"errors"
	"slices"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/pkg/tz"
)

func (d *SqlDb) validateInventoryHostNameIsFree(q gorp.SqlExecutor, host db.InventoryHost) error {
	count, err := q.SelectInt(
		d.PrepareQuery("select count(*) from project__inventory_host where inventory_id=? and name=? and id<>?"),
		host.InventoryID, host.Name, host.ID)
	if err != nil {
		return err
	}

	if count > 0 {
		return common_errors.NewValidationError("a host with this name already exists in the inventory")
	}

	return nil
}

func (d *SqlDb) GetInventoryHosts(projectID, inventoryID int, params db.InventoryHostListParams) (hosts []db.InventoryHost, total int, err error) {
	where := sq.And{sq.Eq{"project_id": projectID}, sq.Eq{"inventory_id": inventoryID}}

	if params.Search != "" {
		escaper := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_")
		where = append(where, sq.Expr("lower(name) like ? escape '!'", "%"+escaper.Replace(strings.ToLower(params.Search))+"%"))
	}

	if params.GroupID != 0 {
		var children map[int][]int
		children, err = d.inventoryGroupChildren(d.Sql(), projectID, inventoryID, nil)
		if err != nil {
			return
		}
		if _, ok := children[params.GroupID]; !ok {
			err = db.ErrNotFound
			return
		}
		groups := append(db.InventoryGroupDescendants(children, params.GroupID), params.GroupID)
		where = append(where, sq.Expr("id in (?)",
			sq.Select("host_id").From("project__inventory_group_host").Where(sq.Eq{"group_id": groups})))
	}

	if params.IDs != nil {
		where = append(where, sq.Eq{"id": params.IDs})
	}

	if params.Ungrouped {
		where = append(where, sq.Expr(
			"not exists (select 1 from project__inventory_group_host gh where gh.host_id = project__inventory_host.id)"))
	}

	query, args, err := sq.Select("count(*)").From("project__inventory_host").Where(where).ToSql()
	if err != nil {
		return
	}

	err = d.selectOne(&total, query, args...)
	if err != nil {
		return
	}

	q := sq.Select("*").From("project__inventory_host").Where(where).OrderBy("name")
	if params.Count > 0 {
		q = q.Limit(uint64(params.Count)).Offset(uint64(params.Offset))
	}

	query, args, err = q.ToSql()
	if err != nil {
		return
	}

	_, err = d.selectAll(&hosts, query, args...)
	if err != nil || len(hosts) == 0 {
		return
	}

	var links []db.InventoryGroupHost
	if params.Count > 0 {
		ids := make([]int, len(hosts))
		for i, host := range hosts {
			ids[i] = host.ID
		}
		query, args, err = sq.Select("group_id", "host_id").From("project__inventory_group_host").
			Where(sq.Eq{"host_id": ids}).
			OrderBy("group_id").
			ToSql()
		if err != nil {
			return
		}
		_, err = d.selectAll(&links, query, args...)
	} else {
		links, err = d.GetInventoryGroupHosts(projectID, inventoryID)
	}
	if err != nil {
		return
	}
	groups := map[int][]int{}
	for _, link := range links {
		groups[link.HostID] = append(groups[link.HostID], link.GroupID)
	}
	for i := range hosts {
		hosts[i].Groups = append([]int{}, groups[hosts[i].ID]...)
	}
	return
}

func (d *SqlDb) GetInventoryHost(projectID, inventoryID, hostID int) (host db.InventoryHost, err error) {
	query, args, err := sq.Select("*").From("project__inventory_host").
		Where(sq.Eq{"id": hostID, "inventory_id": inventoryID, "project_id": projectID}).
		ToSql()
	if err != nil {
		return
	}

	err = d.selectOne(&host, query, args...)
	if err != nil {
		return
	}

	host.Groups = []int{}
	_, err = d.selectAll(&host.Groups,
		"select group_id from project__inventory_group_host where host_id=? order by group_id", hostID)
	return
}

func (d *SqlDb) GetEnabledInventoryHosts(projectID, inventoryID int) (hosts []db.InventoryHost, err error) {
	query, args, err := sq.Select("*").From("project__inventory_host").
		Where(sq.Eq{"project_id": projectID, "inventory_id": inventoryID, "enabled": true}).
		OrderBy("name").
		ToSql()
	if err != nil {
		return
	}

	_, err = d.selectAll(&hosts, query, args...)
	return
}

// inventoryHostWriteError returns the name error when a concurrent request took the name, older nodes write hosts without the lock
func (d *SqlDb) inventoryHostWriteError(host db.InventoryHost, err error) error {
	var nameErr *common_errors.ValidationError
	if errors.As(d.validateInventoryHostNameIsFree(d.Sql(), host), &nameErr) {
		return nameErr
	}
	return err
}

func (d *SqlDb) CreateInventoryHost(host db.InventoryHost, addGroups []int) (db.InventoryHost, error) {
	if err := host.Validate(); err != nil {
		return db.InventoryHost{}, err
	}
	addGroups = slices.Compact(slices.Sorted(slices.Values(addGroups)))

	host.Created = tz.Now()
	host.VariableCount = db.CountInventoryVariables(host.Variables)
	host.Groups = append([]int{}, addGroups...)

	query, args, err := sq.Insert("project__inventory_host").
		SetMap(map[string]any{
			"inventory_id":   host.InventoryID,
			"project_id":     host.ProjectID,
			"name":           host.Name,
			"description":    host.Description,
			"enabled":        host.Enabled,
			"variables":      host.Variables,
			"variable_count": host.VariableCount,
			"created":        host.Created,
		}).
		ToSql()
	if err != nil {
		return db.InventoryHost{}, err
	}

	// Only a change of links takes the inventory lock
	if len(addGroups) == 0 {
		if err = d.validateInventoryHostNameIsFree(d.Sql(), host); err != nil {
			return db.InventoryHost{}, err
		}
		host.ID, err = d.insert("id", query, args...)
		if err != nil {
			return db.InventoryHost{}, d.inventoryHostWriteError(host, err)
		}
		return host, nil
	}

	tx, err := d.Sql().Begin()
	if err != nil {
		return db.InventoryHost{}, err
	}

	if err = d.lockInventory(tx, host.ProjectID, host.InventoryID); err != nil {
		_ = tx.Rollback()
		return db.InventoryHost{}, err
	}

	if _, err = d.inventoryGroupChildren(tx, host.ProjectID, host.InventoryID, addGroups); err != nil {
		_ = tx.Rollback()
		return db.InventoryHost{}, err
	}

	if err = d.validateInventoryHostNameIsFree(tx, host); err != nil {
		_ = tx.Rollback()
		return db.InventoryHost{}, err
	}

	host.ID, err = d.connection.InsertTx(tx, "id", query, args...)
	if err != nil {
		_ = tx.Rollback()
		return db.InventoryHost{}, d.inventoryHostWriteError(host, err)
	}

	links := make([]db.InventoryGroupHost, len(addGroups))
	for i, group := range addGroups {
		links[i] = db.InventoryGroupHost{GroupID: group, HostID: host.ID}
	}
	if err = d.insertInventoryGroupHosts(tx, links); err != nil {
		_ = tx.Rollback()
		return db.InventoryHost{}, err
	}

	return host, tx.Commit()
}

func (d *SqlDb) UpdateInventoryHost(host db.InventoryHost, addGroups, removeGroups []int) (added, removed []int, err error) {
	if err = host.Validate(); err != nil {
		return
	}
	if addGroups, removeGroups, err = addRemoveIDs(addGroups, removeGroups); err != nil {
		return
	}

	query, args, err := sq.Update("project__inventory_host").
		SetMap(map[string]any{
			"name":           host.Name,
			"description":    host.Description,
			"variables":      host.Variables,
			"variable_count": db.CountInventoryVariables(host.Variables),
		}).
		Where(sq.Eq{"id": host.ID, "inventory_id": host.InventoryID, "project_id": host.ProjectID}).
		ToSql()
	if err != nil {
		return
	}

	// Only a change of links takes the inventory lock
	if len(addGroups) == 0 && len(removeGroups) == 0 {
		if err = d.validateInventoryHostNameIsFree(d.Sql(), host); err != nil {
			return
		}
		var res sql.Result
		if res, err = d.exec(query, args...); err != nil {
			err = d.inventoryHostWriteError(host, err)
		}
		if err = validateMutationResult(res, err); err != nil {
			return
		}
		var affected int64
		if affected, err = res.RowsAffected(); err != nil || affected > 0 {
			return
		}

		// MySQL counts only changed rows, so an update without changes also affects none
		var count int
		err = d.selectOne(&count,
			"select count(*) from project__inventory_host where id=? and inventory_id=? and project_id=?",
			host.ID, host.InventoryID, host.ProjectID)
		if err == nil && count == 0 {
			err = db.ErrNotFound
		}
		return
	}

	tx, err := d.Sql().Begin()
	if err != nil {
		return
	}

	if err = d.lockInventory(tx, host.ProjectID, host.InventoryID); err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}

	if len(addGroups) > 0 {
		if _, err = d.inventoryGroupChildren(tx, host.ProjectID, host.InventoryID, addGroups); err != nil {
			_ = tx.Rollback()
			return nil, nil, err
		}
	}

	count, err := tx.SelectInt(
		d.PrepareQuery("select count(*) from project__inventory_host where id=? and inventory_id=? and project_id=?"),
		host.ID, host.InventoryID, host.ProjectID)
	if err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}
	if count == 0 {
		_ = tx.Rollback()
		return nil, nil, db.ErrNotFound
	}

	if err = d.validateInventoryHostNameIsFree(tx, host); err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}

	var current []int
	_, err = tx.Select(&current, d.PrepareQuery("select group_id from project__inventory_group_host where host_id=?"), host.ID)
	if err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}
	var links []db.InventoryGroupHost
	for _, id := range addGroups {
		if !slices.Contains(current, id) {
			added = append(added, id)
			links = append(links, db.InventoryGroupHost{GroupID: id, HostID: host.ID})
		}
	}
	for _, id := range removeGroups {
		if slices.Contains(current, id) {
			removed = append(removed, id)
		}
	}

	if _, err = d.execTx(tx, query, args...); err != nil {
		_ = tx.Rollback()
		return nil, nil, d.inventoryHostWriteError(host, err)
	}

	if err = d.insertInventoryGroupHosts(tx, links); err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}
	for _, group := range removed {
		_, err = d.execTx(tx, "delete from project__inventory_group_host where group_id=? and host_id=?", group, host.ID)
		if err != nil {
			_ = tx.Rollback()
			return nil, nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, nil, err
	}
	return added, removed, nil
}

func (d *SqlDb) SetInventoryHostEnabled(projectID, inventoryID, hostID int, enabled bool) (bool, error) {
	res, err := d.exec(
		"update project__inventory_host set enabled=? where id=? and inventory_id=? and project_id=? and enabled<>?",
		enabled, hostID, inventoryID, projectID, enabled)
	if err != nil {
		return false, err
	}

	affected, err := res.RowsAffected()
	return affected > 0, err
}

func (d *SqlDb) DeleteInventoryHost(projectID, inventoryID, hostID int) error {
	return requireDeletedRow(d.exec(
		"delete from project__inventory_host where id=? and inventory_id=? and project_id=?",
		hostID, inventoryID, projectID))
}
