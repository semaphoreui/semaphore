package sql

import (
	"errors"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/pkg/tz"
)

func (d *SqlDb) validateInventoryHostNameIsFree(host db.InventoryHost) error {
	var count int

	err := d.selectOne(&count,
		"select count(*) from project__inventory_host where inventory_id=? and name=? and id<>?",
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

func (d *SqlDb) CreateInventoryHost(host db.InventoryHost) (db.InventoryHost, error) {
	if err := host.Validate(); err != nil {
		return db.InventoryHost{}, err
	}

	if err := d.validateInventoryHostNameIsFree(host); err != nil {
		return db.InventoryHost{}, err
	}

	host.Created = tz.Now()

	query, args, err := sq.Insert("project__inventory_host").
		SetMap(map[string]any{
			"inventory_id": host.InventoryID,
			"project_id":   host.ProjectID,
			"name":         host.Name,
			"description":  host.Description,
			"enabled":      host.Enabled,
			"variables":    host.Variables,
			"created":      host.Created,
		}).
		ToSql()
	if err != nil {
		return db.InventoryHost{}, err
	}

	host.ID, err = d.insert("id", query, args...)
	if err != nil {
		// A concurrent request may have taken the name after the check
		var nameErr *common_errors.ValidationError
		if errors.As(d.validateInventoryHostNameIsFree(host), &nameErr) {
			return db.InventoryHost{}, nameErr
		}
		return db.InventoryHost{}, err
	}

	return host, nil
}

func (d *SqlDb) UpdateInventoryHost(host db.InventoryHost) error {
	if err := host.Validate(); err != nil {
		return err
	}

	if err := d.validateInventoryHostNameIsFree(host); err != nil {
		return err
	}

	query, args, err := sq.Update("project__inventory_host").
		SetMap(map[string]any{
			"name":        host.Name,
			"description": host.Description,
			"variables":   host.Variables,
		}).
		Where(sq.Eq{"id": host.ID, "inventory_id": host.InventoryID, "project_id": host.ProjectID}).
		ToSql()
	if err != nil {
		return err
	}

	res, err := d.exec(query, args...)
	if err != nil {
		// A concurrent request may have taken the name after the check
		var nameErr *common_errors.ValidationError
		if errors.As(d.validateInventoryHostNameIsFree(host), &nameErr) {
			return nameErr
		}
	}
	if err = validateMutationResult(res, err); err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}

	// MySQL counts only changed rows, so an update without changes also affects none
	var count int
	err = d.selectOne(&count,
		"select count(*) from project__inventory_host where id=? and inventory_id=? and project_id=?",
		host.ID, host.InventoryID, host.ProjectID)
	if err != nil {
		return err
	}
	if count == 0 {
		return db.ErrNotFound
	}

	return nil
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
