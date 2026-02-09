// Package infrastructure provides infrastructure layer implementations.
package infrastructure

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics for the application.
type Metrics struct {
	WoLPacketsSent     prometheus.Counter
	WoLPacketsFailed   prometheus.Counter
	MachineNotFound    prometheus.Counter
	MachinesListed     prometheus.Counter
	MachinesRetrieved  prometheus.Counter
	RequestDuration    prometheus.Histogram
	ConfiguredMachines prometheus.Gauge
}

// NewMetrics creates and registers all Prometheus metrics.
func NewMetrics() (*Metrics, error) {
	m := &Metrics{
		WoLPacketsSent: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "gwaihir_wol_packets_sent_total",
			Help: "Total number of WoL packets successfully sent",
		}),
		WoLPacketsFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "gwaihir_wol_packets_failed_total",
			Help: "Total number of WoL packet send failures",
		}),
		MachineNotFound: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "gwaihir_machine_not_found_total",
			Help: "Total number of machine not found errors",
		}),
		MachinesListed: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "gwaihir_machines_listed_total",
			Help: "Total number of times machines list was requested",
		}),
		MachinesRetrieved: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "gwaihir_machines_retrieved_total",
			Help: "Total number of times a machine was retrieved by ID",
		}),
		RequestDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "gwaihir_request_duration_seconds",
			Help:    "Request latency in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		ConfiguredMachines: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "gwaihir_configured_machines_total",
			Help: "Total number of configured machines in allowlist",
		}),
	}

	// Register all metrics
	if err := prometheus.Register(m.WoLPacketsSent); err != nil {
		return nil, fmt.Errorf("failed to register WoLPacketsSent: %w", err)
	}
	if err := prometheus.Register(m.WoLPacketsFailed); err != nil {
		return nil, fmt.Errorf("failed to register WoLPacketsFailed: %w", err)
	}
	if err := prometheus.Register(m.MachineNotFound); err != nil {
		return nil, fmt.Errorf("failed to register MachineNotFound: %w", err)
	}
	if err := prometheus.Register(m.MachinesListed); err != nil {
		return nil, fmt.Errorf("failed to register MachinesListed: %w", err)
	}
	if err := prometheus.Register(m.MachinesRetrieved); err != nil {
		return nil, fmt.Errorf("failed to register MachinesRetrieved: %w", err)
	}
	if err := prometheus.Register(m.RequestDuration); err != nil {
		return nil, fmt.Errorf("failed to register RequestDuration: %w", err)
	}
	if err := prometheus.Register(m.ConfiguredMachines); err != nil {
		return nil, fmt.Errorf("failed to register ConfiguredMachines: %w", err)
	}

	return m, nil
}

// MetricsHandler returns an HTTP handler that serves Prometheus metrics.
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
