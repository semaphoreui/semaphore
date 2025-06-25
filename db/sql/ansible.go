//go:build !pro

package sql

import (
	"github.com/semaphoreui/semaphore/db"
)

func (d *SqlDb) CreateAnsibleTaskHost(host db.AnsibleTaskHost) error {
	return nil
}

func (d *SqlDb) CreateAnsibleTaskError(error db.AnsibleTaskError) error {
	return nil
}

func (d *SqlDb) GetAnsibleTaskHosts(projectID int, taskID int) (res []db.AnsibleTaskHost, err error) {
	res = []db.AnsibleTaskHost{
		{
			Host:        "192.168.0.1",
			Changed:     3,
			Failed:      0,
			Ignored:     1,
			Ok:          1,
			Rescued:     0,
			Skipped:     3,
			Unreachable: 0,
		},
		{
			Host:        "192.168.0.1",
			Changed:     3,
			Failed:      0,
			Ignored:     1,
			Ok:          1,
			Rescued:     0,
			Skipped:     3,
			Unreachable: 0,
		},
		{
			Host:        "192.168.0.2",
			Changed:     2,
			Failed:      1,
			Ignored:     0,
			Ok:          4,
			Rescued:     1,
			Skipped:     2,
			Unreachable: 0,
		},
		{
			Host:        "10.0.0.5",
			Changed:     5,
			Failed:      0,
			Ignored:     2,
			Ok:          8,
			Rescued:     0,
			Skipped:     1,
			Unreachable: 0,
		},
		{
			Host:        "web-server-01",
			Changed:     1,
			Failed:      2,
			Ignored:     0,
			Ok:          3,
			Rescued:     1,
			Skipped:     0,
			Unreachable: 1,
		},
		{
			Host:        "database-primary",
			Changed:     7,
			Failed:      0,
			Ignored:     0,
			Ok:          12,
			Rescued:     0,
			Skipped:     2,
			Unreachable: 0,
		},
		{
			Host:        "172.16.10.15",
			Changed:     2,
			Failed:      1,
			Ignored:     3,
			Ok:          5,
			Rescued:     2,
			Skipped:     1,
			Unreachable: 0,
		},
		{
			Host:        "app-server-03",
			Changed:     4,
			Failed:      0,
			Ignored:     1,
			Ok:          7,
			Rescued:     0,
			Skipped:     3,
			Unreachable: 0,
		},
		{
			Host:        "192.168.5.100",
			Changed:     0,
			Failed:      3,
			Ignored:     0,
			Ok:          2,
			Rescued:     0,
			Skipped:     1,
			Unreachable: 2,
		},
		{
			Host:        "load-balancer-01",
			Changed:     2,
			Failed:      0,
			Ignored:     0,
			Ok:          5,
			Rescued:     0,
			Skipped:     0,
			Unreachable: 0,
		},
		{
			Host:        "10.10.0.5",
			Changed:     6,
			Failed:      1,
			Ignored:     2,
			Ok:          9,
			Rescued:     1,
			Skipped:     4,
			Unreachable: 0,
		},
		{
			Host:        "cache-server",
			Changed:     3,
			Failed:      0,
			Ignored:     0,
			Ok:          6,
			Rescued:     0,
			Skipped:     1,
			Unreachable: 0,
		},
		{
			Host:        "192.168.2.25",
			Changed:     1,
			Failed:      2,
			Ignored:     1,
			Ok:          3,
			Rescued:     0,
			Skipped:     2,
			Unreachable: 1,
		},
		{
			Host:        "worker-node-01",
			Changed:     4,
			Failed:      0,
			Ignored:     0,
			Ok:          7,
			Rescued:     0,
			Skipped:     2,
			Unreachable: 0,
		},
		{
			Host:        "172.16.20.30",
			Changed:     5,
			Failed:      0,
			Ignored:     3,
			Ok:          10,
			Rescued:     1,
			Skipped:     3,
			Unreachable: 0,
		},
		{
			Host:        "monitoring-server",
			Changed:     2,
			Failed:      0,
			Ignored:     0,
			Ok:          8,
			Rescued:     0,
			Skipped:     1,
			Unreachable: 0,
		},
		{
			Host:        "10.0.1.15",
			Changed:     0,
			Failed:      4,
			Ignored:     1,
			Ok:          2,
			Rescued:     0,
			Skipped:     0,
			Unreachable: 2,
		},
		{
			Host:        "backup-server",
			Changed:     1,
			Failed:      0,
			Ignored:     0,
			Ok:          9,
			Rescued:     0,
			Skipped:     2,
			Unreachable: 0,
		},
		{
			Host:        "192.168.10.50",
			Changed:     3,
			Failed:      1,
			Ignored:     2,
			Ok:          6,
			Rescued:     1,
			Skipped:     2,
			Unreachable: 0,
		},
		{
			Host:        "dev-environment",
			Changed:     8,
			Failed:      0,
			Ignored:     0,
			Ok:          14,
			Rescued:     0,
			Skipped:     3,
			Unreachable: 0,
		},
		{
			Host:        "172.16.30.100",
			Changed:     2,
			Failed:      3,
			Ignored:     1,
			Ok:          4,
			Rescued:     0,
			Skipped:     1,
			Unreachable: 1,
		},
	}
	return
}

func (d *SqlDb) GetAnsibleTaskErrors(projectID int, taskID int) (res []db.AnsibleTaskError, err error) {
	res = []db.AnsibleTaskError{
		{
			Host:  "192.168.0.1",
			Task:  "Check memory",
			Error: "Memory check failed",
		},
		{
			Host:  "192.168.0.1",
			Task:  "Check memory",
			Error: "Memory check failed",
		},
		{
			Host:  "192.168.0.2",
			Task:  "Install packages",
			Error: "Package repository not found",
		},
		{
			Host:  "10.0.0.15",
			Task:  "Restart service",
			Error: "Service failed to restart: timeout",
		},
		{
			Host:  "web-server-01",
			Task:  "Configure firewall",
			Error: "Invalid firewall rule syntax",
		},
		{
			Host:  "192.168.1.50",
			Task:  "Deploy application",
			Error: "Insufficient disk space",
		},
		{
			Host:  "database-01",
			Task:  "Backup database",
			Error: "Permission denied to backup location",
		},
		{
			Host:  "10.10.5.3",
			Task:  "Update system packages",
			Error: "Network connection interrupted",
		},
		{
			Host:  "worker-node-3",
			Task:  "Configure Docker",
			Error: "Docker daemon failed to start",
		},
		{
			Host:  "192.168.0.45",
			Task:  "Clone repository",
			Error: "Git authentication failed",
		},
		{
			Host:  "load-balancer-02",
			Task:  "Configure SSL certificate",
			Error: "Certificate validation error: expired",
		},
	}
	return
}
