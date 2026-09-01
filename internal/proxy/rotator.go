package proxy

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var ProxySources = []string{
	"https://api.proxyscrape.com/v2/?request=displayproxies&protocol=http&timeout=3000&country=all&ssl=yes&anonymity=elite",
	"https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/http.txt",
	"https://raw.githubusercontent.com/TheSpeedX/SOCKS-List/master/http.txt",
}

type Rotator struct {
	customProxies []string
	fetchedList   []string
	mu            sync.Mutex
}

func NewRotator(customList []string) *Rotator {
	return &Rotator{
		customProxies: customList,
	}
}

func (r *Rotator) FetchOnlineProxies() ([]string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	var results []string
	seen := make(map[string]bool)

	for _, src := range ProxySources {
		resp, err := client.Get(src)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") && !seen[line] {
				seen[line] = true
				if !strings.HasPrefix(line, "http://") && !strings.HasPrefix(line, "https://") && !strings.HasPrefix(line, "socks5://") {
					line = "http://" + line
				}
				results = append(results, line)
			}
		}
		_ = resp.Body.Close()
		if len(results) >= 100 {
			break
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("gagal mengambil proxy publik")
	}

	r.mu.Lock()
	r.fetchedList = results
	r.mu.Unlock()

	return results, nil
}

func CheckProxyWorking(proxyURL string, timeout time.Duration) bool {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return false
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(u),
			DialContext: (&net.Dialer{
				Timeout: timeout,
			}).DialContext,
			TLSHandshakeTimeout: timeout,
		},
		Timeout: timeout,
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://seed4.me", nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	return resp.StatusCode == http.StatusOK
}

func (r *Rotator) GetWorkingProxy(logFn func(string)) (string, error) {
	// Check Tor SOCKS5
	if CheckTorRunning("127.0.0.1:9050") {
		if logFn != nil {
			logFn("Terdeteksi Tor aktif lokal di 127.0.0.1:9050. Menggunakan Tor...")
		}
		return "socks5://127.0.0.1:9050", nil
	}

	candidateList := r.customProxies
	if len(candidateList) == 0 {
		if len(r.fetchedList) == 0 {
			if logFn != nil {
				logFn("Mengambil daftar proxy aktif untuk rotasi IP...")
			}
			var err error
			candidateList, err = r.FetchOnlineProxies()
			if err != nil {
				return "", err
			}
		} else {
			candidateList = r.fetchedList
		}
	}

	if logFn != nil {
		logFn(fmt.Sprintf("Mencari proxy aktif dari %d kandidat...", len(candidateList)))
	}

	resultChan := make(chan string, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workerCount := 25
	proxyChan := make(chan string, len(candidateList))
	for _, p := range candidateList {
		proxyChan <- p
	}
	close(proxyChan)

	var wg sync.WaitGroup
	var once sync.Once

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case p, ok := <-proxyChan:
					if !ok {
						return
					}
					if CheckProxyWorking(p, 4*time.Second) {
						once.Do(func() {
							resultChan <- p
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
	case found, ok := <-resultChan:
		if ok && found != "" {
			if logFn != nil {
				logFn(fmt.Sprintf("[✓] Proxy terhubung: %s", found))
			}
			return found, nil
		}
	case <-time.After(15 * time.Second):
	}

	return "", fmt.Errorf("tidak ditemukan proxy aktif")
}

func CheckTorRunning(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
