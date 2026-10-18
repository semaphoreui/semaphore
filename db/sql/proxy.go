package sql

import (
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) GetProxy(projectID int, proxyID int) (proxy db.Proxy, err error) {
	return d.getProxy(projectID, proxyID, db.MaxProxyChainLength)
}

// getProxy loads a proxy with the chain of proxies it requires. depth bounds the
// recursion so a chain which loops back on itself cannot hang the task, even
// though ValidateProxy rejects such a chain when it is created.
func (d *SqlDb) getProxy(projectID int, proxyID int, depth int) (proxy db.Proxy, err error) {
	err = d.getObject(projectID, db.ProxyProps, proxyID, &proxy)
	if err != nil {
		return
	}

	if proxy.SSHKeyID != nil {
		proxy.SSHKey, err = d.GetAccessKey(projectID, *proxy.SSHKeyID)
		if err != nil {
			return
		}
	}

	if proxy.RequiresProxyID == nil || depth <= 1 {
		return
	}

	var required db.Proxy
	required, err = d.getProxy(projectID, *proxy.RequiresProxyID, depth-1)
	if err != nil {
		return
	}

	proxy.RequiresProxy = &required
	return
}

func (d *SqlDb) GetProxies(projectID int, params db.RetrieveQueryParams) (proxies []db.Proxy, err error) {
	proxies = []db.Proxy{}
	err = d.getObjects(projectID, db.ProxyProps, params, nil, &proxies)
	return
}

func (d *SqlDb) GetProxyRefs(projectID int, proxyID int) (db.ObjectReferrers, error) {
	return d.getObjectRefs(projectID, db.ProxyProps, proxyID)
}

func (d *SqlDb) CreateProxy(proxy db.Proxy) (newProxy db.Proxy, err error) {
	insertID, err := d.insert(
		"id",
		"insert into project__proxy (project_id, name, `type`, host, port, `user`, ssh_key_id, requires_proxy_id) "+
			"values (?, ?, ?, ?, ?, ?, ?, ?)",
		proxy.ProjectID,
		proxy.Name,
		proxy.Type,
		proxy.Host,
		proxy.Port,
		proxy.User,
		proxy.SSHKeyID,
		proxy.RequiresProxyID)

	if err != nil {
		return
	}

	newProxy = proxy
	newProxy.ID = insertID
	return
}

func (d *SqlDb) UpdateProxy(proxy db.Proxy) error {
	_, err := d.exec(
		"update project__proxy set name=?, `type`=?, host=?, port=?, `user`=?, ssh_key_id=?, requires_proxy_id=? "+
			"where id=? and project_id=?",
		proxy.Name,
		proxy.Type,
		proxy.Host,
		proxy.Port,
		proxy.User,
		proxy.SSHKeyID,
		proxy.RequiresProxyID,
		proxy.ID,
		proxy.ProjectID)

	return err
}

func (d *SqlDb) DeleteProxy(projectID int, proxyID int) error {
	return d.deleteObject(projectID, db.ProxyProps, proxyID)
}
