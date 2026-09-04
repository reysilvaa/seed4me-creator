package seed4me

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var BaseURL = "https://seed4.me"

type browserProfile struct {
	UserAgent string
	SecChUa   string
}

var browserProfiles = []browserProfile{
	{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36",
		SecChUa:   `"Chromium";v="125", "Google Chrome";v="125", "Not-A.Brand";v="99"`,
	},
	{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
		SecChUa:   `"Chromium";v="124", "Google Chrome";v="124", "Not-A.Brand";v="99"`,
	},
	{
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36 Edg/125.0.0.0",
		SecChUa:   `"Chromium";v="125", "Microsoft Edge";v="125", "Not-A.Brand";v="99"`,
	},
}

var (
	successRegex = regexp.MustCompile(`(?i)(confirmation code has been sent|confirm email address|please check your email|verification code)`)
	alertRegex   = regexp.MustCompile(`(?i)<div[^>]*class="[^"]*alert[^"]*"[^>]*>([\s\S]*?)</div>`)
	tagRegex     = regexp.MustCompile(`<[^>]+>|\s+`)
)

// newHTTP: proxy yang tidak valid = error (bukan diam-diam lewat koneksi langsung).
func newHTTP(proxyURL string, timeout time.Duration) (*http.Client, error) {
	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 20 * time.Second,
	}
	if proxyURL != "" {
		u, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("proxy tidak valid: %w", err)
		}
		transport.Proxy = http.ProxyURL(u)
	}
	return &http.Client{Jar: jar, Transport: transport, Timeout: timeout}, nil
}

func getRandomProfile() browserProfile {
	idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(browserProfiles))))
	return browserProfiles[idx.Int64()]
}

func setBrowserHeaders(req *http.Request, p browserProfile) {
	req.Header.Set("User-Agent", p.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,id;q=0.8")
	req.Header.Set("Sec-Ch-Ua", p.SecChUa)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
}

func Register(email, password, promoCode, proxyURL string) error {
	client, err := newHTTP(proxyURL, 35*time.Second)
	if err != nil {
		return err
	}
	profile := getRandomProfile()

	// 1. Pre-flight GET untuk inisialisasi session cookie
	getReq, err := http.NewRequest("GET", BaseURL+"/users/register", nil)
	if err == nil {
		setBrowserHeaders(getReq, profile)
		if resp, err := client.Do(getReq); err == nil {
			_ = resp.Body.Close()
		}
	}

	// 2. Human jitter delay (1.5 - 2.5s)
	jitter, _ := rand.Int(rand.Reader, big.NewInt(1000))
	time.Sleep(time.Duration(1500+jitter.Int64()) * time.Millisecond)

	// 3. POST form registrasi
	form := url.Values{
		"data[User][username]":        {email},
		"data[User][password]":        {password},
		"data[User][confirmPassword]": {password},
		"data[User][accept]":          {"yes"},
	}
	if promoCode != "" {
		form.Set("data[User][coupon]", promoCode)
	}

	req, err := http.NewRequest("POST", BaseURL+"/users/register", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	setBrowserHeaders(req, profile)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", BaseURL)
	req.Header.Set("Referer", BaseURL+"/users/register")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := string(bodyBytes)

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return fmt.Errorf("registrasi gagal (HTTP %d)", resp.StatusCode)
	}
	if successRegex.MatchString(bodyStr) {
		return nil
	}
	if strings.Contains(bodyStr, "Temporary emails are not supported") {
		return fmt.Errorf("domain email diblokir Seed4Me")
	}
	if strings.Contains(bodyStr, "already exists") {
		return fmt.Errorf("email sudah terdaftar di Seed4Me")
	}
	if strings.Contains(bodyStr, "Please wait for some time") || strings.Contains(bodyStr, "can not register now") {
		return fmt.Errorf("IP terkena rate limit Seed4Me")
	}
	if match := alertRegex.FindStringSubmatch(bodyStr); len(match) > 1 {
		msg := strings.TrimSpace(tagRegex.ReplaceAllString(match[1], " "))
		if msg != "" {
			return fmt.Errorf("%s", msg)
		}
	}

	return fmt.Errorf("gagal submit form registrasi (status %d)", resp.StatusCode)
}

func Confirm(tokenOrURL, proxyURL string) error {
	client, err := newHTTP(proxyURL, 35*time.Second)
	if err != nil {
		return err
	}
	profile := getRandomProfile()

	var confirmURL string
	if strings.HasPrefix(tokenOrURL, "http://") || strings.HasPrefix(tokenOrURL, "https://") {
		confirmURL = tokenOrURL
	} else {
		confirmURL = fmt.Sprintf("%s/users/confirmEmail/%s", BaseURL, tokenOrURL)
	}

	u, err := url.Parse(confirmURL)
	if err != nil {
		return fmt.Errorf("URL konfirmasi tidak valid: %w", err)
	}
	baseHost, _ := url.Parse(BaseURL)
	allowedHost := func(host string) bool {
		host = strings.ToLower(host)
		return host == strings.ToLower(baseHost.Host) || strings.HasSuffix(host, ".spmailtechn.com")
	}
	if !allowedHost(u.Host) {
		return fmt.Errorf("host konfirmasi tidak dikenal: %s", u.Host)
	}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return http.ErrUseLastResponse
		}
		if !allowedHost(req.URL.Host) {
			return fmt.Errorf("redirect ke host tidak dikenal: %s", req.URL.Host)
		}
		return nil
	}

	req, err := http.NewRequest("GET", confirmURL, nil)
	if err != nil {
		return err
	}
	setBrowserHeaders(req, profile)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return nil
	}
	return fmt.Errorf("konfirmasi email gagal (HTTP %d)", resp.StatusCode)
}
