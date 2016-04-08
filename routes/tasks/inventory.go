package tasks

import (
	"io/ioutil"
	"strconv"

	"github.com/ansible-semaphore/semaphore/util"
)

func (t *task) installInventory() error {
	if t.inventory.SshKeyID != nil {
		// write inventory key
		err := t.installKey(t.inventory.SshKey)
		if err != nil {
			return err
		}
	}

	switch t.inventory.Type {
	case "static":
		return t.installStaticInventory()
	}

	return nil
}

func (t *task) installStaticInventory() error {
	t.log("installing static inventory")

	// create inventory file
	return ioutil.WriteFile(util.Config.TmpPath+"/inventory_"+strconv.Itoa(t.task.ID), []byte(t.inventory.Inventory), 0664)
}
