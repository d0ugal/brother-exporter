package config

import (
	"fmt"
	"os"
	"time"

	promexporter_config "github.com/d0ugal/promexporter/config"
	"gopkg.in/yaml.v3"
)

// Duration uses promexporter Duration type
type Duration = promexporter_config.Duration

type Config struct {
	promexporter_config.BaseConfig

	Printer PrinterConfig `yaml:"printer"`
}

type PrinterConfig struct {
	Host       string   `yaml:"host"`
	Community  string   `yaml:"community"`
	Type       string   `yaml:"type"`
	Interfaces []string `yaml:"interfaces"`
}

// LoadConfig loads configuration with priority: env vars > yaml file > defaults.
// The yaml file is optional; if path is empty or the file does not exist it is
// silently skipped. Environment variables are always applied on top.
func LoadConfig(path string) (*Config, error) {
	var cfg Config

	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
			}
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
		}
	}

	if err := promexporter_config.ApplyGenericEnvVars(&cfg.BaseConfig); err != nil {
		return nil, fmt.Errorf("failed to apply generic environment variables: %w", err)
	}

	applyEnvVars(&cfg)
	setDefaults(&cfg)

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &cfg, nil
}

// applyEnvVars overlays Brother-exporter environment variables onto cfg.
// Only variables that are set (non-empty) are applied.
func applyEnvVars(cfg *Config) {
	if host := os.Getenv("BROTHER_EXPORTER_SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if portStr := os.Getenv("BROTHER_EXPORTER_SERVER_PORT"); portStr != "" {
		if port, err := parseInt(portStr); err == nil {
			cfg.Server.Port = port
		}
	}
	if level := os.Getenv("BROTHER_EXPORTER_LOG_LEVEL"); level != "" {
		cfg.Logging.Level = level
	}
	if format := os.Getenv("BROTHER_EXPORTER_LOG_FORMAT"); format != "" {
		cfg.Logging.Format = format
	}
	if intervalStr := os.Getenv("BROTHER_EXPORTER_METRICS_DEFAULT_INTERVAL"); intervalStr != "" {
		if interval, err := time.ParseDuration(intervalStr); err == nil {
			cfg.Metrics.Collection.DefaultInterval = promexporter_config.Duration{Duration: interval}
			cfg.Metrics.Collection.DefaultIntervalSet = true
		}
	}
	if host := os.Getenv("BROTHER_EXPORTER_PRINTER_HOST"); host != "" {
		cfg.Printer.Host = host
	}
	if community := os.Getenv("BROTHER_EXPORTER_PRINTER_COMMUNITY"); community != "" {
		cfg.Printer.Community = community
	}
	if printerType := os.Getenv("BROTHER_EXPORTER_PRINTER_TYPE"); printerType != "" {
		cfg.Printer.Type = printerType
	}
}

// parseInt parses a string to int
func parseInt(s string) (int, error) {
	var i int

	_, err := fmt.Sscanf(s, "%d", &i)
	if err != nil {
		return 0, err
	}
	// Check if there are any remaining characters (like decimal points)
	if len(fmt.Sprintf("%d", i)) != len(s) {
		return 0, fmt.Errorf("invalid integer format: %s", s)
	}

	return i, nil
}

// setDefaults sets default values for configuration
func setDefaults(config *Config) {
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}

	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}

	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}

	if config.Logging.Format == "" {
		config.Logging.Format = "json"
	}

	if !config.Metrics.Collection.DefaultIntervalSet {
		config.Metrics.Collection.DefaultInterval = promexporter_config.Duration{Duration: time.Second * 30}
	}

	if config.Printer.Host == "" {
		config.Printer.Host = "192.168.1.100"
	}

	if config.Printer.Community == "" {
		config.Printer.Community = "public"
	}

	if config.Printer.Type == "" {
		config.Printer.Type = "laser"
	}
}

// Validate performs comprehensive validation of the configuration
func (c *Config) Validate() error {
	// Validate server configuration
	if err := c.validateServerConfig(); err != nil {
		return fmt.Errorf("server config: %w", err)
	}

	// Validate logging configuration
	if err := c.validateLoggingConfig(); err != nil {
		return fmt.Errorf("logging config: %w", err)
	}

	// Validate metrics configuration
	if err := c.validateMetricsConfig(); err != nil {
		return fmt.Errorf("metrics config: %w", err)
	}

	// Validate printer configuration
	if err := c.validatePrinterConfig(); err != nil {
		return fmt.Errorf("printer config: %w", err)
	}

	return nil
}

func (c *Config) validateServerConfig() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", c.Server.Port)
	}

	return nil
}

func (c *Config) validateLoggingConfig() error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalid logging level: %s", c.Logging.Level)
	}

	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}
	if !validFormats[c.Logging.Format] {
		return fmt.Errorf("invalid logging format: %s", c.Logging.Format)
	}

	return nil
}

func (c *Config) validateMetricsConfig() error {
	if c.Metrics.Collection.DefaultInterval.Seconds() < 1 {
		return fmt.Errorf("default interval must be at least 1 second, got %d", c.Metrics.Collection.DefaultInterval.Seconds())
	}

	if c.Metrics.Collection.DefaultInterval.Seconds() > 86400 {
		return fmt.Errorf("default interval must be at most 86400 seconds (24 hours), got %d", c.Metrics.Collection.DefaultInterval.Seconds())
	}

	return nil
}

func (c *Config) validatePrinterConfig() error {
	if c.Printer.Host == "" {
		return fmt.Errorf("printer host is required")
	}

	if c.Printer.Community == "" {
		return fmt.Errorf("printer community is required")
	}

	if c.Printer.Type == "" {
		return fmt.Errorf("printer type is required")
	}

	return nil
}

// GetDefaultInterval returns the default collection interval
func (c *Config) GetDefaultInterval() int {
	return c.Metrics.Collection.DefaultInterval.Seconds()
}
