package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Missing file fallback
	cfg, err := LoadConfig("nonexistent.json")
	if err != nil {
		t.Fatalf("unexpected error on missing file: %v", err)
	}
	if cfg.Count != 1 || cfg.TorSOCKS != "127.0.0.1:9050" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}

	// Valid file parsing
	tmp := filepath.Join(t.TempDir(), "config.json")
	_ = os.WriteFile(tmp, []byte(`{"count": 5, "concurrency": 2, "promo_code": "TEST", "email_domain": "mailmmo.io.vn"}`), 0644)
	custom, err := LoadConfig(tmp)
	if err != nil {
		t.Fatalf("unexpected error on valid file: %v", err)
	}
	if custom.Count != 5 || custom.Concurrency != 2 || custom.PromoCode != "TEST" || custom.EmailDomain != "mailmmo.io.vn" {
		t.Fatalf("failed to parse custom values: %+v", custom)
	}

	// Corrupt file must fail loudly instead of silently using defaults
	bad := filepath.Join(t.TempDir(), "bad.json")
	_ = os.WriteFile(bad, []byte(`{not json`), 0644)
	if _, err := LoadConfig(bad); err == nil {
		t.Fatal("expected error on corrupt file, got nil")
	}
}
