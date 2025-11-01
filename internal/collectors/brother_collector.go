package collectors

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/d0ugal/brother-exporter/internal/config"
	"github.com/d0ugal/brother-exporter/internal/metrics"
	"github.com/d0ugal/promexporter/app"
	"github.com/d0ugal/promexporter/tracing"
	"github.com/gosnmp/gosnmp"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
)

// convertToInt converts various SNMP value types to int using Go generics
// This eliminates the duplicated type switch statements throughout the code
func convertToInt[T any](value T, context string) (int, bool) {
	switch v := any(value).(type) {
	case int:
		return v, true
	case uint:
		if v > math.MaxInt {
			return 0, false
		}

		return int(v), true
	case int32:
		return int(v), true
	case uint32:
		return int(v), true
	case int64:
		return int(v), true
	case uint64:
		if v > math.MaxInt {
			return 0, false
		}

		return int(v), true
	default:
		slog.Debug("Unexpected type for "+context, "type", fmt.Sprintf("%T", v), "value", v)
		return 0, false
	}
}

// bytesToHexString converts a byte slice to a hex string, excluding the last byte (checksum)
func bytesToHexString(bytes []uint8) string {
	if len(bytes) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.Grow((len(bytes) - 1) * 2) // Pre-allocate capacity for efficiency

	for i := 0; i < len(bytes)-1; i++ {
		builder.WriteString(fmt.Sprintf("%02x", bytes[i]))
	}

	return builder.String()
}

// splitIntoChunks splits a string into chunks of the specified size
func splitIntoChunks(s string, chunkSize int) []string {
	if chunkSize <= 0 {
		return []string{s}
	}

	chunks := make([]string, 0, (len(s)+chunkSize-1)/chunkSize)

	for i := 0; i < len(s); i += chunkSize {
		end := i + chunkSize
		if end > len(s) {
			end = len(s)
		}

		chunks = append(chunks, s[i:end])
	}

	return chunks
}

// calculateStatusFromLevel calculates status string and value based on percentage level
func calculateStatusFromLevel(level float64) (status string, statusValue float64) {
	status = "ok"
	statusValue = 1.0

	if level < BrotherLowThreshold {
		status = "low"
		statusValue = 0.0
	} else if level == 0 {
		status = "empty"
		statusValue = 0.0
	}

	return status, statusValue
}

// handleCollectionError is a helper function for the repetitive error handling pattern in collectMetrics
func (bc *BrotherCollector) handleCollectionError(err error, operation string) {
	if err != nil {
		slog.Error("Failed to collect "+operation, "error", err)
		bc.metrics.PrinterConnectionErrors.With(prometheus.Labels{
			"host":      bc.config.Printer.Host,
			"operation": operation,
		}).Inc()
	}
}

// collectColorLevelsWithStatus collects level and status metrics for each color using the specified OID base
func (bc *BrotherCollector) collectColorLevelsWithStatus(ctx context.Context, oidBase string, colors []string, levelMetric, statusMetric *prometheus.GaugeVec, contextName string) {
	tracer := bc.app.GetTracer()

	var span *tracing.CollectorSpan

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "brother-collector", "collect-color-levels")

		span.SetAttributes(
			attribute.String("oid.base", oidBase),
			attribute.String("context", contextName),
			attribute.Int("colors.count", len(colors)),
		)

		defer span.End()
	}

	collectStart := time.Now()
	var colorsCollected int

	for i, color := range colors {
		oid := fmt.Sprintf("%s.%d", oidBase, i+1)

		getStart := time.Now()

		result, err := bc.client.Get([]string{oid})

		getDuration := time.Since(getStart)

		if err != nil {
			slog.Debug("Failed to get "+contextName, "color", color, "oid", oid, "error", err)

			if span != nil {
				span.RecordError(err, attribute.String("color", color), attribute.String("oid", oid))
			}

			continue
		}

		if len(result.Variables) > 0 {
			parseStart := time.Now()

			level, ok := convertToInt(result.Variables[0].Value, contextName)
			if !ok {
				if span != nil {
					span.RecordError(fmt.Errorf("failed to convert level"), attribute.String("color", color))
				}

				continue
			}

			// Convert to percentage and cap at 100
			percentage := float64(level)
			if percentage > 100 {
				percentage = 100
			}

			parseDuration := time.Since(parseStart)

			// Set level metric
			levelMetric.With(prometheus.Labels{
				"host":  bc.config.Printer.Host,
				"color": color,
			}).Set(percentage)

			// Set status based on level
			status, statusValue := calculateStatusFromLevel(percentage)

			statusMetric.With(prometheus.Labels{
				"host":   bc.config.Printer.Host,
				"color":  color,
				"status": status,
			}).Set(statusValue)

			colorsCollected++

			if span != nil {
				span.SetAttributes(
					attribute.Float64("color."+color+".get_duration_seconds", getDuration.Seconds()),
					attribute.Float64("color."+color+".parse_duration_seconds", parseDuration.Seconds()),
					attribute.Float64("color."+color+".percentage", percentage),
					attribute.String("color."+color+".status", status),
				)
			}
		}
	}

	collectDuration := time.Since(collectStart)

	if span != nil {
		span.SetAttributes(
			attribute.Int("collect.colors_collected", colorsCollected),
			attribute.Float64("collect.duration_seconds", collectDuration.Seconds()),
		)
		span.AddEvent("color_levels_collected",
			attribute.Int("colors_collected", colorsCollected),
		)
	}
}

// BrotherCollector collects metrics from Brother printers via SNMP
type BrotherCollector struct {
	config  *config.Config
	metrics *metrics.BrotherRegistry
	app     *app.App
	client  *gosnmp.GoSNMP
	mu      sync.RWMutex
	done    chan struct{}
}

// Brother printer SNMP OIDs
const (
	// OIDSystemDescription is the system information OID
	OIDSystemDescription = "1.3.6.1.2.1.1.1.0"
	OIDSystemUpTime      = "1.3.6.1.2.1.1.3.0"
	OIDSystemContact     = "1.3.6.1.2.1.1.4.0"
	OIDSystemName        = "1.3.6.1.2.1.1.5.0"
	OIDSystemLocation    = "1.3.6.1.2.1.1.6.0"

	// OIDBrotherBase is the Brother specific OIDs base
	OIDBrotherBase = "1.3.6.1.4.1.2435"

	// OIDPrinterStatus is the printer status OID
	OIDPrinterStatus = "1.3.6.1.2.1.25.3.2.1.5.1"

	// OIDBrotherConsumableInfo and related OIDs are Brother-specific consumable OIDs (these work better than standard MIB)
	OIDBrotherConsumableInfo  = "1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.1.0"  // Consumable info
	OIDBrotherConsumableLevel = "1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.4.0"  // Consumable level (104%)
	OIDBrotherStatus          = "1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.7.0"  // Status (1)
	OIDBrotherFirmware        = "1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.17.0" // Firmware version
	OIDBrotherMaintenanceData = "1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.8.0"  // Maintenance data
	OIDBrotherCountersData    = "1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.10.0" // Counters data
	OIDBrotherNextCareData    = "1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.11.0" // Nextcare data

	// OIDBrotherModel and related OIDs are Brother-specific printer info OIDs
	OIDBrotherModel  = "1.3.6.1.4.1.2435.2.3.9.1.1.7.0"       // Model
	OIDBrotherSerial = "1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.1.0" // Serial number
	OIDBrotherMAC    = "1.3.6.1.2.1.2.2.1.6.1"                // MAC address
	OIDBrotherUptime = "1.3.6.1.2.1.1.3.0"                    // System uptime (hundredths of seconds)

	// OIDTonerLevelBase and related OIDs are standard MIB OIDs (these return -2/-3 for Brother printers)
	OIDTonerLevelBase      = "1.3.6.1.2.1.43.11.1.1.9.1"
	OIDDrumLevelBase       = "1.3.6.1.2.1.43.11.1.1.8.1"
	OIDPaperTrayStatusBase = "1.3.6.1.2.1.43.8.2.1.10.1"

	// OIDPageCountTotal is a page counter OID (standard MIB - these work reliably)
	OIDPageCountTotal = "1.3.6.1.2.1.43.10.2.1.4.1.1" // Standard MIB total page count
)

// Brother data parsing constants
const (
	BrotherChunkSize     = 14  // Size of data chunks in Brother hex strings
	BrotherPercentageDiv = 100 // Divisor for percentage values from Brother data
	BrotherLowThreshold  = 10  // Threshold for "low" status (percentage)
)

// Color mappings for Brother printers
var (
	LaserColors = []string{"black", "cyan", "magenta", "yellow"}
	InkColors   = []string{"black", "cyan", "magenta", "yellow"}
)

func NewBrotherCollector(cfg *config.Config, metricsRegistry *metrics.BrotherRegistry, app *app.App) *BrotherCollector {
	return &BrotherCollector{
		config:  cfg,
		metrics: metricsRegistry,
		app:     app,
		done:    make(chan struct{}),
	}
}

func (bc *BrotherCollector) Start(ctx context.Context) {
	go bc.run(ctx)
}

// run handles the main collection loop
func (bc *BrotherCollector) run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(bc.config.GetDefaultInterval()) * time.Second)
	defer ticker.Stop()

	// Initial collection
	bc.collectMetrics(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-bc.done:
			return
		case <-ticker.C:
			bc.collectMetrics(ctx)
		}
	}
}

// collectMetrics performs a single metrics collection cycle
func (bc *BrotherCollector) collectMetrics(ctx context.Context) {
	startTime := time.Now()

	// Create span for collection cycle
	tracer := bc.app.GetTracer()

	var collectorSpan *tracing.CollectorSpan

	if tracer != nil && tracer.IsEnabled() {
		collectorSpan = tracer.NewCollectorSpan(ctx, "brother-collector", "collect-metrics")

		collectorSpan.SetAttributes(
			attribute.String("printer.host", bc.config.Printer.Host),
			attribute.String("printer.type", bc.config.Printer.Type),
		)
		defer collectorSpan.End()
	}

	var spanCtx context.Context //nolint:contextcheck // Extracting context from span for child operations

	if collectorSpan != nil {
		spanCtx = collectorSpan.Context() //nolint:contextcheck // Standard OpenTelemetry pattern: extract context from span
	} else {
		spanCtx = ctx
	}

	if err := bc.connect(spanCtx); err != nil {
		slog.Error("Failed to connect to Brother printer",
			"host", bc.config.Printer.Host,
			"error", err,
		)

		if collectorSpan != nil {
			collectorSpan.RecordError(err, attribute.String("printer.host", bc.config.Printer.Host))
		}

		bc.metrics.PrinterConnectionStatus.With(prometheus.Labels{
			"host": bc.config.Printer.Host,
		}).Set(0)
		bc.metrics.PrinterConnectionErrors.With(prometheus.Labels{
			"host":      bc.config.Printer.Host,
			"operation": "connect",
		}).Inc()

		return
	}

	defer bc.disconnect(spanCtx)

	// Set connection status
	bc.metrics.PrinterConnectionStatus.With(prometheus.Labels{
		"host": bc.config.Printer.Host,
	}).Set(1)

	// Reuse spanCtx from above for child operations

	// Collect printer information
	bc.handleCollectionError(bc.collectPrinterInfo(spanCtx), "printer info")

	// Collect printer status
	bc.handleCollectionError(bc.collectPrinterStatus(spanCtx), "printer status")

	// Collect printer uptime
	bc.handleCollectionError(bc.collectPrinterUptime(spanCtx), "printer uptime")

	// Collect Brother-specific metrics (these work better than standard MIB)
	if err := bc.collectBrotherSpecificMetrics(spanCtx); err != nil {
		bc.handleCollectionError(err, "brother_metrics")

		// Fallback to standard MIB only if Brother-specific collection fails
		switch bc.config.Printer.Type {
		case "laser":
			bc.handleCollectionError(bc.collectLaserMetrics(spanCtx), "laser_metrics")
		case "ink":
			bc.handleCollectionError(bc.collectInkjetMetrics(spanCtx), "inkjet_metrics")
		}
	} else {
		// If Brother-specific collection succeeded, also collect nextcare data
		bc.handleCollectionError(bc.collectBrotherNextCareData(spanCtx), "nextcare_metrics")
	}

	// Collect paper tray status
	bc.handleCollectionError(bc.collectPaperTrayStatus(spanCtx), "paper_tray")

	// Collect page counters using standard MIB OIDs
	bc.handleCollectionError(bc.collectPageCounters(spanCtx), "page_counters")

	duration := time.Since(startTime).Seconds()

	if collectorSpan != nil {
		collectorSpan.SetAttributes(
			attribute.Float64("collection.duration_seconds", duration),
		)
		collectorSpan.AddEvent("collection_completed",
			attribute.String("printer.host", bc.config.Printer.Host),
			attribute.Float64("duration_seconds", duration),
		)
	}

	slog.Info("Collection cycle completed", "host", bc.config.Printer.Host, "duration", duration)
}

// connect establishes SNMP connection to the printer
func (bc *BrotherCollector) connect(ctx context.Context) error {
	tracer := bc.app.GetTracer()

	var span *tracing.CollectorSpan

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "brother-collector", "connect")

		span.SetAttributes(
			attribute.String("snmp.host", bc.config.Printer.Host),
			attribute.Int("snmp.port", 161),
			attribute.String("snmp.version", "v2c"),
			attribute.Int("snmp.timeout_seconds", 10),
			attribute.Int("snmp.retries", 3),
		)

		defer span.End()
	}

	configStart := time.Now()

	bc.mu.Lock()
	defer bc.mu.Unlock()

	bc.client = &gosnmp.GoSNMP{
		Target:    bc.config.Printer.Host,
		Port:      161,
		Community: bc.config.Printer.Community,
		Version:   gosnmp.Version2c,
		Timeout:   time.Duration(10) * time.Second,
		Retries:   3,
	}

	configDuration := time.Since(configStart)

	if span != nil {
		span.SetAttributes(
			attribute.Float64("config.duration_seconds", configDuration.Seconds()),
		)
		span.AddEvent("client_configured")
	}

	connectStart := time.Now()

	err := bc.client.Connect()

	connectDuration := time.Since(connectStart)

	if span != nil {
		span.SetAttributes(
			attribute.Float64("connect.duration_seconds", connectDuration.Seconds()),
			attribute.Bool("connect.success", err == nil),
		)

		if err != nil {
			span.RecordError(err, attribute.String("operation", "snmp_connect"))
		} else {
			span.AddEvent("connection_established",
				attribute.String("host", bc.config.Printer.Host),
			)
		}
	}

	return err
}

// disconnect closes the SNMP connection
func (bc *BrotherCollector) disconnect(ctx context.Context) {
	tracer := bc.app.GetTracer()

	var span *tracing.CollectorSpan

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "brother-collector", "disconnect")
		defer span.End()
	}

	disconnectStart := time.Now()

	bc.mu.Lock()
	defer bc.mu.Unlock()

	if bc.client != nil {
		if err := bc.client.Conn.Close(); err != nil {
			slog.Debug("Error closing SNMP connection", "error", err)

			if span != nil {
				span.RecordError(err, attribute.String("operation", "snmp_disconnect"))
			}
		}

		bc.client = nil
	}

	disconnectDuration := time.Since(disconnectStart)

	if span != nil {
		span.SetAttributes(
			attribute.Float64("disconnect.duration_seconds", disconnectDuration.Seconds()),
		)
		span.AddEvent("connection_closed")
	}
}

// collectPrinterInfo collects basic printer information using Brother-specific OIDs
func (bc *BrotherCollector) collectPrinterInfo(ctx context.Context) error {
	tracer := bc.app.GetTracer()

	var span *tracing.CollectorSpan

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "brother-collector", "collect-printer-info")

		span.SetAttributes(
			attribute.Int("oids.count", 4),
		)

		defer span.End()
	}

	oids := []string{
		OIDBrotherModel,
		OIDBrotherSerial,
		OIDBrotherFirmware,
		OIDBrotherMAC,
	}

	getStart := time.Now()

	result, err := bc.client.Get(oids)

	getDuration := time.Since(getStart)

	if span != nil {
		span.SetAttributes(
			attribute.StringSlice("oids", oids),
			attribute.Float64("get.duration_seconds", getDuration.Seconds()),
			attribute.Bool("get.success", err == nil),
		)

		if err != nil {
			span.RecordError(err, attribute.String("operation", "snmp_get"))
			return fmt.Errorf("failed to get Brother printer info: %w", err)
		}
	} else {
		if err != nil {
			return fmt.Errorf("failed to get Brother printer info: %w", err)
		}
	}
	var model, serial, firmware, mac string
	parseStart := time.Now()

	for _, variable := range result.Variables {
		if variable.Value == nil {
			slog.Info("Variable has nil value", "name", variable.Name)
			continue
		}

		value := string(variable.Value.([]byte))

		// Remove leading dot from OID name for comparison
		oidName := strings.TrimPrefix(variable.Name, ".")
		switch oidName {
		case OIDBrotherModel:
			slog.Debug("Processing Brother model", "raw_value", value, "contains_MDL", strings.Contains(value, "MDL:"))
			// Extract model from the long string (look for MDL:...;)
			if strings.Contains(value, "MDL:") {
				start := strings.Index(value, "MDL:") + 4

				end := strings.Index(value[start:], ";")
				if end != -1 {
					model = strings.TrimSpace(value[start : start+end])
					// Remove " series" suffix if present
					model = strings.TrimSuffix(model, " series")
					slog.Debug("Extracted model from MDL", "start", start, "end", end, "model", model)
				} else {
					model = strings.TrimSpace(value[start:])
					model = strings.TrimSuffix(model, " series")
					slog.Debug("Using model from MDL start", "model", model)
				}
			} else {
				model = strings.TrimSpace(value)
				model = strings.TrimSuffix(model, " series")
				slog.Debug("Using raw model value", "model", model)
			}
		case OIDBrotherSerial:
			serial = strings.TrimSpace(value)
			slog.Debug("Processing Brother serial", "raw_value", value, "serial", serial)
		case OIDBrotherFirmware:
			firmware = strings.TrimSpace(value)
			slog.Debug("Processing Brother firmware", "raw_value", value, "firmware", firmware)
		case OIDBrotherMAC:
			slog.Debug("Processing Brother MAC", "raw_value", value, "length", len(value), "type", fmt.Sprintf("%T", variable.Value))
			// Convert MAC address bytes to hex string format
			if macBytes, ok := variable.Value.([]byte); ok && len(macBytes) == 6 {
				mac = fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
					macBytes[0], macBytes[1], macBytes[2],
					macBytes[3], macBytes[4], macBytes[5])
				slog.Debug("Formatted MAC address from bytes", "mac", mac)
			} else {
				mac = strings.TrimSpace(value)
				slog.Debug("Using MAC as string", "mac", mac)
			}
		default:
			slog.Debug("Unknown OID", "name", variable.Name, "value", value)
		}
	}

	// Set printer info metric
	bc.metrics.PrinterInfo.With(prometheus.Labels{
		"host":     bc.config.Printer.Host,
		"model":    model,
		"serial":   serial,
		"firmware": firmware,
		"type":     bc.config.Printer.Type,
		"mac":      mac,
	}).Set(1)

	parseDuration := time.Since(parseStart)

	if span != nil {
		span.SetAttributes(
			attribute.String("printer.model", model),
			attribute.String("printer.serial", serial),
			attribute.String("printer.firmware", firmware),
			attribute.String("printer.mac", mac),
			attribute.Float64("parse.duration_seconds", parseDuration.Seconds()),
		)
		span.AddEvent("printer_info_collected",
			attribute.String("model", model),
		)
	}

	slog.Debug("Printer info collected",
		"model", model,
		"serial", serial,
		"firmware", firmware,
		"mac", mac)

	return nil
}

// collectPrinterUptime collects printer uptime information
func (bc *BrotherCollector) collectPrinterUptime(ctx context.Context) error {
	tracer := bc.app.GetTracer()

	var span *tracing.CollectorSpan

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "brother-collector", "collect-printer-uptime")

		span.SetAttributes(
			attribute.String("oid", OIDBrotherUptime),
		)

		defer span.End()
	}

	getStart := time.Now()

	result, err := bc.client.Get([]string{OIDBrotherUptime})

	getDuration := time.Since(getStart)

	if span != nil {
		span.SetAttributes(
			attribute.Float64("get.duration_seconds", getDuration.Seconds()),
			attribute.Bool("get.success", err == nil),
		)

		if err != nil {
			span.RecordError(err, attribute.String("operation", "snmp_get"))
			return fmt.Errorf("failed to get printer uptime: %w", err)
		}
	} else {
		if err != nil {
			return fmt.Errorf("failed to get printer uptime: %w", err)
		}
	}
	if len(result.Variables) == 0 || result.Variables[0].Value == nil {
		err := fmt.Errorf("no uptime data received")

		if span != nil {
			span.RecordError(err, attribute.String("operation", "parse_uptime"))
		}

		return err
	}

	variable := result.Variables[0]
	parseStart := time.Now()

	// The uptime OID returns time in hundredths of a second
	// We need to convert it to seconds
	uptimeHundredths, ok := convertToInt(variable.Value, "uptime")
	if !ok {
		err := fmt.Errorf("failed to convert uptime value: %T", variable.Value)

		if span != nil {
			span.RecordError(err, attribute.String("operation", "convert_uptime"))
		}

		return err
	}

	// Convert from hundredths of seconds to seconds
	uptimeSeconds := float64(uptimeHundredths) / 100.0

	// Calculate the Unix timestamp when the printer was last restarted
	// by subtracting the uptime from the current time
	currentTime := float64(time.Now().Unix())
	restartTimestamp := currentTime - uptimeSeconds

	parseDuration := time.Since(parseStart)

	// Set the uptime metric with the restart timestamp
	bc.metrics.PrinterUptime.With(prometheus.Labels{
		"host": bc.config.Printer.Host,
	}).Set(restartTimestamp)

	if span != nil {
		span.SetAttributes(
			attribute.Int64("uptime.hundredths", int64(uptimeHundredths)),
			attribute.Float64("uptime.seconds", uptimeSeconds),
			attribute.Float64("uptime.restart_timestamp", restartTimestamp),
			attribute.Float64("uptime.current_timestamp", currentTime),
			attribute.Float64("parse.duration_seconds", parseDuration.Seconds()),
		)
		span.AddEvent("uptime_collected",
			attribute.Float64("uptime_seconds", uptimeSeconds),
		)
	}

	slog.Debug("Printer uptime collected", "uptime_seconds", uptimeSeconds, "restart_timestamp", restartTimestamp, "current_time", currentTime)

	return nil
}

// collectPrinterStatus collects printer status information
func (bc *BrotherCollector) collectPrinterStatus(ctx context.Context) error {
	tracer := bc.app.GetTracer()

	var span *tracing.CollectorSpan

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "brother-collector", "collect-printer-status")

		span.SetAttributes(
			attribute.String("oid", OIDPrinterStatus),
		)

		defer span.End()
	}

	getStart := time.Now()

	result, err := bc.client.Get([]string{OIDPrinterStatus})

	getDuration := time.Since(getStart)

	if span != nil {
		span.SetAttributes(
			attribute.Float64("get.duration_seconds", getDuration.Seconds()),
			attribute.Bool("get.success", err == nil),
		)

		if err != nil {
			span.RecordError(err, attribute.String("operation", "snmp_get"))
			return fmt.Errorf("failed to get printer status: %w", err)
		}
	} else {
		if err != nil {
			return fmt.Errorf("failed to get printer status: %w", err)
		}
	}
	if len(result.Variables) > 0 {
		parseStart := time.Now()

		status, ok := convertToInt(result.Variables[0].Value, "printer status")
		if !ok {
			if span != nil {
				span.SetAttributes(
					attribute.Bool("parse.success", false),
				)
				span.RecordError(fmt.Errorf("failed to convert status value"), attribute.String("operation", "convert_status"))
			}

			return nil
		}

		// Map status codes to strings (these may vary by printer model)
		var statusStr string

		switch status {
		case 3: // idle
			statusStr = "ready"
		case 4: // printing
			statusStr = "printing"
		case 5: // warmup
			statusStr = "warmup"
		default:
			statusStr = "unknown"
		}

		// Set status metric (1 for ready, 0 for others)
		statusValue := 0.0
		if statusStr == "ready" {
			statusValue = 1.0
		}

		parseDuration := time.Since(parseStart)

		bc.metrics.PrinterStatus.With(prometheus.Labels{
			"host":   bc.config.Printer.Host,
			"status": statusStr,
		}).Set(statusValue)

		if span != nil {
			span.SetAttributes(
				attribute.Int("status.code", status),
				attribute.String("status.string", statusStr),
				attribute.Float64("status.value", statusValue),
				attribute.Float64("parse.duration_seconds", parseDuration.Seconds()),
				attribute.Bool("parse.success", true),
			)
			span.AddEvent("status_collected",
				attribute.String("status", statusStr),
			)
		}
	}

	return nil
}

// collectBrotherSpecificMetrics collects Brother-specific metrics using the proper OIDs and decoding
func (bc *BrotherCollector) collectBrotherSpecificMetrics(ctx context.Context) error {
	tracer := bc.app.GetTracer()

	var (
		span    *tracing.CollectorSpan
		spanCtx context.Context //nolint:contextcheck // Extracting context from span for child operations
	)

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "brother-collector", "collect-brother-specific-metrics")

		span.SetAttributes(
			attribute.String("printer.type", bc.config.Printer.Type),
		)

		spanCtx = span.Context() //nolint:contextcheck // Standard OpenTelemetry pattern: extract context from span
		defer span.End()
	} else {
		spanCtx = ctx
	}

	collectStart := time.Now()

	// Get maintenance data (contains toner and drum levels)
	if err := bc.collectBrotherMaintenanceData(spanCtx); err != nil {
		slog.Error("Failed to collect Brother maintenance data", "error", err)

		if span != nil {
			span.RecordError(err, attribute.String("operation", "maintenance_data"))
		}
	}

	// Get counters data (contains page counts)
	if err := bc.collectBrotherCountersData(spanCtx); err != nil {
		slog.Error("Failed to collect Brother counters data", "error", err)

		if span != nil {
			span.RecordError(err, attribute.String("operation", "counters_data"))
		}
	}

	// Get basic info
	oids := []string{
		OIDBrotherConsumableInfo, // Consumable info (E83216M3N204406)
		OIDBrotherFirmware,       // Firmware version (1.16)
	}

	getStart := time.Now()

	result, err := bc.client.Get(oids)

	getDuration := time.Since(getStart)

	if span != nil {
		span.SetAttributes(
			attribute.StringSlice("oids", oids),
			attribute.Float64("get.duration_seconds", getDuration.Seconds()),
			attribute.Bool("get.success", err == nil),
		)

		if err != nil {
			span.RecordError(err, attribute.String("operation", "snmp_get_basic_info"))
			return fmt.Errorf("failed to get Brother basic info: %w", err)
		}
	} else {
		if err != nil {
			return fmt.Errorf("failed to get Brother basic info: %w", err)
		}
	}

	for i, variable := range result.Variables {
		if variable.Value == nil {
			continue
		}

		switch i {
		case 0: // Consumable info
			if bytes, ok := variable.Value.([]uint8); ok {
				info := string(bytes)
				slog.Debug("Brother consumable info", "info", info)

				if span != nil {
					span.SetAttributes(
						attribute.String("consumable.info", info),
					)
				}
			}
		case 1: // Firmware
			if bytes, ok := variable.Value.([]uint8); ok {
				firmware := string(bytes)
				slog.Debug("Brother firmware", "firmware", firmware)

				if span != nil {
					span.SetAttributes(
						attribute.String("consumable.firmware", firmware),
					)
				}
			}
		}
	}

	collectDuration := time.Since(collectStart)

	if span != nil {
		span.SetAttributes(
			attribute.Float64("collect.duration_seconds", collectDuration.Seconds()),
		)
		span.AddEvent("brother_metrics_collected")
	}

	return nil
}

// collectBrotherMaintenanceData extracts toner and drum levels from Brother maintenance data
func (bc *BrotherCollector) collectBrotherMaintenanceData(ctx context.Context) error {
	tracer := bc.app.GetTracer()

	var span *tracing.CollectorSpan

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "brother-collector", "collect-maintenance-data")

		span.SetAttributes(
			attribute.String("oid", OIDBrotherMaintenanceData),
		)

		defer span.End()
	}

	collectStart := time.Now()
	getStart := time.Now()

	result, err := bc.client.Get([]string{OIDBrotherMaintenanceData})

	getDuration := time.Since(getStart)

	if span != nil {
		span.SetAttributes(
			attribute.Float64("get.duration_seconds", getDuration.Seconds()),
			attribute.Bool("get.success", err == nil),
		)

		if err != nil {
			span.RecordError(err, attribute.String("operation", "snmp_get"))
			return fmt.Errorf("failed to get Brother maintenance data: %w", err)
		}
	} else {
		if err != nil {
			return fmt.Errorf("failed to get Brother maintenance data: %w", err)
		}
	}

	if len(result.Variables) == 0 || result.Variables[0].Value == nil {
		err := fmt.Errorf("no maintenance data received")

		if span != nil {
			span.RecordError(err, attribute.String("operation", "parse_maintenance_data"))
		}

		return err
	}

	variable := result.Variables[0]
	parseStart := time.Now()

	bytes, ok := variable.Value.([]uint8)
	if !ok {
		err := fmt.Errorf("maintenance data is not a byte array: %T", variable.Value)

		if span != nil {
			span.RecordError(err, attribute.String("operation", "convert_maintenance_data"))
		}

		return err
	}

	// Convert bytes to hex string (excluding last byte which is checksum)
	hexString := bytesToHexString(bytes)

	if span != nil {
		span.SetAttributes(
			attribute.Int("maintenance.data_length_bytes", len(bytes)),
			attribute.Int("maintenance.hex_length", len(hexString)),
		)
	}

	slog.Debug("Brother maintenance hex data", "hex", hexString, "length", len(hexString))

	// Split into chunks (CHUNK_SIZE from Python library)
	chunks := splitIntoChunks(hexString, BrotherChunkSize)

	if span != nil {
		span.SetAttributes(
			attribute.Int("maintenance.chunks_count", len(chunks)),
			attribute.Float64("parse.hex_duration_seconds", time.Since(parseStart).Seconds()),
		)
	}

	slog.Debug("Brother maintenance chunks", "chunks", chunks)

	// Extract toner and drum levels using the Brother-specific format
	tonerLevels := make(map[string]int)
	drumLevels := make(map[string]int)

	// Brother laser maintenance mappings (from const.py)
	laserMaintenanceMap := map[string]string{
		"6f": "black_toner_remaining",
		"70": "cyan_toner_remaining",
		"71": "magenta_toner_remaining",
		"72": "yellow_toner_remaining",
		"79": "cyan_drum_remaining",
		"7a": "magenta_drum_remaining",
		"7b": "yellow_drum_remaining",
		"80": "black_drum_remaining",
		"69": "belt_unit_remaining",
		"6a": "fuser_unit_remaining",
		"6b": "laser_unit_remaining",
		"6c": "paper_feeding_kit_remaining",
	}

	// Process each chunk
	for _, chunk := range chunks {
		if len(chunk) < 2 {
			continue
		}

		// First 2 hex chars are the type code
		typeCode := chunk[:2]

		// Check if this is a toner or drum level we care about
		if sensorType, exists := laserMaintenanceMap[typeCode]; exists {
			if len(chunk) >= 10 {
				// Last 8 hex chars contain the value (as per Python library)
				valueHex := chunk[len(chunk)-8:]

				value, err := strconv.ParseInt(valueHex, 16, 64)
				if err != nil {
					slog.Debug("Failed to parse hex value", "hex", valueHex, "error", err)
					continue
				}

				// For percentage values, divide by 100 (as per Python library)
				percentage := int(value / BrotherPercentageDiv)
				if percentage >= 0 && percentage <= 100 {
					switch sensorType {
					case "black_toner_remaining":
						tonerLevels["black"] = percentage
					case "cyan_toner_remaining":
						tonerLevels["cyan"] = percentage
					case "magenta_toner_remaining":
						tonerLevels["magenta"] = percentage
					case "yellow_toner_remaining":
						tonerLevels["yellow"] = percentage
					case "black_drum_remaining":
						drumLevels["black"] = percentage
					case "cyan_drum_remaining":
						drumLevels["cyan"] = percentage
					case "magenta_drum_remaining":
						drumLevels["magenta"] = percentage
					case "yellow_drum_remaining":
						drumLevels["yellow"] = percentage
					case "belt_unit_remaining":
						bc.metrics.BeltUnitRemainingPercent.With(prometheus.Labels{
							"host": bc.config.Printer.Host,
						}).Set(float64(percentage))
					case "fuser_unit_remaining":
						bc.metrics.FuserUnitRemainingPercent.With(prometheus.Labels{
							"host": bc.config.Printer.Host,
						}).Set(float64(percentage))
					case "laser_unit_remaining":
						bc.metrics.LaserUnitRemainingPercent.With(prometheus.Labels{
							"host": bc.config.Printer.Host,
						}).Set(float64(percentage))
					case "paper_feeding_kit_remaining":
						bc.metrics.PaperFeedingKitRemainingPercent.With(prometheus.Labels{
							"host": bc.config.Printer.Host,
						}).Set(float64(percentage))
					}

					slog.Debug("Found sensor", "type", sensorType, "value_hex", valueHex, "value", value, "percentage", percentage)
				}
			}
		}
	}

	// Update toner level metrics
	for color, level := range tonerLevels {
		bc.metrics.TonerLevel.With(prometheus.Labels{
			"host":  bc.config.Printer.Host,
			"color": color,
		}).Set(float64(level))

		// Set toner status based on level
		status, statusValue := calculateStatusFromLevel(float64(level))

		bc.metrics.TonerStatus.With(prometheus.Labels{
			"host":   bc.config.Printer.Host,
			"color":  color,
			"status": status,
		}).Set(statusValue)
	}

	// Update drum level metrics
	for color, level := range drumLevels {
		bc.metrics.DrumLevel.With(prometheus.Labels{
			"host":  bc.config.Printer.Host,
			"color": color,
		}).Set(float64(level))

		// Set drum status based on level
		status, statusValue := calculateStatusFromLevel(float64(level))

		bc.metrics.DrumStatus.With(prometheus.Labels{
			"host":   bc.config.Printer.Host,
			"color":  color,
			"status": status,
		}).Set(statusValue)
	}

	collectDuration := time.Since(collectStart)

	if span != nil {
		span.SetAttributes(
			attribute.Int("maintenance.toner_colors", len(tonerLevels)),
			attribute.Int("maintenance.drum_colors", len(drumLevels)),
			attribute.Float64("collect.duration_seconds", collectDuration.Seconds()),
		)
		span.AddEvent("maintenance_data_collected",
			attribute.Int("toner_colors", len(tonerLevels)),
			attribute.Int("drum_colors", len(drumLevels)),
		)
	}

	slog.Debug("Brother maintenance data collected",
		"toner_levels", tonerLevels,
		"drum_levels", drumLevels)

	return nil
}

// collectBrotherCountersData extracts page counts from Brother counters data
func (bc *BrotherCollector) collectBrotherCountersData(ctx context.Context) error {
	tracer := bc.app.GetTracer()

	var span *tracing.CollectorSpan

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "brother-collector", "collect-counters-data")

		span.SetAttributes(
			attribute.String("oid", OIDBrotherCountersData),
		)

		defer span.End()
	}

	collectStart := time.Now()
	getStart := time.Now()

	result, err := bc.client.Get([]string{OIDBrotherCountersData})

	getDuration := time.Since(getStart)

	if span != nil {
		span.SetAttributes(
			attribute.Float64("get.duration_seconds", getDuration.Seconds()),
			attribute.Bool("get.success", err == nil),
		)

		if err != nil {
			span.RecordError(err, attribute.String("operation", "snmp_get"))
			return fmt.Errorf("failed to get Brother counters data: %w", err)
		}
	} else {
		if err != nil {
			return fmt.Errorf("failed to get Brother counters data: %w", err)
		}
	}

	if len(result.Variables) == 0 || result.Variables[0].Value == nil {
		err := fmt.Errorf("no counters data received")

		if span != nil {
			span.RecordError(err, attribute.String("operation", "parse_counters_data"))
		}

		return err
	}

	variable := result.Variables[0]
	parseStart := time.Now()

	bytes, ok := variable.Value.([]uint8)
	if !ok {
		err := fmt.Errorf("counters data is not a byte array: %T", variable.Value)

		if span != nil {
			span.RecordError(err, attribute.String("operation", "convert_counters_data"))
		}

		return err
	}

	// Convert bytes to hex string (excluding last byte which is checksum)
	hexString := bytesToHexString(bytes)

	if span != nil {
		span.SetAttributes(
			attribute.Int("counters.data_length_bytes", len(bytes)),
			attribute.Int("counters.hex_length", len(hexString)),
		)
	}

	slog.Debug("Brother counters hex data", "hex", hexString, "length", len(hexString))

	// Split into chunks (CHUNK_SIZE from Python library)
	chunks := splitIntoChunks(hexString, BrotherChunkSize)

	if span != nil {
		span.SetAttributes(
			attribute.Int("counters.chunks_count", len(chunks)),
			attribute.Float64("parse.duration_seconds", time.Since(parseStart).Seconds()),
		)
	}

	slog.Debug("Brother counters chunks", "chunks", chunks)

	collectDuration := time.Since(collectStart)

	if span != nil {
		span.SetAttributes(
			attribute.Float64("collect.duration_seconds", collectDuration.Seconds()),
		)
		span.AddEvent("counters_data_collected")
	}

	slog.Debug("Brother counters data collected (page count metrics removed)")

	return nil
}

// collectBrotherNextCareData extracts remaining pages from Brother nextcare data
func (bc *BrotherCollector) collectBrotherNextCareData(ctx context.Context) error {
	tracer := bc.app.GetTracer()

	var span *tracing.CollectorSpan

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "brother-collector", "collect-nextcare-data")

		span.SetAttributes(
			attribute.String("oid", OIDBrotherNextCareData),
		)

		defer span.End()
	}

	collectStart := time.Now()
	getStart := time.Now()

	result, err := bc.client.Get([]string{OIDBrotherNextCareData})

	getDuration := time.Since(getStart)

	if span != nil {
		span.SetAttributes(
			attribute.Float64("get.duration_seconds", getDuration.Seconds()),
			attribute.Bool("get.success", err == nil),
		)

		if err != nil {
			span.RecordError(err, attribute.String("operation", "snmp_get"))
			return fmt.Errorf("failed to get Brother nextcare data: %w", err)
		}
	} else {
		if err != nil {
			return fmt.Errorf("failed to get Brother nextcare data: %w", err)
		}
	}
	if len(result.Variables) == 0 || result.Variables[0].Value == nil {
		err := fmt.Errorf("no nextcare data received")

		if span != nil {
			span.RecordError(err, attribute.String("operation", "parse_nextcare_data"))
		}

		return err
	}

	variable := result.Variables[0]
	parseStart := time.Now()

	bytes, ok := variable.Value.([]uint8)
	if !ok {
		err := fmt.Errorf("nextcare data is not a byte array: %T", variable.Value)

		if span != nil {
			span.RecordError(err, attribute.String("operation", "convert_nextcare_data"))
		}

		return err
	}

	// Convert bytes to hex string (excluding last byte which is checksum)
	hexString := bytesToHexString(bytes)

	if span != nil {
		span.SetAttributes(
			attribute.Int("nextcare.data_length_bytes", len(bytes)),
			attribute.Int("nextcare.hex_length", len(hexString)),
		)
	}

	slog.Debug("Brother nextcare hex data", "hex", hexString, "length", len(hexString))

	// Split into chunks (CHUNK_SIZE from Python library)
	chunks := splitIntoChunks(hexString, BrotherChunkSize)

	if span != nil {
		span.SetAttributes(
			attribute.Int("nextcare.chunks_count", len(chunks)),
			attribute.Float64("parse.hex_duration_seconds", time.Since(parseStart).Seconds()),
		)
	}

	slog.Debug("Brother nextcare chunks", "chunks", chunks)

	// Brother nextcare mappings (from const.py)
	nextcareMap := map[string]string{
		"73": "laser_unit_remaining_pages",
		"77": "paper_feeding_kit_1_remaining_pages",
		"82": "drum_remaining_pages",
		"86": "paper_feeding_kit_mp_remaining_pages",
		"88": "belt_unit_remaining_pages",
		"89": "fuser_unit_remaining_pages",
		"a4": "black_drum_remaining_pages",
		"a5": "cyan_drum_remaining_pages",
		"a6": "magenta_drum_remaining_pages",
		"a7": "yellow_drum_remaining_pages",
	}

	// Process each chunk
	for _, chunk := range chunks {
		if len(chunk) < 2 {
			continue
		}

		// First 2 hex chars are the type code
		typeCode := chunk[:2]

		// Check if this is a nextcare metric we care about
		if sensorType, exists := nextcareMap[typeCode]; exists {
			if len(chunk) >= 10 {
				// Last 8 hex chars contain the value (as per Python library)
				valueHex := chunk[len(chunk)-8:]

				value, err := strconv.ParseInt(valueHex, 16, 64)
				if err != nil {
					slog.Debug("Failed to parse hex value", "hex", valueHex, "error", err)
					continue
				}

				// For page counts, use the raw value (not divided by 100)
				if value >= 0 && value < 10000000 {
					switch sensorType {
					case "belt_unit_remaining_pages":
						bc.metrics.BeltUnitRemainingPages.With(prometheus.Labels{
							"host": bc.config.Printer.Host,
						}).Set(float64(value))
					case "fuser_unit_remaining_pages":
						bc.metrics.FuserUnitRemainingPages.With(prometheus.Labels{
							"host": bc.config.Printer.Host,
						}).Set(float64(value))
					case "laser_unit_remaining_pages":
						bc.metrics.LaserUnitRemainingPages.With(prometheus.Labels{
							"host": bc.config.Printer.Host,
						}).Set(float64(value))
					case "paper_feeding_kit_mp_remaining_pages":
						bc.metrics.PaperFeedingKitRemainingPages.With(prometheus.Labels{
							"host": bc.config.Printer.Host,
						}).Set(float64(value))
					}

					slog.Debug("Found nextcare sensor", "type", sensorType, "value_hex", valueHex, "value", value)
				}
			}
		}
	}

	collectDuration := time.Since(collectStart)

	if span != nil {
		span.SetAttributes(
			attribute.Float64("collect.duration_seconds", collectDuration.Seconds()),
		)
		span.AddEvent("nextcare_data_collected")
	}

	slog.Debug("Brother nextcare data collected")

	return nil
}

// collectLaserMetrics collects metrics specific to laser printers
func (bc *BrotherCollector) collectLaserMetrics(ctx context.Context) error {
	tracer := bc.app.GetTracer()

	var (
		span    *tracing.CollectorSpan
		spanCtx context.Context //nolint:contextcheck // Extracting context from span for child operations
	)

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "brother-collector", "collect-laser-metrics")

		span.SetAttributes(
			attribute.String("printer.type", "laser"),
			attribute.Int("colors.count", len(LaserColors)),
		)

		spanCtx = span.Context() //nolint:contextcheck // Standard OpenTelemetry pattern: extract context from span
		defer span.End()
	} else {
		spanCtx = ctx
	}

	collectStart := time.Now()

	// Collect toner levels and status
	bc.collectColorLevelsWithStatus(spanCtx, OIDTonerLevelBase, LaserColors, bc.metrics.TonerLevel, bc.metrics.TonerStatus, "toner level")

	// Collect drum levels and status
	bc.collectColorLevelsWithStatus(spanCtx, OIDDrumLevelBase, LaserColors, bc.metrics.DrumLevel, bc.metrics.DrumStatus, "drum level")

	collectDuration := time.Since(collectStart)

	if span != nil {
		span.SetAttributes(
			attribute.Float64("collect.duration_seconds", collectDuration.Seconds()),
		)
		span.AddEvent("laser_metrics_collected")
	}

	return nil
}

// collectInkjetMetrics collects metrics specific to inkjet printers
func (bc *BrotherCollector) collectInkjetMetrics(ctx context.Context) error {
	tracer := bc.app.GetTracer()

	var span *tracing.CollectorSpan

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "brother-collector", "collect-inkjet-metrics")

		span.SetAttributes(
			attribute.String("printer.type", "ink"),
			attribute.Int("colors.count", len(InkColors)),
		)

		defer span.End()
	}

	collectStart := time.Now()
	var colorsCollected int

	// Collect ink levels (using Brother-specific OIDs)
	inkOIDs := []string{
		"1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.1.0", // Brother ink info
		"1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.4.0", // Brother ink level
	}

	for i, oid := range inkOIDs {
		if i >= len(InkColors) {
			break
		}

		color := InkColors[i]

		getStart := time.Now()

		result, err := bc.client.Get([]string{oid})

		getDuration := time.Since(getStart)

		if err != nil {
			slog.Debug("Failed to get ink level", "color", color, "oid", oid, "error", err)

			if span != nil {
				span.RecordError(err, attribute.String("color", color), attribute.String("oid", oid))
			}

			continue
		}

		if len(result.Variables) > 0 {
			parseStart := time.Now()

			level, ok := convertToInt(result.Variables[0].Value, "ink level")
			if !ok {
				if span != nil {
					span.RecordError(fmt.Errorf("failed to convert ink level"), attribute.String("color", color))
				}

				continue
			}

			// Convert to percentage
			percentage := float64(level)
			if percentage > 100 {
				percentage = 100
			}

			parseDuration := time.Since(parseStart)

			bc.metrics.InkLevel.With(prometheus.Labels{
				"host":  bc.config.Printer.Host,
				"color": color,
			}).Set(percentage)

			// Set status based on level
			status, statusValue := calculateStatusFromLevel(percentage)

			bc.metrics.InkStatus.With(prometheus.Labels{
				"host":   bc.config.Printer.Host,
				"color":  color,
				"status": status,
			}).Set(statusValue)

			colorsCollected++

			if span != nil {
				span.SetAttributes(
					attribute.Float64("ink."+color+".get_duration_seconds", getDuration.Seconds()),
					attribute.Float64("ink."+color+".parse_duration_seconds", parseDuration.Seconds()),
					attribute.Float64("ink."+color+".percentage", percentage),
					attribute.String("ink."+color+".status", status),
				)
			}
		}
	}

	collectDuration := time.Since(collectStart)

	if span != nil {
		span.SetAttributes(
			attribute.Int("collect.colors_collected", colorsCollected),
			attribute.Float64("collect.duration_seconds", collectDuration.Seconds()),
		)
		span.AddEvent("inkjet_metrics_collected",
			attribute.Int("colors_collected", colorsCollected),
		)
	}

	return nil
}

// collectPaperTrayStatus collects paper tray status
func (bc *BrotherCollector) collectPaperTrayStatus(ctx context.Context) error {
	tracer := bc.app.GetTracer()

	var span *tracing.CollectorSpan

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "brother-collector", "collect-paper-tray-status")

		defer span.End()
	}

	collectStart := time.Now()

	// Check main paper tray
	oid := fmt.Sprintf("%s.1", OIDPaperTrayStatusBase)

	if span != nil {
		span.SetAttributes(
			attribute.String("oid", oid),
		)
	}

	getStart := time.Now()

	result, err := bc.client.Get([]string{oid})

	getDuration := time.Since(getStart)

	if span != nil {
		span.SetAttributes(
			attribute.Float64("get.duration_seconds", getDuration.Seconds()),
			attribute.Bool("get.success", err == nil),
		)

		if err != nil {
			span.RecordError(err, attribute.String("operation", "snmp_get"))
		}
	}

	if err != nil {
		slog.Debug("Failed to get paper tray status", "oid", oid, "error", err)
		return nil
	}

	if len(result.Variables) > 0 {
		parseStart := time.Now()

		status, ok := convertToInt(result.Variables[0].Value, "paper tray status")
		if !ok {
			if span != nil {
				span.SetAttributes(
					attribute.Bool("parse.success", false),
				)
				span.RecordError(fmt.Errorf("failed to convert status value"), attribute.String("operation", "convert_status"))
			}

			return nil
		}

		// Map status codes
		var (
			statusStr   string
			statusValue float64
		)

		switch status {
		case 3: // normal
			statusStr = "ok"
			statusValue = 1.0
		case 4: // empty
			statusStr = "empty"
			statusValue = 0.0
		case 5: // low
			statusStr = "low"
			statusValue = 0.0
		default:
			statusStr = "unknown"
			statusValue = 0.0
		}

		parseDuration := time.Since(parseStart)

		bc.metrics.PaperTrayStatus.With(prometheus.Labels{
			"host":   bc.config.Printer.Host,
			"tray":   "main",
			"status": statusStr,
		}).Set(statusValue)

		if span != nil {
			span.SetAttributes(
				attribute.Int("tray.status_code", status),
				attribute.String("tray.status_string", statusStr),
				attribute.Float64("tray.status_value", statusValue),
				attribute.Float64("parse.duration_seconds", parseDuration.Seconds()),
				attribute.Bool("parse.success", true),
			)
			span.AddEvent("paper_tray_status_collected",
				attribute.String("status", statusStr),
			)
		}
	}

	collectDuration := time.Since(collectStart)

	if span != nil {
		span.SetAttributes(
			attribute.Float64("collect.duration_seconds", collectDuration.Seconds()),
		)
	}

	return nil
}

// collectPageCounters collects page count metrics using Brother-specific counters data
func (bc *BrotherCollector) collectPageCounters(ctx context.Context) error {
	tracer := bc.app.GetTracer()

	var span *tracing.CollectorSpan

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "brother-collector", "collect-page-counters")

		span.SetAttributes(
			attribute.String("oid", OIDBrotherCountersData),
		)

		defer span.End()
	}

	collectStart := time.Now()
	getStart := time.Now()

	// Get the Brother counters data which contains multiple counter types
	result, err := bc.client.Get([]string{OIDBrotherCountersData})

	getDuration := time.Since(getStart)

	if span != nil {
		span.SetAttributes(
			attribute.Float64("get.duration_seconds", getDuration.Seconds()),
			attribute.Bool("get.success", err == nil),
		)

		if err != nil {
			span.RecordError(err, attribute.String("operation", "snmp_get"))
			return fmt.Errorf("failed to get Brother counters data: %w", err)
		}
	} else {
		if err != nil {
			return fmt.Errorf("failed to get Brother counters data: %w", err)
		}
	}
	if len(result.Variables) == 0 || result.Variables[0].Value == nil {
		err := fmt.Errorf("no Brother counters data received")

		if span != nil {
			span.RecordError(err, attribute.String("operation", "parse_counters_data"))
		}

		return err
	}

	variable := result.Variables[0]
	parseStart := time.Now()

	// Parse the hex string data
	hexData, ok := variable.Value.([]byte)
	if !ok {
		err := fmt.Errorf("invalid Brother counters data type: %T", variable.Value)

		if span != nil {
			span.RecordError(err, attribute.String("operation", "convert_counters_data"))
		}

		return err
	}

	if span != nil {
		span.SetAttributes(
			attribute.Int("counters.data_length_bytes", len(hexData)),
		)
	}

	// Parse the hex data to extract individual counters
	counters := bc.parseBrotherCounters(hexData)

	parseDuration := time.Since(parseStart)

	// Update metrics with the parsed counter values
	updateStart := time.Now()

	bc.metrics.PageCountTotal.With(prometheus.Labels{
		"host": bc.config.Printer.Host,
	}).Set(float64(counters["0001"])) // Total page count
	bc.metrics.PageCountBlack.With(prometheus.Labels{
		"host": bc.config.Printer.Host,
	}).Set(float64(counters["0101"])) // B/W count
	bc.metrics.PageCountColor.With(prometheus.Labels{
		"host": bc.config.Printer.Host,
	}).Set(float64(counters["0201"])) // Color count
	bc.metrics.PageCountDuplex.With(prometheus.Labels{
		"host": bc.config.Printer.Host,
	}).Set(float64(counters["0601"])) // Duplex count
	bc.metrics.PageCountDrumBlack.With(prometheus.Labels{
		"host": bc.config.Printer.Host,
	}).Set(float64(counters["1201"])) // Black drum count
	bc.metrics.PageCountDrumCyan.With(prometheus.Labels{
		"host": bc.config.Printer.Host,
	}).Set(float64(counters["1301"])) // Cyan drum count
	bc.metrics.PageCountDrumMagenta.With(prometheus.Labels{
		"host": bc.config.Printer.Host,
	}).Set(float64(counters["1401"])) // Magenta drum count
	bc.metrics.PageCountDrumYellow.With(prometheus.Labels{
		"host": bc.config.Printer.Host,
	}).Set(float64(counters["1501"])) // Yellow drum count

	updateDuration := time.Since(updateStart)
	collectDuration := time.Since(collectStart)

	if span != nil {
		span.SetAttributes(
			attribute.Int64("counters.total", int64(counters["0001"])),
			attribute.Int64("counters.black_white", int64(counters["0101"])),
			attribute.Int64("counters.color", int64(counters["0201"])),
			attribute.Int64("counters.duplex", int64(counters["0601"])),
			attribute.Int64("counters.black_drum", int64(counters["1201"])),
			attribute.Int64("counters.cyan_drum", int64(counters["1301"])),
			attribute.Int64("counters.magenta_drum", int64(counters["1401"])),
			attribute.Int64("counters.yellow_drum", int64(counters["1501"])),
			attribute.Float64("parse.duration_seconds", parseDuration.Seconds()),
			attribute.Float64("update.duration_seconds", updateDuration.Seconds()),
			attribute.Float64("collect.duration_seconds", collectDuration.Seconds()),
		)
		span.AddEvent("page_counters_collected",
			attribute.Int64("total", int64(counters["0001"])),
		)
	}

	slog.Debug("Page counters collected",
		"total", counters["0001"],
		"bw", counters["0101"],
		"color", counters["0201"],
		"duplex", counters["0601"],
		"black_drum", counters["1201"],
		"cyan_drum", counters["1301"],
		"magenta_drum", counters["1401"],
		"yellow_drum", counters["1501"])

	return nil
}

// parseBrotherCounters parses the hex data from Brother counters OID
func (bc *BrotherCollector) parseBrotherCounters(hexData []byte) map[string]int {
	counters := make(map[string]int)

	// Initialize all counter types to 0
	counterTypes := []string{"0001", "0101", "0201", "0601", "1201", "1301", "1401", "1501", "1601"}
	for _, counterType := range counterTypes {
		counters[counterType] = 0
	}

	// Parse the hex data - each counter entry is 7 bytes:
	// 2 bytes: counter type (e.g., "0001", "0101", "0201")
	// 1 byte: unknown/flag (always 04)
	// 4 bytes: counter value (big-endian)

	for i := 0; i < len(hexData)-6; i += 7 {
		if i+6 >= len(hexData) {
			break
		}

		// Extract counter type (first 2 bytes)
		counterType := fmt.Sprintf("%02X%02X", hexData[i], hexData[i+1])

		// Skip the flag byte (should be 04)
		if hexData[i+2] != 0x04 {
			continue
		}

		// Extract counter value (last 4 bytes, big-endian)
		value := int(hexData[i+3])<<24 | int(hexData[i+4])<<16 | int(hexData[i+5])<<8 | int(hexData[i+6])

		// Store the counter value if it's a known type
		if _, exists := counters[counterType]; exists {
			counters[counterType] = value
		}
	}

	return counters
}

// Stop stops the collector
func (bc *BrotherCollector) Stop() {
	close(bc.done)
	bc.disconnect(context.Background())
}
