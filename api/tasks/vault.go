package tasks

import (
	"io/ioutil"
	"os"
	"strconv"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
)

func (t *task) InstallVaultFiles() error {
	var user db.User

	if t.template.UserVault == true {
		if err := t.fetch("User not found", &user, "select * from user where id=?", t.task.UserID); err != nil {
			return err
		}
		password, err := util.UserVaultCache.GetPassword(user.Username)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(util.Config.TmpPath+"/vault_pass_"+strconv.Itoa(t.task.ID), []byte(password), 0600)
		if err != nil {
			return err
		}
		t.log("Installed vault pass file " + util.Config.TmpPath + "/vault_pass_" + strconv.Itoa(t.task.ID))
		err = ioutil.WriteFile(util.Config.TmpPath+"/vault_"+strconv.Itoa(t.task.ID), []byte(user.Vault), 0600)
		if err != nil {
			return err
		}
		t.log("Installed vault file " + util.Config.TmpPath + "/vault_" + strconv.Itoa(t.task.ID))
	}
	return nil
}

func (t *task) RemoveVaultFiles() error {
	if t.template.UserVault == true {
		err := os.Remove(util.Config.TmpPath + "/vault_pass_" + strconv.Itoa(t.task.ID))
		if err != nil {
			t.log("Error removing vault password file: " + util.Config.TmpPath + "/vault_pass_" + strconv.Itoa(t.task.ID) + ": " + err.Error())
		}
		err = os.Remove(util.Config.TmpPath + "/vault_" + strconv.Itoa(t.task.ID))
		if err != nil {
			t.log("Error removing vault file " + util.Config.TmpPath + "/vault_" + strconv.Itoa(t.task.ID) + ": " + err.Error())
		}
		t.log("Removed vault files")
	}
	return nil
}
