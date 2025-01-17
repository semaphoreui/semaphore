package bolt

import (
	"errors"
	"fmt"
	"github.com/semaphoreui/semaphore/db"
	"reflect"
)

type publicAlias struct {
	aliasProps       db.ObjectProps
	publicAliasProps db.ObjectProps
	db               *BoltDb
}

func (d *publicAlias) getAliases(projectID int, filter func(i interface{}) bool, res interface{}) (err error) {

	err = d.db.getObjects(projectID, d.aliasProps, db.RetrieveQueryParams{}, filter, res)

	return
}

func (d *publicAlias) getAlias(projectID int, aliasID int, res interface{}) (err error) {

	err = d.db.getObject(projectID, d.aliasProps, intObjectID(aliasID), res)

	return
}

func (d *publicAlias) getPublicAlias(alias string, aliasObj interface{}) (err error) {

	err = d.db.getObject(-1, d.publicAliasProps, strObjectID(alias), aliasObj)

	return
}

func (d *publicAlias) createAlias(aliasObj interface{}) (newAlias interface{}, err error) {

	alias := aliasObj.(db.Aliasable).ToAlias()

	err = d.getPublicAlias(alias.Alias, newAlias)

	if err == nil {
		err = fmt.Errorf("alias already exists")
	}

	if !errors.Is(err, db.ErrNotFound) {
		return
	}

	newAlias, err = d.db.createObject(alias.ProjectID, d.aliasProps, aliasObj)

	if err != nil {
		return
	}

	_, err = d.db.createObject(-1, d.publicAliasProps, aliasObj)

	if err != nil {
		_ = d.deleteIntegrationAlias(alias.ProjectID, alias.ID)
		return
	}

	return
}

func (d *publicAlias) deleteIntegrationAlias(projectID int, aliasID int) (err error) {
	aliasPtr := reflect.New(d.aliasProps.Type)
	aliasObj := aliasPtr.Elem().Interface()

	alias := aliasObj.(db.Aliasable).ToAlias()

	err = d.db.getObject(projectID, d.aliasProps, intObjectID(aliasID), aliasPtr.Interface())
	if err != nil {
		return
	}

	err = d.db.deleteObject(projectID, d.aliasProps, intObjectID(aliasID), nil)
	if err != nil {
		return
	}

	err = d.db.deleteObject(-1, d.publicAliasProps, strObjectID(alias.Alias), nil)
	if err != nil {
		return
	}

	return
}
