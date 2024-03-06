package sql

import (
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
	err = d.getProjectObjects(projectID, db.IntegrationProps, params, &integrations)
	return integrations, err
}

func (d *SqlDb) GetAllIntegrations() (integrations []db.Integration, err error) {
	var integrationObjects interface{}
	integrationObjects, err = d.GetAllObjects(db.IntegrationProps)
	integrations = integrationObjects.([]db.Integration)
	return
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
	//extractors, err := d.GetIntegrationExtractors(0, db.RetrieveQueryParams{}, integrationID)
	//
	//if err != nil {
	//	return err
	//}
	//
	//for extractor := range extractors {
	//	d.DeleteIntegrationExtractor(0, extractors[extractor].ID, integrationID)
	//}
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

func (d *SqlDb) GetIntegrationExtractValuesByExtractorID(integrationID int) (values []db.IntegrationExtractValue, err error) {
	var sqlError error
	query, args, sqlError := squirrel.Select("v.*").
		From("project__integration_extract_value as v").
		Where(squirrel.Eq{"integration_id": integrationID}).
		OrderBy("v.id").
		ToSql()

	if sqlError != nil {
		return []db.IntegrationExtractValue{}, sqlError
	}

	err = d.selectOne(&values, query, args...)

	return values, err
}

func (d *SqlDb) GetIntegrationMatchersByExtractorID(integrationID int) (matchers []db.IntegrationMatcher, err error) {
	var sqlError error
	query, args, sqlError := squirrel.Select("m.*").
		From("project__integration_matcher as m").
		Where(squirrel.Eq{"integration_id": integrationID}).
		OrderBy("m.id").
		ToSql()

	if sqlError != nil {
		return []db.IntegrationMatcher{}, sqlError
	}

	err = d.selectOne(&matchers, query, args...)

	return matchers, err
}

//func (d *SqlDb) DeleteIntegrationExtractor(projectID int, integrationID int, integrationID int) error {
//	values, err := d.GetIntegrationExtractValuesByExtractorID(integrationID)
//	if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
//		return err
//	}
//
//	for value := range values {
//
//		err = d.DeleteIntegrationExtractValue(0, values[value].ID, integrationID)
//		if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
//			log.Error(err)
//			return err
//		}
//	}
//
//	matchers, errExtractor := d.GetIntegrationMatchersByExtractorID(integrationID)
//	if errExtractor != nil && !strings.Contains(errExtractor.Error(), "no rows in result set") {
//		log.Error(errExtractor)
//		return errExtractor
//	}
//
//	for matcher := range matchers {
//		err = d.DeleteIntegrationMatcher(0, matchers[matcher].ID, integrationID)
//		if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
//			log.Error(err)
//			return err
//		}
//	}
//
//	return d.deleteObjectByReferencedID(integrationID, db.IntegrationProps, db.IntegrationExtractorProps, integrationID)
//}
//
//func (d *SqlDb) UpdateIntegrationExtractor(projectID int, integrationExtractor db.IntegrationExtractor) error {
//	err := integrationExtractor.Validate()
//
//	if err != nil {
//		return err
//	}
//
//	_, err = d.exec(
//		"update project__integration_extractor set name=? where id=?",
//		integrationExtractor.Name,
//		integrationExtractor.ID)
//
//	return err
//}

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

func (d *SqlDb) GetAllIntegrationExtractValues() (values []db.IntegrationExtractValue, err error) {
	var valueObjects interface{}
	valueObjects, err = d.GetAllObjects(db.IntegrationExtractValueProps)
	values = valueObjects.([]db.IntegrationExtractValue)
	return
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

func (d *SqlDb) GetAllIntegrationMatchers() (matchers []db.IntegrationMatcher, err error) {
	var matcherObjects interface{}
	matcherObjects, err = d.GetAllObjects(db.IntegrationMatcherProps)
	matchers = matcherObjects.([]db.IntegrationMatcher)

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
		"update project__integration_matcher set match_type=?, `method`=?, body_data_type=?, `key`=?, `value`=?, `name`=? where `id`=?",
		integrationMatcher.MatchType,
		integrationMatcher.Method,
		integrationMatcher.BodyDataType,
		integrationMatcher.Key,
		integrationMatcher.Value,
		integrationMatcher.Name,
		integrationMatcher.ID)

	return err
}
