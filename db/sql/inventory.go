package sql

import (
	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) GetInventory(projectID int, inventoryID int) (inventory db.Inventory, err error) {
	err = d.getObject(projectID, db.InventoryProps, inventoryID, &inventory)

	return
}

func (d *SqlDb) GetInventories(projectID int, params db.RetrieveQueryParams, types []db.InventoryType) ([]db.Inventory, error) {
	var inventories []db.Inventory
	err := d.getObjects(projectID, db.InventoryProps, params, func(builder squirrel.SelectBuilder) squirrel.SelectBuilder {
		if len(types) == 0 {
			return builder
		}

		return builder.Where(squirrel.Eq{"type": types})
	}, &inventories)
	return inventories, err
}

func (d *SqlDb) GetInventoryRefs(projectID int, inventoryID int) (db.ObjectReferrers, error) {
	return d.getObjectRefs(projectID, db.InventoryProps, inventoryID)
}

func (d *SqlDb) DeleteInventory(projectID int, inventoryID int) error {
	return d.deleteObject(projectID, db.InventoryProps, inventoryID)
}

func (d *SqlDb) UpdateInventory(inventory db.Inventory) error {

	_, err := d.exec(
		"update project__inventory set "+
			"name=?, "+
			"type=?, "+
			"runner_tag=?, "+
			"ssh_key_id=?, "+
			"inventory=?, "+
			"become_key_id=?, "+
			"template_id=?, "+
			"repository_id=? "+
			"where id=?",
		inventory.Name,
		inventory.Type,
		inventory.RunnerTag,
		inventory.SSHKeyID,
		inventory.Inventory,
		inventory.BecomeKeyID,
		inventory.TemplateID,
		inventory.RepositoryID,
		inventory.ID)

	return err
}

func (d *SqlDb) SetInventoryVariables(projectID, inventoryID int, variables string) error {
	_, err := d.exec("update project__inventory set variables=?, variable_count=? where id=? and project_id=?",
		variables, db.CountInventoryVariables(variables), inventoryID, projectID)
	return err
}

func (d *SqlDb) CreateInventory(inventory db.Inventory) (newInventory db.Inventory, err error) {
	inventory.VariableCount = db.CountInventoryVariables(inventory.VariablesText())

	query, args, err := squirrel.Insert("project__inventory").
		SetMap(map[string]any{
			"project_id":     inventory.ProjectID,
			"name":           inventory.Name,
			"type":           inventory.Type,
			"ssh_key_id":     inventory.SSHKeyID,
			"inventory":      inventory.Inventory,
			"become_key_id":  inventory.BecomeKeyID,
			"template_id":    inventory.TemplateID,
			"repository_id":  inventory.RepositoryID,
			"runner_tag":     inventory.RunnerTag,
			"variables":      inventory.Variables,
			"variable_count": inventory.VariableCount,
		}).
		ToSql()
	if err != nil {
		return
	}

	insertID, err := d.insert("id", query, args...)
	if err != nil {
		return
	}

	newInventory = inventory
	newInventory.ID = insertID
	return
}
