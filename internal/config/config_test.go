package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTTL(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr bool
	}{
		{"1h", 3600, false},
		{"24h", 86400, false},
		{"30m", 1800, false},
		{"7d", 604800, false},
		{"30d", 2592000, false},
		{"1d", 86400, false},
		{"3600", 3600, false},
		{"86400", 86400, false},
		{"", 0, true},
		{"0", 0, true},
		{"-1", 0, true},
		{"0d", 0, true},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseTTL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTTL(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseTTL(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.toml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Endpoint != "https://sh.techops.services" {
		t.Errorf("expected default endpoint, got %s", cfg.Endpoint)
	}
	if cfg.DefaultTTL != "24h" {
		t.Errorf("expected default TTL 24h, got %s", cfg.DefaultTTL)
	}
	if !cfg.Clipboard {
		t.Error("expected clipboard to be true by default")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	original := &Config{
		Endpoint:   "https://custom.example.com",
		APIKey:     "sk_test_key",
		DefaultTTL: "7d",
		Clipboard:  false,
	}

	if err := Save(path, original); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if loaded.Endpoint != original.Endpoint {
		t.Errorf("endpoint: got %s, want %s", loaded.Endpoint, original.Endpoint)
	}
	if loaded.APIKey != original.APIKey {
		t.Errorf("api_key: got %s, want %s", loaded.APIKey, original.APIKey)
	}
	if loaded.DefaultTTL != original.DefaultTTL {
		t.Errorf("default_ttl: got %s, want %s", loaded.DefaultTTL, original.DefaultTTL)
	}
	if loaded.Clipboard != original.Clipboard {
		t.Errorf("clipboard: got %v, want %v", loaded.Clipboard, original.Clipboard)
	}
}

func TestEnvOverrides(t *testing.T) {
	t.Setenv("SHARE_ENDPOINT", "https://env.example.com")
	t.Setenv("SHARE_API_KEY", "sk_env_key")
	t.Setenv("SHARE_DEFAULT_TTL", "48h")
	t.Setenv("SHARE_CLIPBOARD", "false")

	cfg, err := Load("/nonexistent/path/config.toml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Endpoint != "https://env.example.com" {
		t.Errorf("expected env endpoint, got %s", cfg.Endpoint)
	}
	if cfg.APIKey != "sk_env_key" {
		t.Errorf("expected env api key, got %s", cfg.APIKey)
	}
	if cfg.DefaultTTL != "48h" {
		t.Errorf("expected env TTL, got %s", cfg.DefaultTTL)
	}
	if cfg.Clipboard {
		t.Error("expected clipboard false from env")
	}
}

func TestSaveCreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "deep", "config.toml")

	cfg := DefaultConfig()
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected config file to exist")
	}
}

func TestConfigFilePermissions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	cfg := &Config{
		APIKey: "sk_secret",
	}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat error: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("expected file mode 0600, got %o", perm)
	}
}
