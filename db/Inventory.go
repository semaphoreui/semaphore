package db

const (
	InventoryStatic     = "static"
	InventoryStaticYaml = "static-yaml"
	InventoryFile       = "file"
)

// Inventory is the model of an ansible inventory file
type Inventory struct {
	ID        int    `db:"id" json:"id"`
	Name      string `db:"name" json:"name" binding:"required"`
	ProjectID int    `db:"project_id" json:"project_id"`
	Inventory string `db:"inventory" json:"inventory"`

	// accesses hosts in inventory
	SSHKeyID *int      `db:"ssh_key_id" json:"ssh_key_id"`
	SSHKey   AccessKey `db:"-" json:"-"`

	BecomeKeyID *int      `db:"become_key_id" json:"become_key_id"`
	BecomeKey   AccessKey `db:"-" json:"-"`

	// static/file
	Type string `db:"type" json:"type"`
}

//func (i *Inventory) StartSshAgent(logger lib.Logger) (lib.SshAgent, error) {
//
//	sshAgent := lib.SshAgent{
//		Logger: logger,
//		Keys: []lib.SshAgentKey{
//			{
//				Key:        []byte(i.SSHKey.SshKey.PrivateKey),
//				Passphrase: []byte(i.SSHKey.SshKey.Passphrase),
//			},
//		},
//		SocketFile: path.Join(util.Config.TmpPath, fmt.Sprintf("ssh-agent-%d-%d.sock", time.Now().Unix(), 0)),
//	}
//
//	if i.BecomeKeyID != nil {
//		sshAgent.Keys = append(sshAgent.Keys, lib.SshAgentKey{
//			Key:        []byte(i.BecomeKey.SshKey.PrivateKey),
//			Passphrase: []byte(i.BecomeKey.SshKey.Passphrase),
//		})
//	}
//
//	return sshAgent, sshAgent.Listen()
//}

func FillInventory(d Store, inventory *Inventory) (err error) {
	if inventory.SSHKeyID != nil {
		inventory.SSHKey, err = d.GetAccessKey(inventory.ProjectID, *inventory.SSHKeyID)
	}

	if err != nil {
		return
	}

	if inventory.BecomeKeyID != nil {
		inventory.BecomeKey, err = d.GetAccessKey(inventory.ProjectID, *inventory.BecomeKeyID)
	}

	return
}
