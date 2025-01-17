package bolt

import (
	"github.com/semaphoreui/semaphore/db"
	"reflect"
)

var integrationAliasProps = db.ObjectProps{
	TableName:         "integration_alias",
	Type:              reflect.TypeOf(db.IntegrationAlias{}),
	PrimaryColumnName: "alias",
}

func (d *BoltDb) GetIntegrationAliases(projectID int, integrationID *int) (res []db.IntegrationAlias, err error) {

	err = d.integrationAlias.getAliases(projectID, func(i interface{}) bool {
		alias := i.(db.IntegrationAlias)
		if alias.IntegrationID == nil && integrationID == nil {
			return true
		} else if alias.IntegrationID != nil && integrationID != nil {
			return *alias.IntegrationID == *integrationID
		}
		return false
	}, &res)

	return
}

func (d *BoltDb) GetIntegrationsByAlias(alias string) (res []db.Integration, err error) {

	var aliasObj db.IntegrationAlias

	err = d.integrationAlias.getPublicAlias(alias, &aliasObj)
	if err != nil {
		return
	}

	if aliasObj.IntegrationID == nil {
		err = d.getObjects(aliasObj.ProjectID, db.IntegrationProps, db.RetrieveQueryParams{}, func(i interface{}) bool {
			integration := i.(db.Integration)
			return integration.Searchable
		}, &res)

		if err != nil {
			return
		}

	} else {
		var integration db.Integration
		integration, err = d.GetIntegration(aliasObj.ProjectID, *aliasObj.IntegrationID)
		if err != nil {
			return
		}
		res = append(res, integration)
	}

	return
}

func (d *BoltDb) CreateIntegrationAlias(alias db.IntegrationAlias) (res db.IntegrationAlias, err error) {

	newAlias, err := d.integrationAlias.createAlias(alias)

	if err != nil {
		return
	}

	res = newAlias.(db.IntegrationAlias)

	return
}

func (d *BoltDb) DeleteIntegrationAlias(projectID int, aliasID int) (err error) {

	err = d.integrationAlias.deleteIntegrationAlias(projectID, aliasID)

	return
}
