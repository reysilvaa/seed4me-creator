package service

import (
	"fmt"
	"strings"
	"time"

	"seed4me-creator/internal/client"
	"seed4me-creator/internal/config"
	"seed4me-creator/internal/model"
	"seed4me-creator/internal/proxy"
	"seed4me-creator/internal/storage"
)

func CreateAccount(cfg config.Config, logFn func(string)) (*model.Account, error) {
	activeProxy := cfg.Proxy
	if activeProxy == "" {
		activeProxy = proxy.GetWorkingProxy(cfg.TorSOCKS, logFn)
		if activeProxy == "" {
			return nil, fmt.Errorf("tidak ada proxy berfungsi (Tor mati & daftar proxy publik kosong) — stop, jangan lanjut lewat IP asli")
		}
	}

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		email := client.GenerateEmail(cfg.EmailDomain)
		password := client.GeneratePassword()

		if logFn != nil {
			logFn(fmt.Sprintf("Mendaftarkan ke Seed4Me: %s (percobaan #%d)...", email, attempt))
		}

		// 1. Registrasi ke Seed4Me (via proxy untuk rotasi IP)
		if err := client.RegisterSeed4Me(email, password, cfg.PromoCode, activeProxy); err != nil {
			if logFn != nil {
				logFn(fmt.Sprintf("Gagal registrasi Seed4Me: %v", err))
			}
			lastErr = err
			if strings.Contains(err.Error(), "diblokir") {
				if cfg.EmailDomain == "" || cfg.EmailDomain == "catchmail.io" {
					return nil, fmt.Errorf("domain '%s' diblokir Seed4Me — gunakan custom domain di config.json (arahkan MX ke smtp.catchmail.io)", email)
				}
			}
			if attempt < 3 && cfg.Proxy == "" {
				activeProxy = proxy.GetWorkingProxy(cfg.TorSOCKS, logFn)
				if activeProxy == "" {
					return nil, fmt.Errorf("proxy berhenti berfungsi: %v", lastErr)
				}
			}
			continue
		}

		// 2. Polling email di CatchMail (koneksi direct)
		if logFn != nil {
			logFn("Menunggu email verifikasi masuk di CatchMail...")
		}

		token, err := client.PollToken(email, 45*time.Second, logFn)
		if err != nil {
			if logFn != nil {
				logFn(fmt.Sprintf("Verifikasi email gagal: %v", err))
			}
			lastErr = err
			continue
		}

		if logFn != nil {
			logFn(fmt.Sprintf("Aktivasi token konfirmasi: %s", token))
		}

		// 3. Konfirmasi email di Seed4Me (via proxy)
		if err := client.ConfirmSeed4Me(token, activeProxy); err != nil {
			if logFn != nil {
				logFn(fmt.Sprintf("Aktivasi token gagal: %v", err))
			}
			lastErr = err
			continue
		}

		acc := model.Account{
			Email:     email,
			Password:  password,
			Status:    "Active",
			PSK:       "seed4me",
			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		if err := storage.SaveAccount(acc, cfg.JSONFile, cfg.TXTFile); err != nil {
			return nil, fmt.Errorf("akun berhasil dibuat tapi GAGAL disimpan: %w", err)
		}

		_ = storage.SaveOVPNProfiles(acc.Email, acc.Password, "ovpn")
		if logFn != nil {
			logFn("[✓] Profil OpenVPN (Linux/Debian & Windows) dibuat di folder ovpn/")
		}

		return &acc, nil
	}

	return nil, fmt.Errorf("gagal membuat akun setelah 3 percobaan: %v", lastErr)
}