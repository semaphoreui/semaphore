package tasks

import (
	"testing"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupExecutorConfig(t *testing.T) {
	t.Helper()
	prevCfg := util.Config
	t.Cleanup(func() { util.Config = prevCfg })
	if util.Config == nil {
		util.Config = &util.ConfigType{}
	}
}

// TestGetShellArgs_PassesSurveySecretVar verifies that Survey variables of type
// "Secret" (delivered via LocalExecutor.Secret) are passed to Bash/Shell tasks,
// alongside plain survey vars that arrive in Environment.JSON.
func TestGetShellArgs_PassesSurveySecretVar(t *testing.T) {
	setupExecutorConfig(t)

	exec := &LocalExecutor{
		Template: db.Template{
			Type:     db.TemplateTask,
			Playbook: "date.sh",
		},
		Environment: db.Environment{
			JSON: `{"PLAIN_VAR":"hello"}`,
		},
		Secret: `{"MY_VAR":"s3cr3t"}`,
	}

	args, err := exec.getShellArgs("admin", nil)
	require.NoError(t, err)

	assert.Contains(t, args, "PLAIN_VAR=hello", "plain survey var must be passed")
	assert.Contains(t, args, "MY_VAR=s3cr3t", "secret survey var must be passed")
}

// TestGetEnvironmentExtraVars_MergesSecret checks the shared helper merges the
// Secret field into the extra vars map used by Shell and Terraform tasks.
func TestGetEnvironmentExtraVars_MergesSecret(t *testing.T) {
	setupExecutorConfig(t)

	exec := &LocalExecutor{
		Environment: db.Environment{JSON: `{"PLAIN_VAR":"hello"}`},
		Secret:      `{"MY_VAR":"s3cr3t"}`,
	}

	extraVars, err := exec.getEnvironmentExtraVars("admin", nil)
	require.NoError(t, err)

	assert.Equal(t, "hello", extraVars["PLAIN_VAR"])
	assert.Equal(t, "s3cr3t", extraVars["MY_VAR"])
}
// TestGetInventorySSHCommonArgs verifies the ssh options ansible receives to
// reach inventory hosts through a jump host.
func TestGetInventorySSHCommonArgs(t *testing.T) {
	setupExecutorConfig(t)

	user := "ansible-proxy"
	port := 2222

	t.Run("no proxy means no ssh args", func(t *testing.T) {
		e := &LocalExecutor{Inventory: db.Inventory{}}
		assert.Empty(t, e.getInventorySSHCommonArgs())
	})

	t.Run("proxy adds ProxyJump", func(t *testing.T) {
		e := &LocalExecutor{Inventory: db.Inventory{
			Proxy: &db.Proxy{
				Type: db.ProxySSH,
				Host: "bastion.example.org",
				User: &user,
				Port: &port,
			},
		}}

		assert.Equal(t, "-o ProxyJump=ansible-proxy@bastion.example.org:2222", e.getInventorySSHCommonArgs())
	})
}

// TestGetPlaybookArgs_Proxy verifies the jump host reaches the ansible-playbook
// command line, and that it is absent when the inventory has no proxy.
func TestGetPlaybookArgs_Proxy(t *testing.T) {
	setupExecutorConfig(t)

	inventoryID := 1

	newExecutor := func(proxy *db.Proxy) *LocalExecutor {
		return &LocalExecutor{
			Template:  db.Template{App: db.AppAnsible, Playbook: "test.yml"},
			Inventory: db.Inventory{ID: inventoryID, Type: db.InventoryStatic, Proxy: proxy},
		}
	}

	t.Run("without proxy", func(t *testing.T) {
		args, _, err := newExecutor(nil).getPlaybookArgs("admin", nil)

		require.NoError(t, err)
		assert.NotContains(t, args, "--ssh-common-args")
	})

	t.Run("with proxy", func(t *testing.T) {
		args, _, err := newExecutor(&db.Proxy{Type: db.ProxySSH, Host: "bastion.example.org"}).
			getPlaybookArgs("admin", nil)

		require.NoError(t, err)
		require.Contains(t, args, "--ssh-common-args")

		i := indexOf(args, "--ssh-common-args")
		require.Less(t, i+1, len(args))
		assert.Equal(t, "-o ProxyJump=bastion.example.org", args[i+1])
	})
}

func indexOf(args []string, value string) int {
	for i, a := range args {
		if a == value {
			return i
		}
	}
	return -1
}
