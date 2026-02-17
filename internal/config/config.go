package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	defaultEndpoint = "https://sh.techops.services"
	defaultTTL      = "24h"
	defaultListSize = 50
)

// Config holds all configuration for the share CLI.
type Config struct {
	Endpoint   string `toml:"endpoint"`
	APIKey     string `toml:"api_key"`
	DefaultTTL string `toml:"default_ttl"`
	Clipboard  bool   `toml:"clipboard"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Endpoint:   defaultEndpoint,
		DefaultTTL: defaultTTL,
		Clipboard:  true,
	}
}

// DefaultConfigPath returns the platform default config path.
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "share", "config.toml")
}

// Load reads the config file from disk. Returns DefaultConfig if the file
// does not exist (config is optional). Environment variables override file values.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path == "" {
		path = DefaultConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			applyEnv(cfg)
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	applyEnv(cfg)
	return cfg, nil
}

// Save writes the config to disk, creating parent directories as needed.
func Save(path string, cfg *Config) error {
	if path == "" {
		path = DefaultConfigPath()
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config directory %s: %w", dir, err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("opening config file %s: %w", path, err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("writing config file %s: %w", path, err)
	}
	return nil
}

// applyEnv overrides config values with environment variables.
func applyEnv(cfg *Config) {
	if v := os.Getenv("SHARE_ENDPOINT"); v != "" {
		cfg.Endpoint = v
	}
	if v := os.Getenv("SHARE_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("SHARE_DEFAULT_TTL"); v != "" {
		cfg.DefaultTTL = v
	}
	if v := os.Getenv("SHARE_CLIPBOARD"); v != "" {
		cfg.Clipboard = v == "true" || v == "1"
	}
}

// ParseTTL converts a human-readable duration string to seconds.
// Supported formats: "1h", "7d", "30d", "3600" (raw seconds).
func ParseTTL(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty TTL value")
	}

	// Try raw seconds first.
	if secs, err := strconv.Atoi(s); err == nil {
		if secs <= 0 {
			return 0, fmt.Errorf("TTL must be positive, got %d", secs)
		}
		return secs, nil
	}

	// Try Go duration (supports "1h", "30m", "2h30m", etc.)
	if d, err := time.ParseDuration(s); err == nil {
		secs := int(d.Seconds())
		if secs <= 0 {
			return 0, fmt.Errorf("TTL must be positive, got %s", s)
		}
		return secs, nil
	}

	// Handle day suffix: "7d", "30d"
	if strings.HasSuffix(s, "d") {
		numStr := strings.TrimSuffix(s, "d")
		days, err := strconv.Atoi(numStr)
		if err != nil {
			return 0, fmt.Errorf("invalid TTL format: %q", s)
		}
		if days <= 0 {
			return 0, fmt.Errorf("TTL must be positive, got %s", s)
		}
		return days * 86400, nil
	}

	return 0, fmt.Errorf("invalid TTL format: %q (use e.g. 1h, 7d, 30d, or raw seconds)", s)
}
