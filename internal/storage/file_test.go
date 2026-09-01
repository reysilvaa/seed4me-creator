package storage

import (
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

	txtData, err := os.ReadFile(txtPath)
	if err != nil {
		t.Fatalf("failed to read txt: %v", err)
	}
	expected := "Email: user.test@catchmail.io | Password: Pass123! | Status: Active | Date: 2026-09-02 00:00:00"
	if !strings.Contains(string(txtData), expected) {
		t.Fatalf("expected txt line '%s', got '%s'", expected, string(txtData))
	}
}

func TestSaveAccountRefusesCorruptJSON(t *testing.T) {
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "accounts.json")
	_ = os.WriteFile(jsonPath, []byte("{corrupt"), 0644)

	acc := model.Account{Email: "x@y.z", Password: "p", Status: "Active"}
	if err := SaveAccount(acc, jsonPath, filepath.Join(tmpDir, "accounts.txt")); err == nil {
		t.Fatal("expected error on corrupt json, got nil")
	}
	data, _ := os.ReadFile(jsonPath)
	if string(data) != "{corrupt" {
		t.Fatalf("corrupt file must not be overwritten, got %q", string(data))
	}
}
