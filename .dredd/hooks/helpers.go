package main

import (
	"encoding/json"
	"fmt"
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/db/sql"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/snikch/goodman/transaction"
	"math/rand"
	"os"
	"strconv"
	"time"
)

var tablesShouldBeTruncated = [...]string {
	"access_key",
	"event",
	"user__token",
	"project",
	"task__output",
	"task",
	"session",
	"project__environment",
	"project__inventory",
	"project__repository",
	"project__template",
	"project__schedule",
	"project__user",
	"user",
}
// Test Runner User
func addTestRunnerUser() {
	uid := getUUID()
	testRunnerUser = &db.User{
		Username: "ITU-" + uid,
		Name:     "ITU-" + uid,
		Email:    uid + "@semaphore.test",
		Created:  db.GetParsedTime(time.Now()),
		Admin:    true,
	}

	dbConnect()
	defer store.Sql().Db.Close()

	truncateAll()

	if err := store.Sql().Insert(testRunnerUser); err != nil {
		panic(err)
	}
	addToken(adminToken, testRunnerUser.ID)
}

func truncateAll() {
	tx, err := store.Sql().Begin()
	if err != nil {
		panic(err)
	}

	_, err = tx.Exec("SET FOREIGN_KEY_CHECKS = 0")
	if err == nil {
		for _, tableName := range tablesShouldBeTruncated {
			tx.Exec("TRUNCATE TABLE " + tableName)
		}
		tx.Exec("SET FOREIGN_KEY_CHECKS = 1")
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}
}

func removeTestRunnerUser(transactions []*transaction.Transaction) {
	dbConnect()
	defer store.Sql().Db.Close()
	deleteToken(adminToken, testRunnerUser.ID)
	deleteObject(testRunnerUser)
}

// Parameter Substitution
func setupObjectsAndPaths(t *transaction.Transaction) {
	alterRequestPath(t)
	alterRequestBody(t)
}

// Object Lifecycle
func addUserProjectRelation(pid int, user int) {
	_, err := store.Sql().Exec("insert into project__user (project_id, user_id, `admin`) values (?, ?, 1)", pid, user)
	if err != nil {
		fmt.Println(err)
	}
}

func deleteUserProjectRelation(pid int, user int) {
	_, err := store.Sql().Exec("delete from project__user where project_id=? and user_id=?", strconv.Itoa(pid), strconv.Itoa(user))
	if err != nil {
		fmt.Println(err)
	}
}

func addAccessKey(pid *int) *db.AccessKey {
	uid := getUUID()
	secret := "5up3r53cr3t\n"
	key := db.AccessKey{
		Name:      "ITK-" + uid,
		Type:      "ssh",
		Secret:	   &secret,
		ProjectID: pid,
	}
	if err := store.Sql().Insert(&key); err != nil {
		fmt.Println(err)
	}
	return &key
}

func addProject() *db.Project {
	uid := getUUID()
	project := db.Project{
		Name:    "ITP-" + uid,
		Created: time.Now(),
	}
	if err := store.Sql().Insert(&project); err != nil {
		fmt.Println(err)
	}
	return &project
}

func addUser() *db.User {
	uid := getUUID()
	user := db.User{
		Created:  time.Now(),
		Username: "ITU-" + uid,
		Email:    "test@semaphore." + uid,
	}
	if err := store.Sql().Insert(&user); err != nil {
		fmt.Println(err)
	}
	return &user
}

func addSchedule() *db.Schedule {
	schedule, err := store.CreateSchedule(db.Schedule{
		TemplateID: int(templateID),
		CronFormat: "* * * 1 *",
		ProjectID: userProject.ID,
	})

	if err != nil {
		fmt.Println(err)
	}

	return &schedule
}


func addTask() *db.Task {
	t := db.Task{
		TemplateID: int(templateID),
		Status: "testing",
		UserID: &userPathTestUser.ID,
		Created: db.GetParsedTime(time.Now()),
	}
	if err := store.Sql().Insert(&t); err != nil {
		fmt.Println(err)
	}
	return &t
}

func deleteObject(i interface{}) {
	_, err := store.Sql().Delete(i)
	if err != nil {
		fmt.Println(err)
	}
}

// Token Handling
func addToken(tok string, user int) {
	token := db.APIToken{
		ID:      tok,
		Created: time.Now(),
		UserID:  user,
		Expired: false,
	}
	if err := store.Sql().Insert(&token); err != nil {
		fmt.Println(err)
	}
}

func deleteToken(tok string, user int) {
	token := db.APIToken{
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

var store sql.SqlDb

func dbConnect() {
	store = sql.SqlDb{}

	if err := store.Connect(); err != nil {
		panic(err)
	}
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
