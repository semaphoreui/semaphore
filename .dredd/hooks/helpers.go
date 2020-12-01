package main

import (
	"encoding/json"
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/snikch/goodman/transaction"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// Test Runner User
func addTestRunnerUser() {
	uid := getUUID()
	testRunnerUser = &models.User{
		Username: "ITU-" + uid,
		Name:     "ITU-" + uid,
		Email:    uid + "@semaphore.test",
		Created:  db.GetParsedTime(time.Now()),
		Admin:    true,
	}

	dbConnect()
	defer db.Sql.Db.Close()
	if err := db.Sql.Insert(testRunnerUser); err != nil {
		panic(err)
	}
	addToken(adminToken, testRunnerUser.ID)
}

func removeTestRunnerUser(transactions []*transaction.Transaction) {
	dbConnect()
	defer db.Sql.Db.Close()
	deleteToken(adminToken, testRunnerUser.ID)
	deleteObject(testRunnerUser)
}

// Parameter Substitution
func setupObjectsAndPaths(t *transaction.Transaction) {
	alterRequestBody(t)
	alterRequestPath(t)
}

// Object Lifecycle
func addUserProjectRelation(pid int, user int) {
	_, err := db.Sql.Exec("insert into project__user set project_id=?, user_id=?, `admin`=1", pid, user)
	if err != nil {
		fmt.Println(err)
	}
}
func deleteUserProjectRelation(pid int, user int) {
	_, err := db.Sql.Exec("delete from project__user where project_id=? and user_id=?", strconv.Itoa(pid), strconv.Itoa(user))
	if err != nil {
		fmt.Println(err)
	}
}

func addAccessKey(pid *int) *models.AccessKey {
	uid := getUUID()
	secret := "5up3r53cr3t"
	key := models.AccessKey{
		Name:      "ITK-" + uid,
		Type:      "ssh",
		Secret:	   &secret,
		ProjectID: pid,
	}
	if err := db.Sql.Insert(&key); err != nil {
		fmt.Println(err)
	}
	return &key
}

func addProject() *models.Project {
	uid := getUUID()
	project := models.Project{
		Name:    "ITP-" + uid,
		Created: time.Now(),
	}
	if err := db.Sql.Insert(&project); err != nil {
		fmt.Println(err)
	}
	return &project
}

func addUser() *models.User {
	uid := getUUID()
	user := models.User{
		Created:  time.Now(),
		Username: "ITU-" + uid,
		Email:    "test@semaphore." + uid,
	}
	if err := db.Sql.Insert(&user); err != nil {
		fmt.Println(err)
	}
	return &user
}

func addTask() *models.Task {
	t := models.Task{
		TemplateID: int(templateID),
		Status: "testing",
		UserID: &userPathTestUser.ID,
		Created: db.GetParsedTime(time.Now()),
	}
	if err := db.Sql.Insert(&t); err != nil {
		fmt.Println(err)
	}
	return &t
}

func deleteObject(i interface{}) {
	_, err := db.Sql.Delete(i)
	if err != nil {
		fmt.Println(err)
	}
}

// Token Handling
func addToken(tok string, user int) {
	token := models.APIToken{
		ID:      tok,
		Created: time.Now(),
		UserID:  user,
		Expired: false,
	}
	if err := db.Sql.Insert(&token); err != nil {
		fmt.Println(err)
	}
}

func deleteToken(tok string, user int) {
	token := models.APIToken{
		ID:     tok,
		UserID: user,
	}
	deleteObject(&token)
}

// HELPERS
var r *rand.Rand
var randSetup = false

func getUUID() string {
	if !randSetup {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
		randSetup = true
	}
	return randomString(8)
}
func randomString(strlen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := ""
	for i := 0; i < strlen; i++ {
		index := r.Intn(len(chars))
		result += chars[index : index+1]
	}
	return result
}

func loadConfig() {
	cwd, _ := os.Getwd()
	file, _ := os.Open(cwd + "/.dredd/config.json")
	if err := json.NewDecoder(file).Decode(&util.Config); err != nil {
		fmt.Println("Could not decode configuration!")
		panic(err)
	}
}

func dbConnect() {
	if err := db.Connect(); err != nil {
		panic(err)
	}
	db.SetupDBLink()
}

func stringInSlice(a string, list []string) (int, bool) {
	for k, b := range list {
		if b == a {
			return k, true
		}
	}
	return 0, false
}

func printError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
