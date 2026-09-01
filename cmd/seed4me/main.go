package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"seed4me-creator/internal/config"
	"seed4me-creator/internal/service"
)

var printMu sync.Mutex

func logMsg(format string, a ...any) {
	printMu.Lock()
	defer printMu.Unlock()
	fmt.Printf(format+"\n", a...)
}

func main() {
	configPath := flag.String("config", "config.json", "Path file konfigurasi")
	countFlag := flag.Int("n", 0, "Jumlah akun yang ingin dibuat")
	concurrencyFlag := flag.Int("c", 0, "Jumlah worker paralel")
	proxyFlag := flag.String("proxy", "", "Proxy manual (misal: http://ip:port)")
	promoFlag := flag.String("promo", "", "Promo code")
	flag.Parse()

	cfg := config.LoadConfig(*configPath)

	// Jika dijalankan tanpa argumen CLI, tampilkan Menu Interaktif
	if len(os.Args) == 1 {
		showInteractiveMenu(&cfg)
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

	runCreator(cfg)
}

func showInteractiveMenu(cfg *config.Config) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n======================================================")
		fmt.Println("       SEED4ME AUTO CREATOR (Go Edition)")
		fmt.Println("======================================================")
		fmt.Println(" [1] Buat 1 Akun Cepat (Instan)")
		fmt.Println(" [2] Buat Banyak Akun (Kustom Jumlah & Worker)")
		fmt.Println(" [3] Lihat Daftar Akun (accounts.txt)")
		fmt.Println(" [4] Cek File OpenVPN (folder ovpn/)")
		fmt.Println(" [0] Keluar")
		fmt.Println("======================================================")
		fmt.Print("Pilih menu [0-4]: ")

		input, _ := reader.ReadString('\n')
		choice := strings.TrimSpace(input)

		switch choice {
		case "1":
			cfg.Count = 1
			cfg.Concurrency = 1
			return
		case "2":
			fmt.Print("Masukkan jumlah akun yang ingin dibuat [default 5]: ")
			nStr, _ := reader.ReadString('\n')
			nStr = strings.TrimSpace(nStr)
			if n, err := strconv.Atoi(nStr); err == nil && n > 0 {
				cfg.Count = n
			} else {
				cfg.Count = 5
			}

			fmt.Print("Masukkan jumlah worker paralel [default 2]: ")
			cStr, _ := reader.ReadString('\n')
			cStr = strings.TrimSpace(cStr)
			if c, err := strconv.Atoi(cStr); err == nil && c > 0 {
				cfg.Concurrency = c
			} else {
				cfg.Concurrency = 2
			}
			return
		case "3":
			data, err := os.ReadFile(cfg.TXTFile)
			if err != nil {
				fmt.Printf("\n[!] Belum ada akun di %s\n", cfg.TXTFile)
			} else {
				fmt.Printf("\n--- ISI %s ---\n%s\n", cfg.TXTFile, string(data))
			}
		case "4":
			files, err := os.ReadDir("ovpn")
			if err != nil || len(files) == 0 {
				fmt.Println("\n[!] Folder ovpn/ belum ada. Buat akun terlebih dahulu.")
			} else {
				fmt.Println("\n--- FILE DALAM FOLDER ovpn/ ---")
				for _, f := range files {
					fmt.Printf("  • %s\n", f.Name())
				}
				fmt.Println("\nGunakan: sudo openvpn --config ovpn/seed4me-sg.ovpn --auth-user-pass ovpn/auth.txt")
			}
		case "0":
			fmt.Println("Keluar.")
			os.Exit(0)
		default:
			fmt.Println("[!] Pilihan tidak valid, silakan coba lagi.")
		}
	}
}

func runCreator(cfg config.Config) {
	if cfg.Concurrency > cfg.Count {
		cfg.Concurrency = cfg.Count
	}

	logMsg("[i] Menjalankan Seed4Me Creator | Target: %d akun | Concurrency: %d worker", cfg.Count, cfg.Concurrency)

	jobs := make(chan int, cfg.Count)
	for i := 1; i <= cfg.Count; i++ {
		jobs <- i
	}
	close(jobs)

	var wg sync.WaitGroup
	for w := 1; w <= cfg.Concurrency; w++ {
		wg.Add(1)
		go func(wid int) {
			defer wg.Done()
			for id := range jobs {
				logFn := func(msg string) {
					logMsg("[Worker %d | #%d/%d] %s", wid, id, cfg.Count, msg)
				}
				acc, err := service.CreateAccount(cfg, logFn)
				if err != nil {
					logMsg("[✗] [Worker %d | #%d] Gagal: %v", wid, id, err)
					continue
				}
				printMu.Lock()
				fmt.Printf("\n\033[92m\033[1m[✓] [Worker %d | #%d] Akun Sukses:\n", wid, id)
				fmt.Printf("Email     : %s\n", acc.Email)
				fmt.Printf("Password  : %s\n", acc.Password)
				fmt.Printf("Status    : ACTIVE (Free Trial 7 Hari)\033[0m\n\n")
				printMu.Unlock()
			}
		}(w)
	}
	wg.Wait()

	logMsg("[✓] Selesai! Akun tersimpan di %s & %s", cfg.JSONFile, cfg.TXTFile)
}
