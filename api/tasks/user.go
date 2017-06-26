package tasks

import (
	"io/ioutil"
	"os"
	"strconv"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/util"
)

func (t *task) InstallUserVarFile() error {
	var user db.User

	if t.template.UserVars {
		if err := t.fetch("User not found", &user, "select * from user where id=?", t.task.UserID); err != nil {
			return err
		}
		err := ioutil.WriteFile(util.Config.TmpPath+"/user_vars_"+strconv.Itoa(t.task.ID), []byte(user.ExtraVars), 0644)
		if err != nil {
			t.log("Error writing user vars file " + util.Config.TmpPath + "/user_vars_" + strconv.Itoa(t.task.ID) + ": " + err.Error())
			return err
		}
		t.log("Created user vars file " + util.Config.TmpPath + "/user_vars_" + strconv.Itoa(t.task.ID))
	}
	return nil
}

func (t *task) RemoveUserVarFile() error {
	if t.template.UserVault == true {
		err := os.Remove(util.Config.TmpPath + "/user_vars_" + strconv.Itoa(t.task.ID))
		if err != nil {
			t.log("Error deleting user vars file " + util.Config.TmpPath + "/user_vars_" + strconv.Itoa(t.task.ID) + ": " + err.Error())
			return err
		}
		t.log("Deleted user vars file " + util.Config.TmpPath + "/user_vars_" + strconv.Itoa(t.task.ID))
	}
	return nil
}
