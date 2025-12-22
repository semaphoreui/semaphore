package tasks

import (
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/db_lib"
	"github.com/semaphoreui/semaphore/pkg/ssh"
	log "github.com/sirupsen/logrus"

	"github.com/semaphoreui/semaphore/util"
)

func (t *LocalJob) installInventory() (err error) {
	processKeyIDs, err := t.getProcessSSHKeyIDs()
	if err != nil {
		return err
	}

	var processKeys []db.AccessKey
	for _, keyID := range processKeyIDs {
		key, err := t.store.GetAccessKey(t.Template.ProjectID, keyID)
		if err != nil {
			return fmt.Errorf("failed to load process SSH key %d: %w", keyID, err)
		}

		if err := t.encryptionService.DeserializeSecret(&key); err != nil {
			return fmt.Errorf("failed to deserialize process SSH key %d: %w", keyID, err)
		}

		if key.Type != db.AccessKeySSH {
			return fmt.Errorf("process key %d is not an SSH key (type: %s)", keyID, key.Type)
		}

		processKeys = append(processKeys, key)
	}

	if t.Inventory.SSHKeyID != nil {
		if t.Inventory.SSHKey.Type == db.AccessKeySSH && len(processKeys) > 0 {
			// Both inventory key and process keys - combine them
			allKeys := append([]db.AccessKey{t.Inventory.SSHKey}, processKeys...)
			agent, err := ssh.StartSSHAgentWithKeys(allKeys, t.Template.ProjectID, t.Logger)
			if err != nil {
				return fmt.Errorf("failed to start SSH agent with combined keys: %w", err)
			}
			t.sshKeyInstallation.SSHAgent = &agent
			t.sshKeyInstallation.Login = t.Inventory.SSHKey.SshKey.Login
			t.Log(fmt.Sprintf("Started SSH Agent with inventory key + %d process key(s)", len(processKeys)))
		} else if len(processKeys) > 0 {
			t.sshKeyInstallation, err = t.KeyInstaller.Install(t.Inventory.SSHKey, db.AccessKeyRoleAnsibleUser, t.Logger)
			if err != nil {
				return err
			}
			agent, err := ssh.StartSSHAgentWithKeys(processKeys, t.Template.ProjectID, t.Logger)
			if err != nil {
				return fmt.Errorf("failed to start SSH agent with process keys: %w", err)
			}
			if t.sshKeyInstallation.SSHAgent != nil {
				t.sshKeyInstallation.SSHAgent.Close()
			}
			t.sshKeyInstallation.SSHAgent = &agent
			t.Log(fmt.Sprintf("Started SSH Agent with %d process key(s)", len(processKeys)))
		} else {
			t.sshKeyInstallation, err = t.KeyInstaller.Install(t.Inventory.SSHKey, db.AccessKeyRoleAnsibleUser, t.Logger)
			if err != nil {
				return err
			}
		}
	} else if len(processKeys) > 0 {
		agent, err := ssh.StartSSHAgentWithKeys(processKeys, t.Template.ProjectID, t.Logger)
		if err != nil {
			return fmt.Errorf("failed to start SSH agent with process keys: %w", err)
		}
		t.sshKeyInstallation.SSHAgent = &agent
		t.Log(fmt.Sprintf("Started SSH Agent with %d process key(s)", len(processKeys)))
	}

	if t.Inventory.BecomeKeyID != nil {
		t.becomeKeyInstallation, err = t.KeyInstaller.Install(t.Inventory.BecomeKey, db.AccessKeyRoleAnsibleBecomeUser, t.Logger)
		if err != nil {
			return
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

func (t *LocalJob) tmpInventoryFilename() string {
	if t.Inventory.Repository == nil {
		return "inventory_" + strconv.Itoa(t.Inventory.ID)
	}
	return t.Inventory.Repository.GetDirName(t.Template.ID) + "_inventory_" + strconv.Itoa(t.Inventory.ID)
}

func (t *LocalJob) tmpInventoryFullPath() string {
	if t.Inventory.Repository != nil && t.Inventory.Repository.GetType() == db.RepositoryLocal {
		return t.Inventory.Repository.GetGitURL(true)
	}
	pathname := path.Join(util.Config.GetProjectTmpDir(t.Template.ProjectID), t.tmpInventoryFilename())
	if t.Inventory.Type == db.InventoryStaticYaml {
		pathname += ".yml"
	}
	return pathname
}

func (t *LocalJob) cloneInventoryRepo(keyInstaller db_lib.AccessKeyInstaller) error {
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

func (t *LocalJob) installStaticInventory() error {
	t.Log("installing static inventory")

	fullPath := t.tmpInventoryFullPath()

	// create inventory file
	return os.WriteFile(fullPath, []byte(t.Inventory.Inventory), 0664)
}

func (t *LocalJob) destroyInventoryFile() {
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

func (t *LocalJob) destroyKeys() {
	err := t.sshKeyInstallation.Destroy()
	if err != nil {
		t.Log("Can't destroy inventory user key, error: " + err.Error())
	}

	err = t.becomeKeyInstallation.Destroy()
	if err != nil {
		t.Log("Can't destroy inventory become user key, error: " + err.Error())
	}

	for _, vault := range t.vaultFileInstallations {
		err = vault.Destroy()
		if err != nil {
			t.Log("Can't destroy inventory vault password file, error: " + err.Error())
		}
	}
}
