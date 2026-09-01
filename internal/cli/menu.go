package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"seed4me-creator/internal/config"
)

// ShowInteractiveMenu menampilkan menu interaktif berbasis terminal (TUI)
func ShowInteractiveMenu(cfg *config.Config) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n======================================================")
		fmt.Println("       SEED4ME AUTO CREATOR (Go Edition)")
		fmt.Println("======================================================")
		fmt.Printf(" Provider Email Aktif: \033[96m%s\033[0m\n", cfg.EmailService)
		fmt.Println("------------------------------------------------------")
		fmt.Println(" [1] Buat 1 Akun Cepat (Instan)")
		fmt.Println(" [2] Buat Banyak Akun (Kustom Jumlah & Worker)")
		fmt.Println(" [3] Lihat Daftar Akun (accounts.txt)")
		fmt.Println(" [4] Ganti Provider Email (TempMail.ing / TempMail.lol / CatchMail)")
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
			fmt.Println("\nPilih Provider Email:")
			fmt.Println("  [1] TempMail.ing (Cloudflare Worker API, Free & Open)")
			fmt.Println("  [2] TempMail.lol (v3 API, Fast Long-Polling)")
			fmt.Println("  [3] CatchMail (catchmail.io & Custom Domain)")
			fmt.Print("Pilihan [1/2/3]: ")
			pInput, _ := reader.ReadString('\n')
			pChoice := strings.TrimSpace(pInput)
			switch pChoice {
			case "2":
				cfg.EmailService = config.EmailServiceTempMailLol
				fmt.Println("[✓] Provider diubah ke TempMail.lol.")
			case "3":
				cfg.EmailService = config.EmailServiceCatchMail
				fmt.Println("[✓] Provider diubah ke CatchMail.")
			default:
				cfg.EmailService = config.EmailServiceTempMailIng
				fmt.Println("[✓] Provider diubah ke TempMail.ing.")
			}
		case "0":
			fmt.Println("Keluar.")
			os.Exit(0)
		default:
			fmt.Println("[!] Pilihan tidak valid, silakan coba lagi.")
		}
	}
}
