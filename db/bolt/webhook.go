package bolt

import (
	"github.com/ansible-semaphore/semaphore/db"
)

/*
Integrations
*/
func (d *BoltDb) CreateIntegration(integration db.Integration) (db.Integration, error) {
	err := integration.Validate()

	if err != nil {
		return db.Integration{}, err
	}

	newIntegration, err := d.createObject(integration.ProjectID, db.IntegrationProps, integration)
	return newIntegration.(db.Integration), err
}

func (d *BoltDb) GetIntegrations(projectID int, params db.RetrieveQueryParams) (integrations []db.Integration, err error) {
	err = d.getObjects(projectID, db.IntegrationProps, params, nil, &integrations)
	return integrations, err
}

func (d *BoltDb) GetIntegration(projectID int, integrationID int) (integration db.Integration, err error) {
	err = d.getObject(projectID, db.IntegrationProps, intObjectID(integrationID), &integration)
	if err != nil {
		return
	}

	return
}

func (d *BoltDb) GetAllIntegrations() ([]db.Integration, error) {
	return []db.Integration{}, nil
}

func (d *BoltDb) UpdateIntegration(integration db.Integration) error {
	err := integration.Validate()

	if err != nil {
		return err
	}

	return d.updateObject(integration.ProjectID, db.IntegrationProps, integration)

}

func (d *BoltDb) GetIntegrationRefs(projectID int, integrationID int) (db.IntegrationReferrers, error) {
	//return d.getObjectRefs(projectID, db.IntegrationProps, integrationID)
	return db.IntegrationReferrers{}, nil
}

/*
Integration Extractors
*/

func (d *BoltDb) CreateIntegrationExtractor(integrationExtractor db.IntegrationExtractor) (db.IntegrationExtractor, error) {
	err := integrationExtractor.Validate()

	if err != nil {
		return db.IntegrationExtractor{}, err
	}

	newIntegrationExtractor, err := d.createObject(integrationExtractor.IntegrationID, db.IntegrationExtractorProps, integrationExtractor)
	return newIntegrationExtractor.(db.IntegrationExtractor), err
}

func (d *BoltDb) GetIntegrationExtractors(integrationID int, params db.RetrieveQueryParams) ([]db.IntegrationExtractor, error) {
	var extractors []db.IntegrationExtractor
	err := d.getObjects(integrationID, db.IntegrationExtractorProps, params, nil, &extractors)

	return extractors, err
}

func (d *BoltDb) GetIntegrationExtractor(integrationID int, extractorID int) (db.IntegrationExtractor, error) {
	var extractor db.IntegrationExtractor
	err := d.getObject(integrationID, db.IntegrationExtractorProps, intObjectID(extractorID), &extractor)

	return extractor, err

}

func (d *BoltDb) UpdateIntegrationExtractor(integrationExtractor db.IntegrationExtractor) error {
	err := integrationExtractor.Validate()

	if err != nil {
		return err
	}

	return d.updateObject(integrationExtractor.IntegrationID, db.IntegrationExtractorProps, integrationExtractor)
}

func (d *BoltDb) GetIntegrationExtractorRefs(integrationID int, extractorID int) (db.IntegrationExtractorReferrers, error) {
	return d.getIntegrationExtractorRefs(integrationID, db.IntegrationExtractorProps, extractorID)
}

/*
Integration ExtractValue
*/
func (d *BoltDb) GetIntegrationExtractValuesByExtractorID(extractorID int) (values []db.IntegrationExtractValue, err error) {
	err = d.getObjects(extractorID, db.IntegrationExtractValueProps, db.RetrieveQueryParams{}, nil, &values)
	return values, err
}

func (d *BoltDb) DeleteIntegrationExtractValue(extractorID int, valueID int) error {
	return d.deleteObject(extractorID, db.IntegrationExtractValueProps, intObjectID(valueID), nil)
}

func (d *BoltDb) GetIntegrationMatchersByExtractorID(extractorID int) (matchers []db.IntegrationMatcher, err error) {
	err = d.getObjects(extractorID, db.IntegrationMatcherProps, db.RetrieveQueryParams{}, nil, &matchers)

	return matchers, err
}

func (d *BoltDb) GetAllIntegrationMatchers() (matchers []db.IntegrationMatcher, err error) {
	err = d.getObjects(0, db.IntegrationMatcherProps, db.RetrieveQueryParams{}, nil, &matchers)

	return matchers, err
}

func (d *BoltDb) DeleteIntegrationExtractor(integrationID int, extractorID int) error {
	values, err := d.GetIntegrationExtractValuesByExtractorID(extractorID)

	if err != nil {
		return err
	}

	for value := range values {
		d.DeleteIntegrationExtractValue(extractorID, values[value].ID)
	}

	matchers, err := d.GetIntegrationMatchersByExtractorID(extractorID)

	if err != nil {
		return err
	}

	for matcher := range matchers {
		d.DeleteIntegrationMatcher(extractorID, matchers[matcher].ID)
	}
	return d.deleteObject(integrationID, db.IntegrationExtractorProps, intObjectID(extractorID), nil)
}

func (d *BoltDb) CreateIntegrationExtractValue(value db.IntegrationExtractValue) (db.IntegrationExtractValue, error) {
	err := value.Validate()

	if err != nil {
		return db.IntegrationExtractValue{}, err
	}

	newValue, err := d.createObject(value.ExtractorID, db.IntegrationExtractValueProps, value)
	return newValue.(db.IntegrationExtractValue), err

}

func (d *BoltDb) GetIntegrationExtractValues(extractorID int, params db.RetrieveQueryParams) (values []db.IntegrationExtractValue, err error) {
	values = make([]db.IntegrationExtractValue, 0)
	var allValues []db.IntegrationExtractValue

	err = d.getObjects(extractorID, db.IntegrationExtractValueProps, db.RetrieveQueryParams{}, nil, &allValues)

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

func (d *BoltDb) GetAllIntegrationExtractValues() (matchers []db.IntegrationExtractValue, err error) {
	err = d.getObjects(0, db.IntegrationExtractValueProps, db.RetrieveQueryParams{}, nil, &matchers)

	return matchers, err
}

func (d *BoltDb) GetIntegrationExtractValue(extractorID int, valueID int) (value db.IntegrationExtractValue, err error) {
	err = d.getObject(extractorID, db.IntegrationExtractValueProps, intObjectID(valueID), &value)
	return value, err
}

func (d *BoltDb) UpdateIntegrationExtractValue(integrationExtractValue db.IntegrationExtractValue) error {
	err := integrationExtractValue.Validate()

	if err != nil {
		return err
	}

	return d.updateObject(integrationExtractValue.ExtractorID, db.IntegrationExtractValueProps, integrationExtractValue)
}

func (d *BoltDb) GetIntegrationExtractValueRefs(extractorID int, valueID int) (db.IntegrationExtractorChildReferrers, error) {
	return d.getIntegrationExtractorChildrenRefs(extractorID, db.IntegrationExtractValueProps, valueID)
}

/*
Integration Matcher
*/
func (d *BoltDb) CreateIntegrationMatcher(matcher db.IntegrationMatcher) (db.IntegrationMatcher, error) {
	err := matcher.Validate()

	if err != nil {
		return db.IntegrationMatcher{}, err
	}
	newMatcher, err := d.createObject(matcher.ExtractorID, db.IntegrationMatcherProps, matcher)
	return newMatcher.(db.IntegrationMatcher), err
}

func (d *BoltDb) GetIntegrationMatchers(extractorID int, params db.RetrieveQueryParams) (matchers []db.IntegrationMatcher, err error) {
	matchers = make([]db.IntegrationMatcher, 0)
	var allMatchers []db.IntegrationMatcher

	err = d.getObjects(extractorID, db.IntegrationMatcherProps, db.RetrieveQueryParams{}, nil, &allMatchers)

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

func (d *BoltDb) GetIntegrationMatcher(extractorID int, matcherID int) (matcher db.IntegrationMatcher, err error) {
	var matchers []db.IntegrationMatcher
	matchers, err = d.GetIntegrationMatchers(extractorID, db.RetrieveQueryParams{})

	for _, v := range matchers {
		if v.ID == matcherID {
			matcher = v
		}
	}

	return
}

func (d *BoltDb) UpdateIntegrationMatcher(integrationMatcher db.IntegrationMatcher) error {
	err := integrationMatcher.Validate()

	if err != nil {
		return err
	}

	return d.updateObject(integrationMatcher.ExtractorID, db.IntegrationMatcherProps, integrationMatcher)
}

func (d *BoltDb) DeleteIntegrationMatcher(extractorID int, matcherID int) error {
	return d.deleteObject(extractorID, db.IntegrationMatcherProps, intObjectID(matcherID), nil)
}
func (d *BoltDb) DeleteIntegration(projectID int, integrationID int) error {
	extractors, err := d.GetIntegrationExtractors(integrationID, db.RetrieveQueryParams{})

	if err != nil {
		return err
	}

	for extractor := range extractors {
		d.DeleteIntegrationExtractor(integrationID, extractors[extractor].ID)
	}

	return d.deleteObject(projectID, db.IntegrationProps, intObjectID(integrationID), nil)
}

func (d *BoltDb) GetIntegrationMatcherRefs(extractorID int, valueID int) (db.IntegrationExtractorChildReferrers, error) {
	return d.getIntegrationExtractorChildrenRefs(extractorID, db.IntegrationMatcherProps, valueID)
}
