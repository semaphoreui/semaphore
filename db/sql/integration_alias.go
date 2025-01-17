package sql

import (
	"database/sql"
	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) CreateIntegrationAlias(alias db.IntegrationAlias) (res db.IntegrationAlias, err error) {

	insertID, err := d.insert(
		"id",
		"insert into project__integration_alias (project_id, integration_id, alias) values (?, ?, ?)",
		alias.ProjectID,
		alias.IntegrationID,
		alias.Alias)

	if err != nil {
		return
	}

	res = alias
	res.ID = insertID
	return
}

func (d *SqlDb) GetIntegrationAliases(projectID int, integrationID *int) (res []db.IntegrationAlias, err error) {

	q := squirrel.Select("*").From(db.IntegrationAliasProps.TableName)

	if integrationID == nil {
		q = q.Where("project_id=? AND integration_id is null", projectID)
	} else {
		q = q.Where("project_id=? AND integration_id=?", projectID, integrationID)
	}

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(&res, query, args...)

	return
}

func (d *SqlDb) GetIntegrationsByAlias(alias string) (res []db.Integration, err error) {

	var aliasObj db.IntegrationAlias

	q := squirrel.Select("*").
		From(db.IntegrationAliasProps.TableName).
		Where("alias=?", alias)

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	err = d.selectOne(&aliasObj, query, args...)

	if err == sql.ErrNoRows {
		err = db.ErrNotFound
	}

	if aliasObj.IntegrationID == nil {
		var projIntegrations []db.Integration
		projIntegrations, err = d.GetIntegrations(aliasObj.ProjectID, db.RetrieveQueryParams{})
		if err != nil {
			return
		}
		for _, integration := range projIntegrations {
			if integration.Searchable {
				res = append(res, integration)
			}
		}
	} else {
		var integration db.Integration
		integration, err = d.GetIntegration(aliasObj.ProjectID, *aliasObj.IntegrationID)
		res = append(res, integration)
	}

	return
}

func (d *SqlDb) DeleteIntegrationAlias(projectID int, aliasID int) error {
	return d.deleteObject(projectID, db.IntegrationAliasProps, aliasID)
}
