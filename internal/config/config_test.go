package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Missing file fallback
	cfg := LoadConfig("nonexistent.json")
	if cfg.Count != 1 || cfg.TorSOCKS != "127.0.0.1:9050" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}

	// Valid file parsing
	tmp := filepath.Join(t.TempDir(), "config.json")
	_ = os.WriteFile(tmp, []byte(`{"count": 5, "concurrency": 2, "promo_code": "TEST", "email_domain": "mailmmo.io.vn"}`), 0644)
	custom := LoadConfig(tmp)
	if custom.Count != 5 || custom.Concurrency != 2 || custom.PromoCode != "TEST" || custom.EmailDomain != "mailmmo.io.vn" {
		t.Fatalf("failed to parse custom values: %+v", custom)
	}
}
