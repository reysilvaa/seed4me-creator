package proxy

import (
	"bufio"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var sources = []string{
	"https://api.proxyscrape.com/v2/?request=displayproxies&protocol=http&timeout=3000&country=all&ssl=yes&anonymity=elite",
	"https://raw.githubusercontent.com/monosans/proxy-list/main/proxies/http.txt",
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
		Transport: &http.Transport{Proxy: http.ProxyURL(u)},
		Timeout:   3 * time.Second,
	}
	resp, err := client.Get("https://seed4.me")
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func GetWorkingProxy(torAddr string, logFn func(string)) string {
	if torAddr != "" && IsTorRunning(torAddr) {
		if logFn != nil {
			logFn("Menggunakan Tor lokal di " + torAddr)
		}
		return "socks5://" + torAddr
	}

	if logFn != nil {
		logFn("Mencari proxy publik aktif...")
	}

	client := &http.Client{Timeout: 5 * time.Second}
	for _, src := range sources {
		resp, err := client.Get(src)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(resp.Body)
		checked := 0
		for scanner.Scan() && checked < 30 {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if !strings.HasPrefix(line, "http") && !strings.HasPrefix(line, "socks5") {
				line = "http://" + line
			}
			checked++
			if TestProxy(line) {
				_ = resp.Body.Close()
				if logFn != nil {
					logFn("[✓] Proxy terhubung: " + line)
				}
				return line
			}
		}
		_ = scanner.Err()
		_ = resp.Body.Close()
	}

	return ""
}
