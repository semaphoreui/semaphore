package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommunityAuditExporterIsNoop(t *testing.T) {
	exporter, err := NewAuditExporter(nil)
	require.NoError(t, err)
	require.NotNil(t, exporter)
	require.NotPanics(t, func() {
		exporter.Stop()
		exporter.Stop()
	})
}
