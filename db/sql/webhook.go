package sql

import (
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/masterminds/squirrel"
	"strings"
	// fmt"
	log "github.com/Sirupsen/logrus"
)

func (d *SqlDb) CreateWebhook(webhook db.Webhook) (newWebhook db.Webhook, err error) {
	err = webhook.Validate()

	if err != nil {
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into project__webhook (project_id, name, template_id) values (?, ?, ?)",
		webhook.ProjectID,
		webhook.Name,
		webhook.TemplateID)

	if err != nil {
		return
	}

	newWebhook = webhook
	newWebhook.ID = insertID

	return
}

func (d *SqlDb) GetWebhooks(projectID int, params db.RetrieveQueryParams) (webhooks []db.Webhook, err error) {
	err = d.getObjects(projectID, db.WebhookProps, params, &webhooks)
	return webhooks, err
}

func (d *SqlDb) GetAllWebhooks() (webhooks []db.Webhook, err error) {
	var webhookObjects interface{}
	webhookObjects, err = d.GetAllObjects(db.WebhookProps)
	webhooks = webhookObjects.([]db.Webhook)
	return
}

func (d *SqlDb) GetWebhook(projectID int, webhookID int) (webhook db.Webhook, err error) {
	err = d.getObject(projectID, db.WebhookProps, webhookID, &webhook)
	return
}

func (d *SqlDb) GetWebhookRefs(projectID int, webhookID int) (referrers db.WebhookReferrers, err error) {
	var extractorReferrer []db.ObjectReferrer
	extractorReferrer, err = d.GetObjectReferences(db.WebhookProps, db.WebhookExtractorProps, webhookID)
	referrers = db.WebhookReferrers{
		WebhookExtractors: extractorReferrer,
	}
	return
}

func (d *SqlDb) GetWebhookExtractorsByWebhookID(webhookID int) ([]db.WebhookExtractor, error) {
	var extractors []db.WebhookExtractor
	err := d.GetObjectsByForeignKeyQuery(db.WebhookExtractorProps, webhookID, db.WebhookProps, db.RetrieveQueryParams{}, &extractors);
	return extractors, err
}

func (d *SqlDb) DeleteWebhook(projectID int, webhookID int) error {
	extractors, err := d.GetWebhookExtractorsByWebhookID(webhookID)

	if err != nil {
		return err
	}

	for extractor := range extractors {
		d.DeleteWebhookExtractor(webhookID, extractors[extractor].ID)
	}
	return d.deleteObject(projectID, db.WebhookProps, webhookID)
}

func (d *SqlDb) UpdateWebhook(webhook db.Webhook) error {
	err := webhook.Validate()

	if err != nil {
		return err
	}

	_, err = d.exec(
		"update project__webhook set name=?, template_id=? where id=?",
		webhook.Name,
		webhook.TemplateID,
		webhook.ID)

	return err
}

func (d *SqlDb) CreateWebhookExtractor(webhookExtractor db.WebhookExtractor) (newWebhookExtractor db.WebhookExtractor, err error) {
	err = webhookExtractor.Validate()

	if err != nil {
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into project__webhook_extractor (name, webhook_id) values (?, ?)",
		webhookExtractor.Name,
		webhookExtractor.WebhookID)

	if err != nil {
		return
	}

	newWebhookExtractor = webhookExtractor
	newWebhookExtractor.ID = insertID

	return
}

func (d *SqlDb) GetWebhookExtractor(extractorID int, webhookID int) (extractor db.WebhookExtractor, err error) {
	query, args, err := squirrel.Select("e.*").
		From("project__webhook_extractor as e").
		Where(squirrel.And{
			squirrel.Eq{"webhook_id": webhookID},
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

func (d *SqlDb) GetAllWebhookExtractors() (extractors []db.WebhookExtractor, err error) {
	var extractorObjects interface{}
	extractorObjects, err = d.GetAllObjects(db.WebhookExtractorProps)
	extractors = extractorObjects.([]db.WebhookExtractor)
	return
}

func (d *SqlDb) GetWebhookExtractors(webhookID int, params db.RetrieveQueryParams) ([]db.WebhookExtractor, error) {
	var extractors []db.WebhookExtractor
	err := d.getObjectsByReferrer(webhookID, db.WebhookProps, db.WebhookExtractorProps, params, &extractors)

	return extractors, err
}

func (d *SqlDb) GetWebhookExtractorRefs(webhookID int, extractorID int) (refs db.WebhookExtractorReferrers, err error) {
	refs.WebhookMatchers, err = d.GetObjectReferences(db.WebhookMatcherProps, db.WebhookExtractorProps, extractorID)
	refs.WebhookExtractValues, err = d.GetObjectReferences(db.WebhookExtractValueProps, db.WebhookExtractorProps, extractorID)

	return
}

func (d *SqlDb) GetWebhookExtractValuesByExtractorID(extractorID int) (values []db.WebhookExtractValue, err error) {
	var sqlError error
	query, args, sqlError := squirrel.Select("v.*").
		From("project__webhook_extract_value as v").
		Where(squirrel.Eq{"extractor_id": extractorID}).
		OrderBy("v.id").
		ToSql()

	if sqlError != nil {
		return []db.WebhookExtractValue{}, sqlError
	}

	err = d.selectOne(&values, query, args...)

	return values, err
}

func (d *SqlDb) GetWebhookMatchersByExtractorID(extractorID int) (matchers []db.WebhookMatcher, err error) {
	var sqlError error
	query, args, sqlError := squirrel.Select("m.*").
		From("project__webhook_matcher as m").
		Where(squirrel.Eq{"extractor_id": extractorID}).
		OrderBy("m.id").
		ToSql()

	if sqlError != nil {
		return []db.WebhookMatcher{}, sqlError
	}

	err = d.selectOne(&matchers, query, args...)

	return matchers, err
}

func (d *SqlDb) DeleteWebhookExtractor(webhookID int, extractorID int) error {
	values, err := d.GetWebhookExtractValuesByExtractorID(extractorID)
	if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
		return err
	}

	for value := range values {

		err = d.DeleteWebhookExtractValue(extractorID, values[value].ID)
		if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
			log.Error(err)
			return err
		}
	}

	matchers, errExtractor := d.GetWebhookMatchersByExtractorID(extractorID)
	if errExtractor != nil && !strings.Contains(errExtractor.Error(), "no rows in result set") {
		log.Error(errExtractor)
		return errExtractor
	}

	for matcher := range matchers {
		err = d.DeleteWebhookMatcher(extractorID, matchers[matcher].ID)
		if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
			log.Error(err)
			return err
		}
	}

	return d.deleteObjectByReferencedID(webhookID, db.WebhookProps, db.WebhookExtractorProps, extractorID)
}


func (d *SqlDb) UpdateWebhookExtractor(webhookExtractor db.WebhookExtractor) error {
	err := webhookExtractor.Validate()

	if err != nil {
		return err
	}

	_, err = d.exec(
		"update project__webhook_extractor set name=? where id=?",
		webhookExtractor.Name,
		webhookExtractor.ID)

	return err
}

func (d *SqlDb) CreateWebhookExtractValue(value db.WebhookExtractValue) (newValue db.WebhookExtractValue, err error) {
	err = value.Validate()

	if err != nil {
		return
	}

	insertID, err := d.insert("id",
		"insert into project__webhook_extract_value (value_source, body_data_type, key, variable, name, extractor_id) values (?, ?, ?, ?, ?, ?)",
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

func (d *SqlDb) GetWebhookExtractValues(extractorID int, params db.RetrieveQueryParams) ([]db.WebhookExtractValue, error) {
	var values []db.WebhookExtractValue
	err := d.getObjectsByReferrer(extractorID, db.WebhookExtractorProps, db.WebhookExtractValueProps, params, &values)
	return values, err
}

func (d *SqlDb) GetAllWebhookExtractValues() (values []db.WebhookExtractValue, err error) {
	var valueObjects interface{}
	valueObjects, err = d.GetAllObjects(db.WebhookExtractValueProps)
	values = valueObjects.([]db.WebhookExtractValue)
	return
}

func (d *SqlDb) GetWebhookExtractValue(valueID int, extractorID int) (value db.WebhookExtractValue, err error) {
	query, args, err := squirrel.Select("v.*").
		From("project__webhook_extract_value as v").
		Where(squirrel.Eq{"id": valueID}).
		OrderBy("v.id").
		ToSql()

	if err != nil {
		return
	}

	err = d.selectOne(&value, query, args...)

	return value, err
}

func (d *SqlDb) GetWebhookExtractValueRefs(extractorID int, valueID int) (refs db.WebhookExtractorChildReferrers, err error) {
	refs.WebhookExtractors, err = d.GetObjectReferences(db.WebhookExtractorProps, db.WebhookExtractValueProps, extractorID)
	return
}

func (d *SqlDb) DeleteWebhookExtractValue(extractorID int, valueID int) error {
	return d.deleteObjectByReferencedID(extractorID, db.WebhookExtractorProps, db.WebhookExtractValueProps, valueID)
}


func (d *SqlDb) UpdateWebhookExtractValue(webhookExtractValue db.WebhookExtractValue) error {
	err := webhookExtractValue.Validate()

	if err != nil {
		return err
	}

	_, err = d.exec(
		"update project__webhook_extract_value set value_source=?, body_data_type=?, key=?, variable=?, name=? where id=?",
		webhookExtractValue.ValueSource,
		webhookExtractValue.BodyDataType,
		webhookExtractValue.Key,
		webhookExtractValue.Variable,
		webhookExtractValue.Name,
		webhookExtractValue.ID)

	return err
}

func (d *SqlDb) CreateWebhookMatcher(matcher db.WebhookMatcher) (newMatcher db.WebhookMatcher, err error) {
	err = matcher.Validate()

	if err != nil {
		return
	}

	insertID, err := d.insert(
		"id",
		"insert into project__webhook_matcher (match_type, method, body_data_type, key, value, extractor_id, name) values (?, ?, ?, ?, ?, ?, ?)",
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

func (d *SqlDb) GetWebhookMatchers(extractorID int, params db.RetrieveQueryParams) (matchers []db.WebhookMatcher, err error) {
	query, args, err := squirrel.Select("m.*").
		From("project__webhook_matcher as m").
		Where(squirrel.Eq{"extractor_id": extractorID}).
		OrderBy("m.id").
		ToSql()

	if err != nil {
		return
	}

	_, err = d.selectAll(&matchers, query, args...)

	return
}

func (d *SqlDb) GetAllWebhookMatchers() (matchers []db.WebhookMatcher, err error) {
	var matcherObjects interface{}
	matcherObjects, err = d.GetAllObjects(db.WebhookMatcherProps)
	matchers = matcherObjects.([]db.WebhookMatcher)

	return
}

func (d *SqlDb) GetWebhookMatcher(matcherID int, extractorID int) (matcher db.WebhookMatcher, err error) {
	query, args, err := squirrel.Select("m.*").
		From("project__webhook_matcher as m").
		Where(squirrel.Eq{"id": matcherID}).
		OrderBy("m.id").
		ToSql()

	if err != nil {
		return
	}

	err = d.selectOne(&matcher, query, args...)

	return matcher, err
}

func (d *SqlDb) GetWebhookMatcherRefs(extractorID int, matcherID int) (refs db.WebhookExtractorChildReferrers, err error) {
	refs.WebhookExtractors, err = d.GetObjectReferences(db.WebhookExtractorProps, db.WebhookMatcherProps, matcherID)

	return
}

func (d *SqlDb) DeleteWebhookMatcher(extractorID int, matcherID int) error {
	return d.deleteObjectByReferencedID(extractorID, db.WebhookExtractorProps, db.WebhookMatcherProps, matcherID)
}


func (d *SqlDb) UpdateWebhookMatcher(webhookMatcher db.WebhookMatcher) error {
	err := webhookMatcher.Validate()

	if err != nil {
		return err
	}

	_, err = d.exec(
		"update project__webhook_matcher set match_type=?, method=?, body_data_type=?, key=?, value=?, name=? where id=?",
		webhookMatcher.MatchType,
		webhookMatcher.Method,
		webhookMatcher.BodyDataType,
		webhookMatcher.Key,
		webhookMatcher.Value,
		webhookMatcher.Name,
		webhookMatcher.ID)

	return err
}
