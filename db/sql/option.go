package sql

import (
	"database/sql"
	"errors"
	"github.com/Masterminds/squirrel"
	"github.com/ansible-semaphore/semaphore/db"
)

func (d *SqlDb) SetOption(key string, value string) error {
	_, err := d.getOption(key)

	if errors.Is(err, db.ErrNotFound) {
		_, err = d.insert(
			"key",
			"insert into `option` (`key`, `value`) values (?, ?)",
			key, value)
	} else if err == nil {
		_, err = d.exec("update `option` set `value`=? where `key`=?", value, key)
	}

	return err
}

func (d *SqlDb) GetOptions(params db.RetrieveQueryParams) (res map[string]string, err error) {
	var options []db.Option
	res = make(map[string]string)

	if params.Filter != "" {
		err = db.ValidateOptionKey(params.Filter)
		if err != nil {
			return
		}
	}

	err = d.getObjects(0, db.OptionProps, params, func(q squirrel.SelectBuilder) squirrel.SelectBuilder {
		if params.Filter == "" {
			return q
		}
		return q.Where("`key` = ? OR `key` LIKE ?", params.Filter, params.Filter+".%")
	}, &options)

	if err != nil {
		return
	}

	for _, opt := range options {
		res[opt.Key] = opt.Value
	}

	return
}

func (d *SqlDb) getOption(key string) (value string, err error) {
	q := squirrel.Select("*").
		From("`"+db.OptionProps.TableName+"`").
		Where("`key`=?", key)

	query, args, err := q.ToSql()

	if err != nil {
		return
	}

	var opt db.Option

	err = d.selectOne(&opt, query, args...)

	if errors.Is(err, sql.ErrNoRows) {
		err = db.ErrNotFound
	}

	value = opt.Value

	return
}

func (d *SqlDb) GetOption(key string) (value string, err error) {

	value, err = d.getOption(key)

	if errors.Is(err, db.ErrNotFound) {
		err = nil
	}

	return
}

func (d *SqlDb) DeleteOption(key string) (err error) {
	err = db.ValidateOptionKey(key)
	if err != nil {
		return
	}

	err = d.deleteObject(0, db.OptionProps, key)

	return
}

func (d *SqlDb) DeleteOptions(filter string) (err error) {
	err = db.ValidateOptionKey(filter)
	if err != nil {
		return
	}

	_, err = d.exec("DELETE FROM `option` WHERE `key` = ? OR `key` LIKE ?", filter, filter+".%")

	return
}
