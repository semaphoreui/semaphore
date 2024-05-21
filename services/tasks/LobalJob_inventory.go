package tasks

import (
	"github.com/ansible-semaphore/semaphore/db"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"

	"github.com/ansible-semaphore/semaphore/util"
)

func (t *LocalJob) installInventory() (err error) {
	if t.Inventory.SSHKeyID != nil {
		t.sshKeyInstallation, err = t.Inventory.SSHKey.Install(db.AccessKeyRoleAnsibleUser, t.Logger)
		if err != nil {
			return
		}
	}

	if t.Inventory.BecomeKeyID != nil {
		t.becomeKeyInstallation, err = t.Inventory.BecomeKey.Install(db.AccessKeyRoleAnsibleBecomeUser, t.Logger)
		if err != nil {
			return
		}
	}

	if t.Inventory.Type == db.InventoryStatic || t.Inventory.Type == db.InventoryStaticYaml {
		err = t.installStaticInventory()
	}

	return
}

func (t *LocalJob) tmpInventoryFilename() string {
	path := util.Config.TmpPath + "/inventory_" + strconv.Itoa(t.Task.ID)
	if t.Inventory.Type == db.InventoryStaticYaml {
		path += ".yml"
	}
	return path
}

func (t *LocalJob) installStaticInventory() error {
	t.Log("installing static inventory")

	path := t.tmpInventoryFilename()

	// create inventory file
	return os.WriteFile(path, []byte(t.Inventory.Inventory), 0664)
}
func (t *LocalJob) destroyInventoryFile() {
	path := t.tmpInventoryFilename()
	if err := os.Remove(path); err != nil {
		log.Error(err)
	}
}

func (t *LocalJob) destroyKeys() {
	err := t.sshKeyInstallation.Destroy()
	if err != nil {
		t.Log("Can't destroy inventory user key, error: " + err.Error())
	}

	err = t.becomeKeyInstallation.Destroy()
	if err != nil {
		t.Log("Can't destroy inventory become user key, error: " + err.Error())
	}

	err = t.vaultFileInstallation.Destroy()
	if err != nil {
		t.Log("Can't destroy inventory vault password file, error: " + err.Error())
	}
}
