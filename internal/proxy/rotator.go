package proxy

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
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
	"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/http.txt",
	"https://api.proxyscrape.com/v2/?request=displayproxies&protocol=http&timeout=3000&country=all&ssl=yes&anonymity=elite",
}

var (
	proxyCacheMu   sync.Mutex
	proxyCache     []string
	proxyCacheTime time.Time
)

const proxyCacheTTL = 60 * time.Second

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

// fetchCandidates mengambil daftar proxy dari semua sumber secara paralel;
// hasil dicache per proses (TTL 60s) supaya tiap worker tidak fetch ulang.
func fetchCandidates() []string {
	proxyCacheMu.Lock()
	defer proxyCacheMu.Unlock()

	if proxyCache != nil && time.Since(proxyCacheTime) < proxyCacheTTL {
		return proxyCache
	}

	lists := make([][]string, len(sources))
	var wg sync.WaitGroup
	for i, src := range sources {
		wg.Add(1)
		go func(i int, src string) {
			defer wg.Done()
			lists[i] = fetchSource(src)
		}(i, src)
	}
	wg.Wait()

	var merged []string
	for _, l := range lists {
		merged = append(merged, l...)
	}
	if len(merged) == 0 {
		return merged // jangan cache hasil kosong — coba lagi di panggilan berikutnya
	}
	proxyCache, proxyCacheTime = merged, time.Now()
	return merged
}

func fetchSource(src string) []string {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(src)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var out []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.HasPrefix(line, "http") && !strings.HasPrefix(line, "socks5") {
			line = "http://" + line
		}
		out = append(out, line)
	}
	_ = scanner.Err() // daftar terpotong karena line panjang/error — abaikan, pakai yang kebaca
	return out
}

// FindFirstWorkingProxy menguji daftar proxy secara paralel (multi-worker) dan berhenti saat menemukan yang aktif pertama
func FindFirstWorkingProxy(candidates []string, workers int) string {
	if len(candidates) == 0 {
		return ""
	}
	if workers <= 0 {
		workers = 30
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
		logFn("Mencari proxy publik aktif (daftar paralel + test paralel)...")
	}

	candidates := fetchCandidates()
	if len(candidates) == 0 {
		if logFn != nil {
			logFn("[!] Daftar proxy kosong — semua sumber gagal diambil (cek jaringan). Set \"proxy\" di config.json atau nyalakan Tor.")
		}
		return ""
	}
	if logFn != nil {
		logFn(fmt.Sprintf("Kandidat proxy: %d, mulai test batch...", len(candidates)))
	}

	// Daftar publik kebanyakan mati (~5-10% hidup) — coba beberapa batch acak
	// sampai ketemu yang hidup, maksimal 4 batch (240 kandidat) atau daftar habis.
	batchSize := 60
	for batch := 0; batch < 4 && batch*batchSize < len(candidates); batch++ {
		offset := 0
		if r, err := rand.Int(rand.Reader, big.NewInt(int64(len(candidates)))); err == nil {
			offset = int(r.Int64())
		}

		size := batchSize
		if size > len(candidates) {
			size = len(candidates)
		}
		batchList := make([]string, size)
		for i := 0; i < size; i++ {
			batchList[i] = candidates[(offset+i)%len(candidates)]
		}

		if working := FindFirstWorkingProxy(batchList, 30); working != "" {
			if logFn != nil {
				logFn("[✓] Proxy terhubung: " + working)
			}
			return working
		}
		if logFn != nil {
			logFn(fmt.Sprintf("Batch #%d tidak ada yang hidup, coba batch berikutnya...", batch+1))
		}
	}
	return ""
}
