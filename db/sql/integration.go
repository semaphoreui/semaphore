package sql

import (
	"github.com/Masterminds/squirrel"
	"github.com/ansible-semaphore/semaphore/db"
	"strings"
	// fmt"
	log "github.com/sirupsen/logrus"
)

func (d *SqlDb) CreateIntegration(integration db.Integration) (newIntegration db.Integration, err error) {
	err = integration.Validate()

	if err != nil {
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into project__integration "+
			"(project_id, name, template_id, auth_method, auth_secret_id, auth_header) values "+
			"(?, ?, ?, ?, ?, ?)",
		integration.ProjectID,
		integration.Name,
		integration.TemplateID,
		integration.AuthMethod,
		integration.AuthSecretID,
		integration.AuthHeader)

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
	var extractorReferrer []db.ObjectReferrer
	extractorReferrer, err = d.GetObjectReferences(db.IntegrationProps, db.IntegrationExtractorProps, integrationID)
	referrers = db.IntegrationReferrers{
		IntegrationExtractors: extractorReferrer,
	}
	return
}

func (d *SqlDb) GetIntegrationExtractorsByIntegrationID(integrationID int) ([]db.IntegrationExtractor, error) {
	var extractors []db.IntegrationExtractor
	err := d.GetObjectsByForeignKeyQuery(db.IntegrationExtractorProps, integrationID, db.IntegrationProps, db.RetrieveQueryParams{}, &extractors)
	return extractors, err
}

func (d *SqlDb) DeleteIntegration(projectID int, integrationID int) error {
	extractors, err := d.GetIntegrationExtractorsByIntegrationID(integrationID)

	if err != nil {
		return err
	}

	for extractor := range extractors {
		d.DeleteIntegrationExtractor(integrationID, extractors[extractor].ID)
	}
	return d.deleteObject(projectID, db.IntegrationProps, integrationID)
}

func (d *SqlDb) UpdateIntegration(integration db.Integration) error {
	err := integration.Validate()

	if err != nil {
		return err
	}

	_, err = d.exec(
		"update project__integration set name=?, template_id=?, auth_method=?, auth_secret_id=?, auth_header=? where id=?",
		integration.Name,
		integration.TemplateID,
		integration.ID,
		integration.AuthMethod,
		integration.AuthSecretID,
		integration.AuthHeader)

	return err
}

func (d *SqlDb) CreateIntegrationExtractor(integrationExtractor db.IntegrationExtractor) (newIntegrationExtractor db.IntegrationExtractor, err error) {
	err = integrationExtractor.Validate()

	if err != nil {
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into project__integration_extractor (name, integration_id) values (?, ?)",
		integrationExtractor.Name,
		integrationExtractor.IntegrationID)

	if err != nil {
		return
	}

	newIntegrationExtractor = integrationExtractor
	newIntegrationExtractor.ID = insertID

	return
}

func (d *SqlDb) GetIntegrationExtractor(integrationID int, extractorID int) (extractor db.IntegrationExtractor, err error) {
	query, args, err := squirrel.Select("e.*").
		From("project__integration_extractor as e").
		Where(squirrel.And{
			squirrel.Eq{"integration_id": integrationID},
			squirrel.Eq{"id": extractorID},
		}).
		OrderBy("e.name").
		ToSql()

	if err != nil {
		return
	}

	err = d.selectOne(&extractor, query, args...)

	return extractor, err
}

func (d *SqlDb) GetAllIntegrationExtractors() (extractors []db.IntegrationExtractor, err error) {
	var extractorObjects interface{}
	extractorObjects, err = d.GetAllObjects(db.IntegrationExtractorProps)
	extractors = extractorObjects.([]db.IntegrationExtractor)
	return
}

func (d *SqlDb) GetIntegrationExtractors(integrationID int, params db.RetrieveQueryParams) ([]db.IntegrationExtractor, error) {
	var extractors []db.IntegrationExtractor
	err := d.getObjectsByReferrer(integrationID, db.IntegrationExtractorProps, db.IntegrationExtractorProps, params, &extractors)

	return extractors, err
}

func (d *SqlDb) GetIntegrationExtractorRefs(integrationID int, extractorID int) (refs db.IntegrationExtractorReferrers, err error) {
	refs.IntegrationMatchers, err = d.GetObjectReferences(db.IntegrationMatcherProps, db.IntegrationExtractorProps, extractorID)
	refs.IntegrationExtractValues, err = d.GetObjectReferences(db.IntegrationExtractValueProps, db.IntegrationExtractorProps, extractorID)

	return
}

func (d *SqlDb) GetIntegrationExtractValuesByExtractorID(extractorID int) (values []db.IntegrationExtractValue, err error) {
	var sqlError error
	query, args, sqlError := squirrel.Select("v.*").
		From("project__integration_extract_value as v").
		Where(squirrel.Eq{"extractor_id": extractorID}).
		OrderBy("v.id").
		ToSql()

	if sqlError != nil {
		return []db.IntegrationExtractValue{}, sqlError
	}

	err = d.selectOne(&values, query, args...)

	return values, err
}

func (d *SqlDb) GetIntegrationMatchersByExtractorID(extractorID int) (matchers []db.IntegrationMatcher, err error) {
	var sqlError error
	query, args, sqlError := squirrel.Select("m.*").
		From("project__integration_matcher as m").
		Where(squirrel.Eq{"extractor_id": extractorID}).
		OrderBy("m.id").
		ToSql()

	if sqlError != nil {
		return []db.IntegrationMatcher{}, sqlError
	}

	err = d.selectOne(&matchers, query, args...)

	return matchers, err
}

func (d *SqlDb) DeleteIntegrationExtractor(integrationID int, extractorID int) error {
	values, err := d.GetIntegrationExtractValuesByExtractorID(extractorID)
	if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
		return err
	}

	for value := range values {

		err = d.DeleteIntegrationExtractValue(extractorID, values[value].ID)
		if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
			log.Error(err)
			return err
		}
	}

	matchers, errExtractor := d.GetIntegrationMatchersByExtractorID(extractorID)
	if errExtractor != nil && !strings.Contains(errExtractor.Error(), "no rows in result set") {
		log.Error(errExtractor)
		return errExtractor
	}

	for matcher := range matchers {
		err = d.DeleteIntegrationMatcher(extractorID, matchers[matcher].ID)
		if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
			log.Error(err)
			return err
		}
	}

	return d.deleteObjectByReferencedID(integrationID, db.IntegrationProps, db.IntegrationExtractorProps, extractorID)
}

func (d *SqlDb) UpdateIntegrationExtractor(integrationExtractor db.IntegrationExtractor) error {
	err := integrationExtractor.Validate()

	if err != nil {
		return err
	}

	_, err = d.exec(
		"update project__integration_extractor set name=? where id=?",
		integrationExtractor.Name,
		integrationExtractor.ID)

	return err
}

func (d *SqlDb) CreateIntegrationExtractValue(value db.IntegrationExtractValue) (newValue db.IntegrationExtractValue, err error) {
	err = value.Validate()

	if err != nil {
		return
	}

	insertID, err := d.insert("id",
		"insert into project__integration_extract_value "+
			"(value_source, body_data_type, `key`, `variable`, `name`, extractor_id) values "+
			"(?, ?, ?, ?, ?, ?)",
		value.ValueSource,
		value.BodyDataType,
		value.Key,
		value.Variable,
		value.Name,
		value.ExtractorID)

	if err != nil {
		return
	}

	newValue = value
	newValue.ID = insertID

	return
}

func (d *SqlDb) GetIntegrationExtractValues(extractorID int, params db.RetrieveQueryParams) ([]db.IntegrationExtractValue, error) {
	var values []db.IntegrationExtractValue
	err := d.getObjectsByReferrer(extractorID, db.IntegrationExtractValueProps, db.IntegrationExtractValueProps, params, &values)
	return values, err
}

func (d *SqlDb) GetAllIntegrationExtractValues() (values []db.IntegrationExtractValue, err error) {
	var valueObjects interface{}
	valueObjects, err = d.GetAllObjects(db.IntegrationExtractValueProps)
	values = valueObjects.([]db.IntegrationExtractValue)
	return
}

func (d *SqlDb) GetIntegrationExtractValue(extractorID int, valueID int) (value db.IntegrationExtractValue, err error) {
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

func (d *SqlDb) GetIntegrationExtractValueRefs(extractorID int, valueID int) (refs db.IntegrationExtractorChildReferrers, err error) {
	refs.IntegrationExtractors, err = d.GetObjectReferences(db.IntegrationExtractorProps, db.IntegrationExtractValueProps, extractorID)
	return
}

func (d *SqlDb) DeleteIntegrationExtractValue(extractorID int, valueID int) error {
	return d.deleteObjectByReferencedID(extractorID, db.IntegrationExtractorProps, db.IntegrationExtractValueProps, valueID)
}

func (d *SqlDb) UpdateIntegrationExtractValue(integrationExtractValue db.IntegrationExtractValue) error {
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

func (d *SqlDb) CreateIntegrationMatcher(matcher db.IntegrationMatcher) (newMatcher db.IntegrationMatcher, err error) {
	err = matcher.Validate()

	if err != nil {
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into project__integration_matcher "+
			"(match_type, `method`, body_data_type, `key`, `value`, extractor_id, `name`) values "+
			"(?, ?, ?, ?, ?, ?, ?)",
		matcher.MatchType,
		matcher.Method,
		matcher.BodyDataType,
		matcher.Key,
		matcher.Value,
		matcher.ExtractorID,
		matcher.Name)

	if err != nil {
		return
	}

	newMatcher = matcher
	newMatcher.ID = insertID

	return
}

func (d *SqlDb) GetIntegrationMatchers(extractorID int, params db.RetrieveQueryParams) (matchers []db.IntegrationMatcher, err error) {
	query, args, err := squirrel.Select("m.*").
		From("project__integration_matcher as m").
		Where(squirrel.Eq{"extractor_id": extractorID}).
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

func (d *SqlDb) GetIntegrationMatcher(extractorID int, matcherID int) (matcher db.IntegrationMatcher, err error) {
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

func (d *SqlDb) GetIntegrationMatcherRefs(extractorID int, matcherID int) (refs db.IntegrationExtractorChildReferrers, err error) {
	refs.IntegrationExtractors, err = d.GetObjectReferences(db.IntegrationExtractorProps, db.IntegrationMatcherProps, matcherID)

	return
}

func (d *SqlDb) DeleteIntegrationMatcher(extractorID int, matcherID int) error {
	return d.deleteObjectByReferencedID(extractorID, db.IntegrationExtractorProps, db.IntegrationMatcherProps, matcherID)
}

func (d *SqlDb) UpdateIntegrationMatcher(integrationMatcher db.IntegrationMatcher) error {
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
