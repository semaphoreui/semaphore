package tasks

import (
	"github.com/ansible-semaphore/semaphore/db"
	"io/ioutil"
	"strconv"

	"github.com/ansible-semaphore/semaphore/util"
)

func (t *LocalJob) installInventory() (err error) {
	if t.Inventory.SSHKeyID != nil {
		err = t.Inventory.SSHKey.Install(db.AccessKeyRoleAnsibleUser)
		if err != nil {
			return
		}
	}

	if t.Inventory.BecomeKeyID != nil {
		err = t.Inventory.BecomeKey.Install(db.AccessKeyRoleAnsibleBecomeUser)
		if err != nil {
			return
		}
	}

	if t.Inventory.Type == db.InventoryStatic || t.Inventory.Type == db.InventoryStaticYaml {
		err = t.installStaticInventory()
	}

	return
}

func (t *LocalJob) installStaticInventory() error {
	t.Log("installing static inventory")

	path := util.Config.TmpPath + "/inventory_" + strconv.Itoa(t.Task.ID)
	if t.Inventory.Type == db.InventoryStaticYaml {
		path += ".yml"
	}

	// create inventory file
	return ioutil.WriteFile(path, []byte(t.Inventory.Inventory), 0664)
}
