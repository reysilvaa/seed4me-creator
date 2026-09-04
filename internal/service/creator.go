package service

import (
	"fmt"
	"strings"
	"time"

	"seed4me-creator/internal/config"
	"seed4me-creator/internal/email"
	"seed4me-creator/internal/model"
	"seed4me-creator/internal/proxy"
	"seed4me-creator/internal/seed4me"
	"seed4me-creator/internal/storage"
)

func CreateAccount(cfg config.Config, logFn func(string)) (*model.Account, error) {
	provider, err := email.GetProvider(cfg)
	if err != nil {
		return nil, err
	}

	activeProxy := cfg.Proxy
	var lastErr error

	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		if activeProxy == "" {
			activeProxy = proxy.GetWorkingProxy(cfg.TorSOCKS, logFn)
			if activeProxy == "" {
				return nil, fmt.Errorf("tidak ada proxy berfungsi — nyalakan Tor atau set proxy manual di config.json")
			}
		}

		emailAddr, err := provider.GenerateEmail()
		if err != nil {
			if logFn != nil {
				logFn(fmt.Sprintf("Gagal generate email: %v", err))
			}
			lastErr = err
			continue
		}

		password := cfg.Password

		if logFn != nil {
			logFn(fmt.Sprintf("Mendaftarkan ke Seed4Me: %s (percobaan #%d/%d)...", emailAddr, attempt, cfg.MaxRetries))
		}

		// 1. Registrasi ke Seed4Me (via proxy untuk rotasi IP)
		if err := seed4me.Register(emailAddr, password, cfg.PromoCode, activeProxy); err != nil {
			if logFn != nil {
				logFn(fmt.Sprintf("Gagal registrasi Seed4Me: %v", err))
			}
			lastErr = err
			if attempt < cfg.MaxRetries {
				if cfg.Proxy != "" {
					if isProxyError(err) {
						cfg.Proxy = "" // proxy manual mati — jangan ulangi di percobaan berikutnya
						activeProxy = ""
					}
				} else if !strings.Contains(strings.ToLower(err.Error()), "domain email diblokir") {
					activeProxy = ""
				}
			}
			continue
		}

		// 2. Polling email konfirmasi
		if logFn != nil {
			logFn("Menunggu email verifikasi masuk...")
		}
		pollTimeout := time.Duration(cfg.PollTimeoutSeconds) * time.Second
		token, err := provider.PollToken(emailAddr, pollTimeout)
		if err != nil {
			if logFn != nil {
				logFn(fmt.Sprintf("Verifikasi email gagal: %v", err))
			}
			lastErr = err
			continue
		}

		// 3. Konfirmasi pendaftaran di Seed4Me
		if err := seed4me.Confirm(token, activeProxy); err != nil {
			if logFn != nil {
				logFn(fmt.Sprintf("Gagal konfirmasi Seed4Me: %v", err))
			}
			lastErr = err
			if attempt < cfg.MaxRetries && isProxyError(err) {
				activeProxy = ""
				if cfg.Proxy != "" {
					cfg.Proxy = ""
				}
			}
			continue
		}

		account := &model.Account{
			Email:     emailAddr,
			Password:  password,
			Status:    "Active",
			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		if err := storage.SaveAccount(*account, cfg.JSONFile, cfg.TXTFile); err != nil {
			return nil, fmt.Errorf("akun berhasil dibuat tapi GAGAL disimpan: %w", err)
		}
		return account, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("tidak ada percobaan berhasil")
	}
	return nil, fmt.Errorf("gagal membuat akun setelah %d percobaan: %v", cfg.MaxRetries, lastErr)
}

// isProxyError: error koneksi yang berarti proxy (bukan site) yang bermasalah.
func isProxyError(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	for _, marker := range []string{
		"proxyconnect", "socks connect", "connection refused", "connection reset", "i/o timeout",
	} {
		if strings.Contains(s, marker) {
			return true
		}
	}
	return false
}
