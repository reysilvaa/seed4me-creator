package service

import (
	"fmt"
	"time"

	"seed4me-linux/internal/client"
	"seed4me-linux/internal/model"
	"seed4me-linux/internal/proxy"
	"seed4me-linux/internal/storage"
)

type AccountCreator struct {
	Rotator     *proxy.Rotator
	ManualProxy string
	PromoCode   string
	LogInfo     func(string)
	LogWarn     func(string)
	LogError    func(string)
	LogSuccess  func(string)
}

func NewAccountCreator(rotator *proxy.Rotator, manualProxy, promoCode string, logInfo, logWarn, logError, logSuccess func(string)) *AccountCreator {
	return &AccountCreator{
		Rotator:     rotator,
		ManualProxy: manualProxy,
		PromoCode:   promoCode,
		LogInfo:     logInfo,
		LogWarn:     logWarn,
		LogError:    logError,
		LogSuccess:  logSuccess,
	}
}

func (s *AccountCreator) CreateOne() (*model.Account, error) {
	activeProxy := s.ManualProxy

	// 1. Ambil proxy aktif jika tidak ada manual proxy
	if activeProxy == "" && s.Rotator != nil {
		if p, err := s.Rotator.GetWorkingProxy(s.LogInfo); err == nil {
			activeProxy = p
		}
	}

	// 2. Coba hingga 3 percobaan dengan rotasi otomatis jika rate limited
	for attempt := 1; attempt <= 3; attempt++ {
		s4m := client.NewSeed4MeClient(activeProxy)
		cm := client.NewCatchMailClient("")

		email := client.GenerateEmail()
		password := client.GeneratePassword()

		if s.LogInfo != nil {
			s.LogInfo(fmt.Sprintf("Mendaftarkan: %s", email))
		}

		err := s4m.Register(email, password, s.PromoCode)
		if err != nil {
			if s.LogWarn != nil {
				s.LogWarn(fmt.Sprintf("Gagal (percobaan #%d): %v", attempt, err))
			}
			if s.Rotator != nil && attempt < 3 {
				if s.LogInfo != nil {
					s.LogInfo("Rotasi proxy baru...")
				}
				if nextProxy, pErr := s.Rotator.GetWorkingProxy(s.LogInfo); pErr == nil {
					activeProxy = nextProxy
					continue
				}
			}
			return nil, fmt.Errorf("registrasi gagal: %w", err)
		}

		if s.LogInfo != nil {
			s.LogInfo("Menunggu email verifikasi di CatchMail.io...")
		}

		code, err := cm.PollVerificationToken(email, 35*time.Second, s.LogInfo)
		if err != nil {
			return nil, fmt.Errorf("verifikasi email gagal: %w", err)
		}

		if s.LogInfo != nil {
			s.LogInfo(fmt.Sprintf("Aktivasi token: %s", code))
		}

		if err := s4m.ConfirmEmail(code); err != nil {
			return nil, fmt.Errorf("aktivasi token gagal: %w", err)
		}

		acc := model.Account{
			Email:     email,
			Password:  password,
			Status:    "Active",
			PSK:       "seed4me",
			Notes:     "7 Days Free Trial",
			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		_ = storage.SaveAccount(acc)

		if s.LogSuccess != nil {
			s.LogSuccess("Akun Seed4Me Berhasil Dibuat!")
		}

		return &acc, nil
	}

	return nil, fmt.Errorf("semua percobaan habis")
}
