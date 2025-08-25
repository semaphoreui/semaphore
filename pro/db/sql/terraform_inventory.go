package sql

import (
	"github.com/semaphoreui/semaphore/db"
)

type TerraformStoreImpl struct {
}

func (d *TerraformStoreImpl) CreateTerraformInventoryAlias(alias db.TerraformInventoryAlias) (res db.TerraformInventoryAlias, err error) {
	return
}

func (d *TerraformStoreImpl) UpdateTerraformInventoryAlias(alias db.TerraformInventoryAlias) (err error) {
	return
}

func (d *TerraformStoreImpl) GetTerraformInventoryAliasByAlias(alias string) (res db.TerraformInventoryAlias, err error) {
	return
}

func (d *TerraformStoreImpl) GetTerraformInventoryAlias(projectID, inventoryID int, aliasID string) (res db.TerraformInventoryAlias, err error) {
	return
}

func (d *TerraformStoreImpl) GetTerraformInventoryAliases(projectID, inventoryID int) (res []db.TerraformInventoryAlias, err error) {
	return
}

func (d *TerraformStoreImpl) DeleteTerraformInventoryAlias(projectID int, inventoryID int, aliasID string) (err error) {
	return
}

func (d *TerraformStoreImpl) GetTerraformInventoryStates(projectID, inventoryID int, params db.RetrieveQueryParams) (res []db.TerraformInventoryState, err error) {
	return
}

func (d *TerraformStoreImpl) CreateTerraformInventoryState(state db.TerraformInventoryState) (res db.TerraformInventoryState, err error) {
	return
}

func (d *TerraformStoreImpl) DeleteTerraformInventoryState(projectID int, inventoryID int, stateID int) (err error) {
	return
}

func (d *TerraformStoreImpl) GetTerraformInventoryState(projectID int, inventoryId int, stateID int) (res db.TerraformInventoryState, err error) {
	return
}
