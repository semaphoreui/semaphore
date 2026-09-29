package audit

import (
	"context"
	"fmt"
	"net/http"
	"net/netip"

	"github.com/semaphoreui/semaphore/db"
	"github.com/semaphoreui/semaphore/pro_interfaces"
	"github.com/semaphoreui/semaphore/util"
)

type Service struct {
	recorder Recorder
	exporter pro_interfaces.AuditExporter
	trusted  []netip.Prefix
	enabled  bool
}

func StartService(
	store db.AuditEventManager,
	conf *util.AuditConfig,
	nodeID string,
	exporter pro_interfaces.AuditExporter,
) (*Service, error) {
	if !conf.IsEnabled() {
		return &Service{recorder: Nop{}}, nil
	}

	trusted, err := conf.TrustedProxies()
	if err != nil {
		return nil, err
	}

	// Started first, so it exports the start event.
	if err = exporter.Start(); err != nil {
		return nil, fmt.Errorf("audit export: %w", err)
	}

	s := &Service{
		recorder: NewRecorder(store, Options{InstanceID: conf.InstanceID, NodeID: nodeID}),
		exporter: exporter,
		trusted:  trusted,
		enabled:  true,
	}

	destinations := exporter.DestinationIDs()
	if destinations == nil {
		destinations = []string{}
	}
	s.recorder.Record(WithActor(context.Background(), SystemActor(ComponentServer)), Event{
		Kind:     AuditLifecycleStart,
		Metadata: LifecycleMetadata{Destinations: destinations},
	})

	return s, nil
}

func (s *Service) Stop() {
	if s.enabled {
		s.exporter.Stop()
	}
}

func (s *Service) Recorder() Recorder {
	return s.recorder
}

func (s *Service) Wrap(h http.Handler) http.Handler {
	if !s.enabled {
		return h
	}
	return RequestMiddleware(s.trusted)(h)
}
