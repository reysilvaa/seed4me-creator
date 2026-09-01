package vpn

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ConnectLinuxVPN starts OpenVPN daemon in the background on Linux
func ConnectLinuxVPN(nodeCode string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("fitur ini khusus sistem operasi Linux")
	}
	if nodeCode == "" {
		nodeCode = "sg"
	}

	configFile := filepath.Join("ovpn", fmt.Sprintf("seed4me-%s.ovpn", nodeCode))
	authFile := filepath.Join("ovpn", "auth.txt")

	// Check if openvpn installed
	if _, err := exec.LookPath("openvpn"); err != nil {
		return fmt.Errorf("openvpn belum terinstall di Linux (jalankan: sudo apt install -y openvpn)")
	}

	// Kill existing session first
	_ = DisconnectLinuxVPN()

	cmd := exec.Command("sudo", "openvpn", "--config", configFile, "--auth-user-pass", authFile, "--daemon", "seed4me-vpn")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gagal menghubungkan OpenVPN Linux: %v (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// DisconnectLinuxVPN terminates the running OpenVPN daemon on Linux
func DisconnectLinuxVPN() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("fitur ini khusus sistem operasi Linux")
	}

	cmd := exec.Command("sudo", "pkill", "-f", "openvpn.*seed4me")
	_ = cmd.Run()
	return nil
}
