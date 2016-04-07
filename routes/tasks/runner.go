package tasks

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"

	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/routes/sockets"
	"github.com/ansible-semaphore/semaphore/util"
)

type task struct {
	task        models.Task
	template    models.Template
	sshKey      models.AccessKey
	inventory   models.Inventory
	repository  models.Repository
	environment models.Environment
}

func (t *task) log(msg string) {
	sockets.Broadcast([]byte(msg))
}

func (t *task) run() {
	pool.running = t

	defer func() {
		fmt.Println("Stopped running tasks")
		pool.running = nil
	}()

	t.log("Started: " + strconv.Itoa(t.task.ID) + "\n")

	if err := t.populateDetails(); err != nil {
		t.log("Error: " + err.Error())
		return
	}

	if err := t.installKey(t.repository.SshKey); err != nil {
		t.log("Failed installing ssh key for repository access: " + err.Error())
		return
	}

	if err := t.updateRepository(); err != nil {
		t.log("Failed updating repository: " + err.Error())
		return
	}

}

func (t *task) fetch(errMsg string, ptr interface{}, query string, args ...interface{}) error {
	err := database.Mysql.SelectOne(ptr, query, args...)
	if err == sql.ErrNoRows {
		t.log(errMsg)
		return err
	}

	if err != nil {
		panic(err)
	}

	return nil
}

func (t *task) populateDetails() error {
	// get template
	if err := t.fetch("Template not found!", &t.template, "select * from project__template where id=?", t.task.TemplateID); err != nil {
		return err
	}

	// get access key
	if err := t.fetch("Template Access Key not found!", &t.sshKey, "select * from access_key where id=?", t.template.SshKeyID); err != nil {
		return err
	}

	if t.sshKey.Type != "ssh" {
		t.log("Non ssh-type keys are currently not supported: " + t.sshKey.Type)
		return errors.New("Unsupported SSH Key")
	}

	// get inventory
	if err := t.fetch("Template Inventory not found!", &t.inventory, "select * from project__inventory where id=?", t.template.InventoryID); err != nil {
		return err
	}

	// get repository
	if err := t.fetch("Repository not found!", &t.repository, "select * from project__repository where id=?", t.template.RepositoryID); err != nil {
		return err
	}

	// get repository access key
	if err := t.fetch("Repository Access Key not found!", &t.repository.SshKey, "select * from access_key where id=?", t.repository.SshKeyID); err != nil {
		return err
	}
	if t.repository.SshKey.Type != "ssh" {
		t.log("Repository Access Key is not 'SSH': " + t.repository.SshKey.Type)
		return errors.New("Unsupported SSH Key")
	}

	// get environment
	if len(t.task.Environment) == 0 && t.template.EnvironmentID != nil {
		err := t.fetch("Environment not found", &t.environment, "select * from project__environment where id=?", *t.template.EnvironmentID)
		if err != nil {
			return err
		}
	} else if len(t.task.Environment) > 0 {
		t.environment.JSON = t.task.Environment
	}

	return nil
}

func (t *task) installKey(key models.AccessKey) error {
	t.log("installing Access key: " + key.Name)

	// create .ssh directory
	err := os.MkdirAll(util.Config.TmpPath+"/.ssh", 448)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(util.Config.TmpPath+"/.ssh/id_rsa", []byte(*key.Secret), 0600)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(util.Config.TmpPath+"/.ssh/id_rsa.pub", []byte(*key.Key), 0644)
	if err != nil {
		return err
	}

	t.log("key " + key.Name + " installed")
	return nil
}

func (t *task) updateRepository() error {
	repoName := "repository_" + strconv.Itoa(t.repository.ID)

	_, err := os.Stat(util.Config.TmpPath + "/" + repoName)
	if err != nil && os.IsNotExist(err) {
		t.log("Cloning repository")

		cmd := exec.Command("git", "clone", t.repository.GitUrl, repoName)
		cmd.Env = []string{
			"HOME=" + util.Config.TmpPath,
			"PWD=" + util.Config.TmpPath,
			"GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no -i " + util.Config.TmpPath + "/.ssh/id_rsa",
			// "GIT_FLUSH=1",
		}
		cmd.Dir = util.Config.TmpPath

		out, err := cmd.CombinedOutput()
		fmt.Println(string(out))

		return err
	} else if err != nil {
		return err
	}

	t.log("Updating repository")

	// update instead of cloning
	cmd := exec.Command("git", "pull", "origin", "master")
	cmd.Env = []string{
		"HOME=" + util.Config.TmpPath,
		"PWD=" + util.Config.TmpPath,
		"GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no -i " + util.Config.TmpPath + "/.ssh/id_rsa",
	}
	cmd.Dir = util.Config.TmpPath + "/" + repoName

	out, err := cmd.CombinedOutput()
	fmt.Println(string(out))

	return nil
}
