// Package config provides configuration loading for the Pangolin Ingress Controller.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Config holds the runtime configuration for PIC.
type Config struct {
	// DefaultTunnelName is the tunnel used when ingressClassName is exactly "pangolin"
	DefaultTunnelName string

	// TunnelMapping maps ingressClass suffixes to tunnel names
	// e.g., "eu" -> "tunnel-eu" for ingressClassName "pangolin-eu"
	TunnelMapping map[string]string

	// BackendScheme is the protocol for backend services ("http" or "https")
	BackendScheme string

	// ResyncPeriod is how often to re-reconcile all resources
	ResyncPeriod time.Duration

	// LogLevel controls logging verbosity ("debug", "info", "warn", "error")
	LogLevel string

	// WatchNamespaces limits which namespaces to watch (empty = all)
	WatchNamespaces []string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		DefaultTunnelName: getEnv("PIC_DEFAULT_TUNNEL_NAME", "default"),
		BackendScheme:     getEnv("PIC_BACKEND_SCHEME", "http"),
		LogLevel:          getEnv("PIC_LOG_LEVEL", "info"),
		TunnelMapping:     make(map[string]string),
	}

	// Parse resync period
	resyncStr := getEnv("PIC_RESYNC_PERIOD", "5m")
	resync, err := time.ParseDuration(resyncStr)
	if err != nil {
		return nil, fmt.Errorf("invalid PIC_RESYNC_PERIOD %q: %w", resyncStr, err)
	}
	cfg.ResyncPeriod = resync

	// Parse watch namespaces
	if ns := getEnv("PIC_WATCH_NAMESPACES", ""); ns != "" {
		cfg.WatchNamespaces = strings.Split(ns, ",")
		for i := range cfg.WatchNamespaces {
			cfg.WatchNamespaces[i] = strings.TrimSpace(cfg.WatchNamespaces[i])
		}
	}

	// Parse tunnel mapping
	if mapping := getEnv("PIC_TUNNEL_CLASS_MAPPING", ""); mapping != "" {
		lines := strings.Split(mapping, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				cfg.TunnelMapping[key] = value
			}
		}
	}

	return cfg, nil
}

// MustLoad loads configuration or panics on error.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
