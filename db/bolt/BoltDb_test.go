package bolt

import (
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"reflect"
	"testing"
)

type test1 struct {
	ID             int    `db:"ID"`
	FirstName      string `db:"first_name" json:"firstName"`
	LastName       string `db:"last_name" json:"lastName"`
	Password       string `db:"-" json:"password"`
	PasswordRepeat string `db:"-" json:"passwordRepeat"`
	PasswordHash   string `db:"password" json:"-"`
	Removed        bool   `db:"removed"`
}

var inventoryID = 10
var environmentID = 10

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
		FirstName:      "Denis",
		LastName:       "Gukov",
		Password:       "1234556",
		PasswordRepeat: "123456",
		PasswordHash:   "9347502348723",
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
	if test1.FirstName != "Denis" ||
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
			ID:   1,
			Name: "x",
		},
		{
			ID:   2,
			Name: "a",
		},
		{
			ID:   3,
			Name: "d",
		},
		{
			ID:   4,
			Name: "b",
		},
		{
			ID:   5,
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
	f, err := getFieldNameByTagSuffix(reflect.TypeOf(test1{}), "db", "first_name")
	if err != nil {
		t.Fatal(err.Error())
	}

	if f != "FirstName" {
		t.Fatal()
	}
}

func TestGetFieldNameByTag2(t *testing.T) {
	f, err := getFieldNameByTagSuffix(reflect.TypeOf(db.UserWithPwd{}), "db", "id")
	if err != nil {
		t.Fatal(err.Error())
	}
	if f != "ID" {
		t.Fatal()
	}
}

func TestIsObjectInUse(t *testing.T) {
	store := CreateTestStore()

	proj, err := store.CreateProject(db.Project{
		Name: "test",
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = store.CreateTemplate(db.Template{
		Name:          "Test",
		Playbook:      "test.yml",
		ProjectID:     proj.ID,
		InventoryID:   &inventoryID,
		EnvironmentID: &environmentID,
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	isUse, err := store.isObjectInUse(proj.ID, db.InventoryProps, intObjectID(10), db.TemplateProps)

	if err != nil {
		t.Fatal(err.Error())
	}

	if !isUse {
		t.Fatal()
	}

}

func TestIsObjectInUse_Environment(t *testing.T) {
	store := CreateTestStore()

	proj, err := store.CreateProject(db.Project{
		Name: "test",
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = store.CreateTemplate(db.Template{
		Name:          "Test",
		Playbook:      "test.yml",
		ProjectID:     proj.ID,
		InventoryID:   &inventoryID,
		EnvironmentID: &environmentID,
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	isUse, err := store.isObjectInUse(proj.ID, db.EnvironmentProps, intObjectID(10), db.TemplateProps)

	if err != nil {
		t.Fatal(err.Error())
	}

	if !isUse {
		t.Fatal()
	}

}

func TestIsObjectInUse_EnvironmentNil(t *testing.T) {
	store := CreateTestStore()

	proj, err := store.CreateProject(db.Project{
		Name: "test",
	})

	if err != nil {
		t.Fatal(err)
	}

	_, err = store.CreateTemplate(db.Template{
		Name:          "Test",
		Playbook:      "test.yml",
		ProjectID:     proj.ID,
		InventoryID:   &inventoryID,
		EnvironmentID: nil,
	})

	if err != nil {
		t.Fatal(err)
	}

	isUse, err := store.isObjectInUse(proj.ID, db.EnvironmentProps, intObjectID(10), db.TemplateProps)

	if err != nil {
		t.Fatal(err)
	}

	if isUse {
		t.Fatal()
	}
}

func TestBoltDb_CreateAPIToken(t *testing.T) {
	store := CreateTestStore()

	user, err := store.CreateUser(db.UserWithPwd{
		Pwd: "3412341234123",
		User: db.User{
			Username: "test",
			Name:     "Test",
			Email:    "test@example.com",
			Admin:    true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	token, err := store.CreateAPIToken(db.APIToken{
		ID:     "f349gyhgqirgysfgsfg34973dsfad",
		UserID: user.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	token2, err := store.GetAPIToken(token.ID)
	if err != nil {
		t.Fatal(err)
	}

	if token2.ID != token.ID {
		t.Fatal()
	}

	tokens, err := store.GetAPITokens(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(tokens) != 1 {
		t.Fatal()
	}

	if tokens[0].ID != token.ID {
		t.Fatal()
	}

	err = store.ExpireAPIToken(user.ID, token.ID)
	if err != nil {
		t.Fatal(err)
	}

	token2, err = store.GetAPIToken(token.ID)
	if err != nil {
		t.Fatal(err)
	}

	if !token2.Expired {
		t.Fatal()
	}

	err = store.DeleteAPIToken(user.ID, token.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.GetAPIToken(token.ID)
	if err == nil {
		t.Fatal("Token not deleted")
	}
}

func TestBoltDb_GetRepositoryRefs(t *testing.T) {
	store := CreateTestStore()

	repo1, err := store.CreateRepository(db.Repository{
		Name:      "repo1",
		GitURL:    "git@example.com/repo1",
		GitBranch: "master",
		ProjectID: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.CreateTemplate(db.Template{
		Type:          db.TemplateBuild,
		Name:          "tpl1",
		Playbook:      "build.yml",
		RepositoryID:  repo1.ID,
		ProjectID:     1,
		InventoryID:   &inventoryID,
		EnvironmentID: &environmentID,
	})
	if err != nil {
		t.Fatal(err)
	}

	tpl2, err := store.CreateTemplate(db.Template{
		Type:          db.TemplateBuild,
		Name:          "tpl12",
		Playbook:      "build.yml",
		ProjectID:     1,
		InventoryID:   &inventoryID,
		EnvironmentID: &environmentID,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.CreateSchedule(db.Schedule{
		CronFormat:   "* * * * *",
		TemplateID:   tpl2.ID,
		ProjectID:    1,
		RepositoryID: &repo1.ID,
	})

	if err != nil {
		t.Fatal(err)
	}

	refs, err := store.GetRepositoryRefs(1, repo1.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(refs.Templates) != 2 {
		t.Fatal()
	}
}
