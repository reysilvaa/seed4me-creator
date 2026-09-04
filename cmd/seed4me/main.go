package main

import (
	"flag"
	"fmt"
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

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	cfg.Normalize()

	if *serviceFlag != "" {
		cfg.EmailService = config.EmailService(*serviceFlag)
	}

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
	cfg.Normalize()

	if err := cfg.Validate(); err != nil {
		fmt.Fprintln(os.Stderr, "Konfigurasi tidak valid:", err)
		os.Exit(2)
	}
	if err := service.RunBatch(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
