package sql

import (
	"github.com/Masterminds/squirrel"
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) CreateIntegration(integration db.Integration) (newIntegration db.Integration, err error) {
	err = integration.Validate()

	if err != nil {
		return
	}

	if integration.TaskParams != nil {
		params := *integration.TaskParams
		params.ProjectID = integration.ProjectID
		err = d.Sql().Insert(&params)
		if err != nil {
			return
		}
		integration.TaskParamsID = &params.ID
	}

	insertID, err := d.insert(
		"id",
		"insert into project__integration "+
			"(project_id, name, template_id, auth_method, auth_secret_id, auth_header, searchable, task_params_id) values "+
			"(?, ?, ?, ?, ?, ?, ?, ?)",
		integration.ProjectID,
		integration.Name,
		integration.TemplateID,
		integration.AuthMethod,
		integration.AuthSecretID,
		integration.AuthHeader,
		integration.Searchable,
		integration.TaskParamsID)

	if err != nil {
		return
	}

	newIntegration = integration
	newIntegration.ID = insertID

	return
}

func (d *SqlDb) GetIntegrations(projectID int, params db.RetrieveQueryParams, includeTaskParams bool) (integrations []db.Integration, err error) {
	err = d.getObjects(projectID, db.IntegrationProps, params, nil, &integrations)

	if includeTaskParams {
		for i := range integrations {
			if integrations[i].TaskParamsID == nil {
				continue
			}

			var taskParams db.TaskParams
			err = d.getObject(projectID, db.TaskParamsProps, *integrations[i].TaskParamsID, &taskParams)
			if err != nil {
				return nil, err
			}
			integrations[i].TaskParams = &taskParams
		}
	}

	return integrations, err
}

func (d *SqlDb) GetIntegration(projectID int, integrationID int) (integration db.Integration, err error) {
	err = d.getObject(projectID, db.IntegrationProps, integrationID, &integration)
	if err != nil {
		return
	}

	if integration.TaskParamsID != nil {
		var taskParams db.TaskParams
		err = d.getObject(projectID, db.TaskParamsProps, *integration.TaskParamsID, &taskParams)
		if err != nil {
			return
		}

		integration.TaskParams = &taskParams
	}

	return
}

func (d *SqlDb) GetIntegrationRefs(projectID int, integrationID int) (referrers db.IntegrationReferrers, err error) {
	referrers.IntegrationMatchers = make([]db.ObjectReferrer, 0)
	referrers.IntegrationExtractValues = make([]db.ObjectReferrer, 0)

	var matchers []db.IntegrationMatcher
	_, err = d.selectAll(&matchers,
		"select id, name from project__integration_matcher "+
			"where integration_id=? "+
			"and integration_id in (select id from project__integration where project_id=?)",
		integrationID,
		projectID)
	if err != nil {
		return
	}

	for _, m := range matchers {
		referrers.IntegrationMatchers = append(referrers.IntegrationMatchers, db.ObjectReferrer{ID: m.ID, Name: m.Name})
	}

	var values []db.IntegrationExtractValue
	_, err = d.selectAll(&values,
		"select id, name from project__integration_extract_value "+
			"where integration_id=? "+
			"and integration_id in (select id from project__integration where project_id=?)",
		integrationID,
		projectID)
	if err != nil {
		return
	}

	for _, v := range values {
		referrers.IntegrationExtractValues = append(referrers.IntegrationExtractValues, db.ObjectReferrer{ID: v.ID, Name: v.Name})
	}

	return
}

func (d *SqlDb) DeleteIntegration(projectID int, integrationID int) (err error) {
	var integration db.Integration
	err = d.getObject(projectID, db.IntegrationProps, integrationID, &integration)
	if err != nil {
		return
	}

	err = d.deleteObject(projectID, db.IntegrationProps, integrationID)
	if err != nil {
		return
	}

	if integration.TaskParamsID != nil {
		err = d.deleteObject(projectID, db.TaskParamsProps, *integration.TaskParamsID)
	}
	return
}

func (d *SqlDb) UpdateIntegration(integration db.Integration) (err error) {

	if err = integration.Validate(); err != nil {
		return
	}

	if integration.TaskParams != nil {
		var curr db.Integration
		err = d.getObject(integration.ProjectID, db.IntegrationProps, integration.ID, &curr)
		if err != nil {
			return
		}

		params := *integration.TaskParams
		params.ProjectID = integration.ProjectID

		if curr.TaskParamsID == nil {
			err = d.Sql().Insert(&params)
		} else {
			params.ID = *curr.TaskParamsID
			_, err = d.Sql().Update(&params)
		}

		if err != nil {
			return
		}

		integration.TaskParamsID = &params.ID
	}

	_, err = d.exec(
		"update project__integration set "+
			"`name`=?, "+
			"template_id=?, "+
			"auth_method=?, "+
			"auth_secret_id=?, "+
			"auth_header=?, "+
			"searchable=?, "+
			"task_params_id=? "+
			"where project_id=? AND `id`=?",
		integration.Name,
		integration.TemplateID,
		integration.AuthMethod,
		integration.AuthSecretID,
		integration.AuthHeader,
		integration.Searchable,
		integration.TaskParamsID,
		integration.ProjectID,
		integration.ID)

	return err
}

// validateIntegrationOwnership returns db.ErrNotFound if the integration
// does not belong to the project.
func (d *SqlDb) validateIntegrationOwnership(projectID int, integrationID int) error {
	var integration db.Integration
	return d.getObject(projectID, db.IntegrationProps, integrationID, &integration)
}

func (d *SqlDb) CreateIntegrationExtractValue(projectId int, value db.IntegrationExtractValue) (newValue db.IntegrationExtractValue, err error) {
	err = value.Validate()

	if err != nil {
		return
	}

	if err = d.validateIntegrationOwnership(projectId, value.IntegrationID); err != nil {
		return
	}

	insertID, err := d.insert("id",
		"insert into project__integration_extract_value "+
			"(value_source, body_data_type, `key`, `variable`, `name`, integration_id, variable_type) values "+
			"(?, ?, ?, ?, ?, ?, ?)",
		value.ValueSource,
		value.BodyDataType,
		value.Key,
		value.Variable,
		value.Name,
		value.IntegrationID,
		value.VariableType)

	if err != nil {
		return
	}

	newValue = value
	newValue.ID = insertID

	return
}

func (d *SqlDb) GetIntegrationExtractValues(projectID int, params db.RetrieveQueryParams, integrationID int) ([]db.IntegrationExtractValue, error) {
	var values []db.IntegrationExtractValue

	q := squirrel.Select("pe.*").
		From("project__integration_extract_value as pe").
		Where("pe.integration_id=?", integrationID).
		Where("pe.integration_id in (select id from project__integration where project_id=?)", projectID)

	q, err := getQueryForParams(q, "pe.", db.IntegrationExtractValueProps, params)
	if err != nil {
		return nil, err
	}

	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = d.selectAll(&values, query, args...)

	return values, err
}

func (d *SqlDb) GetIntegrationExtractValue(projectID int, valueID int, integrationID int) (value db.IntegrationExtractValue, err error) {
	query, args, err := squirrel.Select("v.*").
		From("project__integration_extract_value as v").
		Where(squirrel.Eq{"id": valueID, "integration_id": integrationID}).
		Where("v.integration_id in (select id from project__integration where project_id=?)", projectID).
		OrderBy("v.id").
		ToSql()

	if err != nil {
		return
	}

	err = d.selectOne(&value, query, args...)

	return value, err
}

func (d *SqlDb) GetIntegrationExtractValueRefs(projectID int, valueID int, integrationID int) (refs db.IntegrationExtractorChildReferrers, err error) {
	if err = d.validateIntegrationOwnership(projectID, integrationID); err != nil {
		return
	}

	refs.Integrations, err = d.GetObjectReferences(db.IntegrationProps, db.IntegrationExtractValueProps, integrationID)
	return
}

func (d *SqlDb) DeleteIntegrationExtractValue(projectID int, valueID int, integrationID int) error {
	res, err := d.exec(
		"delete from project__integration_extract_value "+
			"where `id`=? and integration_id=? "+
			"and integration_id in (select id from project__integration where project_id=?)",
		valueID,
		integrationID,
		projectID)

	if err != nil {
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

func (d *SqlDb) UpdateIntegrationExtractValue(projectID int, integrationExtractValue db.IntegrationExtractValue) error {
	err := integrationExtractValue.Validate()

	if err != nil {
		return err
	}

	res, err := d.exec(
		"update project__integration_extract_value set value_source=?, body_data_type=?, `key`=?, `variable`=?, `name`=?, `variable_type`=? "+
			"where integration_id=? and `id`=? "+
			"and integration_id in (select id from project__integration where project_id=?)",
		integrationExtractValue.ValueSource,
		integrationExtractValue.BodyDataType,
		integrationExtractValue.Key,
		integrationExtractValue.Variable,
		integrationExtractValue.Name,
		integrationExtractValue.VariableType,
		integrationExtractValue.IntegrationID,
		integrationExtractValue.ID,
		projectID)

	if err != nil {
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

func (d *SqlDb) CreateIntegrationMatcher(projectID int, matcher db.IntegrationMatcher) (newMatcher db.IntegrationMatcher, err error) {
	err = matcher.Validate()

	if err != nil {
		return
	}

	if err = d.validateIntegrationOwnership(projectID, matcher.IntegrationID); err != nil {
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into project__integration_matcher "+
			"(match_type, `method`, body_data_type, `key`, `value`, integration_id, `name`) values "+
			"(?, ?, ?, ?, ?, ?, ?)",
		matcher.MatchType,
		matcher.Method,
		matcher.BodyDataType,
		matcher.Key,
		matcher.Value,
		matcher.IntegrationID,
		matcher.Name)

	if err != nil {
		return
	}

	newMatcher = matcher
	newMatcher.ID = insertID

	return
}

func (d *SqlDb) GetIntegrationMatchers(projectID int, params db.RetrieveQueryParams, integrationID int) (matchers []db.IntegrationMatcher, err error) {
	query, args, err := squirrel.Select("m.*").
		From("project__integration_matcher as m").
		Where(squirrel.Eq{"integration_id": integrationID}).
		Where("m.integration_id in (select id from project__integration where project_id=?)", projectID).
		OrderBy("m.id").
		ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(&matchers, query, args...)

	return
}

func (d *SqlDb) GetIntegrationMatcher(projectID int, matcherID int, integrationID int) (matcher db.IntegrationMatcher, err error) {
	query, args, err := squirrel.Select("m.*").
		From("project__integration_matcher as m").
		Where(squirrel.Eq{"id": matcherID, "integration_id": integrationID}).
		Where("m.integration_id in (select id from project__integration where project_id=?)", projectID).
		OrderBy("m.id").
		ToSql()

	if err != nil {
		return
	}

	err = d.selectOne(&matcher, query, args...)

	return matcher, err
}

func (d *SqlDb) GetIntegrationMatcherRefs(projectID int, matcherID int, integrationID int) (refs db.IntegrationExtractorChildReferrers, err error) {
	if err = d.validateIntegrationOwnership(projectID, integrationID); err != nil {
		return
	}

	refs.Integrations, err = d.GetObjectReferences(db.IntegrationProps, db.IntegrationMatcherProps, matcherID)

	return
}

func (d *SqlDb) DeleteIntegrationMatcher(projectID int, matcherID int, integrationID int) error {
	res, err := d.exec(
		"delete from project__integration_matcher "+
			"where `id`=? and integration_id=? "+
			"and integration_id in (select id from project__integration where project_id=?)",
		matcherID,
		integrationID,
		projectID)

	if err != nil {
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

func (d *SqlDb) UpdateIntegrationMatcher(projectID int, integrationMatcher db.IntegrationMatcher) error {
	err := integrationMatcher.Validate()

	if err != nil {
		return err
	}

	res, err := d.exec(
		"update project__integration_matcher set match_type=?, `method`=?, body_data_type=?, `key`=?, `value`=?, `name`=? "+
			"where integration_id=? and `id`=? "+
			"and integration_id in (select id from project__integration where project_id=?)",
		integrationMatcher.MatchType,
		integrationMatcher.Method,
		integrationMatcher.BodyDataType,
		integrationMatcher.Key,
		integrationMatcher.Value,
		integrationMatcher.Name,
		integrationMatcher.IntegrationID,
		integrationMatcher.ID,
		projectID)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return db.ErrNotFound
	}

	return err
}
