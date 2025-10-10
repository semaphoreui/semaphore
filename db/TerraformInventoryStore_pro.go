package db

type TerraformStore interface {
	CreateTerraformInventoryAlias(alias TerraformInventoryAlias) (TerraformInventoryAlias, error)
	GetTerraformInventoryAliasByAlias(alias string) (TerraformInventoryAlias, error)
	GetTerraformInventoryAlias(projectID int, inventoryID int, aliasID string) (TerraformInventoryAlias, error)
	GetTerraformInventoryAliases(projectID, inventoryID int) ([]TerraformInventoryAlias, error)
	UpdateTerraformInventoryAlias(alias TerraformInventoryAlias) error
	DeleteTerraformInventoryAlias(projectID int, inventoryID int, aliasID string) error

	CreateTerraformInventoryState(State TerraformInventoryState) (TerraformInventoryState, error)
	GetTerraformInventoryState(projectID int, inventoryId int, stateID int) (TerraformInventoryState, error)
	GetTerraformInventoryStates(projectID, inventoryID int, params RetrieveQueryParams) ([]TerraformInventoryState, error)
	DeleteTerraformInventoryState(projectID int, inventoryId int, stateID int) error
}
