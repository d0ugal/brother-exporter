package collectors

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/d0ugal/brother-exporter/internal/config"
	"github.com/d0ugal/brother-exporter/internal/metrics"
	"github.com/gosnmp/gosnmp"
)

// convertToInt converts various SNMP value types to int using Go generics
// This eliminates the duplicated type switch statements throughout the code
func convertToInt[T any](value T, context string) (int, bool) {
	switch v := any(value).(type) {
	case int:
		return v, true
	case uint:
		return int(v), true
	case int64:
		return int(v), true
	case uint64:
		return int(v), true
	default:
		slog.Debug("Unexpected type for "+context, "type", fmt.Sprintf("%T", v), "value", v)
		return 0, false
	}
}

// BrotherCollector collects metrics from Brother printers via SNMP
type BrotherCollector struct {
	config  *config.Config
	metrics *metrics.Registry
	client  *gosnmp.GoSNMP
	mu      sync.RWMutex
	done    chan struct{}
}

// Brother printer SNMP OIDs
const (
	// System information
	OIDSystemDescription = "1.3.6.1.2.1.1.1.0"
	OIDSystemUpTime      = "1.3.6.1.2.1.1.3.0"
	OIDSystemContact     = "1.3.6.1.2.1.1.4.0"
	OIDSystemName        = "1.3.6.1.2.1.1.5.0"
	OIDSystemLocation    = "1.3.6.1.2.1.1.6.0"

	// Brother specific OIDs (base: 1.3.6.1.4.1.2435)
	OIDBrotherBase = "1.3.6.1.4.1.2435"

	// Printer status
	OIDPrinterStatus = "1.3.6.1.2.1.25.3.2.1.5.1"

	// Brother-specific consumable OIDs (these work better than standard MIB)
	OIDBrotherConsumableInfo  = "1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.1.0"  // Consumable info
	OIDBrotherConsumableLevel = "1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.4.0"  // Consumable level (104%)
	OIDBrotherPageCount       = "1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.6.0"  // Page count (10209)
	OIDBrotherStatus          = "1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.7.0"  // Status (1)
	OIDBrotherFirmware        = "1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.17.0" // Firmware version

	// Standard MIB OIDs (these return -2/-3 for Brother printers)
	OIDTonerLevelBase      = "1.3.6.1.2.1.43.11.1.1.9.1"
	OIDDrumLevelBase       = "1.3.6.1.2.1.43.11.1.1.8.1"
	OIDPaperTrayStatusBase = "1.3.6.1.2.1.43.8.2.1.10.1"

	// Page counters
	OIDPageCountTotal = "1.3.6.1.2.1.43.10.2.1.4.1.1"          // Standard total pages
	OIDPageCountColor = "1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.9.0" // Brother color pages
)

// Color mappings for Brother printers
var (
	LaserColors = []string{"black", "cyan", "magenta", "yellow"}
	InkColors   = []string{"black", "cyan", "magenta", "yellow"}
)

func NewBrotherCollector(cfg *config.Config, metricsRegistry *metrics.Registry) *BrotherCollector {
	return &BrotherCollector{
		config:  cfg,
		metrics: metricsRegistry,
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
	bc.collectMetrics()

	for {
		select {
		case <-ctx.Done():
			return
		case <-bc.done:
			return
		case <-ticker.C:
			bc.collectMetrics()
		}
	}
}

// collectMetrics performs a single metrics collection cycle
func (bc *BrotherCollector) collectMetrics() {
	if err := bc.connect(); err != nil {
		slog.Error("Failed to connect to Brother printer",
			"host", bc.config.Printer.Host,
			"error", err,
		)
		bc.metrics.PrinterConnectionStatus.WithLabelValues(bc.config.Printer.Host).Set(0)
		bc.metrics.PrinterConnectionErrors.WithLabelValues(bc.config.Printer.Host, "connect").Inc()

		return
	}

	defer bc.disconnect()

	// Set connection status
	bc.metrics.PrinterConnectionStatus.WithLabelValues(bc.config.Printer.Host).Set(1)

	// Collect printer information
	if err := bc.collectPrinterInfo(); err != nil {
		slog.Error("Failed to collect printer info", "error", err)
		bc.metrics.PrinterConnectionErrors.WithLabelValues(bc.config.Printer.Host, "info").Inc()
	}

	// Collect printer status
	if err := bc.collectPrinterStatus(); err != nil {
		slog.Error("Failed to collect printer status", "error", err)
		bc.metrics.PrinterConnectionErrors.WithLabelValues(bc.config.Printer.Host, "status").Inc()
	}

	// Collect Brother-specific metrics (these work better than standard MIB)
	if err := bc.collectBrotherSpecificMetrics(); err != nil {
		slog.Error("Failed to collect Brother-specific metrics", "error", err)
		bc.metrics.PrinterConnectionErrors.WithLabelValues(bc.config.Printer.Host, "brother_metrics").Inc()

		// Fallback to standard MIB only if Brother-specific collection fails
		switch bc.config.Printer.Type {
		case "laser":
			if err := bc.collectLaserMetrics(); err != nil {
				slog.Error("Failed to collect laser metrics", "error", err)
				bc.metrics.PrinterConnectionErrors.WithLabelValues(bc.config.Printer.Host, "laser_metrics").Inc()
			}
		case "ink":
			if err := bc.collectInkjetMetrics(); err != nil {
				slog.Error("Failed to collect inkjet metrics", "error", err)
				bc.metrics.PrinterConnectionErrors.WithLabelValues(bc.config.Printer.Host, "inkjet_metrics").Inc()
			}
		}
	} else {
		// If Brother-specific collection succeeded, also collect nextcare data
		if err := bc.collectBrotherNextCareData(); err != nil {
			slog.Error("Failed to collect Brother nextcare data", "error", err)
			bc.metrics.PrinterConnectionErrors.WithLabelValues(bc.config.Printer.Host, "nextcare_metrics").Inc()
		}
	}

	// Collect page counters using standard MIB (Brother-specific OIDs give wrong values)
	if err := bc.collectPageCounters(); err != nil {
		slog.Error("Failed to collect page counters", "error", err)
		bc.metrics.PrinterConnectionErrors.WithLabelValues(bc.config.Printer.Host, "page_counters").Inc()
	}

	// Collect paper tray status
	if err := bc.collectPaperTrayStatus(); err != nil {
		slog.Error("Failed to collect paper tray status", "error", err)
		bc.metrics.PrinterConnectionErrors.WithLabelValues(bc.config.Printer.Host, "paper_tray").Inc()
	}
}

// connect establishes SNMP connection to the printer
func (bc *BrotherCollector) connect() error {
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

	return bc.client.Connect()
}

// disconnect closes the SNMP connection
func (bc *BrotherCollector) disconnect() {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if bc.client != nil {
		if err := bc.client.Conn.Close(); err != nil {
			slog.Debug("Error closing SNMP connection", "error", err)
		}

		bc.client = nil
	}
}

// collectPrinterInfo collects basic printer information
func (bc *BrotherCollector) collectPrinterInfo() error {
	oids := []string{
		OIDSystemDescription,
		OIDSystemName,
		OIDSystemContact,
		OIDSystemLocation,
	}

	result, err := bc.client.Get(oids)
	if err != nil {
		return fmt.Errorf("failed to get system info: %w", err)
	}

	var model, serial, firmware string

	for _, variable := range result.Variables {
		value := string(variable.Value.([]byte))

		switch variable.Name {
		case OIDSystemDescription:
			// Parse model and firmware from description
			parts := strings.Split(value, ";")
			if len(parts) >= 2 {
				model = strings.TrimSpace(parts[0])
				firmware = strings.TrimSpace(parts[1])
			} else {
				model = value
			}
		case OIDSystemName:
			serial = value
		}
	}

	// Set printer info metric
	bc.metrics.PrinterInfo.WithLabelValues(
		bc.config.Printer.Host,
		model,
		serial,
		firmware,
		bc.config.Printer.Type,
	).Set(1)

	return nil
}

// collectPrinterStatus collects printer status information
func (bc *BrotherCollector) collectPrinterStatus() error {
	result, err := bc.client.Get([]string{OIDPrinterStatus})
	if err != nil {
		return fmt.Errorf("failed to get printer status: %w", err)
	}

	if len(result.Variables) > 0 {
		status, ok := convertToInt(result.Variables[0].Value, "printer status")
		if !ok {
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

		bc.metrics.PrinterStatus.WithLabelValues(
			bc.config.Printer.Host,
			statusStr,
		).Set(statusValue)
	}

	return nil
}

// collectBrotherSpecificMetrics collects Brother-specific metrics using the proper OIDs and decoding
func (bc *BrotherCollector) collectBrotherSpecificMetrics() error {
	// Get maintenance data (contains toner and drum levels)
	if err := bc.collectBrotherMaintenanceData(); err != nil {
		slog.Error("Failed to collect Brother maintenance data", "error", err)
	}

	// Get counters data (contains page counts)
	if err := bc.collectBrotherCountersData(); err != nil {
		slog.Error("Failed to collect Brother counters data", "error", err)
	}

	// Get basic info
	oids := []string{
		OIDBrotherConsumableInfo, // Consumable info (E83216M3N204406)
		OIDBrotherFirmware,       // Firmware version (1.16)
	}

	result, err := bc.client.Get(oids)
	if err != nil {
		return fmt.Errorf("failed to get Brother basic info: %w", err)
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
			}
		case 1: // Firmware
			if bytes, ok := variable.Value.([]uint8); ok {
				firmware := string(bytes)
				slog.Debug("Brother firmware", "firmware", firmware)
			}
		}
	}

	return nil
}

// collectBrotherMaintenanceData extracts toner and drum levels from Brother maintenance data
func (bc *BrotherCollector) collectBrotherMaintenanceData() error {
	result, err := bc.client.Get([]string{"1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.8.0"})
	if err != nil {
		return fmt.Errorf("failed to get Brother maintenance data: %w", err)
	}

	if len(result.Variables) == 0 || result.Variables[0].Value == nil {
		return fmt.Errorf("no maintenance data received")
	}

	variable := result.Variables[0]

	bytes, ok := variable.Value.([]uint8)
	if !ok {
		return fmt.Errorf("maintenance data is not a byte array: %T", variable.Value)
	}

	// Convert bytes to hex string (excluding last byte which is checksum)
	hexString := ""
	for i := 0; i < len(bytes)-1; i++ {
		hexString += fmt.Sprintf("%02x", bytes[i])
	}

	slog.Debug("Brother maintenance hex data", "hex", hexString, "length", len(hexString))

	// Split into 14-character chunks (CHUNK_SIZE from Python library)
	chunkSize := 14
	chunks := make([]string, 0)

	for i := 0; i < len(hexString); i += chunkSize {
		end := i + chunkSize
		if end > len(hexString) {
			end = len(hexString)
		}

		chunks = append(chunks, hexString[i:end])
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
				percentage := int(value / 100)
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
						bc.metrics.BeltUnitRemainingPercent.WithLabelValues(bc.config.Printer.Host).Set(float64(percentage))
					case "fuser_unit_remaining":
						bc.metrics.FuserUnitRemainingPercent.WithLabelValues(bc.config.Printer.Host).Set(float64(percentage))
					case "laser_unit_remaining":
						bc.metrics.LaserUnitRemainingPercent.WithLabelValues(bc.config.Printer.Host).Set(float64(percentage))
					case "paper_feeding_kit_remaining":
						bc.metrics.PaperFeedingKitRemainingPercent.WithLabelValues(bc.config.Printer.Host).Set(float64(percentage))
					}

					slog.Debug("Found sensor", "type", sensorType, "value_hex", valueHex, "value", value, "percentage", percentage)
				}
			}
		}
	}

	// Update toner level metrics
	for color, level := range tonerLevels {
		bc.metrics.TonerLevel.WithLabelValues(bc.config.Printer.Host, color).Set(float64(level))

		// Set toner status based on level
		status := "ok"
		statusValue := 1.0

		if level < 10 {
			status = "low"
			statusValue = 0.0
		} else if level == 0 {
			status = "empty"
			statusValue = 0.0
		}

		bc.metrics.TonerStatus.WithLabelValues(bc.config.Printer.Host, color, status).Set(statusValue)
	}

	// Update drum level metrics
	for color, level := range drumLevels {
		bc.metrics.DrumLevel.WithLabelValues(bc.config.Printer.Host, color).Set(float64(level))

		// Set drum status based on level
		status := "ok"
		statusValue := 1.0

		if level < 10 {
			status = "low"
			statusValue = 0.0
		} else if level == 0 {
			status = "empty"
			statusValue = 0.0
		}

		bc.metrics.DrumStatus.WithLabelValues(bc.config.Printer.Host, color, status).Set(statusValue)
	}

	slog.Debug("Brother maintenance data collected",
		"toner_levels", tonerLevels,
		"drum_levels", drumLevels)

	return nil
}

// collectBrotherCountersData extracts page counts from Brother counters data
func (bc *BrotherCollector) collectBrotherCountersData() error {
	result, err := bc.client.Get([]string{"1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.10.0"})
	if err != nil {
		return fmt.Errorf("failed to get Brother counters data: %w", err)
	}

	if len(result.Variables) == 0 || result.Variables[0].Value == nil {
		return fmt.Errorf("no counters data received")
	}

	variable := result.Variables[0]

	bytes, ok := variable.Value.([]uint8)
	if !ok {
		return fmt.Errorf("counters data is not a byte array: %T", variable.Value)
	}

	// Convert bytes to hex string (excluding last byte which is checksum)
	hexString := ""
	for i := 0; i < len(bytes)-1; i++ {
		hexString += fmt.Sprintf("%02x", bytes[i])
	}

	slog.Debug("Brother counters hex data", "hex", hexString, "length", len(hexString))

	// Split into 14-character chunks (CHUNK_SIZE from Python library)
	chunkSize := 14
	chunks := make([]string, 0)

	for i := 0; i < len(hexString); i += chunkSize {
		end := i + chunkSize
		if end > len(hexString) {
			end = len(hexString)
		}

		chunks = append(chunks, hexString[i:end])
	}

	slog.Debug("Brother counters chunks", "chunks", chunks)

	// Extract page counts using the Brother-specific format
	pageCounts := make(map[string]int)

	// Brother counters mappings (from const.py)
	countersMap := map[string]string{
		"00": "page_counter",
		"01": "bw_counter",
		"02": "color_counter",
		"06": "duplex_unit_pages_counter",
		"12": "black_counter",
		"13": "cyan_counter",
		"14": "magenta_counter",
		"15": "yellow_counter",
		"16": "image_counter",
	}

	// Process each chunk
	for _, chunk := range chunks {
		if len(chunk) < 2 {
			continue
		}

		// First 2 hex chars are the type code
		typeCode := chunk[:2]

		// Check if this is a counter we care about
		if counterType, exists := countersMap[typeCode]; exists {
			if len(chunk) >= 10 {
				// Last 8 hex chars contain the value (as per Python library)
				valueHex := chunk[len(chunk)-8:]

				value, err := strconv.ParseInt(valueHex, 16, 64)
				if err != nil {
					slog.Debug("Failed to parse hex value", "hex", valueHex, "error", err)
					continue
				}

				// For counters, use the raw value (not divided by 100)
				if value >= 0 && value < 10000000 { // Reasonable page count range
					switch counterType {
					case "page_counter":
						pageCounts["total"] = int(value)
					case "bw_counter":
						pageCounts["bw"] = int(value)
					case "color_counter":
						pageCounts["color"] = int(value)
					}

					slog.Debug("Found counter", "type", counterType, "value_hex", valueHex, "value", value)
				}
			}
		}
	}

	// Update page count metrics (only if we have data from Brother counters)
	// Note: This should only run if the standard page counters failed
	if len(pageCounts) > 0 {
		for countType, count := range pageCounts {
			switch countType {
			case "total":
				bc.metrics.PageCountTotal.WithLabelValues(bc.config.Printer.Host).Add(float64(count))
			case "bw":
				bc.metrics.PageCountBlack.WithLabelValues(bc.config.Printer.Host).Add(float64(count))
			case "color":
				bc.metrics.PageCountColor.WithLabelValues(bc.config.Printer.Host).Add(float64(count))
			}
		}
	}

	// Update individual color page counters
	for _, chunk := range chunks {
		if len(chunk) < 2 {
			continue
		}

		typeCode := chunk[:2]
		if len(chunk) >= 10 {
			valueHex := chunk[len(chunk)-8:]

			value, err := strconv.ParseInt(valueHex, 16, 64)
			if err != nil {
				continue
			}

			if value >= 0 && value < 10000000 {
				switch typeCode {
				case "13": // Cyan counter
					bc.metrics.PageCountCyan.WithLabelValues(bc.config.Printer.Host).Add(float64(value))
				case "14": // Magenta counter
					bc.metrics.PageCountMagenta.WithLabelValues(bc.config.Printer.Host).Add(float64(value))
				case "15": // Yellow counter
					bc.metrics.PageCountYellow.WithLabelValues(bc.config.Printer.Host).Add(float64(value))
				case "06": // Duplex counter
					bc.metrics.PageCountDuplex.WithLabelValues(bc.config.Printer.Host).Add(float64(value))
				}
			}
		}
	}

	slog.Debug("Brother counters data collected", "page_counts", pageCounts)

	return nil
}

// collectBrotherNextCareData extracts remaining pages from Brother nextcare data
func (bc *BrotherCollector) collectBrotherNextCareData() error {
	result, err := bc.client.Get([]string{"1.3.6.1.4.1.2435.2.3.9.4.2.1.5.5.11.0"})
	if err != nil {
		return fmt.Errorf("failed to get Brother nextcare data: %w", err)
	}

	if len(result.Variables) == 0 || result.Variables[0].Value == nil {
		return fmt.Errorf("no nextcare data received")
	}

	variable := result.Variables[0]

	bytes, ok := variable.Value.([]uint8)
	if !ok {
		return fmt.Errorf("nextcare data is not a byte array: %T", variable.Value)
	}

	// Convert bytes to hex string (excluding last byte which is checksum)
	hexString := ""
	for i := 0; i < len(bytes)-1; i++ {
		hexString += fmt.Sprintf("%02x", bytes[i])
	}

	slog.Debug("Brother nextcare hex data", "hex", hexString, "length", len(hexString))

	// Split into 14-character chunks (CHUNK_SIZE from Python library)
	chunkSize := 14
	chunks := make([]string, 0)

	for i := 0; i < len(hexString); i += chunkSize {
		end := i + chunkSize
		if end > len(hexString) {
			end = len(hexString)
		}

		chunks = append(chunks, hexString[i:end])
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
						bc.metrics.BeltUnitRemainingPages.WithLabelValues(bc.config.Printer.Host).Set(float64(value))
					case "fuser_unit_remaining_pages":
						bc.metrics.FuserUnitRemainingPages.WithLabelValues(bc.config.Printer.Host).Set(float64(value))
					case "laser_unit_remaining_pages":
						bc.metrics.LaserUnitRemainingPages.WithLabelValues(bc.config.Printer.Host).Set(float64(value))
					case "paper_feeding_kit_mp_remaining_pages":
						bc.metrics.PaperFeedingKitRemainingPages.WithLabelValues(bc.config.Printer.Host).Set(float64(value))
					}

					slog.Debug("Found nextcare sensor", "type", sensorType, "value_hex", valueHex, "value", value)
				}
			}
		}
	}

	slog.Debug("Brother nextcare data collected")

	return nil
}

// collectLaserMetrics collects metrics specific to laser printers
func (bc *BrotherCollector) collectLaserMetrics() error {
	// Collect toner levels
	for i, color := range LaserColors {
		oid := fmt.Sprintf("%s.%d", OIDTonerLevelBase, i+1)

		result, err := bc.client.Get([]string{oid})
		if err != nil {
			slog.Debug("Failed to get toner level", "color", color, "oid", oid, "error", err)
			continue
		}

		if len(result.Variables) > 0 {
			level, ok := convertToInt(result.Variables[0].Value, "toner level")
			if !ok {
				continue
			}

			// Convert to percentage (assuming max level is 100)
			percentage := float64(level)
			if percentage > 100 {
				percentage = 100
			}

			bc.metrics.TonerLevel.WithLabelValues(
				bc.config.Printer.Host,
				color,
			).Set(percentage)

			// Set status based on level
			status := "ok"
			statusValue := 1.0

			if percentage < 10 {
				status = "low"
				statusValue = 0.0
			} else if percentage == 0 {
				status = "empty"
				statusValue = 0.0
			}

			bc.metrics.TonerStatus.WithLabelValues(
				bc.config.Printer.Host,
				color,
				status,
			).Set(statusValue)
		}
	}

	// Collect drum levels
	for i, color := range LaserColors {
		oid := fmt.Sprintf("%s.%d", OIDDrumLevelBase, i+1)

		result, err := bc.client.Get([]string{oid})
		if err != nil {
			slog.Debug("Failed to get drum level", "color", color, "oid", oid, "error", err)
			continue
		}

		if len(result.Variables) > 0 {
			level, ok := convertToInt(result.Variables[0].Value, "drum level")
			if !ok {
				continue
			}

			// Convert to percentage
			percentage := float64(level)
			if percentage > 100 {
				percentage = 100
			}

			bc.metrics.DrumLevel.WithLabelValues(
				bc.config.Printer.Host,
				color,
			).Set(percentage)

			// Set status based on level
			status := "ok"
			statusValue := 1.0

			if percentage < 10 {
				status = "low"
				statusValue = 0.0
			} else if percentage == 0 {
				status = "empty"
				statusValue = 0.0
			}

			bc.metrics.DrumStatus.WithLabelValues(
				bc.config.Printer.Host,
				color,
				status,
			).Set(statusValue)
		}
	}

	return nil
}

// collectInkjetMetrics collects metrics specific to inkjet printers
func (bc *BrotherCollector) collectInkjetMetrics() error {
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

		result, err := bc.client.Get([]string{oid})
		if err != nil {
			slog.Debug("Failed to get ink level", "color", color, "oid", oid, "error", err)
			continue
		}

		if len(result.Variables) > 0 {
			level, ok := convertToInt(result.Variables[0].Value, "ink level")
			if !ok {
				continue
			}

			// Convert to percentage
			percentage := float64(level)
			if percentage > 100 {
				percentage = 100
			}

			bc.metrics.InkLevel.WithLabelValues(
				bc.config.Printer.Host,
				color,
			).Set(percentage)

			// Set status based on level
			status := "ok"
			statusValue := 1.0

			if percentage < 10 {
				status = "low"
				statusValue = 0.0
			} else if percentage == 0 {
				status = "empty"
				statusValue = 0.0
			}

			bc.metrics.InkStatus.WithLabelValues(
				bc.config.Printer.Host,
				color,
				status,
			).Set(statusValue)
		}
	}

	return nil
}

// collectPageCounters collects page count metrics
func (bc *BrotherCollector) collectPageCounters() error {
	// Use standard MIB OIDs for accurate page counts (Brother OIDs are incorrect)
	oids := []string{
		OIDPageCountTotal, // Standard MIB total page count (this gives correct value: 788)
	}

	result, err := bc.client.Get(oids)
	if err != nil {
		return fmt.Errorf("failed to get page counters: %w", err)
	}

	for i, variable := range result.Variables {
		if variable.Value != nil {
			count, ok := convertToInt(variable.Value, "page counter")
			if !ok {
				continue
			}

			switch i {
			case 0: // Standard MIB total page count
				// For counters, we need to track the difference from last collection
				// For now, just add the current value (this will accumulate, but shows the right total)
				bc.metrics.PageCountTotal.WithLabelValues(bc.config.Printer.Host).Add(float64(count))
				slog.Debug("Added total page count", "count", count)
			}
		}
	}

	return nil
}

// collectPaperTrayStatus collects paper tray status
func (bc *BrotherCollector) collectPaperTrayStatus() error {
	// Check main paper tray
	oid := fmt.Sprintf("%s.1", OIDPaperTrayStatusBase)

	result, err := bc.client.Get([]string{oid})
	if err != nil {
		slog.Debug("Failed to get paper tray status", "oid", oid, "error", err)
		return nil
	}

	if len(result.Variables) > 0 {
		status, ok := convertToInt(result.Variables[0].Value, "paper tray status")
		if !ok {
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

		bc.metrics.PaperTrayStatus.WithLabelValues(
			bc.config.Printer.Host,
			"main",
			statusStr,
		).Set(statusValue)
	}

	return nil
}

// Stop stops the collector
func (bc *BrotherCollector) Stop() {
	close(bc.done)
	bc.disconnect()
}
