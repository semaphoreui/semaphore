package bolt

import (
	"encoding/json"
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"go.etcd.io/bbolt"
	"reflect"
	"sort"
	"strconv"
)

type BoltDb struct {
	db *bbolt.DB
}

func makeBucketId(obj db.ObjectProperties, ids ...int) []byte {
	n := len(ids)

	id := obj.TableName
	for i := 0; i < n; i++ {
		id += fmt.Sprintf("_%010d", ids[i])
	}

	return []byte(id)
}

func (d *BoltDb) Migrate() error {
	return nil
}

func (d *BoltDb) Connect() error {
	config, err := util.Config.GetDBConfig()
	if err != nil {
		return err
	}
	d.db, err = bbolt.Open(config.Hostname, 0666, nil)
	if err != nil {
		return err
	}
	return nil
}

func (d *BoltDb) Close() error {
	return d.db.Close()
}

func (d *BoltDb) getObject(projectID int, props db.ObjectProperties, objectID int, object interface{}) (err error) {
	err = d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(makeBucketId(props, projectID))
		if b == nil {
			return db.ErrNotFound
		}

		id := []byte(strconv.Itoa(objectID))
		str := b.Get(id)
		if str == nil {
			return db.ErrNotFound
		}

		return json.Unmarshal(str, &object)
	})

	return
}

func getFieldNameByTag(t reflect.Type, tag string, value string) (string, error) {
	n := t.NumField()
	for i := 0; i < n; i++ {
		if t.Field(i).Tag.Get(tag) == value {
			return t.Field(i).Name, nil
		}
	}
	return "", fmt.Errorf("")
}

func sortObjects(objects interface{}, sortBy string, sortInverted bool) error {
	objectsValue := reflect.ValueOf(objects).Elem()
	objType := objectsValue.Type().Elem()
	fieldName, err := getFieldNameByTag(objType, "db", sortBy)
	if err != nil {
		return err
	}

	sort.SliceStable(objectsValue.Interface(), func (i, j int) bool {
		fieldI := objectsValue.Index(i).FieldByName(fieldName)
		fieldJ := objectsValue.Index(j).FieldByName(fieldName)
		switch fieldJ.Kind() {
		case reflect.Int:
		case reflect.Int8:
		case reflect.Int16:
		case reflect.Int32:
		case reflect.Int64:
		case reflect.Uint:
		case reflect.Uint8:
		case reflect.Uint16:
		case reflect.Uint32:
		case reflect.Uint64:
			return fieldI.Int() < fieldJ.Int()
		case reflect.Float32:
		case reflect.Float64:
			return fieldI.Float() < fieldJ.Float()
		case reflect.String:
			return fieldI.String() < fieldJ.String()
		}
		return false
	})

	return nil
}

func (d *BoltDb) getObjects(projectID int, props db.ObjectProperties, params db.RetrieveQueryParams, objects interface{}) (err error) {
	objectsValue := reflect.ValueOf(objects).Elem()
	objType := objectsValue.Type().Elem()

	// Read elements from database
	err = d.db.View(func(tx *bbolt.Tx) error {

		b := tx.Bucket(makeBucketId(props, projectID))
		c := b.Cursor()
		i := 0 // current item index
		n := 0 // number of added items

		for k, v := c.First(); k != nil; k, v = c.Next() {
			if i < params.Offset {
				continue
			}

			obj := reflect.New(objType).Elem()
			err2 := json.Unmarshal(v, &obj)
			if err2 == nil {
				return err2
			}

			objectsValue.Set(reflect.Append(objectsValue, obj))

			n++

			if n > params.Count {
				break
			}
		}

		return nil
	})

	if err != nil {
		return
	}


	// Sort elements
	err = sortObjects(objects, params.SortBy, params.SortInverted)

	return
}


func (d *BoltDb) isObjectInUse(projectID int, props db.ObjectProperties, objectID int) (inUse bool, err error) {
	err = d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(makeBucketId(props, projectID))
		inUse = b != nil && b.Get([]byte(strconv.Itoa(objectID))) != nil
		return nil
	})

	return
}

func (d *BoltDb) deleteObject(projectID int, props db.ObjectProperties, objectID int) error {
	inUse, err := d.isObjectInUse(projectID, props, objectID)

	if err != nil {
		return err
	}

	if inUse {
		return db.ErrInvalidOperation
	}

	return d.db.Update(func (tx *bbolt.Tx) error {
		b := tx.Bucket(makeBucketId(db.InventoryObject, projectID))
		if b == nil {
			return db.ErrNotFound
		}
		return b.Delete([]byte(strconv.Itoa(objectID)))
	})
}

func (d *BoltDb) deleteObjectSoft(projectID int, props db.ObjectProperties, objectID int) error {
	return d.deleteObject(projectID, props, objectID)
}
