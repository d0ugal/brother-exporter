package config

import (
	"os"
	"testing"
	"time"

	promexporter_config "github.com/d0ugal/promexporter/config"
)

func TestDuration_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected time.Duration
		wantErr  bool
	}{
		{
			name:     "valid duration string",
			input:    "30s",
			expected: 30 * time.Second,
			wantErr:  false,
		},
		{
			name:     "valid duration string with minutes",
			input:    "2m",
			expected: 2 * time.Minute,
			wantErr:  false,
		},
		{
			name:     "valid duration string with hours",
			input:    "1h",
			expected: time.Hour,
			wantErr:  false,
		},
		{
			name:     "integer seconds (backward compatibility)",
			input:    60,
			expected: 60 * time.Second,
			wantErr:  false,
		},
		{
			name:     "int64 seconds (backward compatibility)",
			input:    int64(120),
			expected: 120 * time.Second,
			wantErr:  false,
		},
		{
			name:    "invalid duration string",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "unsupported type",
			input:   []string{"invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Duration{}
			err := d.UnmarshalYAML(func(v interface{}) error {
				*v.(*interface{}) = tt.input
				return nil
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("Duration.UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && d.Duration != tt.expected {
				t.Errorf("Duration.UnmarshalYAML() = %v, want %v", d.Duration, tt.expected)
			}
		})
	}
}

func TestDuration_Seconds(t *testing.T) {
	d := Duration{Duration: 2 * time.Minute}
	if got := d.Seconds(); got != 120 {
		t.Errorf("Duration.Seconds() = %v, want 120", got)
	}
}

func TestLoad(t *testing.T) {
	// Create a temporary config file
	configContent := `
server:
  host: "127.0.0.1"
  port: 8080

logging:
  level: "debug"
  format: "text"

metrics:
  collection:
    default_interval: "60s"

printer:
  host: "192.168.1.100"
  community: "private"
  type: "laser"
`

	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	_ = tmpFile.Close()

	config, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Test loaded values
	if config.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host = %v, want 127.0.0.1", config.Server.Host)
	}

	if config.Server.Port != 8080 {
		t.Errorf("Server.Port = %v, want 8080", config.Server.Port)
	}

	if config.Logging.Level != "debug" {
		t.Errorf("Logging.Level = %v, want debug", config.Logging.Level)
	}

	if config.Logging.Format != "text" {
		t.Errorf("Logging.Format = %v, want text", config.Logging.Format)
	}

	if config.Metrics.Collection.DefaultInterval.Seconds() != 60 {
		t.Errorf("Metrics.Collection.DefaultInterval = %v, want 60s", config.Metrics.Collection.DefaultInterval.Seconds())
	}

	if config.Printer.Host != "192.168.1.100" {
		t.Errorf("Printer.Host = %v, want 192.168.1.100", config.Printer.Host)
	}

	if config.Printer.Community != "private" {
		t.Errorf("Printer.Community = %v, want private", config.Printer.Community)
	}

	if config.Printer.Type != "laser" {
		t.Errorf("Printer.Type = %v, want laser", config.Printer.Type)
	}
}

func TestLoad_Defaults(t *testing.T) {
	// Create a minimal config file
	configContent := `
printer:
  host: "192.168.1.100"
`

	tmpFile, err := os.CreateTemp("", "test-config-minimal-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	_ = tmpFile.Close()

	config, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Test default values
	if config.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %v, want 0.0.0.0", config.Server.Host)
	}

	if config.Server.Port != 8080 {
		t.Errorf("Server.Port = %v, want 8080", config.Server.Port)
	}

	if config.Logging.Level != "info" {
		t.Errorf("Logging.Level = %v, want info", config.Logging.Level)
	}

	if config.Logging.Format != "json" {
		t.Errorf("Logging.Format = %v, want json", config.Logging.Format)
	}

	if config.Metrics.Collection.DefaultInterval.Seconds() != 30 {
		t.Errorf("Metrics.Collection.DefaultInterval = %v, want 30s", config.Metrics.Collection.DefaultInterval.Seconds())
	}

	if config.Printer.Community != "public" {
		t.Errorf("Printer.Community = %v, want public", config.Printer.Community)
	}

	if config.Printer.Type != "laser" {
		t.Errorf("Printer.Type = %v, want laser", config.Printer.Type)
	}
}

func TestLoad_InvalidFile(t *testing.T) {
	_, err := Load("nonexistent-file.yaml")
	if err == nil {
		t.Error("Load() expected error for nonexistent file")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	configContent := `invalid: yaml: content: [`

	tmpFile, err := os.CreateTemp("", "test-config-invalid-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	_ = tmpFile.Close()

	_, err = Load(tmpFile.Name())
	if err == nil {
		t.Error("Load() expected error for invalid YAML")
	}
}

func TestLoadConfig_FileMode(t *testing.T) {
	configContent := `
printer:
  host: "192.168.1.100"
`

	tmpFile, err := os.CreateTemp("", "test-config-file-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	_ = tmpFile.Close()

	config, err := LoadConfig(tmpFile.Name(), false)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if config.Printer.Host != "192.168.1.100" {
		t.Errorf("Printer.Host = %v, want 192.168.1.100", config.Printer.Host)
	}
}

func TestLoadConfig_EnvMode(t *testing.T) {
	// Set environment variables
	_ = os.Setenv("BROTHER_EXPORTER_SERVER_HOST", "127.0.0.1")
	_ = os.Setenv("BROTHER_EXPORTER_SERVER_PORT", "9090")
	_ = os.Setenv("BROTHER_EXPORTER_LOG_LEVEL", "debug")
	_ = os.Setenv("BROTHER_EXPORTER_LOG_FORMAT", "text")
	_ = os.Setenv("BROTHER_EXPORTER_METRICS_DEFAULT_INTERVAL", "45s")
	_ = os.Setenv("BROTHER_EXPORTER_PRINTER_HOST", "192.168.1.200")
	_ = os.Setenv("BROTHER_EXPORTER_PRINTER_COMMUNITY", "private")
	_ = os.Setenv("BROTHER_EXPORTER_PRINTER_TYPE", "ink")

	defer func() {
		_ = os.Unsetenv("BROTHER_EXPORTER_SERVER_HOST")
		_ = os.Unsetenv("BROTHER_EXPORTER_SERVER_PORT")
		_ = os.Unsetenv("BROTHER_EXPORTER_LOG_LEVEL")
		_ = os.Unsetenv("BROTHER_EXPORTER_LOG_FORMAT")
		_ = os.Unsetenv("BROTHER_EXPORTER_METRICS_DEFAULT_INTERVAL")
		_ = os.Unsetenv("BROTHER_EXPORTER_PRINTER_HOST")
		_ = os.Unsetenv("BROTHER_EXPORTER_PRINTER_COMMUNITY")
		_ = os.Unsetenv("BROTHER_EXPORTER_PRINTER_TYPE")
	}()

	config, err := LoadConfig("", true)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Test environment-loaded values
	if config.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host = %v, want 127.0.0.1", config.Server.Host)
	}

	if config.Server.Port != 9090 {
		t.Errorf("Server.Port = %v, want 9090", config.Server.Port)
	}

	if config.Logging.Level != "debug" {
		t.Errorf("Logging.Level = %v, want debug", config.Logging.Level)
	}

	if config.Logging.Format != "text" {
		t.Errorf("Logging.Format = %v, want text", config.Logging.Format)
	}

	if config.Metrics.Collection.DefaultInterval.Seconds() != 45 {
		t.Errorf("Metrics.Collection.DefaultInterval = %v, want 45s", config.Metrics.Collection.DefaultInterval.Seconds())
	}

	if config.Printer.Host != "192.168.1.200" {
		t.Errorf("Printer.Host = %v, want 192.168.1.200", config.Printer.Host)
	}

	if config.Printer.Community != "private" {
		t.Errorf("Printer.Community = %v, want private", config.Printer.Community)
	}

	if config.Printer.Type != "ink" {
		t.Errorf("Printer.Type = %v, want ink", config.Printer.Type)
	}
}

func TestLoadConfig_EnvMode_MissingPrinterHost(t *testing.T) {
	// Don't set BROTHER_EXPORTER_PRINTER_HOST
	_ = os.Setenv("BROTHER_EXPORTER_SERVER_HOST", "127.0.0.1")

	defer func() {
		_ = os.Unsetenv("BROTHER_EXPORTER_SERVER_HOST")
	}()

	_, err := LoadConfig("", true)
	if err == nil {
		t.Error("LoadConfig() expected error for missing printer host")
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"valid integer", "123", 123, false},
		{"zero", "0", 0, false},
		{"negative", "-456", -456, false},
		{"invalid format", "12.34", 0, true},
		{"invalid characters", "abc", 0, true},
		{"empty string", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInt(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("parseInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				BaseConfig: promexporter_config.BaseConfig{
					Server: promexporter_config.ServerConfig{
						Host: "0.0.0.0",
						Port: 8080,
					},
					Logging: promexporter_config.LoggingConfig{
						Level:  "info",
						Format: "json",
					},
					Metrics: promexporter_config.MetricsConfig{
						Collection: promexporter_config.CollectionConfig{
							DefaultInterval: Duration{Duration: 30 * time.Second},
						},
					},
				},
				Printer: PrinterConfig{
					Host:      "192.168.1.100",
					Community: "public",
					Type:      "laser",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid port - too low",
			config: &Config{
				Server: ServerConfig{
					Host: "0.0.0.0",
					Port: 0,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				Metrics: MetricsConfig{
					Collection: CollectionConfig{
						DefaultInterval: Duration{Duration: 30 * time.Second},
					},
				},
				Printer: PrinterConfig{
					Host:      "192.168.1.100",
					Community: "public",
					Type:      "laser",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid port - too high",
			config: &Config{
				Server: ServerConfig{
					Host: "0.0.0.0",
					Port: 65536,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				Metrics: MetricsConfig{
					Collection: CollectionConfig{
						DefaultInterval: Duration{Duration: 30 * time.Second},
					},
				},
				Printer: PrinterConfig{
					Host:      "192.168.1.100",
					Community: "public",
					Type:      "laser",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid logging level",
			config: &Config{
				Server: ServerConfig{
					Host: "0.0.0.0",
					Port: 8080,
				},
				Logging: LoggingConfig{
					Level:  "invalid",
					Format: "json",
				},
				Metrics: MetricsConfig{
					Collection: CollectionConfig{
						DefaultInterval: Duration{Duration: 30 * time.Second},
					},
				},
				Printer: PrinterConfig{
					Host:      "192.168.1.100",
					Community: "public",
					Type:      "laser",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid logging format",
			config: &Config{
				Server: ServerConfig{
					Host: "0.0.0.0",
					Port: 8080,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "invalid",
				},
				Metrics: MetricsConfig{
					Collection: CollectionConfig{
						DefaultInterval: Duration{Duration: 30 * time.Second},
					},
				},
				Printer: PrinterConfig{
					Host:      "192.168.1.100",
					Community: "public",
					Type:      "laser",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid metrics interval - too low",
			config: &Config{
				Server: ServerConfig{
					Host: "0.0.0.0",
					Port: 8080,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				Metrics: MetricsConfig{
					Collection: CollectionConfig{
						DefaultInterval: Duration{Duration: 0},
					},
				},
				Printer: PrinterConfig{
					Host:      "192.168.1.100",
					Community: "public",
					Type:      "laser",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid metrics interval - too high",
			config: &Config{
				Server: ServerConfig{
					Host: "0.0.0.0",
					Port: 8080,
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "json",
				},
				Metrics: MetricsConfig{
					Collection: CollectionConfig{
						DefaultInterval: Duration{Duration: 25 * time.Hour},
					},
				},
				Printer: PrinterConfig{
					Host:      "192.168.1.100",
					Community: "public",
					Type:      "laser",
				},
			},
			wantErr: true,
		},
		{
			name: "missing printer host",
			config: &Config{
				BaseConfig: promexporter_config.BaseConfig{
					Server: promexporter_config.ServerConfig{
						Host: "0.0.0.0",
						Port: 8080,
					},
					Logging: promexporter_config.LoggingConfig{
						Level:  "info",
						Format: "json",
					},
					Metrics: promexporter_config.MetricsConfig{
						Collection: promexporter_config.CollectionConfig{
							DefaultInterval: Duration{Duration: 30 * time.Second},
						},
					},
				},
				Printer: PrinterConfig{
					Host:      "",
					Community: "public",
					Type:      "laser",
				},
			},
			wantErr: true,
		},
		{
			name: "missing printer community",
			config: &Config{
				BaseConfig: promexporter_config.BaseConfig{
					Server: promexporter_config.ServerConfig{
						Host: "0.0.0.0",
						Port: 8080,
					},
					Logging: promexporter_config.LoggingConfig{
						Level:  "info",
						Format: "json",
					},
					Metrics: promexporter_config.MetricsConfig{
						Collection: promexporter_config.CollectionConfig{
							DefaultInterval: Duration{Duration: 30 * time.Second},
						},
					},
				},
				Printer: PrinterConfig{
					Host:      "192.168.1.100",
					Community: "",
					Type:      "laser",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid printer type",
			config: &Config{
				BaseConfig: promexporter_config.BaseConfig{
					Server: promexporter_config.ServerConfig{
						Host: "0.0.0.0",
						Port: 8080,
					},
					Logging: promexporter_config.LoggingConfig{
						Level:  "info",
						Format: "json",
					},
					Metrics: promexporter_config.MetricsConfig{
						Collection: promexporter_config.CollectionConfig{
							DefaultInterval: Duration{Duration: 30 * time.Second},
						},
					},
				},
				Printer: PrinterConfig{
					Host:      "192.168.1.100",
					Community: "public",
					Type:      "invalid",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_GetDefaultInterval(t *testing.T) {
	config := &Config{
		Metrics: MetricsConfig{
			Collection: CollectionConfig{
				DefaultInterval: Duration{Duration: 45 * time.Second},
			},
		},
	}

	if got := config.GetDefaultInterval(); got != 45 {
		t.Errorf("Config.GetDefaultInterval() = %v, want 45", got)
	}
}
