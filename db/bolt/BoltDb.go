package bolt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"go.etcd.io/bbolt"
	"reflect"
	"sort"
)

const MaxID = 2147483647

type enumerable interface {
	First() (key []byte, value []byte)
	Next() (key []byte, value []byte)
}

type emptyEnumerable struct {}

func (d emptyEnumerable) First() (key []byte, value []byte) {
	return nil, nil
}

func (d emptyEnumerable) Next() (key []byte, value []byte) {
	return nil, nil
}

type BoltDb struct {
	Filename string
	db *bbolt.DB
}

type objectID interface {
	ToBytes() []byte
}

type intObjectID int
type strObjectID string

func (d intObjectID) ToBytes() []byte {
	return []byte(fmt.Sprintf("%010d", d))
}

func (d strObjectID) ToBytes() []byte {
	return []byte(d)
}

func makeBucketId(props db.ObjectProperties, ids ...int) []byte {
	n := len(ids)

	id := props.TableName

	if !props.IsGlobal {
		for i := 0; i < n; i++ {
			id += fmt.Sprintf("_%010d", ids[i])
		}
	}

	return []byte(id)
}

func (d *BoltDb) Migrate() error {
	return nil
}

func (d *BoltDb) Connect() error {
	var filename string
	if d.Filename == "" {
		config, err := util.Config.GetDBConfig()
		if err != nil {
			return err
		}
		filename = config.Hostname
	} else {
		filename = d.Filename
	}

	var err error
	d.db, err = bbolt.Open(filename, 0666, nil)
	if err != nil {
		return err
	}

	return nil
}

func (d *BoltDb) Close() error {
	return d.db.Close()
}

func (d *BoltDb) getObject(bucketID int, props db.ObjectProperties, objectID objectID, object interface{}) (err error) {
	err = d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(makeBucketId(props, bucketID))
		if b == nil {
			return db.ErrNotFound
		}

		str := b.Get(objectID.ToBytes())
		if str == nil {
			return db.ErrNotFound
		}

		return unmarshalObject(str, object)
	})

	return
}

// getFieldNameByTag tries to find field by tag name and value in provided type.
// It returns error if field not found.
func getFieldNameByTag(t reflect.Type, tagName string, tagValue string) (string, error) {
	n := t.NumField()
	for i := 0; i < n; i++ {
		if t.Field(i).Tag.Get(tagName) == tagValue {
			return t.Field(i).Name, nil
		}
	}
	for i := 0; i < n; i++ {
		if t.Field(i).Tag != "" || t.Field(i).Type.Kind() != reflect.Struct {
			continue
		}
		str, err := getFieldNameByTag(t.Field(i).Type, tagName, tagValue)
		if err == nil {
			return str, nil
		}
	}
	return "", fmt.Errorf("field not found")
}

func sortObjects(objects interface{}, sortBy string, sortInverted bool) error {
	objectsValue := reflect.ValueOf(objects).Elem()
	objType := objectsValue.Type().Elem()

	fieldName, err := getFieldNameByTag(objType, "db", sortBy)
	if err != nil {
		return err
	}

	sort.SliceStable(objectsValue.Interface(), func (i, j int) bool {
		valueI := objectsValue.Index(i).FieldByName(fieldName)
		valueJ := objectsValue.Index(j).FieldByName(fieldName)

		less := false

		switch valueI.Kind() {
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64,
			reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			less = valueI.Int() < valueJ.Int()
		case reflect.Float32:
		case reflect.Float64:
			less = valueI.Float() < valueJ.Float()
		case reflect.String:
			less = valueI.String() < valueJ.String()
		}

		if sortInverted {
			less = !less
		}

		return less
	})

	return nil
}

func createObjectType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	n := t.NumField()

	fields := make([]reflect.StructField, n)

	for i := 0; i < n; i++ {
		f := t.Field(i)
		tag := f.Tag.Get("db")
		if tag != "" {
			f.Tag = reflect.StructTag(`json:"` + tag + `"`)
		} else {
			if f.Type.Kind() == reflect.Struct {
				f.Type = createObjectType(f.Type)
			}
		}
		fields[i] = f
	}

	return reflect.StructOf(fields)
}

func unmarshalObject(data []byte, obj interface{}) error {
	newType := createObjectType(reflect.TypeOf(obj))
	ptr := reflect.New(newType).Interface()

	err := json.Unmarshal(data, ptr)
	if err != nil {
		return err
	}

	value := reflect.ValueOf(ptr).Elem()

	objValue := reflect.ValueOf(obj).Elem()

	for i := 0; i < newType.NumField(); i++ {
		objValue.Field(i).Set(value.Field(i))
	}

	return nil
}

func copyObject(obj interface{}, newType reflect.Type) interface{} {
	newValue := reflect.New(newType).Elem()

	oldValue := reflect.ValueOf(obj)

	for i := 0; i < newType.NumField(); i++ {
		var v interface{}
		if newValue.Field(i).Kind() == reflect.Struct &&
			newValue.Field(i).Type().PkgPath() == "" {
			v = copyObject(oldValue.Field(i).Interface(), newValue.Field(i).Type())
		} else {
			v = oldValue.Field(i).Interface()
		}
		newValue.Field(i).Set(reflect.ValueOf(v))
	}

	return newValue.Interface()
}

func marshalObject(obj interface{}) ([]byte, error) {
	newType := createObjectType(reflect.TypeOf(obj))
	return json.Marshal(copyObject(obj, newType))
}

func unmarshalObjects(rawData enumerable, props db.ObjectProperties, params db.RetrieveQueryParams, filter func(interface{}) bool, objects interface{}) (err error) {
	objectsValue := reflect.ValueOf(objects).Elem()
	objType := objectsValue.Type().Elem()

	objectsValue.Set(reflect.MakeSlice(objectsValue.Type(), 0, 0))

	i := 0 // offset counter
	n := 0 // number of added items

	for k, v := rawData.First(); k != nil; k, v = rawData.Next() {
		if params.Offset > 0 && i < params.Offset {
			i++
			continue
		}

		tmp := reflect.New(objType)
		ptr := tmp.Interface()
		err = unmarshalObject(v, ptr)
		obj := reflect.ValueOf(ptr).Elem().Interface()

		if err != nil {
			return
		}

		if filter != nil {
			if !filter(obj) {
				continue
			}
		}

		newObjectValues := reflect.Append(objectsValue, reflect.ValueOf(obj))
		objectsValue.Set(newObjectValues)

		n++

		if params.Count > 0 && n > params.Count {
			break
		}
	}

	sortable := false

	if params.SortBy != "" {
		for _, v := range props.SortableColumns {
			if v == params.SortBy {
				sortable = true
				break
			}
		}
	}

	if sortable {
		err = sortObjects(objects, params.SortBy, params.SortInverted)
	}


	return
}

func (d *BoltDb) getObjects(bucketID int, props db.ObjectProperties, params db.RetrieveQueryParams, filter func(interface{}) bool, objects interface{}) error {
	return d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(makeBucketId(props, bucketID))
		var c enumerable
		if b == nil {
			c = emptyEnumerable{}
		} else {
			c = b.Cursor()
		}
		return unmarshalObjects(c, props, params, filter, objects)
	})
}

func (d *BoltDb) isObjectInUse(bucketID int, props db.ObjectProperties, objID objectID, userProps db.ObjectProperties) (inUse bool, err error) {
	var templates []db.Template

	err = d.getObjects(bucketID, userProps, db.RetrieveQueryParams{}, func (tpl interface{}) bool {
		if props.ForeignColumnName == "" {
			return false
		}

		fieldName, err := getFieldNameByTag(reflect.TypeOf(tpl), "db", props.ForeignColumnName)

		if err != nil {
			return false
		}

		f := reflect.ValueOf(tpl).FieldByName(fieldName)

		if f.IsZero() {
			return false
		}

		if f.Kind() == reflect.Ptr {
			if f.IsNil() {
				return false
			}

			f = f.Elem()
		}

		var fVal objectID
		switch f.Kind() {
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64,
			reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			fVal = intObjectID(2147483647 - f.Int())
		case reflect.String:
			fVal = strObjectID(f.String())
		}

		if fVal == nil {
			return false
		}

		return bytes.Equal(fVal.ToBytes(), objID.ToBytes())
	}, &templates)

	if err != nil {
		return
	}

	inUse = len(templates) > 0

	return
}

func (d *BoltDb) deleteObject(bucketID int, props db.ObjectProperties, objectID objectID) error {
	for _, u := range []db.ObjectProperties{ db.TemplateProps, db.EnvironmentProps, db.InventoryProps, db.RepositoryProps } {
		inUse, err := d.isObjectInUse(bucketID, props, objectID, u)
		if err != nil {
			return err
		}
		if inUse {
			return db.ErrInvalidOperation
		}
	}

	return d.db.Update(func (tx *bbolt.Tx) error {
		b := tx.Bucket(makeBucketId(props, bucketID))
		if b == nil {
			return db.ErrNotFound
		}
		return b.Delete(objectID.ToBytes())
	})
}

func (d *BoltDb) deleteObjectSoft(bucketID int, props db.ObjectProperties, objectID objectID) error {
	var data map[string]interface{}

	// load data
	err := d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(makeBucketId(props, bucketID))
		if b == nil {
			return db.ErrNotFound
		}

		d := b.Get(objectID.ToBytes())

		if d == nil {
			return db.ErrNotFound
		}

		return json.Unmarshal(d, &data)
	})

	if err != nil {
		return err
	}

	// mark as removed if "removed" exists
	if _, ok := data["removed"]; !ok {
		return fmt.Errorf("removed field not exists")
	}

	data["removed"] = true

	// store data back
	res, err := json.Marshal(data)

	if err != nil {
		return err
	}

	return d.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(makeBucketId(props, bucketID))
		if b == nil {
			return db.ErrNotFound
		}

		return b.Put(objectID.ToBytes(), res)
	})
}

// updateObject updates data for object in database.
func (d *BoltDb) updateObject(bucketID int, props db.ObjectProperties, object interface{}) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(makeBucketId(props, bucketID))
		if b == nil {
			return db.ErrNotFound
		}

		idFieldName, err := getFieldNameByTag(reflect.TypeOf(object), "db", props.PrimaryColumnName)

		if err != nil {
			return err
		}

		idValue := reflect.ValueOf(object).FieldByName(idFieldName)

		var objectID objectID

		switch idValue.Kind() {
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64,
			reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			objectID = intObjectID(idValue.Int())
		case reflect.String:
			objectID = strObjectID(idValue.String())
		}

		if objectID == nil {
			return fmt.Errorf("unsupported ID type")
		}

		if b.Get(objectID.ToBytes()) == nil {
			return db.ErrNotFound
		}

		str, err := marshalObject(object)
		if err != nil {
			return err
		}

		return b.Put(objectID.ToBytes(), str)
	})
}

func (d *BoltDb) createObject(bucketID int, props db.ObjectProperties, object interface{}) (interface{}, error) {
	err := d.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(makeBucketId(props, bucketID))

		if err != nil {
			return err
		}

		objPtr := reflect.ValueOf(&object).Elem()

		tmpObj := reflect.New(objPtr.Elem().Type()).Elem()
		tmpObj.Set(objPtr.Elem())

		var objectID objectID

		if props.PrimaryColumnName != "" {
			idFieldName, err := getFieldNameByTag(reflect.TypeOf(object), "db", props.PrimaryColumnName)

			if err != nil {
				return err
			}

			idValue := tmpObj.FieldByName(idFieldName)

			switch idValue.Kind() {
			case reflect.Int,
				reflect.Int8,
				reflect.Int16,
				reflect.Int32,
				reflect.Int64,
				reflect.Uint,
				reflect.Uint8,
				reflect.Uint16,
				reflect.Uint32,
				reflect.Uint64:
				if idValue.Int() == 0 {
					id, err2 := b.NextSequence()
					if err2 != nil {
						return err2
					}
					if props.SortInverted {
						id = MaxID - id
					}
					idValue.SetInt(int64(id))
				}

				objectID = intObjectID(idValue.Int())
			case reflect.String:
				if idValue.String() == "" {
					return fmt.Errorf("object ID can not be empty string")
				}
				objectID = strObjectID(idValue.String())
			case reflect.Invalid:
				id, err2 := b.NextSequence()
				if err2 != nil {
					return err2
				}
				objectID = intObjectID(id)
			default:
				return fmt.Errorf("unsupported ID type")
			}
		} else {
			id, err2 := b.NextSequence()
			if err2 != nil {
				return err2
			}
			if props.SortInverted {
				id = MaxID - id
			}
			objectID = intObjectID(id)
		}

		if objectID == nil {
			return fmt.Errorf("object ID can not be nil")
		}

		objPtr.Set(tmpObj)
		str, err := marshalObject(object)
		if err != nil {
			return err
		}

		return b.Put(objectID.ToBytes(), str)
	})

	return object, err
}
