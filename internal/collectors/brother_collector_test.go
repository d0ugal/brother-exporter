package collectors

import (
	"testing"

	"github.com/d0ugal/brother-exporter/internal/metrics"
	promexporter_metrics "github.com/d0ugal/promexporter/metrics"
	"github.com/prometheus/client_golang/prometheus"
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
