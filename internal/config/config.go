package config

import (
	"encoding/json"
	"os"
)

// EmailService mendefinisikan tipe layanan email sementara
type EmailService string

const (
	EmailServiceTempMailIng EmailService = "tempmailing"
	EmailServiceTempMailLol EmailService = "tempmaillol"
	EmailServiceCatchMail   EmailService = "catchmail"
)

// Config menyimpan seluruh parameter pembuatan akun dan konfigurasi sistem
type Config struct {
	// Pengaturan Akun
	Count       int    `json:"count"`
	Concurrency int    `json:"concurrency"`
	Password    string `json:"password"`
	PromoCode   string `json:"promo_code"`
	MaxRetries  int    `json:"max_retries"`

	// Jaringan & Proxy
	Proxy    string `json:"proxy"`
	TorSOCKS string `json:"tor_socks"`

	// Provider Email
	EmailService       EmailService `json:"email_service"`
	TempMailLolKey     string       `json:"tempmail_lol_key"`
	EmailDomain        string       `json:"email_domain"`
	PollTimeoutSeconds int          `json:"poll_timeout_seconds"`

	// Penyimpanan Output
	JSONFile string `json:"json_file"`
	TXTFile  string `json:"txt_file"`
}

// DefaultConfig mengembalikan konfigurasi default standar
func DefaultConfig() Config {
	return Config{
		Count:              1,
		Concurrency:        4,
		Password:           "12345678",
		TorSOCKS:           "127.0.0.1:9050",
		EmailService:       EmailServiceTempMailLol,
		EmailDomain:        "catchmail.io",
		MaxRetries:         3,
		PollTimeoutSeconds: 45,
		JSONFile:           "accounts.json",
		TXTFile:            "accounts.txt",
	}
}

// Normalize memvalidasi nilai konfigurasi agar selalu dalam batas aman (single source of truth dari DefaultConfig)
func (cfg *Config) Normalize() {
	def := DefaultConfig()
	if cfg.Count < 1 {
		cfg.Count = def.Count
	}
	if cfg.Concurrency < 1 {
		cfg.Concurrency = def.Concurrency
	}
	if cfg.Password == "" {
		cfg.Password = def.Password
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

// LoadConfig membaca file konfigurasi JSON dan menerapkan fallback default
// Key TempMail.lol diambil dari config.json, fallback ke env TEMPMAIL_LOL_KEY (jangan hardcode di source)
func LoadConfig(path string) Config {
	cfg := DefaultConfig()
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &cfg)
	}
	if cfg.TempMailLolKey == "" {
		cfg.TempMailLolKey = os.Getenv("TEMPMAIL_LOL_KEY")
	}
	cfg.Normalize()
	return cfg
}
