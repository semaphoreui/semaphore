package bolt

import (
	"errors"
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"reflect"
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

func (d *BoltDb) DeleteIntegrationExtractValue(projectID int, valueID int, integrationID int) error {
	return d.deleteObject(projectID, db.IntegrationExtractValueProps, intObjectID(valueID), nil)
}

func (d *BoltDb) CreateIntegrationExtractValue(projectId int, value db.IntegrationExtractValue) (db.IntegrationExtractValue, error) {
	err := value.Validate()

	if err != nil {
		return db.IntegrationExtractValue{}, err
	}

	newValue, err := d.createObject(projectId, db.IntegrationExtractValueProps, value)
	return newValue.(db.IntegrationExtractValue), err

}

func (d *BoltDb) GetIntegrationExtractValues(projectID int, params db.RetrieveQueryParams, integrationID int) (values []db.IntegrationExtractValue, err error) {
	values = make([]db.IntegrationExtractValue, 0)

	err = d.getObjects(projectID, db.IntegrationExtractValueProps, params, func(i interface{}) bool {
		v := i.(db.IntegrationExtractValue)
		return v.IntegrationID == integrationID
	}, &values)

	return
}

func (d *BoltDb) GetIntegrationExtractValue(projectID int, valueID int, integrationID int) (value db.IntegrationExtractValue, err error) {
	err = d.getObject(projectID, db.IntegrationExtractValueProps, intObjectID(valueID), &value)
	return value, err
}

func (d *BoltDb) UpdateIntegrationExtractValue(projectID int, integrationExtractValue db.IntegrationExtractValue) error {
	err := integrationExtractValue.Validate()

	if err != nil {
		return err
	}

	return d.updateObject(projectID, db.IntegrationExtractValueProps, integrationExtractValue)
}

func (d *BoltDb) GetIntegrationExtractValueRefs(projectID int, valueID int, integrationID int) (db.IntegrationExtractorChildReferrers, error) {
	return d.getIntegrationExtractorChildrenRefs(projectID, db.IntegrationExtractValueProps, valueID)
}

/*
Integration Matcher
*/
func (d *BoltDb) CreateIntegrationMatcher(projectID int, matcher db.IntegrationMatcher) (db.IntegrationMatcher, error) {
	err := matcher.Validate()

	if err != nil {
		return db.IntegrationMatcher{}, err
	}
	newMatcher, err := d.createObject(projectID, db.IntegrationMatcherProps, matcher)
	return newMatcher.(db.IntegrationMatcher), err
}

func (d *BoltDb) GetIntegrationMatchers(projectID int, params db.RetrieveQueryParams, integrationID int) (matchers []db.IntegrationMatcher, err error) {
	matchers = make([]db.IntegrationMatcher, 0)

	err = d.getObjects(projectID, db.IntegrationMatcherProps, db.RetrieveQueryParams{}, func(i interface{}) bool {
		v := i.(db.IntegrationMatcher)
		return v.IntegrationID == integrationID
	}, &matchers)

	return
}

func (d *BoltDb) GetIntegrationMatcher(projectID int, matcherID int, integrationID int) (matcher db.IntegrationMatcher, err error) {
	var matchers []db.IntegrationMatcher
	matchers, err = d.GetIntegrationMatchers(projectID, db.RetrieveQueryParams{}, integrationID)

	for _, v := range matchers {
		if v.ID == matcherID {
			matcher = v
		}
	}

	return
}

func (d *BoltDb) UpdateIntegrationMatcher(projectID int, integrationMatcher db.IntegrationMatcher) error {
	err := integrationMatcher.Validate()

	if err != nil {
		return err
	}

	return d.updateObject(projectID, db.IntegrationMatcherProps, integrationMatcher)
}

func (d *BoltDb) DeleteIntegrationMatcher(projectID int, matcherID int, integrationID int) error {
	return d.deleteObject(projectID, db.IntegrationMatcherProps, intObjectID(matcherID), nil)
}
func (d *BoltDb) DeleteIntegration(projectID int, integrationID int) error {
	matchers, err := d.GetIntegrationMatchers(projectID, db.RetrieveQueryParams{}, integrationID)

	if err != nil {
		return err
	}

	for m := range matchers {
		d.DeleteIntegrationMatcher(projectID, matchers[m].ID, integrationID)
	}

	return d.deleteObject(projectID, db.IntegrationProps, intObjectID(integrationID), nil)
}

func (d *BoltDb) GetIntegrationMatcherRefs(projectID int, matcherID int, integrationID int) (db.IntegrationExtractorChildReferrers, error) {
	return d.getIntegrationExtractorChildrenRefs(projectID, db.IntegrationMatcherProps, matcherID)
}

var integrationAliasProps = db.ObjectProps{
	TableName:         "integration_alias",
	Type:              reflect.TypeOf(db.IntegrationAlias{}),
	PrimaryColumnName: "alias",
}

func (d *BoltDb) GetIntegrationAliases(projectID int, integrationID *int) (res []db.IntegrationAlias, err error) {

	err = d.getObjects(projectID, db.IntegrationAliasProps, db.RetrieveQueryParams{}, func(i interface{}) bool {
		alias := i.(db.IntegrationAlias)
		if alias.IntegrationID == nil && integrationID == nil {
			return true
		} else if alias.IntegrationID != nil && integrationID != nil {
			return *alias.IntegrationID == *integrationID
		}
		return false
	}, &res)

	return
}

func (d *BoltDb) GetIntegrationsByAlias(alias string) (res []db.Integration, err error) {

	var aliasObj db.IntegrationAlias
	err = d.getObject(-1, integrationAliasProps, strObjectID(alias), &aliasObj)

	if err != nil {
		return
	}

	if aliasObj.IntegrationID == nil {
		err = d.getObjects(aliasObj.ProjectID, db.IntegrationProps, db.RetrieveQueryParams{}, func(i interface{}) bool {
			integration := i.(db.Integration)
			return integration.Searchable
		}, &res)

		if err != nil {
			return
		}

	} else {
		var integration db.Integration
		integration, err = d.GetIntegration(aliasObj.ProjectID, *aliasObj.IntegrationID)
		if err != nil {
			return
		}
		res = append(res, integration)
	}

	return
}

func (d *BoltDb) CreateIntegrationAlias(alias db.IntegrationAlias) (res db.IntegrationAlias, err error) {

	_, err = d.GetIntegrationsByAlias(alias.Alias)

	if err == nil {
		err = fmt.Errorf("alias already exists")
	}

	if !errors.Is(err, db.ErrNotFound) {
		return
	}

	newAlias, err := d.createObject(alias.ProjectID, db.IntegrationAliasProps, alias)

	if err != nil {
		return
	}

	res = newAlias.(db.IntegrationAlias)

	_, err = d.createObject(-1, integrationAliasProps, alias)

	if err != nil {
		_ = d.DeleteIntegrationAlias(alias.ProjectID, alias.ID)
		return
	}

	return
}

func (d *BoltDb) DeleteIntegrationAlias(projectID int, aliasID int) (err error) {

	var alias db.IntegrationAlias
	err = d.getObject(projectID, db.IntegrationAliasProps, intObjectID(aliasID), &alias)
	if err != nil {
		return
	}

	err = d.deleteObject(projectID, db.IntegrationAliasProps, intObjectID(aliasID), nil)
	if err != nil {
		return
	}

	err = d.deleteObject(-1, integrationAliasProps, strObjectID(alias.Alias), nil)
	if err != nil {
		return
	}

	return
}

func (d *BoltDb) GetAllSearchableIntegrations() (integrations []db.Integration, err error) {
	integrations = make([]db.Integration, 0)

	projects, err := d.GetAllProjects()
	if err != nil {
		return
	}

	for _, project := range projects {
		var projectIntegrations []db.Integration
		projectIntegrations, err = d.GetIntegrations(project.ID, db.RetrieveQueryParams{})
		if err != nil {
			return
		}

		integrations = append(projectIntegrations)
	}

	return
}
