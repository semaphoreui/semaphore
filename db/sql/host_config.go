package sql

import "github.com/semaphoreui/semaphore/db"

func (d *SqlDb) GetHostConfig(projectID int, hostConfigID int) (hostConfig db.HostConfig, err error) {
	err = d.getObject(projectID, db.HostConfigProps, hostConfigID, &hostConfig)
	return
}

func (d *SqlDb) GetHostConfigs(projectID int, params db.RetrieveQueryParams) (hostConfigs []db.HostConfig, err error) {
	err = d.getObjects(projectID, db.HostConfigProps, params, nil, &hostConfigs)
	return
}

func (d *SqlDb) CreateHostConfig(hostConfig db.HostConfig) (newHostConfig db.HostConfig, err error) {
	insertID, err := d.insert(
		"id",
		"insert into project__host_config (project_id, `type`, name, ssh_key_id) values (?, ?, ?, ?)",
		hostConfig.ProjectID,
		hostConfig.Type,
		hostConfig.Name,
		hostConfig.SSHKeyID)

	if err != nil {
		return
	}

	newHostConfig = hostConfig
	newHostConfig.ID = insertID
	return
}

func (d *SqlDb) UpdateHostConfig(hostConfig db.HostConfig) error {
	_, err := d.exec(
		"update project__host_config set `type`=?, name=?, ssh_key_id=? where id=? and project_id=?",
		hostConfig.Type,
		hostConfig.Name,
		hostConfig.SSHKeyID,
		hostConfig.ID,
		hostConfig.ProjectID)

	return err
}

func (d *SqlDb) DeleteHostConfig(projectID int, hostConfigID int) error {
	return d.deleteObject(projectID, db.HostConfigProps, hostConfigID)
}
