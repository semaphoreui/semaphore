package sql

import (
	"errors"
	"fmt"
	"slices"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/pkg/tz"
)

// inventoryGroupChildren maps every group of the inventory to its direct children and rejects a named group from elsewhere
func (d *SqlDb) inventoryGroupChildren(q gorp.SqlExecutor, projectID, inventoryID int, named []int) (map[int][]int, error) {
	var ids []int
	_, err := q.Select(&ids, d.PrepareQuery("select id from project__inventory_group where project_id=? and inventory_id=?"),
		projectID, inventoryID)
	if err != nil {
		return nil, err
	}

	var links []struct {
		ParentID int `db:"parent_id"`
		ChildID  int `db:"child_id"`
	}
	_, err = q.Select(&links, d.PrepareQuery(`
		select c.parent_id, c.child_id
		from project__inventory_group_child c
		join project__inventory_group g on g.id = c.parent_id
		where g.project_id=? and g.inventory_id=?
		order by c.parent_id, c.child_id
	`), projectID, inventoryID)
	if err != nil {
		return nil, err
	}

	children := make(map[int][]int, len(ids))
	for _, id := range ids {
		children[id] = nil
	}
	for _, link := range links {
		children[link.ParentID] = append(children[link.ParentID], link.ChildID)
	}
	for _, id := range named {
		if _, ok := children[id]; !ok {
			return nil, common_errors.NewValidationError(fmt.Sprintf("group %d is not in this inventory", id))
		}
	}
	return children, nil
}

// lockInventory starts every change of links, so the changes of one inventory wait for each other
func (d *SqlDb) lockInventory(tx *gorp.Transaction, projectID, inventoryID int) error {
	_, err := d.execTx(tx, "update project__inventory set name=name where id=? and project_id=?", inventoryID, projectID)
	if err != nil {
		return err
	}

	// MySQL counts only changed rows, so the update can not tell whether the inventory exists
	count, err := tx.SelectInt(d.PrepareQuery("select count(*) from project__inventory where id=? and project_id=?"),
		inventoryID, projectID)
	if err != nil {
		return err
	}
	if count == 0 {
		return db.ErrNotFound
	}
	return nil
}

// addRemoveIDs drops repeated ids and rejects an id that is both added and removed
func addRemoveIDs(add, remove []int) ([]int, []int, error) {
	add = slices.Compact(slices.Sorted(slices.Values(add)))
	remove = slices.Compact(slices.Sorted(slices.Values(remove)))
	for _, id := range remove {
		if slices.Contains(add, id) {
			return nil, nil, common_errors.NewValidationError(fmt.Sprintf("%d can not be both added and removed", id))
		}
	}
	return add, remove, nil
}

func (d *SqlDb) validateInventoryGroupNameIsFree(q gorp.SqlExecutor, group db.InventoryGroup) error {
	count, err := q.SelectInt(
		d.PrepareQuery("select count(*) from project__inventory_group where inventory_id=? and name=? and id<>?"),
		group.InventoryID, group.Name, group.ID)
	if err != nil {
		return err
	}

	if count > 0 {
		return common_errors.NewValidationError("a group with this name already exists in the inventory")
	}

	return nil
}

func (d *SqlDb) GetInventoryGroups(projectID, inventoryID int) (groups []db.InventoryGroup, err error) {
	_, err = d.selectAll(&groups,
		"select * from project__inventory_group where project_id=? and inventory_id=? order by name",
		projectID, inventoryID)
	if err != nil {
		return
	}

	children, err := d.inventoryGroupChildren(d.Sql(), projectID, inventoryID, nil)
	if err != nil {
		return
	}

	links, err := d.GetInventoryGroupHosts(projectID, inventoryID)
	if err != nil {
		return
	}
	hosts := map[int][]int{}
	for _, link := range links {
		hosts[link.GroupID] = append(hosts[link.GroupID], link.HostID)
	}

	for i := range groups {
		id := groups[i].ID
		counted := map[int]bool{}
		for _, group := range append(db.InventoryGroupDescendants(children, id), id) {
			for _, host := range hosts[group] {
				counted[host] = true
			}
		}
		groups[i].Children = append([]int{}, children[id]...)
		groups[i].HostCount = len(counted)
	}
	return
}

func (d *SqlDb) GetInventoryGroup(projectID, inventoryID, groupID int) (group db.InventoryGroup, err error) {
	err = d.selectOne(&group,
		"select * from project__inventory_group where id=? and inventory_id=? and project_id=?",
		groupID, inventoryID, projectID)
	if err != nil {
		return
	}

	children, err := d.inventoryGroupChildren(d.Sql(), projectID, inventoryID, nil)
	if err != nil {
		return
	}
	group.Children = append([]int{}, children[groupID]...)

	query, args, err := sq.Select("count(distinct host_id)").From("project__inventory_group_host").
		Where(sq.Eq{"group_id": append(db.InventoryGroupDescendants(children, groupID), groupID)}).
		ToSql()
	if err != nil {
		return
	}
	err = d.selectOne(&group.HostCount, query, args...)
	return
}

func (d *SqlDb) GetInventoryGroupHosts(projectID, inventoryID int) (links []db.InventoryGroupHost, err error) {
	_, err = d.selectAll(&links, `
		select gh.group_id, gh.host_id
		from project__inventory_group_host gh
		join project__inventory_group g on g.id = gh.group_id
		where g.project_id=? and g.inventory_id=?
		order by gh.group_id, gh.host_id
	`, projectID, inventoryID)
	return
}

func (d *SqlDb) CreateInventoryGroup(group db.InventoryGroup, addChildren []int) (db.InventoryGroup, error) {
	if err := group.Validate(); err != nil {
		return db.InventoryGroup{}, err
	}
	addChildren = slices.Compact(slices.Sorted(slices.Values(addChildren)))

	tx, err := d.Sql().Begin()
	if err != nil {
		return db.InventoryGroup{}, err
	}

	if err = d.lockInventory(tx, group.ProjectID, group.InventoryID); err != nil {
		_ = tx.Rollback()
		return db.InventoryGroup{}, err
	}

	if len(addChildren) > 0 {
		if _, err = d.inventoryGroupChildren(tx, group.ProjectID, group.InventoryID, addChildren); err != nil {
			_ = tx.Rollback()
			return db.InventoryGroup{}, err
		}
	}

	if err = d.validateInventoryGroupNameIsFree(tx, group); err != nil {
		_ = tx.Rollback()
		return db.InventoryGroup{}, err
	}

	group.Created = tz.Now()
	group.VariableCount = db.CountInventoryVariables(group.Variables)

	query, args, err := sq.Insert("project__inventory_group").
		SetMap(map[string]any{
			"inventory_id":   group.InventoryID,
			"project_id":     group.ProjectID,
			"name":           group.Name,
			"description":    group.Description,
			"variables":      group.Variables,
			"variable_count": group.VariableCount,
			"created":        group.Created,
		}).
		ToSql()
	if err != nil {
		_ = tx.Rollback()
		return db.InventoryGroup{}, err
	}

	group.ID, err = d.connection.InsertTx(tx, "id", query, args...)
	if err != nil {
		_ = tx.Rollback()
		// A concurrent request may have taken the name after the check
		var nameErr *common_errors.ValidationError
		if errors.As(d.validateInventoryGroupNameIsFree(d.Sql(), group), &nameErr) {
			return db.InventoryGroup{}, nameErr
		}
		return db.InventoryGroup{}, err
	}

	for _, child := range addChildren {
		_, err = d.execTx(tx, "insert into project__inventory_group_child (parent_id, child_id) values (?, ?)", group.ID, child)
		if err != nil {
			_ = tx.Rollback()
			return db.InventoryGroup{}, err
		}
	}

	group.Children = append([]int{}, addChildren...)
	return group, tx.Commit()
}

func (d *SqlDb) UpdateInventoryGroup(group db.InventoryGroup, addChildren, removeChildren []int) (added, removed []int, err error) {
	if err = group.Validate(); err != nil {
		return
	}
	if addChildren, removeChildren, err = addRemoveIDs(addChildren, removeChildren); err != nil {
		return
	}

	tx, err := d.Sql().Begin()
	if err != nil {
		return
	}

	if err = d.lockInventory(tx, group.ProjectID, group.InventoryID); err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}

	children, err := d.inventoryGroupChildren(tx, group.ProjectID, group.InventoryID, slices.Concat(addChildren, removeChildren))
	if err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}

	current, ok := children[group.ID]
	if !ok {
		_ = tx.Rollback()
		return nil, nil, db.ErrNotFound
	}

	if err = d.validateInventoryGroupNameIsFree(tx, group); err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}

	for _, id := range addChildren {
		if !slices.Contains(current, id) {
			added = append(added, id)
		}
	}
	for _, id := range removeChildren {
		if slices.Contains(current, id) {
			removed = append(removed, id)
		}
	}

	// Only a new link can close a cycle, so a cycle already in the table does not block other changes
	for _, id := range added {
		if id == group.ID || slices.Contains(db.InventoryGroupDescendants(children, id), group.ID) {
			_ = tx.Rollback()
			return nil, nil, common_errors.NewValidationError("a group can not be its own ancestor")
		}
	}

	query, args, err := sq.Update("project__inventory_group").
		SetMap(map[string]any{
			"name":           group.Name,
			"description":    group.Description,
			"variables":      group.Variables,
			"variable_count": db.CountInventoryVariables(group.Variables),
		}).
		Where(sq.Eq{"id": group.ID, "inventory_id": group.InventoryID, "project_id": group.ProjectID}).
		ToSql()
	if err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}

	if _, err = d.execTx(tx, query, args...); err != nil {
		_ = tx.Rollback()
		// A concurrent request may have taken the name after the check
		var nameErr *common_errors.ValidationError
		if errors.As(d.validateInventoryGroupNameIsFree(d.Sql(), group), &nameErr) {
			return nil, nil, nameErr
		}
		return nil, nil, err
	}

	for _, child := range added {
		_, err = d.execTx(tx, "insert into project__inventory_group_child (parent_id, child_id) values (?, ?)", group.ID, child)
		if err != nil {
			_ = tx.Rollback()
			return nil, nil, err
		}
	}
	for _, child := range removed {
		_, err = d.execTx(tx, "delete from project__inventory_group_child where parent_id=? and child_id=?", group.ID, child)
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

func (d *SqlDb) DeleteInventoryGroup(projectID, inventoryID, groupID int) error {
	return requireDeletedRow(d.exec(
		"delete from project__inventory_group where id=? and inventory_id=? and project_id=?",
		groupID, inventoryID, projectID))
}

func (d *SqlDb) ChangeInventoryHostGroups(projectID, inventoryID int, hostIDs, add, remove []int) (added, removed map[int][]int, err error) {
	if add, remove, err = addRemoveIDs(add, remove); err != nil {
		return
	}
	hostIDs = slices.Compact(slices.Sorted(slices.Values(hostIDs)))

	tx, err := d.Sql().Begin()
	if err != nil {
		return
	}

	if err = d.lockInventory(tx, projectID, inventoryID); err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}

	if _, err = d.inventoryGroupChildren(tx, projectID, inventoryID, slices.Concat(add, remove)); err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}

	query, args, err := sq.Select("id").From("project__inventory_host").
		Where(sq.Eq{"id": hostIDs, "inventory_id": inventoryID, "project_id": projectID}).
		ToSql()
	if err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}
	var found []int
	if _, err = tx.Select(&found, d.PrepareQuery(query), args...); err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}
	for _, id := range hostIDs {
		if !slices.Contains(found, id) {
			_ = tx.Rollback()
			return nil, nil, common_errors.NewValidationError(fmt.Sprintf("host %d is not in this inventory", id))
		}
	}

	query, args, err = sq.Select("group_id", "host_id").From("project__inventory_group_host").
		Where(sq.Eq{"host_id": hostIDs}).
		ToSql()
	if err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}
	var links []db.InventoryGroupHost
	if _, err = tx.Select(&links, d.PrepareQuery(query), args...); err != nil {
		_ = tx.Rollback()
		return nil, nil, err
	}
	linked := map[db.InventoryGroupHost]bool{}
	for _, link := range links {
		linked[link] = true
	}

	added, removed = map[int][]int{}, map[int][]int{}
	for _, group := range add {
		for _, host := range hostIDs {
			if linked[db.InventoryGroupHost{GroupID: group, HostID: host}] {
				continue
			}
			_, err = d.execTx(tx, "insert into project__inventory_group_host (group_id, host_id) values (?, ?)", group, host)
			if err != nil {
				_ = tx.Rollback()
				return nil, nil, err
			}
			added[group] = append(added[group], host)
		}
	}
	for _, group := range remove {
		for _, host := range hostIDs {
			if !linked[db.InventoryGroupHost{GroupID: group, HostID: host}] {
				continue
			}
			_, err = d.execTx(tx, "delete from project__inventory_group_host where group_id=? and host_id=?", group, host)
			if err != nil {
				_ = tx.Rollback()
				return nil, nil, err
			}
			removed[group] = append(removed[group], host)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, nil, err
	}
	return added, removed, nil
}
