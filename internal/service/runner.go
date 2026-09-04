package service

import (
	"fmt"
	"sync"

	"seed4me-creator/internal/config"
)

var printMu sync.Mutex

func logMsg(format string, a ...any) {
	printMu.Lock()
	defer printMu.Unlock()
	fmt.Printf(format+"\n", a...)
}

// RunBatch menjalankan pembuatan akun secara paralel sesuai worker pool.
func RunBatch(cfg config.Config) error {
	if cfg.Count < 1 || cfg.Concurrency < 1 {
		return fmt.Errorf("count dan concurrency harus lebih besar dari 0")
	}
	if cfg.Concurrency > cfg.Count {
		cfg.Concurrency = cfg.Count
	}

	logMsg("[i] Menjalankan Seed4Me Creator | Target: %d akun | Concurrency: %d worker", cfg.Count, cfg.Concurrency)

	jobs := make(chan int)
	go func() {
		defer close(jobs)
		for i := 1; i <= cfg.Count; i++ {
			jobs <- i
		}
	}()

	var wg sync.WaitGroup
	succeeded := 0
	for w := 1; w <= cfg.Concurrency; w++ {
		wg.Add(1)
		go func(wid int) {
			defer wg.Done()
			for id := range jobs {
				logFn := func(msg string) {
					logMsg("[Worker %d | #%d/%d] %s", wid, id, cfg.Count, msg)
				}
				acc, err := CreateAccount(cfg, logFn)
				if err != nil {
					logMsg("[✗] [Worker %d | #%d] Gagal: %v", wid, id, err)
					continue
				}
				printMu.Lock()
				succeeded++
				fmt.Printf("\n\033[92m\033[1m[✓] [Worker %d | #%d] Akun Sukses:\n", wid, id)
				fmt.Printf("Email     : %s\n", acc.Email)
				fmt.Printf("Password  : %s\n", acc.Password)
				fmt.Printf("Status    : ACTIVE (Free Trial 7 Hari)\033[0m\n\n")
				printMu.Unlock()
			}
		}(w)
	}
	wg.Wait()

	if succeeded == 0 {
		logMsg("[✗] Selesai. 0/%d akun sukses — tidak ada yang tersimpan.", cfg.Count)
		return fmt.Errorf("semua %d akun gagal dibuat", cfg.Count)
	}
	logMsg("[✓] Selesai! %d/%d akun sukses, tersimpan di %s & %s", succeeded, cfg.Count, cfg.JSONFile, cfg.TXTFile)
	return nil
}
