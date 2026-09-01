package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Count       int    `json:"count"`
	Concurrency int    `json:"concurrency"`
	PromoCode   string `json:"promo_code"`
	Proxy       string `json:"proxy"`
	TorSOCKS    string `json:"tor_socks"`
	EmailDomain string `json:"email_domain"`
	JSONFile    string `json:"json_file"`
	TXTFile     string `json:"txt_file"`
}

func LoadConfig(path string) Config {
	cfg := Config{
		Count:       1,
		Concurrency: 1,
		TorSOCKS:    "127.0.0.1:9050",
		EmailDomain: "catchmail.io",
		JSONFile:    "accounts.json",
		TXTFile:     "accounts.txt",
	}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &cfg)
	}
	if cfg.Count < 1 {
		cfg.Count = 1
	}
	if cfg.Concurrency < 1 {
		cfg.Concurrency = 1
	}
	if cfg.EmailDomain == "" {
		cfg.EmailDomain = "catchmail.io"
	}
	if cfg.JSONFile == "" {
		cfg.JSONFile = "accounts.json"
	}
	if cfg.TXTFile == "" {
		cfg.TXTFile = "accounts.txt"
	}
	return cfg
}
