package sql

import (
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

// validateAuthSecret checks that the key an endpoint or a group verifies
// webhooks with belongs to the same project. The foreign key only guarantees
// the key exists; without this a project could authorize against another
// project's secret.
func (d *SqlDb) validateAuthSecret(projectID int, auth db.IntegrationAuth) error {
	if auth.AuthSecretID == nil {
		return nil
	}

	count, err := d.Sql().SelectInt(
		d.PrepareQuery("select count(1) from access_key where id=? and project_id=?"),
		*auth.AuthSecretID,
		projectID)
	if err != nil {
		return err
	}

	if count == 0 {
		return common_errors.NewValidationError("Auth secret does not exist in the project")
	}

	return nil
}

func (d *SqlDb) CreateEndpoint(endpoint db.Endpoint) (newEndpoint db.Endpoint, err error) {
	if err = endpoint.Validate(); err != nil {
		return
	}

	if err = d.validateAuthSecret(endpoint.ProjectID, endpoint.IntegrationAuth); err != nil {
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into project__endpoint "+
			"(project_id, `name`, description, auth_method, auth_secret_id, auth_header) values "+
			"(?, ?, ?, ?, ?, ?)",
		endpoint.ProjectID,
		endpoint.Name,
		endpoint.Description,
		endpoint.AuthMethod,
		endpoint.AuthSecretID,
		endpoint.AuthHeader)
	if err != nil {
		return
	}

	newEndpoint = endpoint
	newEndpoint.ID = insertID

	return
}

func (d *SqlDb) GetEndpoints(projectID int, params db.RetrieveQueryParams) (endpoints []db.Endpoint, err error) {
	err = d.getObjects(projectID, db.EndpointProps, params, nil, &endpoints)
	return
}

func (d *SqlDb) GetEndpoint(projectID int, endpointID int) (endpoint db.Endpoint, err error) {
	err = d.getObject(projectID, db.EndpointProps, endpointID, &endpoint)
	return
}

func (d *SqlDb) UpdateEndpoint(endpoint db.Endpoint) error {
	if err := endpoint.Validate(); err != nil {
		return err
	}

	if err := d.validateAuthSecret(endpoint.ProjectID, endpoint.IntegrationAuth); err != nil {
		return err
	}

	res, err := d.exec(
		"update project__endpoint set "+
			"`name`=?, "+
			"description=?, "+
			"auth_method=?, "+
			"auth_secret_id=?, "+
			"auth_header=? "+
			"where project_id=? and `id`=?",
		endpoint.Name,
		endpoint.Description,
		endpoint.AuthMethod,
		endpoint.AuthSecretID,
		endpoint.AuthHeader,
		endpoint.ProjectID,
		endpoint.ID)
	if err = validateMutationResult(res, err); err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return db.ErrNotFound
	}

	return nil
}

func (d *SqlDb) DeleteEndpoint(projectID int, endpointID int) error {
	refs, err := d.GetEndpointRefs(projectID, endpointID)
	if err != nil {
		return err
	}

	// The foreign key refuses too, but as a driver error; this one names the reason.
	if len(refs.IntegrationGroups) > 0 {
		return common_errors.NewValidationError("Endpoint is used by integration groups")
	}

	return d.deleteObject(projectID, db.EndpointProps, endpointID)
}

func (d *SqlDb) GetEndpointRefs(projectID int, endpointID int) (refs db.EndpointReferrers, err error) {
	refs.IntegrationGroups, err = d.getObjectRefsFrom(projectID, db.EndpointProps, endpointID, db.IntegrationGroupProps)
	return
}
