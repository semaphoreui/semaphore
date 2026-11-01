package sql

import (
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) CreateEndpointAlias(alias db.EndpointAlias) (res db.EndpointAlias, err error) {
	// The endpoint must be the project's own, or a project could add a URL to
	// another project's endpoint.
	var endpoint db.Endpoint
	if err = d.getObject(alias.ProjectID, db.EndpointProps, alias.EndpointID, &endpoint); err != nil {
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into project__endpoint_alias (project_id, endpoint_id, alias) values (?, ?, ?)",
		alias.ProjectID,
		alias.EndpointID,
		alias.Alias)
	if err != nil {
		return
	}

	res = alias
	res.ID = insertID
	return
}

func (d *SqlDb) GetEndpointAliases(projectID int, endpointID int) (res []db.EndpointAlias, err error) {
	res = make([]db.EndpointAlias, 0)

	_, err = d.selectAll(&res,
		"select * from project__endpoint_alias where project_id=? and endpoint_id=? order by id",
		projectID,
		endpointID)

	return
}

func (d *SqlDb) DeleteEndpointAlias(projectID int, endpointID int, aliasID int) error {
	return requireDeletedRow(d.exec(
		"delete from project__endpoint_alias where `id`=? and endpoint_id=? and project_id=?",
		aliasID,
		endpointID,
		projectID))
}

func (d *SqlDb) GetEndpointByAlias(alias string) (endpoint db.Endpoint, err error) {
	err = d.selectOne(&endpoint,
		"select e.* from project__endpoint e "+
			"join project__endpoint_alias a on a.endpoint_id = e.id "+
			"where a.alias=?",
		alias)

	return
}
