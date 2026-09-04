package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"seed4me-creator/internal/model"
)

func TestSaveAccount(t *testing.T) {
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "accounts.json")
	txtPath := filepath.Join(tmpDir, "accounts.txt")

	acc := model.Account{
		Email:     "user.test@catchmail.io",
		Password:  "Pass123!",
		Status:    "Active",
		CreatedAt: "2026-09-02 00:00:00",
	}

	if err := SaveAccount(acc, jsonPath, txtPath); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	// JSON must stay a valid array.
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("failed to read json: %v", err)
	}
	var parsed []model.Account
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("accounts.json bukan array JSON valid: %v\nisi: %s", err, string(jsonData))
	}
	if len(parsed) != 1 || parsed[0].Email != acc.Email {
		t.Fatalf("unexpected accounts content: %+v", parsed)
	}

	txtData, err := os.ReadFile(txtPath)
	if err != nil {
		t.Fatalf("failed to read txt: %v", err)
	}
	expected := "Email: user.test@catchmail.io | Password: Pass123! | Status: Active | Date: 2026-09-02 00:00:00"
	if !strings.Contains(string(txtData), expected) {
		t.Fatalf("expected txt line '%s', got '%s'", expected, string(txtData))
	}
}

func TestSaveAccountMigratesLegacyJSONL(t *testing.T) {
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "accounts.json")

	legacy := "{\"email\":\"old@x.io\",\"password\":\"oldpass1\",\"status\":\"Active\",\"created_at\":\"2026-01-01 00:00:00\"}\n"
	if err := os.WriteFile(jsonPath, []byte(legacy), 0600); err != nil {
		t.Fatalf("write legacy: %v", err)
	}

	acc := model.Account{Email: "new@x.io", Password: "newpass1", Status: "Active"}
	if err := SaveAccount(acc, jsonPath, filepath.Join(tmpDir, "accounts.txt")); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	data, _ := os.ReadFile(jsonPath)
	var parsed []model.Account
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("expected valid JSON array after migration, got %q: %v", string(data), err)
	}
	if len(parsed) != 2 || parsed[0].Email != "old@x.io" || parsed[1].Email != "new@x.io" {
		t.Fatalf("unexpected migrated accounts: %+v", parsed)
	}
}

func TestSaveAccountRefusesCorruptFile(t *testing.T) {
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "accounts.json")
	if err := os.WriteFile(jsonPath, []byte("{bukan json\n"), 0600); err != nil {
		t.Fatalf("write corrupt: %v", err)
	}

	acc := model.Account{Email: "x@y.z", Password: "password1", Status: "Active"}
	if err := SaveAccount(acc, jsonPath, filepath.Join(tmpDir, "accounts.txt")); err == nil {
		t.Fatal("expected error on corrupt file, got nil")
	}
}
