package sql

import (
	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pkg/common_errors"
)

// validateIntegrationGroupEndpoint checks that the endpoint the group runs
// from belongs to the group's own project, and that the group's own key, if
// it overrides the authorization, does too.
func (d *SqlDb) validateIntegrationGroupEndpoint(group db.IntegrationGroup) error {
	var endpoint db.Endpoint
	err := d.getObject(group.ProjectID, db.EndpointProps, group.EndpointID, &endpoint)
	if err == db.ErrNotFound {
		return common_errors.NewValidationError("Endpoint does not exist in the project")
	}
	if err != nil {
		return err
	}

	if group.AuthOverride {
		return d.validateAuthSecret(group.ProjectID, group.IntegrationAuth)
	}

	return nil
}

func (d *SqlDb) CreateIntegrationGroup(group db.IntegrationGroup) (newGroup db.IntegrationGroup, err error) {
	if err = group.Validate(); err != nil {
		return
	}

	if err = d.validateIntegrationGroupEndpoint(group); err != nil {
		return
	}

	group.NormalizeAuth()

	insertID, err := d.insert(
		"id",
		"insert into project__integration_group "+
			"(project_id, `name`, endpoint_id, auth_override, auth_method, auth_secret_id, auth_header) values "+
			"(?, ?, ?, ?, ?, ?, ?)",
		group.ProjectID,
		group.Name,
		group.EndpointID,
		group.AuthOverride,
		group.AuthMethod,
		group.AuthSecretID,
		group.AuthHeader)
	if err != nil {
		return
	}

	newGroup = group
	newGroup.ID = insertID

	return
}

func (d *SqlDb) GetIntegrationGroups(projectID int, params db.RetrieveQueryParams) (groups []db.IntegrationGroup, err error) {
	err = d.getObjects(projectID, db.IntegrationGroupProps, params, nil, &groups)
	return
}

func (d *SqlDb) GetIntegrationGroup(projectID int, groupID int) (group db.IntegrationGroup, err error) {
	err = d.getObject(projectID, db.IntegrationGroupProps, groupID, &group)
	return
}

func (d *SqlDb) GetIntegrationGroupsByEndpoint(projectID int, endpointID int) (groups []db.IntegrationGroup, err error) {
	err = d.getObjects(projectID, db.IntegrationGroupProps, db.RetrieveQueryParams{},
		func(q squirrel.SelectBuilder) squirrel.SelectBuilder {
			return q.Where("endpoint_id=?", endpointID)
		},
		&groups)
	return
}

func (d *SqlDb) UpdateIntegrationGroup(group db.IntegrationGroup) error {
	if err := group.Validate(); err != nil {
		return err
	}

	if err := d.validateIntegrationGroupEndpoint(group); err != nil {
		return err
	}

	group.NormalizeAuth()

	res, err := d.exec(
		"update project__integration_group set "+
			"`name`=?, "+
			"endpoint_id=?, "+
			"auth_override=?, "+
			"auth_method=?, "+
			"auth_secret_id=?, "+
			"auth_header=? "+
			"where project_id=? and `id`=?",
		group.Name,
		group.EndpointID,
		group.AuthOverride,
		group.AuthMethod,
		group.AuthSecretID,
		group.AuthHeader,
		group.ProjectID,
		group.ID)
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

// DeleteIntegrationGroup removes the group with its integrations. The
// integrations go through DeleteIntegration so their task params rows, which
// no foreign key covers, are removed too.
func (d *SqlDb) DeleteIntegrationGroup(projectID int, groupID int) error {
	var group db.IntegrationGroup
	if err := d.getObject(projectID, db.IntegrationGroupProps, groupID, &group); err != nil {
		return err
	}

	integrations, err := d.GetIntegrationsByGroup(projectID, groupID, false)
	if err != nil {
		return err
	}

	for _, integration := range integrations {
		if err = d.DeleteIntegration(projectID, integration.ID); err != nil {
			return err
		}
	}

	return d.deleteObject(projectID, db.IntegrationGroupProps, groupID)
}

func (d *SqlDb) GetIntegrationGroupRefs(projectID int, groupID int) (refs db.IntegrationGroupReferrers, err error) {
	refs.Integrations, err = d.getObjectRefsFrom(projectID, db.IntegrationGroupProps, groupID, db.IntegrationProps)
	return
}
