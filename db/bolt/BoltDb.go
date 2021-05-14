package bolt

import (
	"encoding/json"
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"go.etcd.io/bbolt"
	"reflect"
	"sort"
)


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

func makeObjectId(ids ...int) []byte {
	n := len(ids)

	id := ""
	for i := 0; i < n; i++ {
		if id != "" {
			id += "_"
		}
		id += fmt.Sprintf("%010d", ids[i])
	}

	return []byte(id)
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
		valueI := objectsValue.Index(i).FieldByName(fieldName)
		valueJ := objectsValue.Index(j).FieldByName(fieldName)

		less := false

		switch valueI.Kind() {
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


func (d *BoltDb) isObjectInUse(bucketID int, props db.ObjectProperties, objectID objectID) (inUse bool, err error) {
	return false, nil
}

func (d *BoltDb) deleteObject(bucketID int, props db.ObjectProperties, objectID objectID) error {
	inUse, err := d.isObjectInUse(bucketID, props, objectID)

	if err != nil {
		return err
	}

	if inUse {
		return db.ErrInvalidOperation
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
	return d.deleteObject(bucketID, props, objectID)
}

// updateObject updates data for object in database.
func (d *BoltDb) updateObject(bucketID int, props db.ObjectProperties, object interface{}) error {
	return d.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(makeBucketId(props, bucketID))
		if b == nil {
			return db.ErrNotFound
		}

		idValue := reflect.ValueOf(object).FieldByName("ID")

		id := makeObjectId(int(idValue.Int()))
		if b.Get(id) == nil {
			return db.ErrNotFound
		}

		str, err := marshalObject(object)
		if err != nil {
			return err
		}

		return b.Put(id, str)
	})
}

func (d *BoltDb) createObject(bucketID int, props db.ObjectProperties, object interface{}) (interface{}, error) {
	err := d.db.Update(func(tx *bbolt.Tx) error {
		b, err2 := tx.CreateBucketIfNotExists(makeBucketId(props, bucketID))

		if err2 != nil {
			return err2
		}

		objPtr := reflect.ValueOf(&object).Elem()

		tmpObj := reflect.New(objPtr.Elem().Type()).Elem()
		tmpObj.Set(objPtr.Elem())

		idValue := tmpObj.FieldByName("ID")
		var objectID objectID
		idKind := idValue.Kind()
		switch {
		case idKind >= reflect.Int && idKind <= reflect.Uint64:
			if idValue.Int() == 0 {
				id, err2 := b.NextSequence()
				if err2 != nil {
					return err2
				}
				idValue.SetInt(int64(id))
			}
			objectID = intObjectID(idValue.Int())
		case idKind == reflect.String:
			if idValue.String() == "" {
				return fmt.Errorf("object ID can not be empty string")
			}
			objectID = strObjectID(idValue.String())
		case idKind == reflect.Invalid:
			id, err2 := b.NextSequence()
			if err2 != nil {
				return err2
			}
			objectID = intObjectID(id)
		default:
			return fmt.Errorf("unsupported ID type")
		}

		if objectID == nil {
			return fmt.Errorf("object ID can not be nil")
		}


		objPtr.Set(tmpObj)
		str, err2 := marshalObject(object)
		if err2 != nil {
			return err2
		}

		return b.Put(objectID.ToBytes(), str)
	})

	return object, err
}
