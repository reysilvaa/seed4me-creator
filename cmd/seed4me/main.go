package main

import (
	"flag"
	"os"

	"seed4me-creator/internal/cli"
	"seed4me-creator/internal/config"
	"seed4me-creator/internal/service"
)

func main() {
	configPath := flag.String("config", "config.json", "Path file konfigurasi")
	countFlag := flag.Int("n", 0, "Jumlah akun yang ingin dibuat")
	concurrencyFlag := flag.Int("c", 0, "Jumlah worker paralel")
	proxyFlag := flag.String("proxy", "", "Proxy manual (misal: http://ip:port)")
	promoFlag := flag.String("promo", "", "Promo code")
	serviceFlag := flag.String("service", "", "Provider email: tempmailing, tempmaillol, atau catchmail")
	flag.Parse()

	cfg := config.LoadConfig(*configPath)

	if *serviceFlag != "" {
		cfg.EmailService = config.EmailService(*serviceFlag)
	}

	// Jika dijalankan tanpa argumen CLI, tampilkan Menu Interaktif
	if len(os.Args) == 1 {
		cli.ShowInteractiveMenu(&cfg)
	}

	if *countFlag > 0 {
		cfg.Count = *countFlag
	}
	if *concurrencyFlag > 0 {
		cfg.Concurrency = *concurrencyFlag
	}
	if *proxyFlag != "" {
		cfg.Proxy = *proxyFlag
	}
	if *promoFlag != "" {
		cfg.PromoCode = *promoFlag
	}

	service.RunBatch(cfg)
}
