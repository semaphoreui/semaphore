package tasks

import (
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db_lib"
	log "github.com/sirupsen/logrus"

	"github.com/semaphoreui/semaphore/util"
)

func (t *LocalExecutor) installInventory() (err error) {
	if t.Inventory.SSHKeyID != nil {
		t.sshKeyInstallation, err = t.KeyInstaller.Install(t.Inventory.SSHKey, db.AccessKeyRoleAnsibleUser, t.Logger)
		if err != nil {
			return
		}
	}

	if t.Inventory.BecomeKeyID != nil {
		t.becomeKeyInstallation, err = t.KeyInstaller.Install(t.Inventory.BecomeKey, db.AccessKeyRoleAnsibleBecomeUser, t.Logger)
		if err != nil {
			return
		}
	}

	// The jump hosts usually need different keys than the target hosts, so the
	// whole proxy chain gets its own agent.
	if t.Inventory.Proxy != nil && t.Inventory.Proxy.Type.IsSSH() {
		keys := db_lib.ProxyChainKeys(*t.Inventory.Proxy)
		if len(keys) > 0 {
			t.proxyKeyInstallation, err = t.KeyInstaller.InstallAll(keys, db.AccessKeyRoleAnsibleUser, t.Logger)
			if err != nil {
				return
			}
		}
	}

	switch t.Inventory.Type {
	case db.InventoryFile:
		err = t.cloneInventoryRepo(t.KeyInstaller)
	case db.InventoryStatic, db.InventoryStaticYaml:
		err = t.installStaticInventory()
	}

	return
}

// getInventorySSHCommonArgs returns the ssh options ansible must use to reach
// the hosts of the inventory, currently the proxy chain of the assigned proxy.
func (t *LocalExecutor) getInventorySSHCommonArgs() string {
	if t.Inventory.Proxy == nil || t.Inventory.Proxy.Type != db.ProxySSH {
		return ""
	}

	var socket string
	if t.proxyKeyInstallation.SSHAgent != nil {
		socket = t.proxyKeyInstallation.SSHAgent.SocketFile
	}

	return fmt.Sprintf("-o %q", db_lib.ProxyCommandOption(*t.Inventory.Proxy, socket))
}

func (t *LocalExecutor) tmpInventoryFilename() string {
	if t.Inventory.Repository == nil {
		return "inventory_" + strconv.Itoa(t.Inventory.ID)
	}
	return t.Inventory.Repository.GetDirName(t.Template.ID) + "_inventory_" + strconv.Itoa(t.Inventory.ID)
}

func (t *LocalExecutor) tmpInventoryFullPath() string {
	if t.Inventory.Repository != nil && t.Inventory.Repository.GetType() == db.RepositoryLocal {
		return t.Inventory.Repository.GetGitURL(true)
	}
	pathname := path.Join(util.Config.GetProjectTmpDir(t.Template.ProjectID), t.tmpInventoryFilename())
	if t.Inventory.Type == db.InventoryStaticYaml {
		pathname += ".yml"
	}
	return pathname
}

func (t *LocalExecutor) cloneInventoryRepo(keyInstaller db_lib.AccessKeyInstaller) error {
	if t.Inventory.Repository == nil {
		return nil
	}

	if t.Inventory.Repository.GetType() == db.RepositoryLocal {
		return nil
	}

	t.Log("cloning inventory repository")

	repo := db_lib.GitRepository{
		Logger:     t.Logger,
		TmpDirName: t.tmpInventoryFilename(),
		Repository: *t.Inventory.Repository,
		Client:     db_lib.CreateDefaultGitClient(keyInstaller),
	}

	// Parallel tasks of the same template share this inventory directory —
	// serialize the pull/remove/clone sequence, same as the main repo.
	unlock := t.RepoLock.Lock(repo.GetFullPath())
	defer unlock()

	// Try to pull the repo before trying to clone it
	if repo.CanBePulled() {
		err := repo.Pull()
		if err == nil {
			return nil
		}
	}

	err := os.RemoveAll(repo.GetFullPath())
	if err != nil {
		return err
	}

	return repo.Clone()
}

func (t *LocalExecutor) installStaticInventory() error {
	t.Log("installing static inventory")

	fullPath := t.tmpInventoryFullPath()

	// create inventory file
	return os.WriteFile(fullPath, []byte(t.Inventory.Inventory), 0664)
}

func (t *LocalExecutor) destroyInventoryFile() {
	if !t.Inventory.Type.IsStatic() {
		return
	}

	fullPath := t.tmpInventoryFullPath()
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return
		}

		log.WithError(err).WithFields(log.Fields{
			"context": "task_running",
			"task_id": t.Task.ID,
		}).Warn("failed to remove inventory file")
	}
}

func (t *LocalExecutor) destroyKeys() {
	err := t.sshKeyInstallation.Destroy()
	if err != nil {
		t.Log("Can't destroy inventory user key, error: " + err.Error())
	}

	err = t.becomeKeyInstallation.Destroy()
	if err != nil {
		t.Log("Can't destroy inventory become user key, error: " + err.Error())
	}

	err = t.proxyKeyInstallation.Destroy()
	if err != nil {
		t.Log("Can't destroy inventory proxy key, error: " + err.Error())
	}

	for _, vault := range t.vaultFileInstallations {
		err = vault.Destroy()
		if err != nil {
			t.Log("Can't destroy inventory vault password file, error: " + err.Error())
		}
	}
}
