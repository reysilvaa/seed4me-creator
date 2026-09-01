package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"seed4me-linux/internal/model"
)

const (
	JSONFile = "accounts.json"
	TXTFile  = "accounts.txt"
)

var mu sync.Mutex

func SaveAccount(acc model.Account) error {
	mu.Lock()
	defer mu.Unlock()

	if acc.CreatedAt == "" {
		acc.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
	}

	// 1. Simpan ke JSON
	accounts, _ := LoadAccounts()
	accounts = append(accounts, acc)
	jsonData, err := json.MarshalIndent(accounts, "", "  ")
	if err == nil {
		_ = os.WriteFile(JSONFile, jsonData, 0644)
	}

	// 2. Append ke TXT
	f, err := os.OpenFile(TXTFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		line := fmt.Sprintf("Email: %s | Password: %s | Status: %s | PSK: %s | Date: %s\n",
			acc.Email, acc.Password, acc.Status, acc.PSK, acc.CreatedAt)
		_, _ = f.WriteString(line)
	}

	return nil
}

func LoadAccounts() ([]model.Account, error) {
	data, err := os.ReadFile(JSONFile)
	if err != nil {
		return []model.Account{}, nil
	}
	var accounts []model.Account
	if err := json.Unmarshal(data, &accounts); err != nil {
		return []model.Account{}, nil
	}
	return accounts, nil
}
