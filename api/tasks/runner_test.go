package tasks

import (
	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

func TestTaskGetPlaybookArgs(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	inventoryID := 1

	tsk := task{
		task: db.Task{},
		inventory: db.Inventory{
			SSHKeyID: &inventoryID,
			SSHKey: db.AccessKey{
				ID: 12345,
				Type: db.AccessKeySSH,
			},
		},
		template: db.Template{
			Playbook: "test.yml",
		},
	}

	args, err := tsk.getPlaybookArgs()

	if err != nil {
		t.Fatal(err)
	}

	res := strings.Join(args, " ")
	if res != "-i /tmp/inventory_0 --private-key=/tmp/access_key_12345 --extra-vars {} test.yml" {
		t.Fatal("incorrect result")
	}
}

func TestTaskGetPlaybookArgs2(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	inventoryID := 1

	tsk := task{
		task: db.Task{},
		inventory: db.Inventory{
			SSHKeyID: &inventoryID,
			SSHKey: db.AccessKey{
				ID: 12345,
				Type: db.AccessKeyLoginPassword,
				LoginPassword: db.LoginPassword{
					Password: "123456",
					Login: "root",
				},
			},
		},
		template: db.Template{
			Playbook: "test.yml",
		},
	}

	args, err := tsk.getPlaybookArgs()

	if err != nil {
		t.Fatal(err)
	}

	res := strings.Join(args, " ")
	if res != "-i /tmp/inventory_0 --extra-vars=@/tmp/access_key_12345 --extra-vars {} test.yml" {
		t.Fatal("incorrect result")
	}
}

func TestTaskGetPlaybookArgs3(t *testing.T) {
	util.Config = &util.ConfigType{
		TmpPath: "/tmp",
	}

	inventoryID := 1

	tsk := task{
		task: db.Task{},
		inventory: db.Inventory{
			BecomeKeyID: &inventoryID,
			BecomeKey: db.AccessKey{
				ID: 12345,
				Type: db.AccessKeyLoginPassword,
				LoginPassword: db.LoginPassword{
					Password: "123456",
					Login: "root",
				},
			},
		},
		template: db.Template{
			Playbook: "test.yml",
		},
	}

	args, err := tsk.getPlaybookArgs()

	if err != nil {
		t.Fatal(err)
	}

	res := strings.Join(args, " ")
	if res != "-i /tmp/inventory_0 --extra-vars=@/tmp/access_key_12345 --extra-vars {} test.yml" {
		t.Fatal("incorrect result")
	}
}


func TestCheckTmpDir(t *testing.T) {
	//It should be able to create a random dir in /tmp
	dirName := os.TempDir()+ "/" + randString(rand.Intn(10 - 4) + 4)
	err := checkTmpDir(dirName)
	if err != nil {
		t.Fatal(err)
	}

	//checking again for this directory should return no error, as it exists
	err = checkTmpDir(dirName)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Chmod(dirName,os.FileMode(int(0550)))
	if err != nil {
		t.Fatal(err)
	}

	//nolint: vetshadow
	if stat, err := os.Stat(dirName); err != nil {
		t.Fatal(err)
	} else if stat.Mode() != os.FileMode(int(0550)) {
		// File System is not support 0550 mode, skip this test
		return
	}

	err = checkTmpDir(dirName+"/noway")
	if err == nil {
		t.Fatal("You should not be able to write in this folder, causing an error")
	}
	err = os.Remove(dirName)
	if err != nil {
		t.Log(err)
	}
}


//HELPERS

//https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
var src = rand.NewSource(time.Now().UnixNano())
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)
func randString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}