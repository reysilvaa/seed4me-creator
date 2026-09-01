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

// Append-only JSONL: O(1) per akun, tidak perlu read-modify-write seluruh file.
// File lama (format array JSON atau apa pun) tidak akan dirusak — hanya ditambahi.
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

	// 1. Append ke JSON (satu objek per baris)
	line, err := json.Marshal(acc)
	if err != nil {
		return fmt.Errorf("gagal encode JSON: %w", err)
	}
	if err := appendFile(jsonPath, append(line, '\n')); err != nil {
		return err
	}

	// 2. Append ke TXT
	txt := fmt.Sprintf("Email: %s | Password: %s | Status: %s | Date: %s\n",
		acc.Email, acc.Password, acc.Status, acc.CreatedAt)
	if err := appendFile(txtPath, []byte(txt)); err != nil {
		return err
	}

	return nil
}

func appendFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("gagal membuka %s: %w", path, err)
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("gagal menulis %s: %w", path, err)
	}
	return nil
}
