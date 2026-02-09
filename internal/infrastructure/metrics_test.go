package infrastructure

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNewMetrics(t *testing.T) {
	// Clear default registry to avoid "already registered" errors
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	metrics, err := NewMetrics()
	if err != nil {
		t.Fatalf("Failed to create metrics: %v", err)
	}

	if metrics == nil {
		t.Fatal("Expected non-nil metrics")
	}

	if metrics.WoLPacketsSent == nil {
		t.Fatal("Expected non-nil WoLPacketsSent")
	}
	if metrics.WoLPacketsFailed == nil {
		t.Fatal("Expected non-nil WoLPacketsFailed")
	}
	if metrics.MachineNotFound == nil {
		t.Fatal("Expected non-nil MachineNotFound")
	}
	if metrics.MachinesListed == nil {
		t.Fatal("Expected non-nil MachinesListed")
	}
	if metrics.MachinesRetrieved == nil {
		t.Fatal("Expected non-nil MachinesRetrieved")
	}
	if metrics.RequestDuration == nil {
		t.Fatal("Expected non-nil RequestDuration")
	}
	if metrics.ConfiguredMachines == nil {
		t.Fatal("Expected non-nil ConfiguredMachines")
	}
}

func TestMetricsCounterIncrement(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	metrics, err := NewMetrics()
	if err != nil {
		t.Fatalf("Failed to create metrics: %v", err)
	}

	// Test counter increments
	metrics.WoLPacketsSent.Inc()
	metrics.WoLPacketsSent.Inc()
	metrics.WoLPacketsFailed.Inc()
	metrics.MachineNotFound.Inc()
	metrics.MachinesListed.Inc()
	metrics.MachinesRetrieved.Inc()

	// Should not panic - counters are incremented
}

func TestMetricsGaugeSet(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	metrics, err := NewMetrics()
	if err != nil {
		t.Fatalf("Failed to create metrics: %v", err)
	}

	// Test gauge set
	metrics.ConfiguredMachines.Set(5)
	metrics.ConfiguredMachines.Set(10)
	metrics.ConfiguredMachines.Add(2)
	metrics.ConfiguredMachines.Sub(3)

	// Should not panic - gauge is updated
}

func TestMetricsHistogramObserve(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	metrics, err := NewMetrics()
	if err != nil {
		t.Fatalf("Failed to create metrics: %v", err)
	}

	// Test histogram observe
	metrics.RequestDuration.Observe(0.1)
	metrics.RequestDuration.Observe(0.5)
	metrics.RequestDuration.Observe(1.0)

	// Should not panic - histogram is recorded
}

func TestMetricsHandler(t *testing.T) {
	handler := MetricsHandler()
	if handler == nil {
		t.Fatal("Expected non-nil handler")
	}
}

func TestMetricsRegistrationError(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	// Create first set of metrics successfully
	_, err := NewMetrics()
	if err != nil {
		t.Fatalf("Failed to create initial metrics: %v", err)
	}

	// Try to create again - should fail because metrics are already registered
	_, err = NewMetrics()
	if err == nil {
		t.Fatal("Expected error when registering duplicate metrics")
	}
}
