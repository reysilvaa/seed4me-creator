package vpn

import (
	"runtime"
	"testing"
)

func TestNonWindowsGuard(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Testing on Windows")
	}
	err := CopyToClipboard("test")
	if err == nil {
		t.Fatal("expected error on non-windows OS")
	}
}
