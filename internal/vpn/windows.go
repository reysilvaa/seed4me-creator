package vpn

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// CopyToClipboard copies a string to Windows clipboard
func CopyToClipboard(text string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("fitur clipboard ini khusus sistem operasi Windows")
	}
	cmd := exec.Command("powershell", "-NoProfile", "-Command", "Set-Clipboard", "-Value", text)
	return cmd.Run()
}

// AutoLoginDesktopApp opens Seed4.Me Desktop App, copies credentials, and focuses the login window
func AutoLoginDesktopApp(email, password string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("fitur ini khusus sistem operasi Windows")
	}

	// 1. Launch Seed4.Me app if not running
	appPath := filepath.Join("C:\\Program Files", "Seed4.Me VPN", "bin", "Seed4.Me_VPN.exe")
	_ = exec.Command(appPath).Start()

	time.Sleep(500 * time.Millisecond)

	// 2. Focus window, clear fields (^a), and paste credentials via Clipboard (^v)
	psScript := fmt.Sprintf(`
$wshell = New-Object -ComObject WScript.Shell
if ($wshell.AppActivate('Seed4.Me') -or $wshell.AppActivate('Seed4')) {
    Start-Sleep -Milliseconds 300
    Set-Clipboard -Value '%s'
    $wshell.SendKeys('^a')
    Start-Sleep -Milliseconds 100
    $wshell.SendKeys('^v')
    Start-Sleep -Milliseconds 200
    $wshell.SendKeys('{TAB}')
    Start-Sleep -Milliseconds 200
    Set-Clipboard -Value '%s'
    $wshell.SendKeys('^a')
    Start-Sleep -Milliseconds 100
    $wshell.SendKeys('^v')
    Start-Sleep -Milliseconds 200
    $wshell.SendKeys('{ENTER}')
}
`, email, password)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
	return cmd.Run()
}

// DisconnectWindowsVPN disconnects any active Windows VPN connection
func DisconnectWindowsVPN(name string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("fitur VPN native ini khusus sistem operasi Windows")
	}
	if name == "" {
		name = "Seed4Me-SG"
	}

	cmd := exec.Command("rasdial", name, "/disconnect")
	out, _ := cmd.CombinedOutput()
	outStr := strings.TrimSpace(string(out))
	if strings.Contains(strings.ToLower(outStr), "no active") {
		return fmt.Errorf("tidak ada koneksi VPN yang sedang aktif")
	}
	return nil
}
