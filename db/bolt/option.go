package bolt

import (
	"errors"
	"github.com/ansible-semaphore/semaphore/db"
	"go.etcd.io/bbolt"
	"strings"
)

func (d *BoltDb) GetOptions(params db.RetrieveQueryParams) (res map[string]string, err error) {
	res = make(map[string]string)
	var options []db.Option
	err = d.getObjects(0, db.OptionProps, db.RetrieveQueryParams{}, func(i interface{}) bool {

		option := i.(db.Option)
		if params.Filter == "" {
			return true
		}

		return option.Key == params.Filter || strings.HasPrefix(option.Key, params.Filter+".")

	}, &options)
	for _, opt := range options {
		res[opt.Key] = opt.Value
	}
	return
}

func (d *BoltDb) SetOption(key string, value string) error {

	opt := db.Option{
		Key:   key,
		Value: value,
	}

	_, err := d.getOption(key)

	if errors.Is(err, db.ErrNotFound) {
		_, err = d.createObject(-1, db.OptionProps, opt)
		return err
	} else {
		err = d.updateObject(-1, db.OptionProps, opt)
	}

	return err
}

func (d *BoltDb) getOption(key string) (value string, err error) {
	var option db.Option
	err = d.getObject(-1, db.OptionProps, strObjectID(key), &option)
	value = option.Value
	return
}

func (d *BoltDb) GetOption(key string) (value string, err error) {
	var option db.Option
	err = d.getObject(-1, db.OptionProps, strObjectID(key), &option)
	value = option.Value

	if errors.Is(err, db.ErrNotFound) {
		err = nil
	}

	return
}

func (d *BoltDb) DeleteOption(key string) (err error) {
	err = db.ValidateOptionKey(key)
	if err != nil {
		return
	}

	return d.db.Update(func(tx *bbolt.Tx) error {
		return d.deleteObject(-1, db.OptionProps, strObjectID(key), tx)
	})
}

func (d *BoltDb) DeleteOptions(filter string) (err error) {
	err = db.ValidateOptionKey(filter)
	if err != nil {
		return
	}

	var options []db.Option

	err = d.getObjects(0, db.OptionProps, db.RetrieveQueryParams{}, func(i interface{}) bool {
		opt := i.(db.Option)
		return opt.Key == filter || strings.HasPrefix(opt.Key, filter+".")
	}, &options)

	for _, opt := range options {
		err = d.DeleteOption(opt.Key)
		if err != nil {
			return
		}
	}

	return
}
