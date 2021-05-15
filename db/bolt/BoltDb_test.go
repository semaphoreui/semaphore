package bolt

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"
)

type test1 struct {
	ID int `db:"ID"`
	FistName string `db:"first_name" json:"firstName"`
	LastName string `db:"last_name" json:"lastName"`
	Password string `db:"-" json:"password"`
	PasswordRepeat string `db:"-" json:"passwordRepeat"`
	PasswordHash string `db:"password" json:"-"`
	Removed bool `db:"removed"`
}

func createBoltDb() BoltDb {
	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	fn := "/tmp/test_semaphore_db_" + strconv.Itoa(r.Int())
	return BoltDb{
		Filename: fn,
	}
}

func createStore() db.Store {
	store := createBoltDb()
	return &store
}

func TestDeleteObjectSoft(t *testing.T) {
	store := createBoltDb()
	err := store.Connect()

	if err != nil {
		t.Fatal()
	}

	obj := test1{
		FistName: "Denis",
		LastName: "Gukov",
	}

	props := db.ObjectProperties{
		IsGlobal: true,
		TableName: "test1",
	}

	newObj, err := store.createObject(0, props, obj)

	if err != nil {
		t.Fatal()
	}

	objID := intObjectID(newObj.(test1).ID)

	err = store.deleteObjectSoft(0, props, objID)

	if err != nil {
		t.Fatal()
	}

	var found test1
	err = store.getObject(0, props, objID, &found)

	if err != nil {
		t.Fatal()
	}

	if found.ID != int(objID) ||
		found.Removed != true ||
		found.Password != obj.Password ||
		found.LastName != obj.LastName {

		t.Fatal()
	}
}

func TestMarshalObject_UserWithPwd(t *testing.T) {
	user := db.UserWithPwd{
		Pwd: "123456",
		User: db.User{
			Username: "fiftin",
			Password: "345345234523452345234",
		},
	}

	bytes, err := marshalObject(user)

	if err != nil {
		t.Fatal(fmt.Errorf("function returns error: " + err.Error()))
	}

	str := string(bytes)

	if str != `{"id":0,"created":"0001-01-01T00:00:00Z","username":"fiftin","name":"","email":"","password":"345345234523452345234","admin":false,"external":false,"alert":false}` {
		t.Fatal(fmt.Errorf("incorrect marshalling result"))
	}

	fmt.Println(str)
}

func TestMarshalObject(t *testing.T) {
	test1 := test1{
		FistName: "Denis",
		LastName: "Gukov",
		Password: "1234556",
		PasswordRepeat: "123456",
		PasswordHash: "9347502348723",
	}

	bytes, err := marshalObject(test1)

	if err != nil {
		t.Fatal(fmt.Errorf("function returns error: " + err.Error()))
	}

	str := string(bytes)
	if str != `{"ID":0,"first_name":"Denis","last_name":"Gukov","password":"9347502348723","removed":false}` {
		t.Fatal(fmt.Errorf("incorrect marshalling result"))
	}

	fmt.Println(str)
}

func TestUnmarshalObject(t *testing.T) {
	test1 := test1{}
	data := `{
	"first_name": "Denis", 
	"last_name": "Gukov",
	"password": "9347502348723"
}`
	err := unmarshalObject([]byte(data), &test1)
	if err != nil {
		t.Fatal(fmt.Errorf("function returns error: " + err.Error()))
	}
	if test1.FistName != "Denis" ||
		test1.LastName != "Gukov" ||
		test1.Password != "" ||
		test1.PasswordRepeat != "" ||
		test1.PasswordHash != "9347502348723" {
		t.Fatal(fmt.Errorf("object unmarshalled incorrectly"))
	}
}

func TestSortObjects(t *testing.T) {
	objects := []db.Inventory{
		{
			ID: 1,
			Name: "x",
		},
		{
			ID: 2,
			Name: "a",
		},
		{
			ID: 3,
			Name: "d",
		},
		{
			ID: 4,
			Name: "b",
		},
		{
			ID: 5,
			Name: "r",
		},
	}

	err := sortObjects(&objects, "name", false)
	if err != nil {
		t.Fatal(err)
	}

	expected := objects[0].Name == "a" &&
		objects[1].Name == "b" &&
		objects[2].Name == "d" &&
		objects[3].Name == "r" &&
		objects[4].Name == "x"


	if !expected {
		t.Fatal(fmt.Errorf("objects not sorted"))
	}
}

func TestGetFieldNameByTag(t *testing.T) {
	f, err := getFieldNameByTag(reflect.TypeOf(test1{}), "db", "first_name")
	if err != nil {
		t.Fatal()
	}
	if f != "FistName" {
		t.Fatal()
	}
}
