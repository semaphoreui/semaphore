package sql

import (
	"database/sql"
	"github.com/Masterminds/squirrel"
	"github.com/ansible-semaphore/semaphore/db"
)

func (d *SqlDb) CreateIntegration(integration db.Integration) (newIntegration db.Integration, err error) {
	err = integration.Validate()

	if err != nil {
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into project__integration "+
			"(project_id, name, template_id, auth_method, auth_secret_id, auth_header, searchable) values "+
			"(?, ?, ?, ?, ?, ?, ?)",
		integration.ProjectID,
		integration.Name,
		integration.TemplateID,
		integration.AuthMethod,
		integration.AuthSecretID,
		integration.AuthHeader,
		integration.Searchable)

	if err != nil {
		return
	}

	newIntegration = integration
	newIntegration.ID = insertID

	return
}

func (d *SqlDb) GetIntegrations(projectID int, params db.RetrieveQueryParams) (integrations []db.Integration, err error) {
	err = d.getObjects(projectID, db.IntegrationProps, params, nil, &integrations)
	return integrations, err
}

func (d *SqlDb) GetIntegration(projectID int, integrationID int) (integration db.Integration, err error) {
	err = d.getObject(projectID, db.IntegrationProps, integrationID, &integration)
	return
}

func (d *SqlDb) GetIntegrationRefs(projectID int, integrationID int) (referrers db.IntegrationReferrers, err error) {
	//var extractorReferrer []db.ObjectReferrer
	//extractorReferrer, err = d.GetObjectReferences(db.IntegrationProps, db.IntegrationExtractorProps, integrationID)
	//referrers = db.IntegrationReferrers{
	//	IntegrationExtractors: extractorReferrer,
	//}
	return
}

func (d *SqlDb) DeleteIntegration(projectID int, integrationID int) error {
	return d.deleteObject(projectID, db.IntegrationProps, integrationID)
}

func (d *SqlDb) UpdateIntegration(integration db.Integration) error {
	err := integration.Validate()

	if err != nil {
		return err
	}

	_, err = d.exec(
		"update project__integration set `name`=?, template_id=?, auth_method=?, auth_secret_id=?, auth_header=?, searchable=? where `id`=?",
		integration.Name,
		integration.TemplateID,
		integration.AuthMethod,
		integration.AuthSecretID,
		integration.AuthHeader,
		integration.Searchable,
		integration.ID)

	return err
}

func (d *SqlDb) CreateIntegrationExtractValue(projectId int, value db.IntegrationExtractValue) (newValue db.IntegrationExtractValue, err error) {
	err = value.Validate()

	if err != nil {
		return
	}

	insertID, err := d.insert("id",
		"insert into project__integration_extract_value "+
			"(value_source, body_data_type, `key`, `variable`, `name`, integration_id) values "+
			"(?, ?, ?, ?, ?, ?)",
		value.ValueSource,
		value.BodyDataType,
		value.Key,
		value.Variable,
		value.Name,
		value.IntegrationID)

	if err != nil {
		return
	}

	newValue = value
	newValue.ID = insertID

	return
}

func (d *SqlDb) GetIntegrationExtractValues(projectID int, params db.RetrieveQueryParams, integrationID int) ([]db.IntegrationExtractValue, error) {
	var values []db.IntegrationExtractValue
	err := d.getObjectsByReferrer(integrationID, db.IntegrationProps, db.IntegrationExtractValueProps, params, &values)
	return values, err
}

func (d *SqlDb) GetIntegrationExtractValue(projectID int, valueID int, integrationID int) (value db.IntegrationExtractValue, err error) {
	query, args, err := squirrel.Select("v.*").
		From("project__integration_extract_value as v").
		Where(squirrel.Eq{"id": valueID}).
		OrderBy("v.id").
		ToSql()

	if err != nil {
		return
	}

	err = d.selectOne(&value, query, args...)

	return value, err
}

func (d *SqlDb) GetIntegrationExtractValueRefs(projectID int, valueID int, integrationID int) (refs db.IntegrationExtractorChildReferrers, err error) {
	refs.Integrations, err = d.GetObjectReferences(db.IntegrationProps, db.IntegrationExtractValueProps, integrationID)
	return
}

func (d *SqlDb) DeleteIntegrationExtractValue(projectID int, valueID int, integrationID int) error {
	return d.deleteObjectByReferencedID(integrationID, db.IntegrationProps, db.IntegrationExtractValueProps, valueID)
}

func (d *SqlDb) UpdateIntegrationExtractValue(projectID int, integrationExtractValue db.IntegrationExtractValue) error {
	err := integrationExtractValue.Validate()

	if err != nil {
		return err
	}

	_, err = d.exec(
		"update project__integration_extract_value set value_source=?, body_data_type=?, `key`=?, `variable`=?, `name`=? where `id`=?",
		integrationExtractValue.ValueSource,
		integrationExtractValue.BodyDataType,
		integrationExtractValue.Key,
		integrationExtractValue.Variable,
		integrationExtractValue.Name,
		integrationExtractValue.ID)

	return err
}

func (d *SqlDb) CreateIntegrationMatcher(projectID int, matcher db.IntegrationMatcher) (newMatcher db.IntegrationMatcher, err error) {
	err = matcher.Validate()

	if err != nil {
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
		Where(squirrel.Eq{"id": matcherID}).
		OrderBy("m.id").
		ToSql()

	if err != nil {
		return
	}

	err = d.selectOne(&matcher, query, args...)

	return matcher, err
}

func (d *SqlDb) GetIntegrationMatcherRefs(projectID int, matcherID int, integrationID int) (refs db.IntegrationExtractorChildReferrers, err error) {
	refs.Integrations, err = d.GetObjectReferences(db.IntegrationProps, db.IntegrationMatcherProps, matcherID)

	return
}

func (d *SqlDb) DeleteIntegrationMatcher(projectID int, matcherID int, integrationID int) error {
	return d.deleteObjectByReferencedID(integrationID, db.IntegrationProps, db.IntegrationMatcherProps, matcherID)
}

func (d *SqlDb) UpdateIntegrationMatcher(projectID int, integrationMatcher db.IntegrationMatcher) error {
	err := integrationMatcher.Validate()

	if err != nil {
		return err
	}

	_, err = d.exec(
		"update project__integration_matcher set match_type=?, `method`=?, body_data_type=?, `key`=?, `value`=?, `name`=? where integration_id=? and `id`=?",
		integrationMatcher.MatchType,
		integrationMatcher.Method,
		integrationMatcher.BodyDataType,
		integrationMatcher.Key,
		integrationMatcher.Value,
		integrationMatcher.Name,
		integrationMatcher.IntegrationID,
		integrationMatcher.ID)

	return err
}

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

func (d *SqlDb) GetAllSearchableIntegrations() (integrations []db.Integration, err error) {
	q := squirrel.Select("*").From(db.IntegrationProps.TableName)

	q = q.Where("searchable")

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(&integrations, query, args...)

	return
}
