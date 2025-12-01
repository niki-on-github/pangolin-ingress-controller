package unit

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wizzz/pangolin-ingress-controller/internal/config"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear environment
	os.Clearenv()

	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, "default", cfg.DefaultTunnelName)
	assert.Equal(t, "http", cfg.BackendScheme)
	assert.Equal(t, 5*time.Minute, cfg.ResyncPeriod)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Empty(t, cfg.WatchNamespaces)
	assert.Empty(t, cfg.TunnelMapping)
}

func TestLoadConfig_FromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("PIC_DEFAULT_TUNNEL_NAME", "my-tunnel")
	os.Setenv("PIC_BACKEND_SCHEME", "https")
	os.Setenv("PIC_RESYNC_PERIOD", "10m")
	os.Setenv("PIC_LOG_LEVEL", "debug")
	os.Setenv("PIC_WATCH_NAMESPACES", "ns1,ns2,ns3")
	defer os.Clearenv()

	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, "my-tunnel", cfg.DefaultTunnelName)
	assert.Equal(t, "https", cfg.BackendScheme)
	assert.Equal(t, 10*time.Minute, cfg.ResyncPeriod)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, []string{"ns1", "ns2", "ns3"}, cfg.WatchNamespaces)
}

func TestLoadConfig_TunnelMapping(t *testing.T) {
	os.Setenv("PIC_TUNNEL_CLASS_MAPPING", "eu=tunnel-eu\nus=tunnel-us\nstaging=staging-tunnel")
	defer os.Clearenv()

	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, "tunnel-eu", cfg.TunnelMapping["eu"])
	assert.Equal(t, "tunnel-us", cfg.TunnelMapping["us"])
	assert.Equal(t, "staging-tunnel", cfg.TunnelMapping["staging"])
}

func TestLoadConfig_InvalidResyncPeriod(t *testing.T) {
	os.Setenv("PIC_RESYNC_PERIOD", "invalid")
	defer os.Clearenv()

	_, err := config.Load()
	assert.Error(t, err)
}
