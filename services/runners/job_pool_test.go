package runners

import (
	"testing"

	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPClient_NilRunnerConnection(t *testing.T) {
	util.Config = util.NewConfigType()
	util.Config.Runner = &util.RunnerConfig{}

	require.NotPanics(t, func() {
		client := newHTTPClient()
		require.NotNil(t, client)
	})
}
