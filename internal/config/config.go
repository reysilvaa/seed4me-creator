package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// EmailService mendefinisikan tipe layanan email sementara.
type EmailService string

const (
	EmailServiceTempMailIng EmailService = "tempmailing"
	EmailServiceTempMailLol EmailService = "tempmaillol"
	EmailServiceCatchMail   EmailService = "catchmail"
)

// Config menyimpan seluruh parameter pembuatan akun dan konfigurasi sistem.
type Config struct {
	Count       int    `json:"count"`
	Concurrency int    `json:"concurrency"`
	Password    string `json:"password"`
	PromoCode   string `json:"promo_code"`
	MaxRetries  int    `json:"max_retries"`

	Proxy    string `json:"proxy"`
	TorSOCKS string `json:"tor_socks"`

	EmailService       EmailService `json:"email_service"`
	TempMailLolKey     string       `json:"tempmail_lol_key"`
	EmailDomain        string       `json:"email_domain"`
	PollTimeoutSeconds int          `json:"poll_timeout_seconds"`

	JSONFile string `json:"json_file"`
	TXTFile  string `json:"txt_file"`
}

// DefaultConfig mengembalikan konfigurasi default standar.
func DefaultConfig() Config {
	return Config{
		Count:              1,
		Concurrency:        4,
		TorSOCKS:           "127.0.0.1:9050",
		EmailService:       EmailServiceTempMailLol,
		MaxRetries:         3,
		PollTimeoutSeconds: 45,
		JSONFile:           "accounts.json",
		TXTFile:            "accounts.txt",
	}
}

// Normalize mengisi nilai kosong/tidak valid dengan default aman untuk CLI.
func (cfg *Config) Normalize() {
	def := DefaultConfig()
	if cfg.Count < 1 {
		cfg.Count = def.Count
	}
	if cfg.Concurrency < 1 {
		cfg.Concurrency = def.Concurrency
	}
	if cfg.MaxRetries < 1 {
		cfg.MaxRetries = def.MaxRetries
	}
	if cfg.PollTimeoutSeconds < 5 {
		cfg.PollTimeoutSeconds = def.PollTimeoutSeconds
	}
	if cfg.EmailService == "" {
		cfg.EmailService = def.EmailService
	}
	if cfg.JSONFile == "" {
		cfg.JSONFile = def.JSONFile
	}
	if cfg.TXTFile == "" {
		cfg.TXTFile = def.TXTFile
	}
}

// Validate menolak konfigurasi yang tidak aman atau tidak didukung.
func (cfg Config) Validate() error {
	if cfg.Count < 1 || cfg.Concurrency < 1 {
		return fmt.Errorf("count dan concurrency harus lebih besar dari 0")
	}
	if cfg.MaxRetries < 1 {
		return fmt.Errorf("max_retries harus lebih besar dari 0")
	}
	if cfg.PollTimeoutSeconds < 5 {
		return fmt.Errorf("poll_timeout_seconds minimal 5")
	}
	if len(cfg.Password) < 8 {
		return fmt.Errorf("password minimal 8 karakter")
	}
	if cfg.JSONFile == "" || cfg.TXTFile == "" {
		return fmt.Errorf("json_file dan txt_file wajib diisi")
	}
	if filepath.Clean(cfg.JSONFile) == filepath.Clean(cfg.TXTFile) {
		return fmt.Errorf("json_file dan txt_file harus berbeda")
	}
	if cfg.Proxy != "" {
		if err := validateProxyURL(cfg.Proxy); err != nil {
			return fmt.Errorf("proxy tidak valid: %w", err)
		}
	}
	if cfg.EmailService != EmailServiceTempMailIng &&
		cfg.EmailService != EmailServiceTempMailLol &&
		cfg.EmailService != EmailServiceCatchMail {
		return fmt.Errorf("provider email tidak dikenal: %q", cfg.EmailService)
	}
	if cfg.EmailService == EmailServiceTempMailLol && strings.TrimSpace(cfg.TempMailLolKey) == "" {
		return fmt.Errorf("tempmail_lol_key belum diset")
	}
	if domain := strings.TrimSpace(cfg.EmailDomain); domain != "" && strings.ContainsAny(domain, "@/\\ \t\r\n") {
		return fmt.Errorf("email_domain tidak valid")
	}
	return nil
}

func validateProxyURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https", "socks5", "socks5h":
	default:
		return fmt.Errorf("scheme harus http, https, socks5, atau socks5h")
	}
	if u.Host == "" {
		return fmt.Errorf("host kosong")
	}
	return nil
}

// LoadConfig membaca JSON; file yang hilang tetap memakai default, file korup gagal jelas.
func LoadConfig(path string) (Config, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	if err == nil {
		if len(strings.TrimSpace(string(data))) > 0 {
			if err := json.Unmarshal(data, &cfg); err != nil {
				return Config{}, fmt.Errorf("gagal membaca %s: %w", path, err)
			}
		}
	} else if !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("gagal membaca %s: %w", path, err)
	}

	if strings.TrimSpace(cfg.TempMailLolKey) == "" {
		cfg.TempMailLolKey = strings.TrimSpace(os.Getenv("TEMPMAIL_LOL_KEY"))
	}
	cfg.Normalize()
	return cfg, nil
}
