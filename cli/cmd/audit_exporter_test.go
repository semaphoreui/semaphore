package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/semaphoreui/semaphore/db"
	proServer "github.com/semaphoreui/semaphore/pro/services/server"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAuditExporter struct {
	calls    []string
	startCtx context.Context
	startErr error
	stopErr  error
}

func (exporter *fakeAuditExporter) Start(ctx context.Context) error {
	exporter.calls = append(exporter.calls, "start")
	exporter.startCtx = ctx
	return exporter.startErr
}

func (exporter *fakeAuditExporter) Stop(context.Context) error {
	select {
	case <-exporter.startCtx.Done():
		exporter.calls = append(exporter.calls, "cancel")
	default:
		exporter.calls = append(exporter.calls, "stop without cancel")
	}
	exporter.calls = append(exporter.calls, "stop")
	return exporter.stopErr
}

func auditExporterFactoryFor(
	exporter pro_interfaces.AuditExporter,
	err error,
	called *bool,
) auditExporterFactory {
	return func(util.AuditConfig, db.AuditExporterStore) (pro_interfaces.AuditExporter, error) {
		*called = true
		return exporter, err
	}
}

func TestStartAuditExporter(t *testing.T) {
	t.Run("disabled does not call factory", func(t *testing.T) {
		called := false
		stop, err := startAuditExporter(
			&util.AuditConfig{},
			nil,
			auditExporterFactoryFor(nil, nil, &called),
		)

		require.NoError(t, err)
		assert.Nil(t, stop)
		assert.False(t, called)
	})

	t.Run("enabled without Pro implementation", func(t *testing.T) {
		stop, err := startAuditExporter(
			&util.AuditConfig{Enabled: true},
			nil,
			proServer.NewAuditExporter,
		)

		assert.Nil(t, stop)
		assert.ErrorContains(t, err, "audit export is only available in the proprietary build")
	})

	t.Run("successful Start stops once after cancellation", func(t *testing.T) {
		called := false
		exporter := &fakeAuditExporter{}
		stop, err := startAuditExporter(
			&util.AuditConfig{Enabled: true},
			nil,
			auditExporterFactoryFor(exporter, nil, &called),
		)

		require.NoError(t, err)
		require.NotNil(t, stop)
		stop()
		assert.True(t, called)
		assert.Equal(t, []string{"start", "cancel", "stop"}, exporter.calls)
	})

	t.Run("partial Start error stops once after cancellation", func(t *testing.T) {
		called := false
		startErr := errors.New("start failed")
		exporter := &fakeAuditExporter{startErr: startErr}
		stop, err := startAuditExporter(
			&util.AuditConfig{Enabled: true},
			nil,
			auditExporterFactoryFor(exporter, nil, &called),
		)

		assert.Nil(t, stop)
		assert.ErrorIs(t, err, startErr)
		assert.True(t, called)
		assert.Equal(t, []string{"start", "cancel", "stop"}, exporter.calls)
	})

	t.Run("Start error remains primary after Stop error", func(t *testing.T) {
		called := false
		startErr := errors.New("start failed")
		exporter := &fakeAuditExporter{
			startErr: startErr,
			stopErr:  errors.New("stop failed"),
		}
		stop, err := startAuditExporter(
			&util.AuditConfig{Enabled: true},
			nil,
			auditExporterFactoryFor(exporter, nil, &called),
		)

		assert.Nil(t, stop)
		assert.ErrorIs(t, err, startErr)
		assert.True(t, called)
		assert.Equal(t, []string{"start", "cancel", "stop"}, exporter.calls)
	})
}
