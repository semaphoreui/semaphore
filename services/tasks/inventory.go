package tasks

import (
	"github.com/ansible-semaphore/semaphore/db"
	"io/ioutil"
	"strconv"

	"github.com/ansible-semaphore/semaphore/util"
)

func (t *AnsibleJobRunner) installInventory() (err error) {
	if t.inventory.SSHKeyID != nil {
		err = t.inventory.SSHKey.Install(db.AccessKeyRoleAnsibleUser)
		if err != nil {
			return
		}
	}

	if t.inventory.BecomeKeyID != nil {
		err = t.inventory.BecomeKey.Install(db.AccessKeyRoleAnsibleBecomeUser)
		if err != nil {
			return
		}
	}

	if t.inventory.Type == db.InventoryStatic || t.inventory.Type == db.InventoryStaticYaml {
		err = t.installStaticInventory()
	}

	return
}

func (t *AnsibleJobRunner) installStaticInventory() error {
	t.Log("installing static inventory")

	path := util.Config.TmpPath + "/inventory_" + strconv.Itoa(t.task.ID)
	if t.inventory.Type == db.InventoryStaticYaml {
		path += ".yml"
	}

	// create inventory file
	return ioutil.WriteFile(path, []byte(t.inventory.Inventory), 0664)
}
