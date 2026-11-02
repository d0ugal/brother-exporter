package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression test: the embedded promexporter BaseConfig must be inlined so
// that yaml.v3 populates the server, logging and metrics sections. Without
// the `,inline` tag yaml.v3 keys the embedded struct under a non-existent
// "baseconfig" key and silently drops those sections, leaving defaults.
func TestLoadConfigParsesEmbeddedBaseConfig(t *testing.T) {
	yaml := `
server:
  host: "127.0.0.1"
  port: 8092

logging:
  level: "debug"
  format: "text"

metrics:
  collection:
    default_interval: "45s"

printer:
  host: "192.168.2.123"
  community: "public"
  type: "laser"
`
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(yaml), 0o600))

	cfg, err := LoadConfig(path)
	require.NoError(t, err)

	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, 8092, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "text", cfg.Logging.Format)
	assert.Equal(t, 45, cfg.Metrics.Collection.DefaultInterval.Seconds())
	assert.Equal(t, "192.168.2.123", cfg.Printer.Host)
	assert.Equal(t, "laser", cfg.Printer.Type)
}
