package tasks

import (
	"github.com/ansible-semaphore/semaphore/db"
	"io/ioutil"
	"strconv"

	"github.com/ansible-semaphore/semaphore/util"
)

func (t *TaskRunner) installInventory() (err error) {
	if t.inventory.SSHKeyID != nil {
		err = t.inventory.SSHKey.Install(db.AccessKeyUsageAnsibleUser)
		if err != nil {
			return
		}
	}

	if t.inventory.BecomeKeyID != nil {
		err = t.inventory.BecomeKey.Install(db.AccessKeyUsageAnsibleBecomeUser)
		if err != nil {
			return
		}
	}

	if t.inventory.Type == db.InventoryStatic {
		err = t.installStaticInventory()
	}

	return
}

func (t *TaskRunner) installStaticInventory() error {
	t.Log("installing static inventory")

	// create inventory file
	return ioutil.WriteFile(util.Config.TmpPath+"/inventory_"+strconv.Itoa(t.task.ID), []byte(t.inventory.Inventory), 0664)
}
