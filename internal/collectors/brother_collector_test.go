package collectors

import (
	"testing"

	"github.com/d0ugal/brother-exporter/internal/config"
	"github.com/d0ugal/brother-exporter/internal/metrics"
	promexporter_metrics "github.com/d0ugal/promexporter/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestBrotherCollector_Constants(t *testing.T) {
	// Test that all OID constants are properly defined
	assert.NotEmpty(t, OIDSystemDescription)
	assert.NotEmpty(t, OIDSystemUpTime)
	assert.NotEmpty(t, OIDSystemContact)
	assert.NotEmpty(t, OIDSystemName)
	assert.NotEmpty(t, OIDSystemLocation)
	assert.NotEmpty(t, OIDBrotherBase)
	assert.NotEmpty(t, OIDPrinterStatus)
	assert.NotEmpty(t, OIDBrotherConsumableInfo)
	assert.NotEmpty(t, OIDBrotherConsumableLevel)
	assert.NotEmpty(t, OIDBrotherStatus)
	assert.NotEmpty(t, OIDBrotherFirmware)
	assert.NotEmpty(t, OIDTonerLevelBase)
	assert.NotEmpty(t, OIDDrumLevelBase)
	assert.NotEmpty(t, OIDPaperTrayStatusBase)
}

func TestBrotherCollector_ColorMappings(t *testing.T) {
	// Test that color mappings are properly defined
	assert.NotEmpty(t, LaserColors)
	assert.NotEmpty(t, InkColors)
	assert.Contains(t, LaserColors, "black")
	assert.Contains(t, LaserColors, "cyan")
	assert.Contains(t, LaserColors, "magenta")
	assert.Contains(t, LaserColors, "yellow")
	assert.Contains(t, InkColors, "black")
	assert.Contains(t, InkColors, "cyan")
	assert.Contains(t, InkColors, "magenta")
	assert.Contains(t, InkColors, "yellow")
}

// TestPrinterConnectionErrors_LabelsMatchRegistry guards against the label-name
// mismatch between the registry definition (host, error_type) and the call
// sites in handleCollectionError / the connect-error path. If the labels
// drift again, prometheus.Labels.With() panics here.
func TestPrinterConnectionErrors_LabelsMatchRegistry(t *testing.T) {
	baseRegistry := promexporter_metrics.NewRegistry("brother_exporter_info_test")
	brotherMetrics := metrics.NewBrotherRegistry(baseRegistry)

	assert.NotPanics(t, func() {
		brotherMetrics.PrinterConnectionErrors.With(prometheus.Labels{
			"host":       "test-host",
			"error_type": "connect",
		}).Inc()
		brotherMetrics.PrinterConnectionErrors.With(prometheus.Labels{
			"host":       "test-host",
			"error_type": "deviceinfo",
		}).Inc()
	})
}

func TestPrinterUnreachable(t *testing.T) {
	assert.True(t, printerUnreachable(true, true, true))
	assert.False(t, printerUnreachable(false, false, false))
	assert.False(t, printerUnreachable(true, false, true))
}

func TestSetConnectionStatus(t *testing.T) {
	baseRegistry := promexporter_metrics.NewRegistry("brother_exporter_info_test_status")
	brotherMetrics := metrics.NewBrotherRegistry(baseRegistry)
	bc := &BrotherCollector{
		config:  &config.Config{Printer: config.PrinterConfig{Host: "test-host"}},
		metrics: brotherMetrics,
	}
	gauge := brotherMetrics.PrinterConnectionStatus.With(prometheus.Labels{"host": "test-host"})

	bc.setConnectionStatus(false)
	assert.InDelta(t, 0.0, testutil.ToFloat64(gauge), 0)

	bc.setConnectionStatus(true)
	assert.InDelta(t, 1.0, testutil.ToFloat64(gauge), 0)

	bc.setConnectionStatus(!printerUnreachable(true, true, true))
	assert.InDelta(t, 0.0, testutil.ToFloat64(gauge), 0)
}
