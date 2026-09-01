package proxy

import (
	"bufio"
	"context"
	"crypto/rand"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var sources = []string{
	"https://raw.githubusercontent.com/proxifly/free-proxy-list/main/proxies/all/data.txt",
	"https://raw.githubusercontent.com/proxifly/free-proxy-list/main/proxies/protocols/http/data.txt",
	"https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/http.txt",
	"https://api.proxyscrape.com/v2/?request=displayproxies&protocol=http&timeout=3000&country=all&ssl=yes&anonymity=elite",
}

func IsTorRunning(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 400*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func TestProxy(proxyURL string) bool {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return false
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(u),
			DialContext: (&net.Dialer{
				Timeout: 2 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 2 * time.Second,
		},
		Timeout: 3 * time.Second,
	}
	resp, err := client.Get("https://seed4.me")
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// FindFirstWorkingProxy menguji daftar proxy secara paralel (multi-worker) dan berhenti saat menemukan yang aktif pertama
func FindFirstWorkingProxy(candidates []string, workers int) string {
	if len(candidates) == 0 {
		return ""
	}
	if workers <= 0 {
		workers = 20
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	jobs := make(chan string, len(candidates))
	for _, c := range candidates {
		jobs <- c
	}
	close(jobs)

	resultChan := make(chan string, 1)
	var once sync.Once
	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case candidate, ok := <-jobs:
					if !ok {
						return
					}
					if TestProxy(candidate) {
						once.Do(func() {
							resultChan <- candidate
							cancel()
						})
						return
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	select {
	case res := <-resultChan:
		return res
	case <-ctx.Done():
		return ""
	}
}

func GetWorkingProxy(torAddr string, logFn func(string)) string {
	if torAddr != "" && IsTorRunning(torAddr) {
		if logFn != nil {
			logFn("Menggunakan Tor lokal di " + torAddr)
		}
		return "socks5://" + torAddr
	}

	if logFn != nil {
		logFn("Mencari proxy publik aktif (multi-worker parallel check)...")
	}

	client := &http.Client{Timeout: 5 * time.Second}
	for _, src := range sources {
		resp, err := client.Get(src)
		if err != nil {
			continue
		}
		var candidates []string
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if !strings.HasPrefix(line, "http") && !strings.HasPrefix(line, "socks5") {
				line = "http://" + line
			}
			candidates = append(candidates, line)
		}
		_ = scanner.Err()
		_ = resp.Body.Close()

		if len(candidates) == 0 {
			continue
		}

		offset := 0
		if r, err := rand.Int(rand.Reader, big.NewInt(int64(len(candidates)))); err == nil {
			offset = int(r.Int64())
		}

		// Ambil batch 60 proxy acak untuk diuji paralel
		batchSize := 60
		if batchSize > len(candidates) {
			batchSize = len(candidates)
		}
		batch := make([]string, batchSize)
		for i := 0; i < batchSize; i++ {
			batch[i] = candidates[(offset+i)%len(candidates)]
		}

		if working := FindFirstWorkingProxy(batch, 20); working != "" {
			if logFn != nil {
				logFn("[✓] Proxy terhubung: " + working)
			}
			return working
		}
	}

	return ""
}
