package sql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/semaphoreui/semaphore/util"
	"maps"
	"slices"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-gorp/gorp/v3"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
	"github.com/semaphoreui/semaphore/pkg/tz"
)

// inventorySourceRow is a source as the table keeps it
type inventorySourceRow struct {
	db.InventorySource
	LastResultJSON string `db:"last_result"`
	Hosts          int    `db:"host_count"`
	Groups         int    `db:"group_count"`
}

func (r inventorySourceRow) source() (db.InventorySource, error) {
	source := r.InventorySource
	source.HostCount = r.Hosts
	source.GroupCount = r.Groups
	if r.LastResultJSON != "" {
		source.LastResult = &db.InventorySyncResult{}
		return source, json.Unmarshal([]byte(r.LastResultJSON), source.LastResult)
	}
	return source, nil
}

func inventorySourceSelect() sq.SelectBuilder {
	return sq.Select("s.*",
		"(select count(*) from project__inventory_host h where h.source_id = s.id) as host_count",
		"(select count(*) from project__inventory_group g where g.source_id = s.id) as group_count").
		From("project__inventory_source s")
}

// inventorySourceWriteError returns the name error when the name is taken
func (d *SqlDb) inventorySourceWriteError(source db.InventorySource, err error) error {
	count, countErr := d.Sql().SelectInt(
		d.PrepareQuery("select count(*) from project__inventory_source where inventory_id=? and name=? and id<>?"),
		source.InventoryID, source.Name, source.ID)
	if countErr == nil && count > 0 {
		return common_errors.NewValidationError("a source with this name already exists in the inventory")
	}
	return err
}

func (d *SqlDb) GetInventorySources(projectID, inventoryID int) (sources []db.InventorySource, err error) {
	query, args, err := inventorySourceSelect().
		Where(sq.Eq{"s.project_id": projectID, "s.inventory_id": inventoryID}).
		OrderBy("s.name").
		ToSql()
	if err != nil {
		return
	}

	var rows []inventorySourceRow
	_, err = d.selectAll(&rows, query, args...)
	if err != nil {
		return
	}

	sources = make([]db.InventorySource, len(rows))
	for i, row := range rows {
		if sources[i], err = row.source(); err != nil {
			return nil, err
		}
	}
	return
}

func (d *SqlDb) GetInventorySource(projectID, inventoryID, sourceID int) (db.InventorySource, error) {
	query, args, err := inventorySourceSelect().
		Where(sq.Eq{"s.id": sourceID, "s.project_id": projectID, "s.inventory_id": inventoryID}).
		ToSql()
	if err != nil {
		return db.InventorySource{}, err
	}

	var row inventorySourceRow
	if err = d.selectOne(&row, query, args...); err != nil {
		return db.InventorySource{}, err
	}
	return row.source()
}

func (d *SqlDb) CreateInventorySource(source db.InventorySource) (db.InventorySource, error) {
	if err := source.Validate(); err != nil {
		return db.InventorySource{}, err
	}

	source.Created = tz.Now()

	query, args, err := sq.Insert("project__inventory_source").
		SetMap(map[string]any{
			"project_id":     source.ProjectID,
			"inventory_id":   source.InventoryID,
			"name":           source.Name,
			"type":           source.Type,
			"repository_id":  source.RepositoryID,
			"key_id":         source.KeyID,
			"params":         source.Params,
			"variables":      source.Variables,
			"variable_count": source.VariableCount,
			"status":         source.Status,
			"last_error":     source.LastError,
			"last_result":    "",
			"created":        source.Created,
		}).
		ToSql()
	if err != nil {
		return db.InventorySource{}, err
	}

	source.ID, err = d.insert("id", query, args...)
	if err != nil {
		return db.InventorySource{}, d.inventorySourceWriteError(source, err)
	}
	return source, nil
}

func (d *SqlDb) UpdateInventorySource(source db.InventorySource) error {
	if err := source.Validate(); err != nil {
		return err
	}

	tx, err := d.Sql().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	// A sync writes under this lock, the edit must not come between its check of the source and its commit
	if err = d.lockInventory(tx, source.ProjectID, source.InventoryID); err != nil {
		return err
	}

	var current inventorySourceRow
	err = tx.SelectOne(&current, d.PrepareQuery(
		"select * from project__inventory_source where id=? and inventory_id=? and project_id=?"),
		source.ID, source.InventoryID, source.ProjectID)
	if errors.Is(err, sql.ErrNoRows) {
		return db.ErrNotFound
	}
	if err != nil {
		return err
	}

	set := map[string]any{"name": source.Name, "repository_id": source.RepositoryID, "key_id": source.KeyID, "params": source.Params}
	if current.Status == db.InventorySourceSyncing {
		if current.SyncStarted != nil && current.SyncStarted.After(tz.Now().Add(-db.InventorySourceStaleAfter)) {
			return db.ErrInventorySourceSyncing
		}
		// The run that did not end must not write under the new settings
		set["status"], set["last_error"] = db.InventorySourceFailed, db.InventorySourceInterrupted
	}

	query, args, err := sq.Update("project__inventory_source").SetMap(set).Where(sq.Eq{"id": source.ID}).ToSql()
	if err != nil {
		return err
	}

	if _, err = d.execTx(tx, query, args...); err != nil {
		_ = tx.Rollback()
		return d.inventorySourceWriteError(source, err)
	}
	return tx.Commit()
}

// DeleteInventorySource removes the hosts and groups of the source with their links, also links in manual groups
func (d *SqlDb) DeleteInventorySource(projectID, inventoryID, sourceID int) error {
	tx, err := d.Sql().Begin()
	if err != nil {
		return err
	}

	if err = d.lockInventory(tx, projectID, inventoryID); err != nil {
		_ = tx.Rollback()
		return err
	}

	for _, query := range []string{
		"delete from project__inventory_group_host where host_id in (select id from project__inventory_host where source_id=? and inventory_id=?)",
		"delete from project__inventory_group_host where group_id in (select id from project__inventory_group where source_id=? and inventory_id=?)",
		"delete from project__inventory_group_child where parent_id in (select id from project__inventory_group where source_id=? and inventory_id=?)",
		"delete from project__inventory_group_child where child_id in (select id from project__inventory_group where source_id=? and inventory_id=?)",
		"delete from project__inventory_host where source_id=? and inventory_id=?",
		"delete from project__inventory_group where source_id=? and inventory_id=?",
	} {
		if _, err = d.execTx(tx, query, sourceID, inventoryID); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	err = requireDeletedRow(d.execTx(tx,
		"delete from project__inventory_source where id=? and inventory_id=? and project_id=?",
		sourceID, inventoryID, projectID))
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (d *SqlDb) StartInventorySourceSync(projectID, inventoryID, sourceID int, now, staleBefore time.Time) (int, error) {
	tx, err := d.Sql().Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck

	// An edit of the source holds this lock, the start waits for it and then reads the new settings
	if err = d.lockInventory(tx, projectID, inventoryID); err != nil {
		return 0, err
	}

	res, err := d.execTx(tx, `
		update project__inventory_source set status=?, sync_started=?, sync_run=sync_run+1
		where id=? and inventory_id=? and project_id=? and (status<>? or sync_started<?)
	`, db.InventorySourceSyncing, now, sourceID, inventoryID, projectID, db.InventorySourceSyncing, staleBefore)
	if err != nil {
		return 0, err
	}

	affected, err := res.RowsAffected()
	if err != nil || affected == 0 {
		return 0, err
	}

	run, err := tx.SelectInt(d.PrepareQuery("select sync_run from project__inventory_source where id=?"), sourceID)
	if err != nil {
		return 0, err
	}
	return int(run), tx.Commit()
}

func (d *SqlDb) ApplyInventoryDocument(projectID, inventoryID, sourceID, run int, doc db.InventoryDocument) (db.InventorySyncResult, error) {
	tx, err := d.Sql().Begin()
	if err != nil {
		return db.InventorySyncResult{}, err
	}

	// A panic must not leave the inventory locked
	defer tx.Rollback() //nolint:errcheck

	result, err := d.applyInventoryDocument(tx, projectID, inventoryID, sourceID, run, doc)
	if err != nil {
		return db.InventorySyncResult{}, err
	}

	// The result goes with the hosts, a server stopped after the commit must not lose it
	text, err := json.Marshal(result)
	if err != nil {
		return db.InventorySyncResult{}, err
	}
	// The sync ends here, not when the write began
	now := tz.Now()
	_, err = d.execTx(tx, `
		update project__inventory_source set status=?, last_sync=?, last_success=?, last_result=?, last_error=''
		where id=?
	`, db.InventorySourceSuccess, now, now, string(text), sourceID)
	if err != nil {
		return db.InventorySyncResult{}, err
	}

	if err = tx.Commit(); err != nil {
		return db.InventorySyncResult{}, err
	}
	return result, nil
}

func (d *SqlDb) applyInventoryDocument(tx *gorp.Transaction, projectID, inventoryID, sourceID, run int, doc db.InventoryDocument) (db.InventorySyncResult, error) {
	if err := d.lockInventory(tx, projectID, inventoryID); err != nil {
		return db.InventorySyncResult{}, err
	}

	// A source deleted during the sync would leave rows without an owner, a later run replaced this one
	var variables []string
	_, err := tx.Select(&variables, d.PrepareQuery(`
		select variables from project__inventory_source
		where id=? and inventory_id=? and project_id=? and status=? and sync_run=?
	`), sourceID, inventoryID, projectID, db.InventorySourceSyncing, run)
	if err != nil {
		return db.InventorySyncResult{}, err
	}
	if len(variables) == 0 {
		return db.InventorySyncResult{}, db.ErrNotFound
	}

	state, err := d.inventoryState(tx, projectID, inventoryID)
	if err != nil {
		return db.InventorySyncResult{}, err
	}

	plan, err := db.PlanInventorySync(sourceID, state, doc)
	if err != nil {
		return db.InventorySyncResult{}, err
	}

	if err = d.writeInventoryRows(tx, projectID, inventoryID, sourceID, plan); err != nil {
		return db.InventorySyncResult{}, err
	}

	if err = d.writeInventoryLinks(tx, projectID, inventoryID, state, plan); err != nil {
		return db.InventorySyncResult{}, err
	}

	result := plan.Result
	result.Skipped = result.Skipped[:min(len(result.Skipped), 20)]
	result.VariablesChanged = variables[0] != plan.Variables
	return result, nil
}

// inventoryState reads every host, group and link of the inventory
func (d *SqlDb) inventoryState(tx *gorp.Transaction, projectID, inventoryID int) (state db.InventoryState, err error) {
	_, err = tx.Select(&state.Hosts, d.PrepareQuery("select * from project__inventory_host where project_id=? and inventory_id=?"),
		projectID, inventoryID)
	if err != nil {
		return
	}

	_, err = tx.Select(&state.Groups, d.PrepareQuery("select * from project__inventory_group where project_id=? and inventory_id=?"),
		projectID, inventoryID)
	if err != nil {
		return
	}

	_, err = tx.Select(&state.Members, d.PrepareQuery(`
		select gh.group_id, gh.host_id
		from project__inventory_group_host gh
		join project__inventory_group g on g.id = gh.group_id
		where g.project_id=? and g.inventory_id=?
	`), projectID, inventoryID)
	if err != nil {
		return
	}

	state.Children, err = d.inventoryGroupChildren(tx, projectID, inventoryID, nil)
	return
}

func (d *SqlDb) GetInventoryContent(projectID, inventoryID int) (db.InventoryContent, error) {
	tx, err := d.Sql().Begin()
	if err != nil {
		return db.InventoryContent{}, err
	}
	defer tx.Rollback() //nolint:errcheck

	// The reads must see the inventory before or after a sync, and must not wait for it: a task stop comes here too
	// MySQL gives one snapshot for a transaction by default, SQLite has one connection
	if d.GetDialect() == util.DbDriverPostgres {
		if _, err = tx.Exec("set transaction isolation level repeatable read"); err != nil {
			return db.InventoryContent{}, err
		}
	}

	state, err := d.inventoryState(tx, projectID, inventoryID)
	if err != nil {
		return db.InventoryContent{}, err
	}

	content := db.InventoryContent{Hosts: state.Hosts, Groups: state.Groups, Members: state.Members}
	for i, group := range content.Groups {
		content.Groups[i].Children = state.Children[group.ID]
	}

	query, args, err := inventorySourceSelect().
		Where(sq.Eq{"s.project_id": projectID, "s.inventory_id": inventoryID}).
		OrderBy("s.id").
		ToSql()
	if err != nil {
		return db.InventoryContent{}, err
	}

	var rows []inventorySourceRow
	if _, err = tx.Select(&rows, d.PrepareQuery(query), args...); err != nil {
		return db.InventoryContent{}, err
	}
	for _, row := range rows {
		source, err := row.source()
		if err != nil {
			return db.InventoryContent{}, err
		}
		content.Sources = append(content.Sources, source)
	}

	return content, tx.Commit()
}

// writeInventoryRows deletes, creates and updates the hosts and groups of the plan and the variables of the source
func (d *SqlDb) writeInventoryRows(tx *gorp.Transaction, projectID, inventoryID, sourceID int, plan db.InventorySyncPlan) error {
	var writes []sq.Sqlizer
	for chunk := range slices.Chunk(plan.DeleteHosts, 500) {
		writes = append(writes,
			sq.Delete("project__inventory_group_host").Where(sq.Eq{"host_id": chunk}),
			sq.Delete("project__inventory_host").Where(sq.Eq{"id": chunk, "inventory_id": inventoryID}))
	}
	for chunk := range slices.Chunk(plan.DeleteGroups, 500) {
		writes = append(writes,
			sq.Delete("project__inventory_group_host").Where(sq.Eq{"group_id": chunk}),
			sq.Delete("project__inventory_group_child").Where(sq.Or{sq.Eq{"parent_id": chunk}, sq.Eq{"child_id": chunk}}),
			sq.Delete("project__inventory_group").Where(sq.Eq{"id": chunk, "inventory_id": inventoryID}))
	}

	created := tz.Now()
	for chunk := range slices.Chunk(plan.CreateHosts, 500) {
		insert := sq.Insert("project__inventory_host").Columns("inventory_id", "project_id", "name", "description",
			"enabled", "variables", "variable_count", "created", "source_id")
		for _, host := range chunk {
			insert = insert.Values(inventoryID, projectID, host.Name, "",
				host.Enabled, host.Variables, db.CountInventoryVariables(host.Variables), created, sourceID)
		}
		writes = append(writes, insert)
	}
	for chunk := range slices.Chunk(plan.CreateGroups, 500) {
		insert := sq.Insert("project__inventory_group").Columns("inventory_id", "project_id", "name", "description",
			"variables", "variable_count", "created", "source_id")
		for _, group := range chunk {
			insert = insert.Values(inventoryID, projectID, group.Name, "",
				group.Variables, db.CountInventoryVariables(group.Variables), created, sourceID)
		}
		writes = append(writes, insert)
	}

	for _, host := range plan.UpdateHosts {
		writes = append(writes, sq.Update("project__inventory_host").
			SetMap(map[string]any{"variables": host.Variables, "variable_count": db.CountInventoryVariables(host.Variables), "source_id": sourceID}).
			Where(sq.Eq{"id": host.ID}))
	}
	for _, group := range plan.UpdateGroups {
		writes = append(writes, sq.Update("project__inventory_group").
			SetMap(map[string]any{"variables": group.Variables, "variable_count": db.CountInventoryVariables(group.Variables), "source_id": sourceID}).
			Where(sq.Eq{"id": group.ID}))
	}

	writes = append(writes, sq.Update("project__inventory_source").
		SetMap(map[string]any{"variables": plan.Variables, "variable_count": db.CountInventoryVariables(plan.Variables)}).
		Where(sq.Eq{"id": sourceID}))

	for _, write := range writes {
		query, args, err := write.ToSql()
		if err != nil {
			return err
		}
		if _, err = d.execTx(tx, query, args...); err != nil {
			return err
		}
	}
	return nil
}

// writeInventoryLinks brings the hosts and child groups of every source group to the plan
func (d *SqlDb) writeInventoryLinks(tx *gorp.Transaction, projectID, inventoryID int, state db.InventoryState, plan db.InventorySyncPlan) error {
	var hosts, groups []struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}
	_, err := tx.Select(&hosts, d.PrepareQuery("select id, name from project__inventory_host where project_id=? and inventory_id=?"),
		projectID, inventoryID)
	if err != nil {
		return err
	}
	_, err = tx.Select(&groups, d.PrepareQuery("select id, name from project__inventory_group where project_id=? and inventory_id=?"),
		projectID, inventoryID)
	if err != nil {
		return err
	}
	hostIDs, groupIDs := map[string]int{}, map[string]int{}
	for _, host := range hosts {
		hostIDs[host.Name] = host.ID
	}
	for _, group := range groups {
		groupIDs[group.Name] = group.ID
	}

	// Links of deleted rows are gone already
	deletedHosts, deletedGroups := map[int]bool{}, map[int]bool{}
	for _, id := range plan.DeleteHosts {
		deletedHosts[id] = true
	}
	for _, id := range plan.DeleteGroups {
		deletedGroups[id] = true
	}
	oldHosts, oldChildren := map[int][]int{}, map[int][]int{}
	for _, member := range state.Members {
		if !deletedHosts[member.HostID] {
			oldHosts[member.GroupID] = append(oldHosts[member.GroupID], member.HostID)
		}
	}
	for parent, children := range state.Children {
		for _, child := range children {
			if !deletedGroups[child] {
				oldChildren[parent] = append(oldChildren[parent], child)
			}
		}
	}

	diff := func(old []int, names []string, ids map[string]int) (add, drop []int) {
		had := map[int]bool{}
		for _, id := range old {
			had[id] = true
		}
		for _, name := range names {
			if had[ids[name]] {
				delete(had, ids[name])
			} else {
				add = append(add, ids[name])
			}
		}
		return add, slices.Sorted(maps.Keys(had))
	}

	var writes []sq.Sqlizer
	var hostLinks []db.InventoryGroupHost
	var childLinks [][2]int
	for name, names := range plan.GroupHosts {
		group := groupIDs[name]
		add, drop := diff(oldHosts[group], names, hostIDs)
		for _, host := range add {
			hostLinks = append(hostLinks, db.InventoryGroupHost{GroupID: group, HostID: host})
		}
		for chunk := range slices.Chunk(drop, 500) {
			writes = append(writes, sq.Delete("project__inventory_group_host").Where(sq.Eq{"group_id": group, "host_id": chunk}))
		}
	}
	for name, names := range plan.GroupChildren {
		group := groupIDs[name]
		add, drop := diff(oldChildren[group], names, groupIDs)
		for _, child := range add {
			childLinks = append(childLinks, [2]int{group, child})
		}
		for chunk := range slices.Chunk(drop, 500) {
			writes = append(writes, sq.Delete("project__inventory_group_child").Where(sq.Eq{"parent_id": group, "child_id": chunk}))
		}
	}
	for chunk := range slices.Chunk(childLinks, 500) {
		insert := sq.Insert("project__inventory_group_child").Columns("parent_id", "child_id")
		for _, link := range chunk {
			insert = insert.Values(link[0], link[1])
		}
		writes = append(writes, insert)
	}

	for _, write := range writes {
		query, args, err := write.ToSql()
		if err != nil {
			return err
		}
		if _, err = d.execTx(tx, query, args...); err != nil {
			return err
		}
	}
	return d.insertInventoryGroupHosts(tx, hostLinks)
}

func (d *SqlDb) FailInventorySourceSync(projectID, inventoryID, sourceID, run int, now time.Time, message string) error {
	tx, err := d.Sql().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	// An edit of the source that closes a run as interrupted holds this lock
	err = d.lockInventory(tx, projectID, inventoryID)
	if errors.Is(err, db.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = d.execTx(tx, `
		update project__inventory_source set status=?, last_sync=?, last_error=?
		where id=? and inventory_id=? and project_id=? and status=? and sync_run=?
	`, db.InventorySourceFailed, now, message, sourceID, inventoryID, projectID, db.InventorySourceSyncing, run)
	if err != nil {
		return err
	}
	return tx.Commit()
}
