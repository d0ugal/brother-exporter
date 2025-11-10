# Brother Printer Exporter

A Prometheus exporter for Brother printers that collects metrics via SNMP.

**Image**: `ghcr.io/d0ugal/brother-exporter:v1.11.5`

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

### Endpoints
- `GET /`: HTML dashboard with service status and metrics information
- `GET /metrics`: Prometheus metrics endpoint
- `GET /health`: Health check endpoint

## Quick Start

### Docker Compose

```yaml
version: '3.8'
services:
  brother-exporter:
    image: ghcr.io/d0ugal/brother-exporter:v1.11.5
    ports:
      - "8080:8080"
    environment:
      - BROTHER_EXPORTER_PRINTER_HOST=192.168.1.100
      - BROTHER_EXPORTER_PRINTER_TYPE=laser
      - BROTHER_EXPORTER_PRINTER_COMMUNITY=public
    restart: unless-stopped
```

1. Update the printer IP address in the environment variables
2. Run: `docker-compose up -d`
3. Access metrics: `curl http://localhost:8080/metrics`

## Configuration

Create a `config.yaml` file to configure the printer connection:

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

## Deployment

### Docker Compose (Environment Variables)

```yaml
version: '3.8'
services:
  brother-exporter:
    image: ghcr.io/d0ugal/brother-exporter:v1.11.5
    ports:
      - "8080:8080"
    environment:
      - BROTHER_EXPORTER_PRINTER_HOST=192.168.1.100
      - BROTHER_EXPORTER_PRINTER_TYPE=laser
      - BROTHER_EXPORTER_PRINTER_COMMUNITY=public
    restart: unless-stopped
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: brother-exporter
spec:
  replicas: 1
  selector:
    matchLabels:
      app: brother-exporter
  template:
    metadata:
      labels:
        app: brother-exporter
    spec:
      containers:
      - name: brother-exporter
        image: ghcr.io/d0ugal/brother-exporter:v1.11.5
        ports:
        - containerPort: 8080
        env:
        - name: BROTHER_EXPORTER_PRINTER_HOST
          value: "192.168.1.100"
        - name: BROTHER_EXPORTER_PRINTER_TYPE
          value: "laser"
        - name: BROTHER_EXPORTER_PRINTER_COMMUNITY
          value: "public"
```

## Prometheus Integration

Add to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'brother-exporter'
    static_configs:
      - targets: ['brother-exporter:8080']
```

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

## License

This project is licensed under the MIT License.