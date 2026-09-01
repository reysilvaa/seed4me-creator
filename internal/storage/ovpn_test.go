package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveOVPNProfiles(t *testing.T) {
	tmpDir := filepath.Join(t.TempDir(), "ovpn")
	err := SaveOVPNProfiles("testuser@catchmail.io", "12345678", tmpDir)
	if err != nil {
		t.Fatalf("unexpected error saving ovpn profiles: %v", err)
	}

	// Check auth.txt
	authPath := filepath.Join(tmpDir, "auth.txt")
	authBytes, err := os.ReadFile(authPath)
	if err != nil {
		t.Fatalf("auth.txt not created: %v", err)
	}
	if !strings.Contains(string(authBytes), "testuser@catchmail.io") || !strings.Contains(string(authBytes), "12345678") {
		t.Fatalf("invalid auth.txt content: %s", string(authBytes))
	}

	// Check sg ovpn
	sgPath := filepath.Join(tmpDir, "seed4me-sg.ovpn")
	sgBytes, err := os.ReadFile(sgPath)
	if err != nil {
		t.Fatalf("seed4me-sg.ovpn not created: %v", err)
	}
	if !strings.Contains(string(sgBytes), "sg.seed4.me") || !strings.Contains(string(sgBytes), "data-ciphers") {
		t.Fatalf("invalid seed4me-sg.ovpn content: %s", string(sgBytes))
	}

	// Check connect.sh
	shPath := filepath.Join(tmpDir, "connect.sh")
	if _, err := os.Stat(shPath); err != nil {
		t.Fatalf("connect.sh not created: %v", err)
	}
}
