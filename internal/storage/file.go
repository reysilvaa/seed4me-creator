package storage

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"seed4me-creator/internal/model"
)

var mu sync.Mutex

// LoadAccounts reads canonical JSON arrays and accepts legacy JSONL during migration.
func LoadAccounts(path string) ([]model.Account, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("gagal membaca %s: %w", path, err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil
	}

	var accounts []model.Account
	if err := json.Unmarshal(data, &accounts); err == nil {
		return accounts, nil
	}

	var account model.Account
	if err := json.Unmarshal(data, &account); err == nil && strings.TrimSpace(account.Email) != "" {
		return []model.Account{account}, nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for line := 1; scanner.Scan(); line++ {
		lineData := bytes.TrimSpace(scanner.Bytes())
		if len(lineData) == 0 {
			continue
		}
		var account model.Account
		if err := json.Unmarshal(lineData, &account); err != nil {
			return nil, fmt.Errorf("%s bukan JSON valid pada baris %d: %w", path, line, err)
		}
		if strings.TrimSpace(account.Email) == "" {
			return nil, fmt.Errorf("%s memiliki akun tanpa email pada baris %d", path, line)
		}
		accounts = append(accounts, account)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("gagal membaca %s: %w", path, err)
	}
	return accounts, nil
}

// SaveAccount adds an account and rewrites the canonical JSON/TXT projections safely.
func SaveAccount(acc model.Account, jsonPath, txtPath string) error {
	mu.Lock()
	defer mu.Unlock()

	if strings.TrimSpace(acc.Email) == "" || acc.Password == "" {
		return fmt.Errorf("email dan password akun wajib diisi")
	}
	if jsonPath == "" || txtPath == "" {
		return fmt.Errorf("json_file dan txt_file wajib diisi")
	}
	if filepath.Clean(jsonPath) == filepath.Clean(txtPath) {
		return fmt.Errorf("json_file dan txt_file harus berbeda")
	}
	if acc.Status == "" {
		acc.Status = "Active"
	}
	if acc.CreatedAt == "" {
		acc.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
	}

	for _, path := range []string{jsonPath, txtPath} {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return fmt.Errorf("gagal buat direktori untuk %s: %w", path, err)
		}
	}

	accounts, err := LoadAccounts(jsonPath)
	if err != nil {
		return err
	}
	for _, existing := range accounts {
		if strings.EqualFold(existing.Email, acc.Email) {
			return fmt.Errorf("akun dengan email %s sudah tersimpan", acc.Email)
		}
	}
	accounts = append(accounts, acc)

	jsonData, err := json.MarshalIndent(accounts, "", "  ")
	if err != nil {
		return fmt.Errorf("gagal encode %s: %w", jsonPath, err)
	}
	jsonData = append(jsonData, '\n')
	if err := writeAtomic(jsonPath, jsonData, 0600); err != nil {
		return fmt.Errorf("gagal menulis %s: %w", jsonPath, err)
	}

	var txt strings.Builder
	for _, account := range accounts {
		fmt.Fprintf(&txt, "Email: %s | Password: %s | Status: %s | Date: %s\n",
			account.Email, account.Password, account.Status, account.CreatedAt)
	}
	if err := writeAtomic(txtPath, []byte(txt.String()), 0600); err != nil {
		return fmt.Errorf("gagal menulis %s: %w", txtPath, err)
	}
	return nil
}

func writeAtomic(path string, data []byte, mode os.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	backupPath := ""
	if _, err := os.Stat(path); err == nil {
		backup, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".bak-*")
		if err != nil {
			return err
		}
		backupPath = backup.Name()
		if err := backup.Close(); err != nil {
			_ = os.Remove(backupPath)
			return err
		}
		_ = os.Remove(backupPath)
		if err := os.Rename(path, backupPath); err != nil {
			return err
		}
		defer os.Remove(backupPath)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		if backupPath != "" {
			_ = os.Rename(backupPath, path)
		}
		return err
	}
	if backupPath != "" {
		_ = os.Remove(backupPath)
	}
	return os.Chmod(path, mode)
}
