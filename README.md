# Brother Printer Exporter

A Prometheus exporter for Brother printers that collects metrics via SNMP.

## Features

- **Printer Information**: Model, serial number, firmware version
- **Printer Status**: Ready, printing, warmup, error states
- **Consumable Levels**: Toner/ink levels and status for all colors
- **Drum Levels**: Drum life remaining (laser printers)
- **Paper Tray Status**: Paper availability and status
- **Page Counters**: Total, black, and color page counts
- **Connection Monitoring**: SNMP connection status and error tracking

## Supported Printer Types

- **Laser Printers**: Monitors toner levels, drum levels, and page counters
- **Inkjet Printers**: Monitors ink levels and page counters

## Installation

### Build from Source

```bash
git clone <repository-url>
cd brother-exporter
make build
```

### Configuration

Create a configuration file `config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

logging:
  level: "info"
  format: "json"

metrics:
  collection:
    default_interval: "30s"

printer:
  host: "192.168.1.100"  # Your Brother printer IP
  community: "public"    # SNMP community string
  type: "laser"          # "laser" or "ink"
```

### Environment Variables

You can also configure the exporter using environment variables:

```bash
export BROTHER_EXPORTER_PRINTER_HOST="192.168.1.100"
export BROTHER_EXPORTER_PRINTER_TYPE="laser"
export BROTHER_EXPORTER_PRINTER_COMMUNITY="public"
export BROTHER_EXPORTER_SERVER_PORT="8080"
```

## Usage

### Run with Configuration File

```bash
./brother-exporter -config config.yaml
```

### Run with Environment Variables

```bash
./brother-exporter -config-from-env
```

### Docker

```bash
docker run -d \
  -p 8080:8080 \
  -e BROTHER_EXPORTER_PRINTER_HOST="192.168.1.100" \
  -e BROTHER_EXPORTER_PRINTER_TYPE="laser" \
  brother-exporter
```

## Metrics

### Printer Information
- `brother_printer_info` - Printer model, serial, firmware, and type information

### Connection Status
- `brother_printer_connection_status` - SNMP connection status (1 = connected, 0 = disconnected)
- `brother_printer_connection_errors_total` - Total connection errors by type

### Printer Status
- `brother_printer_status` - Printer operational status

### Consumable Levels (Laser Printers)
- `brother_toner_level_percent` - Toner level percentage by color
- `brother_toner_status` - Toner status (ok/low/empty) by color
- `brother_drum_level_percent` - Drum level percentage by color
- `brother_drum_status` - Drum status (ok/low/empty) by color

### Consumable Levels (Inkjet Printers)
- `brother_ink_level_percent` - Ink level percentage by color
- `brother_ink_status` - Ink status (ok/low/empty) by color

### Paper Tray
- `brother_paper_tray_status` - Paper tray status by tray

### Page Counters
- `brother_page_count_total` - Total pages printed
- `brother_page_count_black_total` - Black pages printed
- `brother_page_count_color_total` - Color pages printed

## SNMP Requirements

The Brother printer must have SNMP enabled with the following settings:

- **SNMP Version**: v2c (recommended) or v1
- **Community String**: Usually "public" (default)
- **Port**: 161 (default SNMP port)

### Enabling SNMP on Brother Printers

1. Access the printer's web interface
2. Navigate to Network Settings → SNMP
3. Enable SNMP
4. Set the community string (default: "public")
5. Save settings

## Troubleshooting

### Connection Issues

1. **Check SNMP is enabled** on the printer
2. **Verify community string** matches configuration
3. **Test SNMP connectivity**:
   ```bash
   snmpwalk -v2c -c public 192.168.1.100 1.3.6.1.2.1.1.1.0
   ```

### No Metrics

1. **Check printer type** configuration (laser vs ink)
2. **Verify OID support** - some older printers may not support all OIDs
3. **Check logs** for SNMP errors

### Common SNMP OIDs

- System Description: `1.3.6.1.2.1.1.1.0`
- Printer Status: `1.3.6.1.2.1.25.3.2.1.5.1`
- Toner Levels: `1.3.6.1.2.1.43.11.1.1.9.1.x`
- Page Counters: `1.3.6.1.2.1.43.10.2.1.4.1.1`

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Linting

```bash
make lint
```

### Formatting

```bash
make fmt
```

## License

This project is licensed under the MIT License.
