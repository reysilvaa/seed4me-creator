package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"seed4me-creator/internal/proxy"
	"seed4me-creator/internal/service"
)

const (
	colorGreen = "\033[92m"
	colorCyan  = "\033[96m"
	colorBold  = "\033[1m"
	colorReset = "\033[0m"
)

func logInfo(msg string)    { fmt.Printf("[i] %s\n", msg) }
func logSuccess(msg string) { fmt.Printf("%s[✓] %s%s\n", colorGreen, msg, colorReset) }
func logWarn(msg string)    { fmt.Printf("[!] %s\n", msg) }
func logError(msg string)   { fmt.Printf("[✗] %s\n", msg) }

func main() {
	countFlag := flag.Int("n", 1, "Jumlah akun yang ingin dibuat")
	proxyFlag := flag.String("proxy", "", "Proxy manual (opsional, misal: http://127.0.0.1:8080)")
	proxyFile := flag.String("proxy-file", "", "File daftar proxy custom (opsional)")
	promoFlag := flag.String("promo", "", "Promo code (opsional)")
	flag.Parse()

	var customProxies []string
	if *proxyFile != "" {
		if data, err := os.ReadFile(*proxyFile); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "#") {
					customProxies = append(customProxies, line)
				}
			}
		}
	}

	rotator := proxy.NewRotator(customProxies)
	creator := service.NewAccountCreator(
		rotator,
		*proxyFlag,
		*promoFlag,
		logInfo,
		logWarn,
		logError,
		logSuccess,
	)

	for i := 1; i <= *countFlag; i++ {
		if *countFlag > 1 {
			fmt.Printf("\n=== [%d/%d] Membuat Akun Seed4Me ===\n", i, *countFlag)
		}

		acc, err := creator.CreateOne()
		if err != nil {
			logError(err.Error())
			continue
		}

		fmt.Printf("%s%s", colorGreen, colorBold)
		fmt.Printf("Email     : %s\n", acc.Email)
		fmt.Printf("Password  : %s\n", acc.Password)
		fmt.Printf("IPSec PSK : %s\n", acc.PSK)
		fmt.Printf("Status    : ACTIVE (Free Trial 7 Hari)\n")
		fmt.Printf("%s", colorReset)
	}
}
