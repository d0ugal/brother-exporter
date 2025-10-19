package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/d0ugal/brother-exporter/internal/collectors"
	"github.com/d0ugal/brother-exporter/internal/config"
	"github.com/d0ugal/brother-exporter/internal/metrics"
	"github.com/d0ugal/promexporter/app"
	"github.com/d0ugal/promexporter/logging"
	promexporter_metrics "github.com/d0ugal/promexporter/metrics"
	"github.com/d0ugal/promexporter/version"
)

// hasEnvironmentVariables checks if any BROTHER_EXPORTER_* environment variables are set
func hasEnvironmentVariables() bool {
	envVars := []string{
		"BROTHER_EXPORTER_SERVER_HOST",
		"BROTHER_EXPORTER_SERVER_PORT",
		"BROTHER_EXPORTER_LOG_LEVEL",
		"BROTHER_EXPORTER_LOG_FORMAT",
		"BROTHER_EXPORTER_METRICS_DEFAULT_INTERVAL",
		"BROTHER_EXPORTER_PRINTER_HOST",
		"BROTHER_EXPORTER_PRINTER_COMMUNITY",
		"BROTHER_EXPORTER_PRINTER_TYPE",
	}

	for _, envVar := range envVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}

	return false
}

func main() {
	// Parse command line flags
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "v", false, "Show version information")

	var (
		configPath    string
		configFromEnv bool
	)

	flag.StringVar(&configPath, "config", "config.yaml", "Path to configuration file")
	flag.BoolVar(&configFromEnv, "config-from-env", false, "Load configuration from environment variables only")
	flag.Parse()

	// Show version if requested
	if showVersion {
		versionInfo := version.Get()
		fmt.Printf("brother-exporter %s\n", versionInfo.Version)
		fmt.Printf("Commit: %s\n", versionInfo.Commit)
		fmt.Printf("Build Date: %s\n", versionInfo.BuildDate)
		fmt.Printf("Go Version: %s\n", versionInfo.GoVersion)
		os.Exit(0)
	}

	// Use environment variable if config flag is not provided
	if configPath == "config.yaml" && !configFromEnv {
		if envConfig := os.Getenv("CONFIG_PATH"); envConfig != "" {
			configPath = envConfig
		}
	}

	// Check if we should use environment-only configuration
	if !configFromEnv {
		// Check explicit flag first
		if os.Getenv("BROTHER_EXPORTER_CONFIG_FROM_ENV") == "true" {
			configFromEnv = true
		} else if hasEnvironmentVariables() {
			// Auto-detect environment variables and use them
			configFromEnv = true
		}
	}

	// Load configuration
	cfg, err := config.LoadConfig(configPath, configFromEnv)
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Configure logging using promexporter
	logging.Configure(&logging.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
	})

	// Initialize metrics registry using promexporter
	metricsRegistry := promexporter_metrics.NewRegistry("brother_exporter_info")

	// Add custom metrics to the registry
	brotherRegistry := metrics.NewBrotherRegistry(metricsRegistry)

	// Create collector
	brotherCollector := collectors.NewBrotherCollector(cfg, brotherRegistry)

	// Create and run application using promexporter
	application := app.New("Brother Exporter").
		WithConfig(&cfg.BaseConfig).
		WithMetrics(metricsRegistry).
		WithCollector(brotherCollector).
		Build()

	if err := application.Run(); err != nil {
		slog.Error("Application failed", "error", err)
		os.Exit(1)
	}
}
