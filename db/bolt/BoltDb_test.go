package bolt

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	str := string(bytes)
	expected := `{"id":0,"created":"0001-01-01T00:00:00Z","username":"fiftin","name":"","email":"","password":"345345234523452345234","admin":false,"external":false,"alert":false}`
	assert.Equal(t, expected, str)

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
	require.NoError(t, err)

	str := string(bytes)
	expected := `{"ID":0,"first_name":"Denis","last_name":"Gukov","password":"9347502348723","removed":false}`
	assert.Equal(t, expected, str)

	fmt.Println(str)
}

func TestUnmarshalObject(t *testing.T) {
	test1 := test1{}
	data := `{
	"first_name": "Denis", 
	"last_name": "Gukov",
	"password": "9347502348723"
}`
	err := unmarshalObject([]byte(data), &test1, nil)
	require.NoError(t, err)

	assert.Equal(t, "Denis", test1.FirstName)
	assert.Equal(t, "Gukov", test1.LastName)
	assert.Equal(t, "", test1.Password)
	assert.Equal(t, "", test1.PasswordRepeat)
	assert.Equal(t, "9347502348723", test1.PasswordHash)
}

func TestSortObjects(t *testing.T) {
	objects := []db.Inventory{
		{ID: 1, Name: "x"},
		{ID: 2, Name: "a"},
		{ID: 3, Name: "d"},
		{ID: 4, Name: "b"},
		{ID: 5, Name: "r"},
	}

	err := sortObjects(&objects, "name", false)
	require.NoError(t, err)

	expected := []string{"a", "b", "d", "r", "x"}
	for i, obj := range objects {
		assert.Equal(t, expected[i], obj.Name)
	}
}

func TestGetFieldNameByTag(t *testing.T) {
	f, err := getFieldNameByTagSuffix(reflect.TypeOf(test1{}), "db", "first_name")
	require.NoError(t, err)
	assert.Equal(t, "FirstName", f)
}

func TestGetFieldNameByTag2(t *testing.T) {
	f, err := getFieldNameByTagSuffix(reflect.TypeOf(db.UserWithPwd{}), "db", "id")
	require.NoError(t, err)
	assert.Equal(t, "ID", f)
}

func TestIsObjectInUse(t *testing.T) {
	store := CreateTestStore()

	proj, err := store.CreateProject(db.Project{Name: "test"})
	require.NoError(t, err)

	_, err = store.CreateTemplate(db.Template{
		Name:          "Test",
		Playbook:      "test.yml",
		ProjectID:     proj.ID,
		InventoryID:   &inventoryID,
		EnvironmentID: &environmentID,
	})
	require.NoError(t, err)

	isUse, err := store.isObjectInUse(proj.ID, db.InventoryProps, intObjectID(10), db.TemplateProps)
	require.NoError(t, err)
	assert.True(t, isUse)
}

func TestIsObjectInUse_Environment(t *testing.T) {
	store := CreateTestStore()

	proj, err := store.CreateProject(db.Project{Name: "test"})
	require.NoError(t, err)

	_, err = store.CreateTemplate(db.Template{
		Name:          "Test",
		Playbook:      "test.yml",
		ProjectID:     proj.ID,
		InventoryID:   &inventoryID,
		EnvironmentID: &environmentID,
	})
	require.NoError(t, err)

	isUse, err := store.isObjectInUse(proj.ID, db.EnvironmentProps, intObjectID(10), db.TemplateProps)
	require.NoError(t, err)
	assert.True(t, isUse)
}

func TestIsObjectInUse_EnvironmentNil(t *testing.T) {
	store := CreateTestStore()

	proj, err := store.CreateProject(db.Project{Name: "test"})
	require.NoError(t, err)

	_, err = store.CreateTemplate(db.Template{
		Name:          "Test",
		Playbook:      "test.yml",
		ProjectID:     proj.ID,
		InventoryID:   &inventoryID,
		EnvironmentID: nil,
	})
	require.NoError(t, err)

	isUse, err := store.isObjectInUse(proj.ID, db.EnvironmentProps, intObjectID(10), db.TemplateProps)
	require.NoError(t, err)
	assert.False(t, isUse)
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
	require.NoError(t, err)

	token, err := store.CreateAPIToken(db.APIToken{
		ID:     "f349gyhgqirgysfgsfg34973dsfad",
		UserID: user.ID,
	})
	require.NoError(t, err)

	token2, err := store.GetAPIToken(token.ID)
	require.NoError(t, err)
	assert.Equal(t, token.ID, token2.ID)

	tokens, err := store.GetAPITokens(user.ID)
	require.NoError(t, err)
	assert.Len(t, tokens, 1)
	assert.Equal(t, token.ID, tokens[0].ID)

	err = store.ExpireAPIToken(user.ID, token.ID)
	require.NoError(t, err)

	token2, err = store.GetAPIToken(token.ID)
	require.NoError(t, err)
	assert.True(t, token2.Expired)

	err = store.DeleteAPIToken(user.ID, token.ID)
	require.NoError(t, err)

	_, err = store.GetAPIToken(token.ID)
	assert.Error(t, err)
}

func TestBoltDb_GetRepositoryRefs(t *testing.T) {
	store := CreateTestStore()

	repo1, err := store.CreateRepository(db.Repository{
		Name:      "repo1",
		GitURL:    "git@example.com/repo1",
		GitBranch: "master",
		ProjectID: 1,
	})
	require.NoError(t, err)

	_, err = store.CreateTemplate(db.Template{
		Type:          db.TemplateBuild,
		Name:          "tpl1",
		Playbook:      "build.yml",
		RepositoryID:  repo1.ID,
		ProjectID:     1,
		InventoryID:   &inventoryID,
		EnvironmentID: &environmentID,
	})
	require.NoError(t, err)

	tpl2, err := store.CreateTemplate(db.Template{
		Type:          db.TemplateBuild,
		Name:          "tpl12",
		Playbook:      "build.yml",
		ProjectID:     1,
		InventoryID:   &inventoryID,
		EnvironmentID: &environmentID,
	})
	require.NoError(t, err)

	_, err = store.CreateSchedule(db.Schedule{
		CronFormat:   "* * * * *",
		TemplateID:   tpl2.ID,
		ProjectID:    1,
		RepositoryID: &repo1.ID,
	})
	require.NoError(t, err)

	refs, err := store.GetRepositoryRefs(1, repo1.ID)
	require.NoError(t, err)
	assert.Len(t, refs.Templates, 1)
	assert.Len(t, refs.Schedules, 1)
}
