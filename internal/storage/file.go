package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"seed4me-creator/internal/model"
)

var mu sync.Mutex

// ponytail: JSON read-modify-write per akun, O(n²) kalau akun ribuan;
// ganti ke append-only JSONL kalau throughput mulai penting.
func SaveAccount(acc model.Account, jsonPath, txtPath string) error {
	mu.Lock()
	defer mu.Unlock()

	if acc.CreatedAt == "" {
		acc.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
	}

	for _, p := range []string{jsonPath, txtPath} {
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			return fmt.Errorf("gagal buat direktori untuk %s: %w", p, err)
		}
	}

	// 1. Simpan ke JSON
	var accounts []model.Account
	if data, err := os.ReadFile(jsonPath); err == nil {
		if err := json.Unmarshal(data, &accounts); err != nil {
			return fmt.Errorf("%s korup — akun TIDAK ditulis agar tidak menimpa data lama: %w", jsonPath, err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("gagal membaca %s: %w", jsonPath, err)
	}

	accounts = append(accounts, acc)
	jsonData, err := json.MarshalIndent(accounts, "", "  ")
	if err != nil {
		return fmt.Errorf("gagal encode JSON: %w", err)
	}
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("gagal menulis %s: %w", jsonPath, err)
	}

	// 2. Append ke TXT
	f, err := os.OpenFile(txtPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("gagal membuka %s: %w", txtPath, err)
	}
	defer f.Close()

	line := fmt.Sprintf("Email: %s | Password: %s | Status: %s | Date: %s\n",
		acc.Email, acc.Password, acc.Status, acc.CreatedAt)
	if _, err := f.WriteString(line); err != nil {
		return fmt.Errorf("gagal menulis %s: %w", txtPath, err)
	}

	return nil
}