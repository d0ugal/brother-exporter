package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Registry holds all the metrics for the Brother printer exporter
type Registry struct {
	// Version info metric
	VersionInfo *prometheus.GaugeVec

	// Printer connection metrics
	PrinterConnectionStatus *prometheus.GaugeVec
	PrinterConnectionErrors *prometheus.CounterVec

	// Printer information
	PrinterInfo *prometheus.GaugeVec

	// Printer status
	PrinterStatus *prometheus.GaugeVec

	// Toner/Cartridge levels (for laser printers)
	TonerLevel  *prometheus.GaugeVec
	TonerStatus *prometheus.GaugeVec

	// Ink levels (for inkjet printers)
	InkLevel  *prometheus.GaugeVec
	InkStatus *prometheus.GaugeVec

	// Drum levels (for laser printers)
	DrumLevel  *prometheus.GaugeVec
	DrumStatus *prometheus.GaugeVec

	// Paper tray status
	PaperTrayStatus *prometheus.GaugeVec

	// Page counters (using standard MIB OIDs)
	PageCountTotal       *prometheus.GaugeVec
	PageCountBlack       *prometheus.GaugeVec
	PageCountColor       *prometheus.GaugeVec
	PageCountDuplex      *prometheus.GaugeVec
	PageCountDrumBlack   *prometheus.GaugeVec
	PageCountDrumCyan    *prometheus.GaugeVec
	PageCountDrumMagenta *prometheus.GaugeVec
	PageCountDrumYellow  *prometheus.GaugeVec

	// Maintenance component life remaining (pages)
	BeltUnitRemainingPages        *prometheus.GaugeVec
	FuserUnitRemainingPages       *prometheus.GaugeVec
	LaserUnitRemainingPages       *prometheus.GaugeVec
	PaperFeedingKitRemainingPages *prometheus.GaugeVec

	// Maintenance component life remaining (percentage)
	BeltUnitRemainingPercent        *prometheus.GaugeVec
	FuserUnitRemainingPercent       *prometheus.GaugeVec
	LaserUnitRemainingPercent       *prometheus.GaugeVec
	PaperFeedingKitRemainingPercent *prometheus.GaugeVec

	// Maintenance counters
	MaintenanceCount *prometheus.CounterVec

	// Metric information for UI
	metricInfo []MetricInfo
}

// MetricInfo contains information about a metric for the UI
type MetricInfo struct {
	Name         string
	Help         string
	Labels       []string
	ExampleValue string
}

// addMetricInfo adds metric information to the registry
func (r *Registry) addMetricInfo(name, help string, labels []string) {
	r.metricInfo = append(r.metricInfo, MetricInfo{
		Name:         name,
		Help:         help,
		Labels:       labels,
		ExampleValue: "",
	})
}

// NewRegistry creates a new metrics registry
func NewRegistry() *Registry {
	r := &Registry{}

	r.VersionInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_exporter_info",
			Help: "Information about the Brother printer exporter",
		},
		[]string{"version", "commit", "build_date"},
	)
	r.addMetricInfo("brother_exporter_info", "Information about the Brother printer exporter", []string{"version", "commit", "build_date"})

	r.PrinterConnectionStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_connection_status",
			Help: "Brother printer connection status (1 = connected, 0 = disconnected)",
		},
		[]string{"host"},
	)
	r.addMetricInfo("brother_printer_connection_status", "Brother printer connection status (1 = connected, 0 = disconnected)", []string{"host"})

	r.PrinterConnectionErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "brother_printer_connection_errors_total",
			Help: "Total number of Brother printer connection errors",
		},
		[]string{"host", "error_type"},
	)
	r.addMetricInfo("brother_printer_connection_errors_total", "Total number of Brother printer connection errors", []string{"host", "error_type"})

	r.PrinterInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_info",
			Help: "Information about the Brother printer",
		},
		[]string{"host", "model", "serial", "firmware", "type", "mac"},
	)
	r.addMetricInfo("brother_printer_info", "Information about the Brother printer", []string{"host", "model", "serial", "firmware", "type", "mac"})

	r.PrinterStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_status",
			Help: "Brother printer status (1 = ready, 0 = not ready)",
		},
		[]string{"host", "status"},
	)
	r.addMetricInfo("brother_printer_status", "Brother printer status (1 = ready, 0 = not ready)", []string{"host", "status"})

	r.TonerLevel = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_toner_level_percent",
			Help: "Brother printer toner level as a percentage",
		},
		[]string{"host", "color"},
	)
	r.addMetricInfo("brother_toner_level_percent", "Brother printer toner level as a percentage", []string{"host", "color"})

	r.TonerStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_toner_status",
			Help: "Brother printer toner status (1 = ok, 0 = low/empty)",
		},
		[]string{"host", "color", "status"},
	)
	r.addMetricInfo("brother_toner_status", "Brother printer toner status (1 = ok, 0 = low/empty)", []string{"host", "color", "status"})

	r.InkLevel = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_ink_level_percent",
			Help: "Brother printer ink level as a percentage",
		},
		[]string{"host", "color"},
	)
	r.addMetricInfo("brother_ink_level_percent", "Brother printer ink level as a percentage", []string{"host", "color"})

	r.InkStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_ink_status",
			Help: "Brother printer ink status (1 = ok, 0 = low/empty)",
		},
		[]string{"host", "color", "status"},
	)
	r.addMetricInfo("brother_ink_status", "Brother printer ink status (1 = ok, 0 = low/empty)", []string{"host", "color", "status"})

	r.DrumLevel = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_drum_level_percent",
			Help: "Brother printer drum level as a percentage",
		},
		[]string{"host", "color"},
	)
	r.addMetricInfo("brother_drum_level_percent", "Brother printer drum level as a percentage", []string{"host", "color"})

	r.DrumStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_drum_status",
			Help: "Brother printer drum status (1 = ok, 0 = low/empty)",
		},
		[]string{"host", "color", "status"},
	)
	r.addMetricInfo("brother_drum_status", "Brother printer drum status (1 = ok, 0 = low/empty)", []string{"host", "color", "status"})

	r.PaperTrayStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_paper_tray_status",
			Help: "Brother printer paper tray status (1 = ok, 0 = empty/error)",
		},
		[]string{"host", "tray", "status"},
	)
	r.addMetricInfo("brother_paper_tray_status", "Brother printer paper tray status (1 = ok, 0 = empty/error)", []string{"host", "tray", "status"})

	// Page counters (using standard MIB OIDs)
	r.PageCountTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_pages",
			Help: "Total number of pages printed",
		},
		[]string{"host"},
	)
	r.addMetricInfo("brother_pages", "Total number of pages printed", []string{"host"})

	r.PageCountBlack = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_page_count_black",
			Help: "Total number of black and white pages printed",
		},
		[]string{"host"},
	)
	r.addMetricInfo("brother_page_count_black", "Total number of black and white pages printed", []string{"host"})

	r.PageCountColor = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_page_count_color",
			Help: "Total number of color pages printed",
		},
		[]string{"host"},
	)
	r.addMetricInfo("brother_page_count_color", "Total number of color pages printed", []string{"host"})

	r.PageCountDuplex = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_page_count_duplex",
			Help: "Total number of duplex pages printed",
		},
		[]string{"host"},
	)
	r.addMetricInfo("brother_page_count_duplex", "Total number of duplex pages printed", []string{"host"})

	r.PageCountDrumBlack = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_page_count_drum_black",
			Help: "Total number of pages printed with black drum",
		},
		[]string{"host"},
	)
	r.addMetricInfo("brother_page_count_drum_black", "Total number of pages printed with black drum", []string{"host"})

	r.PageCountDrumCyan = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_page_count_drum_cyan",
			Help: "Total number of pages printed with cyan drum",
		},
		[]string{"host"},
	)
	r.addMetricInfo("brother_page_count_drum_cyan", "Total number of pages printed with cyan drum", []string{"host"})

	r.PageCountDrumMagenta = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_page_count_drum_magenta",
			Help: "Total number of pages printed with magenta drum",
		},
		[]string{"host"},
	)
	r.addMetricInfo("brother_page_count_drum_magenta", "Total number of pages printed with magenta drum", []string{"host"})

	r.PageCountDrumYellow = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_page_count_drum_yellow",
			Help: "Total number of pages printed with yellow drum",
		},
		[]string{"host"},
	)
	r.addMetricInfo("brother_page_count_drum_yellow", "Total number of pages printed with yellow drum", []string{"host"})

	// Maintenance component life remaining (pages)
	r.BeltUnitRemainingPages = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_belt_unit_remaining_pages",
			Help: "Belt unit remaining life in pages",
		},
		[]string{"host"},
	)
	r.addMetricInfo("brother_belt_unit_remaining_pages", "Belt unit remaining life in pages", []string{"host"})

	r.FuserUnitRemainingPages = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_fuser_unit_remaining_pages",
			Help: "Fuser unit remaining life in pages",
		},
		[]string{"host"},
	)
	r.addMetricInfo("brother_fuser_unit_remaining_pages", "Fuser unit remaining life in pages", []string{"host"})

	r.LaserUnitRemainingPages = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_laser_unit_remaining_pages",
			Help: "Laser unit remaining life in pages",
		},
		[]string{"host"},
	)
	r.addMetricInfo("brother_laser_unit_remaining_pages", "Laser unit remaining life in pages", []string{"host"})

	r.PaperFeedingKitRemainingPages = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_paper_feeding_kit_remaining_pages",
			Help: "Paper feeding kit remaining life in pages",
		},
		[]string{"host"},
	)
	r.addMetricInfo("brother_paper_feeding_kit_remaining_pages", "Paper feeding kit remaining life in pages", []string{"host"})

	// Maintenance component life remaining (percentage)
	r.BeltUnitRemainingPercent = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_belt_unit_remaining_percent",
			Help: "Belt unit remaining life as a percentage",
		},
		[]string{"host"},
	)
	r.addMetricInfo("brother_belt_unit_remaining_percent", "Belt unit remaining life as a percentage", []string{"host"})

	r.FuserUnitRemainingPercent = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_fuser_unit_remaining_percent",
			Help: "Fuser unit remaining life as a percentage",
		},
		[]string{"host"},
	)
	r.addMetricInfo("brother_fuser_unit_remaining_percent", "Fuser unit remaining life as a percentage", []string{"host"})

	r.LaserUnitRemainingPercent = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_laser_unit_remaining_percent",
			Help: "Laser unit remaining life as a percentage",
		},
		[]string{"host"},
	)
	r.addMetricInfo("brother_laser_unit_remaining_percent", "Laser unit remaining life as a percentage", []string{"host"})

	r.PaperFeedingKitRemainingPercent = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_paper_feeding_kit_remaining_percent",
			Help: "Paper feeding kit remaining life as a percentage",
		},
		[]string{"host"},
	)
	r.addMetricInfo("brother_paper_feeding_kit_remaining_percent", "Paper feeding kit remaining life as a percentage", []string{"host"})

	r.MaintenanceCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "brother_maintenance_count_total",
			Help: "Total number of maintenance operations",
		},
		[]string{"host", "type"},
	)
	r.addMetricInfo("brother_maintenance_count_total", "Total number of maintenance operations", []string{"host", "type"})

	return r
}

// GetMetricsInfo returns information about all metrics for the UI
func (r *Registry) GetMetricsInfo() []MetricInfo {
	return r.metricInfo
}

// GetRegistry returns the Prometheus registry
func (r *Registry) GetRegistry() *prometheus.Registry {
	return prometheus.DefaultRegisterer.(*prometheus.Registry)
}
