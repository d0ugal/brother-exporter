package server

import (
	"testing"

	"github.com/d0ugal/brother-exporter/internal/config"
	"github.com/d0ugal/brother-exporter/internal/metrics"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "127.0.0.1",
			Port: 8080,
		},
		Printer: config.PrinterConfig{
			Host:      "192.168.1.100",
			Community: "public",
			Type:      "laser",
		},
	}
	metricsRegistry := metrics.NewRegistry()

	server := New(cfg, metricsRegistry)

	assert.NotNil(t, server)
	assert.Equal(t, cfg, server.config)
	assert.Equal(t, metricsRegistry, server.metrics)
	assert.NotNil(t, server.router)
	// Note: server.server is set in Start(), not in New()
}
