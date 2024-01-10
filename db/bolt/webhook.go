package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
)

/*
	 Webhooks
*/
func (d *BoltDb) CreateWebhook(webhook db.Webhook) (db.Webhook, error) {
	err := webhook.Validate()

	if err != nil {
		return db.Webhook{}, err
	}

	newWebhook, err := d.createObject(webhook.ProjectID, db.WebhookProps, webhook)
	return newWebhook.(db.Webhook), err
}

func (d *BoltDb) GetWebhooks(projectID int, params db.RetrieveQueryParams) (webhooks []db.Webhook, err error) {
	err = d.getObjects(projectID, db.WebhookProps, params, nil, &webhooks)
	return webhooks, err
}

func (d *BoltDb) GetWebhook(projectID int, webhookID int) (webhook db.Webhook, err error) {
	err = d.getObject(projectID, db.WebhookProps, intObjectID(webhookID), &webhook)
	if err != nil {
		return
	}

	return
}

func (d *BoltDb) GetAllWebhooks() ([]db.Webhook, error) {
	return []db.Webhook{}, nil
}

func (d *BoltDb) UpdateWebhook(webhook db.Webhook) error {
	err := webhook.Validate()

	if err != nil {
		return err
	}

	return d.updateObject(webhook.ProjectID, db.WebhookProps, webhook)

}

func (d *BoltDb) GetWebhookRefs(projectID int, webhookID int) (db.WebhookReferrers, error) {
	//return d.getObjectRefs(projectID, db.WebhookProps, webhookID)
	return db.WebhookReferrers{}, nil
}

/*
	 Webhook Extractors
*/
func (d *BoltDb) GetWebhookExtractorsByWebhookID(webhookID int) (extractors []db.WebhookExtractor, err error) {
	err = d.getObjects(webhookID, db.WebhookExtractorProps, db.RetrieveQueryParams{}, nil, &extractors)
	return extractors, err
}

func (d *BoltDb) CreateWebhookExtractor(webhookExtractor db.WebhookExtractor) (db.WebhookExtractor, error) {
	err := webhookExtractor.Validate()

	if err != nil {
		return db.WebhookExtractor{}, err
	}

	newWebhookExtractor, err := d.createObject(webhookExtractor.WebhookID, db.WebhookExtractorProps, webhookExtractor)
	return newWebhookExtractor.(db.WebhookExtractor), err
}

func (d *BoltDb) GetAllWebhookExtractors() ([]db.WebhookExtractor, error) {

	return []db.WebhookExtractor{}, nil
}

func (d *BoltDb) GetWebhookExtractors(webhookID int, params db.RetrieveQueryParams) ([]db.WebhookExtractor, error) {
	var extractors []db.WebhookExtractor
	err := d.getObjects(webhookID, db.WebhookExtractorProps, params, nil, &extractors)

	return extractors, err
}

func (d *BoltDb) GetWebhookExtractor(webhookID int, extractorID int) (db.WebhookExtractor, error) {
	var extractor db.WebhookExtractor
	err := d.getObject(webhookID, db.WebhookExtractorProps, intObjectID(extractorID), &extractor)

	return extractor, err

}

func (d *BoltDb) UpdateWebhookExtractor(webhookExtractor db.WebhookExtractor) error {
	err := webhookExtractor.Validate()

	if err != nil {
		return err
	}

	return d.updateObject(webhookExtractor.WebhookID, db.WebhookExtractorProps, webhookExtractor)
}

func (d *BoltDb) GetWebhookExtractorRefs(webhookID int, extractorID int) (db.WebhookExtractorReferrers, error) {
	return d.getWebhookExtractorRefs(webhookID, db.WebhookExtractorProps, extractorID)
}

/*
   Webhook ExtractValue
*/
func (d *BoltDb) GetWebhookExtractValuesByExtractorID(extractorID int) (values []db.WebhookExtractValue, err error) {
	err = d.getObjects(extractorID, db.WebhookExtractValueProps, db.RetrieveQueryParams{}, nil, &values)
	return values, err
}

func (d *BoltDb) DeleteWebhookExtractValue(extractorID int, valueID int) error {
	return d.deleteObject(extractorID, db.WebhookExtractValueProps, intObjectID(valueID), nil)
}

func (d *BoltDb) GetWebhookMatchersByExtractorID(extractorID int) (matchers []db.WebhookMatcher, err error) {
	err = d.getObjects(extractorID, db.WebhookMatcherProps, db.RetrieveQueryParams{}, nil, &matchers)

	return matchers, err
}

func (d *BoltDb) GetAllWebhookMatchers() (matchers []db.WebhookMatcher, err error) {
	err = d.getObjects(0, db.WebhookMatcherProps, db.RetrieveQueryParams{}, nil, &matchers)

	return matchers, err
}


func (d *BoltDb) DeleteWebhookExtractor(webhookID int, extractorID int) error {
	values, err := d.GetWebhookExtractValuesByExtractorID(extractorID)

	if err != nil {
		return err
	}

	for value := range values {
		d.DeleteWebhookExtractValue(extractorID, values[value].ID)
	}

	matchers, err := d.GetWebhookMatchersByExtractorID(extractorID)

	if err != nil {
		return err
	}

	for matcher := range matchers {
		d.DeleteWebhookMatcher(extractorID, matchers[matcher].ID)
	}
	return d.deleteObject(webhookID, db.WebhookExtractorProps, intObjectID(extractorID), nil)
}


func (d *BoltDb) CreateWebhookExtractValue(value db.WebhookExtractValue) (db.WebhookExtractValue, error) {
	err := value.Validate()

	if err != nil {
		return db.WebhookExtractValue{}, err
	}

	newValue, err := d.createObject(value.ExtractorID, db.WebhookExtractValueProps, value)
	return newValue.(db.WebhookExtractValue), err

}

func (d *BoltDb) GetWebhookExtractValues(extractorID int, params db.RetrieveQueryParams) (values []db.WebhookExtractValue, err error) {
	values = make([]db.WebhookExtractValue, 0)
	var allValues []db.WebhookExtractValue

	err = d.getObjects(extractorID, db.WebhookExtractValueProps, db.RetrieveQueryParams{}, nil, &allValues)

	if err != nil {
		return
	}

	for _, v := range allValues {
		if v.ExtractorID == extractorID {
			values = append(values, v)
		}
	}

	return
}

func (d *BoltDb) GetAllWebhookExtractValues() (matchers []db.WebhookExtractValue, err error) {
	err = d.getObjects(0, db.WebhookExtractValueProps, db.RetrieveQueryParams{}, nil, &matchers)

	return matchers, err
}


func (d *BoltDb) GetWebhookExtractValue(extractorID int, valueID int) (value db.WebhookExtractValue, err error) {
	err = d.getObject(extractorID, db.WebhookExtractValueProps, intObjectID(valueID), &value)
	return value, err
}

func (d *BoltDb) UpdateWebhookExtractValue(webhookExtractValue db.WebhookExtractValue) error {
	err := webhookExtractValue.Validate()

	if err != nil {
		return err
	}

	return d.updateObject(webhookExtractValue.ExtractorID, db.WebhookExtractValueProps, webhookExtractValue)
}

func (d *BoltDb) GetWebhookExtractValueRefs(extractorID int, valueID int) (db.WebhookExtractorChildReferrers, error) {
	return d.getWebhookExtractorChildrenRefs(extractorID, db.WebhookExtractValueProps, valueID)
}
/*
   Webhook Matcher
*/
func (d *BoltDb) CreateWebhookMatcher(matcher db.WebhookMatcher) (db.WebhookMatcher, error) {
	err := matcher.Validate()

	if err != nil {
		return db.WebhookMatcher{}, err
	}
	newMatcher, err := d.createObject(matcher.ExtractorID, db.WebhookMatcherProps, matcher)
	return newMatcher.(db.WebhookMatcher), err
}

func (d *BoltDb) GetWebhookMatchers(extractorID int, params db.RetrieveQueryParams) (matchers []db.WebhookMatcher, err error) {
	matchers = make([]db.WebhookMatcher, 0)
	var allMatchers []db.WebhookMatcher

	err = d.getObjects(extractorID, db.WebhookMatcherProps, db.RetrieveQueryParams{}, nil, &allMatchers)

	if err != nil {
		return
	}

	for _, v := range allMatchers {
		if v.ExtractorID == extractorID {
			matchers = append(matchers, v)
		}
	}

	return
}

func (d *BoltDb) GetWebhookMatcher(extractorID int,matcherID int) (matcher db.WebhookMatcher, err error) {
	var matchers []db.WebhookMatcher
	matchers, err = d.GetWebhookMatchers(extractorID, db.RetrieveQueryParams{})

	for _, v := range matchers {
		if v.ID == matcherID {
			matcher = v
		}
	}

	return
}

func (d *BoltDb) UpdateWebhookMatcher(webhookMatcher db.WebhookMatcher) error {
	err := webhookMatcher.Validate()

	if err != nil {
		return err
	}

	return d.updateObject(webhookMatcher.ExtractorID, db.WebhookMatcherProps, webhookMatcher)
}

func (d *BoltDb) DeleteWebhookMatcher(extractorID int, matcherID int) error {
	return d.deleteObject(extractorID, db.WebhookMatcherProps, intObjectID(matcherID), nil)
}
func (d *BoltDb) DeleteWebhook(projectID int, webhookID int) error {
	extractors, err := d.GetWebhookExtractorsByWebhookID(webhookID)

	if err != nil {
		return err
	}

	for extractor := range extractors {
		d.DeleteWebhookExtractor(webhookID, extractors[extractor].ID)
	}

	return d.deleteObject(projectID, db.WebhookProps, intObjectID(webhookID), nil)
}

func (d *BoltDb) GetWebhookMatcherRefs(extractorID int, valueID int) (db.WebhookExtractorChildReferrers, error) {
	return d.getWebhookExtractorChildrenRefs(extractorID, db.WebhookMatcherProps, valueID)
}
