package metrics

import (
	promexporter_metrics "github.com/d0ugal/promexporter/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// BrotherRegistry wraps the promexporter registry with Brother-specific metrics
type BrotherRegistry struct {
	*promexporter_metrics.Registry

	// Printer connection metrics
	PrinterConnectionStatus *prometheus.GaugeVec
	PrinterConnectionErrors *prometheus.CounterVec

	// Printer information
	PrinterInfo *prometheus.GaugeVec

	// Printer uptime
	PrinterUptime *prometheus.GaugeVec

	// Printer status
	PrinterStatus *prometheus.GaugeVec

	// Toner/Cartridge levels (for laser hosts)
	TonerLevel  *prometheus.GaugeVec
	TonerStatus *prometheus.GaugeVec

	// Ink levels (for inkjet hosts)
	InkLevel  *prometheus.GaugeVec
	InkStatus *prometheus.GaugeVec

	// Drum levels (for laser hosts)
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
}

// NewBrotherRegistry creates a new Brother metrics registry
func NewBrotherRegistry(baseRegistry *promexporter_metrics.Registry) *BrotherRegistry {
	// Get the underlying Prometheus registry
	promRegistry := baseRegistry.GetRegistry()
	factory := promauto.With(promRegistry)

	brother := &BrotherRegistry{
		Registry: baseRegistry,
	}

	// Printer connection metrics
	brother.PrinterConnectionStatus = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_connection_status",
			Help: "Brother host connection status (1=connected, 0=disconnected)",
		},
		[]string{"host"},
	)

	baseRegistry.AddMetricInfo("brother_printer_connection_status", "Brother host connection status (1=connected, 0=disconnected)", []string{"host"})

	brother.PrinterConnectionErrors = factory.NewCounterVec(
		prometheus.CounterOpts{
			Name: "brother_printer_connection_errors_total",
			Help: "Total number of connection errors to Brother host",
		},
		[]string{"host", "error_type"},
	)

	baseRegistry.AddMetricInfo("brother_printer_connection_errors_total", "Total number of connection errors to Brother host", []string{"host", "error_type"})

	// Printer information
	brother.PrinterInfo = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_info",
			Help: "Information about the Brother host",
		},
		[]string{"host", "model", "serial", "firmware", "type", "mac"},
	)

	baseRegistry.AddMetricInfo("brother_printer_info", "Information about the Brother host", []string{"host", "model", "serial", "firmware", "type", "mac"})

	// Printer uptime
	brother.PrinterUptime = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_restart_timestamp",
			Help: "Unix timestamp when Brother host was last restarted (use time() - brother_printer_restart_timestamp for uptime)",
		},
		[]string{"host"},
	)

	baseRegistry.AddMetricInfo("brother_printer_restart_timestamp", "Unix timestamp when Brother host was last restarted (use time() - brother_printer_restart_timestamp for uptime)", []string{"host"})

	// Printer status
	brother.PrinterStatus = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_status",
			Help: "Brother host status (1=ready, 0=not_ready)",
		},
		[]string{"host", "status"},
	)

	baseRegistry.AddMetricInfo("brother_printer_status", "Brother host status (1=ready, 0=not_ready)", []string{"host", "status"})

	// Toner/Cartridge levels (for laser hosts)
	brother.TonerLevel = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_toner_level_percent",
			Help: "Brother host toner level percentage",
		},
		[]string{"host", "color"},
	)

	baseRegistry.AddMetricInfo("brother_printer_toner_level_percent", "Brother host toner level percentage", []string{"host", "color"})

	brother.TonerStatus = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_toner_status",
			Help: "Brother host toner status (1=ok, 0=low/empty)",
		},
		[]string{"host", "color", "status"},
	)

	baseRegistry.AddMetricInfo("brother_printer_toner_status", "Brother host toner status (1=ok, 0=low/empty)", []string{"host", "color", "status"})

	// Ink levels (for inkjet hosts)
	brother.InkLevel = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_ink_level_percent",
			Help: "Brother host ink level percentage",
		},
		[]string{"host", "color"},
	)

	baseRegistry.AddMetricInfo("brother_printer_ink_level_percent", "Brother host ink level percentage", []string{"host", "color"})

	brother.InkStatus = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_ink_status",
			Help: "Brother host ink status (1=ok, 0=low/empty)",
		},
		[]string{"host", "color", "status"},
	)

	baseRegistry.AddMetricInfo("brother_printer_ink_status", "Brother host ink status (1=ok, 0=low/empty)", []string{"host", "color", "status"})

	// Drum levels (for laser hosts)
	brother.DrumLevel = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_drum_level_percent",
			Help: "Brother host drum level percentage",
		},
		[]string{"host", "color"},
	)

	baseRegistry.AddMetricInfo("brother_printer_drum_level_percent", "Brother host drum level percentage", []string{"host", "color"})

	brother.DrumStatus = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_drum_status",
			Help: "Brother host drum status (1=ok, 0=low/empty)",
		},
		[]string{"host", "color", "status"},
	)

	baseRegistry.AddMetricInfo("brother_printer_drum_status", "Brother host drum status (1=ok, 0=low/empty)", []string{"host", "color", "status"})

	// Paper tray status
	brother.PaperTrayStatus = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_paper_tray_status",
			Help: "Brother host paper tray status (1=ok, 0=empty/error)",
		},
		[]string{"host", "tray", "status"},
	)

	baseRegistry.AddMetricInfo("brother_printer_paper_tray_status", "Brother host paper tray status (1=ok, 0=empty/error)", []string{"host", "tray", "status"})

	// Page counters
	brother.PageCountTotal = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_pages",
			Help: "Total number of pages printed",
		},
		[]string{"host"},
	)

	baseRegistry.AddMetricInfo("brother_printer_pages", "Total number of pages printed", []string{"host"})

	brother.PageCountBlack = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_page_count_black",
			Help: "Number of black pages printed",
		},
		[]string{"host"},
	)

	baseRegistry.AddMetricInfo("brother_printer_page_count_black", "Number of black pages printed", []string{"host"})

	brother.PageCountColor = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_page_count_color",
			Help: "Number of color pages printed",
		},
		[]string{"host"},
	)

	baseRegistry.AddMetricInfo("brother_printer_page_count_color", "Number of color pages printed", []string{"host"})

	brother.PageCountDuplex = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_page_count_duplex",
			Help: "Number of duplex pages printed",
		},
		[]string{"host"},
	)

	baseRegistry.AddMetricInfo("brother_printer_page_count_duplex", "Number of duplex pages printed", []string{"host"})

	// Drum page counts
	brother.PageCountDrumBlack = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_page_count_drum_black",
			Help: "Number of pages printed with black drum",
		},
		[]string{"host"},
	)

	baseRegistry.AddMetricInfo("brother_printer_page_count_drum_black", "Number of pages printed with black drum", []string{"host"})

	brother.PageCountDrumCyan = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_page_count_drum_cyan",
			Help: "Number of pages printed with cyan drum",
		},
		[]string{"host"},
	)

	baseRegistry.AddMetricInfo("brother_printer_page_count_drum_cyan", "Number of pages printed with cyan drum", []string{"host"})

	brother.PageCountDrumMagenta = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_page_count_drum_magenta",
			Help: "Number of pages printed with magenta drum",
		},
		[]string{"host"},
	)

	baseRegistry.AddMetricInfo("brother_printer_page_count_drum_magenta", "Number of pages printed with magenta drum", []string{"host"})

	brother.PageCountDrumYellow = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_page_count_drum_yellow",
			Help: "Number of pages printed with yellow drum",
		},
		[]string{"host"},
	)

	baseRegistry.AddMetricInfo("brother_printer_page_count_drum_yellow", "Number of pages printed with yellow drum", []string{"host"})

	// Maintenance component life remaining (pages)
	brother.BeltUnitRemainingPages = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_belt_unit_remaining_pages",
			Help: "Belt unit remaining pages",
		},
		[]string{"host"},
	)

	baseRegistry.AddMetricInfo("brother_printer_belt_unit_remaining_pages", "Belt unit remaining pages", []string{"host"})

	brother.FuserUnitRemainingPages = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_fuser_unit_remaining_pages",
			Help: "Fuser unit remaining pages",
		},
		[]string{"host"},
	)

	baseRegistry.AddMetricInfo("brother_printer_fuser_unit_remaining_pages", "Fuser unit remaining pages", []string{"host"})

	brother.LaserUnitRemainingPages = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_laser_unit_remaining_pages",
			Help: "Laser unit remaining pages",
		},
		[]string{"host"},
	)

	baseRegistry.AddMetricInfo("brother_printer_laser_unit_remaining_pages", "Laser unit remaining pages", []string{"host"})

	brother.PaperFeedingKitRemainingPages = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_paper_feeding_kit_remaining_pages",
			Help: "Paper feeding kit remaining pages",
		},
		[]string{"host"},
	)

	baseRegistry.AddMetricInfo("brother_printer_paper_feeding_kit_remaining_pages", "Paper feeding kit remaining pages", []string{"host"})

	// Maintenance component life remaining (percentage)
	brother.BeltUnitRemainingPercent = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_belt_unit_remaining_percent",
			Help: "Belt unit remaining percentage",
		},
		[]string{"host"},
	)

	baseRegistry.AddMetricInfo("brother_printer_belt_unit_remaining_percent", "Belt unit remaining percentage", []string{"host"})

	brother.FuserUnitRemainingPercent = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_fuser_unit_remaining_percent",
			Help: "Fuser unit remaining percentage",
		},
		[]string{"host"},
	)

	baseRegistry.AddMetricInfo("brother_printer_fuser_unit_remaining_percent", "Fuser unit remaining percentage", []string{"host"})

	brother.LaserUnitRemainingPercent = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_laser_unit_remaining_percent",
			Help: "Laser unit remaining percentage",
		},
		[]string{"host"},
	)

	baseRegistry.AddMetricInfo("brother_printer_laser_unit_remaining_percent", "Laser unit remaining percentage", []string{"host"})

	brother.PaperFeedingKitRemainingPercent = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "brother_printer_paper_feeding_kit_remaining_percent",
			Help: "Paper feeding kit remaining percentage",
		},
		[]string{"host"},
	)

	baseRegistry.AddMetricInfo("brother_printer_paper_feeding_kit_remaining_percent", "Paper feeding kit remaining percentage", []string{"host"})

	// Maintenance counters
	brother.MaintenanceCount = factory.NewCounterVec(
		prometheus.CounterOpts{
			Name: "brother_printer_maintenance_count_total",
			Help: "Total number of maintenance operations",
		},
		[]string{"host", "operation"},
	)

	baseRegistry.AddMetricInfo("brother_printer_maintenance_count_total", "Total number of maintenance operations", []string{"host", "operation"})

	return brother
}
