package telemetry

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

const prometheusNamespace = "z4"

var (
	ReceivedLogs = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: prometheusNamespace,
		Name:      "received_log_count",
		Help:      "The total number of Raft logs sent for application to the fsm",
	})
	LastFSMSnapshot = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prometheusNamespace,
		Name:      "last_fsm_snapshot",
		Help:      "The unix time in seconds when the last fsm snapshot was taken",
	})
	PulledTasks = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: prometheusNamespace,
		Name:      "pulled_task_total",
		Help:      "The total number of tasks sent to clients",
	}, []string{"method", "queue"})
	PushedTasks = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: prometheusNamespace,
		Name:      "pushed_task_total",
		Help:      "The total number of tasks pushed from clients to the server",
	}, []string{"method", "queue"})
	RemovedTasks = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: prometheusNamespace,
		Name:      "removed_task_total",
		Help:      "The total number of tasks removed from the queue",
	}, []string{"method", "queue"})
	IsLeader = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: prometheusNamespace,
		Name:      "is_leader",
		Help:      "A boolean value that indicates if a peer is the cluster leader",
	})
)

// StartPromServer exposes prometheus metrics on the given port.
func StartPromServer(port int) error {
	http.Handle("/metrics", promhttp.Handler())
	addr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(addr, nil)
}
