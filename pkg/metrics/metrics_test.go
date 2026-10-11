package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/semaphoreui/semaphore/pkg/task_logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// scrape renders the current metrics in Prometheus text exposition format,
// the same way a real scrape would, so tests assert on it exactly like an
// external Prometheus server would see it - no extra test-only dependency
// on prometheus/client_golang/prometheus/testutil required.
func scrape(m *Metrics) string {
	req := httptest.NewRequest("GET", "/api/metrics", nil)
	w := httptest.NewRecorder()
	m.ServeHTTP(w, req)
	return w.Body.String()
}

func TestMetrics_ServeHTTP(t *testing.T) {
	m := NewMetrics()

	req := httptest.NewRequest("GET", "/api/metrics", nil)
	w := httptest.NewRecorder()

	m.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "go_goroutines")
}

func TestMetrics_RecordTaskStatusChange_RunningGauge(t *testing.T) {
	m := NewMetrics()

	m.RecordTaskStatusChange(task_logger.TaskWaitingStatus, task_logger.TaskRunningStatus)
	assert.Contains(t, scrape(m), "semaphore_tasks_running 1")

	m.RecordTaskStatusChange(task_logger.TaskRunningStatus, task_logger.TaskSuccessStatus)
	assert.Contains(t, scrape(m), "semaphore_tasks_running 0")
}

func TestMetrics_RecordTaskStatusChange_SkipsRunningGaugeWhenNeverRan(t *testing.T) {
	m := NewMetrics()

	// A task rejected before ever running must not push the gauge negative.
	m.RecordTaskStatusChange(task_logger.TaskWaitingConfirmation, task_logger.TaskRejected)
	m.RecordTaskStatusChange(task_logger.TaskRejected, task_logger.TaskStoppedStatus)

	assert.Contains(t, scrape(m), "semaphore_tasks_running 0")
}

func TestMetrics_RecordTaskStatusChange_OutcomeCounter(t *testing.T) {
	tests := []struct {
		name   string
		status task_logger.TaskStatus
	}{
		{"success", task_logger.TaskSuccessStatus},
		{"error", task_logger.TaskFailStatus},
		{"stopped", task_logger.TaskStoppedStatus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMetrics()

			m.RecordTaskStatusChange(task_logger.TaskRunningStatus, tt.status)

			assert.Contains(t, scrape(m), `semaphore_tasks_total{status="`+string(tt.status)+`"} 1`)
		})
	}
}

func TestMetrics_RecordTaskStatusChange_TerminalToTerminalNotDoubleCounted(t *testing.T) {
	m := NewMetrics()

	m.RecordTaskStatusChange(task_logger.TaskRunningStatus, task_logger.TaskSuccessStatus)
	m.RecordTaskStatusChange(task_logger.TaskSuccessStatus, task_logger.TaskFailStatus)

	body := scrape(m)
	assert.Contains(t, body, `semaphore_tasks_total{status="success"} 1`)
	assert.NotContains(t, body, `semaphore_tasks_total{status="error"}`)
}

func TestMetrics_RecordTaskStatusChange_NonFinishedStatusNotCounted(t *testing.T) {
	m := NewMetrics()

	m.RecordTaskStatusChange(task_logger.TaskWaitingStatus, task_logger.TaskRunningStatus)

	assert.NotContains(t, scrape(m), "semaphore_tasks_total{")
}

func TestMetrics_RecordTaskStatusChange_NilReceiverIsNoop(t *testing.T) {
	var m *Metrics

	assert.NotPanics(t, func() {
		m.RecordTaskStatusChange(task_logger.TaskWaitingStatus, task_logger.TaskRunningStatus)
	})
}

func TestRegister_ExposesForeignCollectors(t *testing.T) {
	m := NewMetrics()
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{Name: "semaphore_test_gauge", Help: "Test."})
	gauge.Set(7)

	require.NoError(t, m.Register(gauge))

	w := httptest.NewRecorder()
	m.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/metrics", nil))
	assert.Contains(t, w.Body.String(), "semaphore_test_gauge 7")
	assert.Error(t, m.Register(gauge), "a second registration of the same collector")
}

func TestRegister_NilMetrics(t *testing.T) {
	var m *Metrics
	assert.NoError(t, m.Register(prometheus.NewGauge(prometheus.GaugeOpts{Name: "x", Help: "x"})))
}
