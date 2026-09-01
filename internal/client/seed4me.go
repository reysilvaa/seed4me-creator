package client

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	seed4MeUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
)

var (
	successRegex = regexp.MustCompile(`(?i)(confirmation code has been sent|confirm email address|please check your email|verification code)`)
	alertRegex   = regexp.MustCompile(`(?i)<div[^>]*class="[^"]*alert[^"]*"[^>]*>([\s\S]*?)</div>`)
	tagRegex     = regexp.MustCompile(`<[^>]+>|\s+`)
)

var Seed4MeBaseURL = "https://seed4.me"

func newHTTP(proxyURL string, timeout time.Duration) *http.Client {
	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{}
	if proxyURL != "" {
		if u, err := url.Parse(proxyURL); err == nil {
			transport.Proxy = http.ProxyURL(u)
		}
	}
	return &http.Client{Jar: jar, Transport: transport, Timeout: timeout}
}

func GeneratePassword() string {
	return "12345678"
}

func RegisterSeed4Me(email, password, promoCode, proxyURL string) error {
	client := newHTTP(proxyURL, 15*time.Second)

	form := url.Values{
		"data[User][username]":        {email},
		"data[User][password]":        {password},
		"data[User][confirmPassword]": {password},
		"data[User][promoCode]":       {strings.TrimSpace(promoCode)},
		"data[User][accept]":          {"yes"},
	}

	req, err := http.NewRequest("POST", Seed4MeBaseURL+"/users/register", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", seed4MeUA)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", Seed4MeBaseURL)
	req.Header.Set("Referer", Seed4MeBaseURL+"/users/register")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("gagal baca respon: %w", err)
	}
	body := string(bodyBytes)

	if resp.StatusCode >= http.StatusBadRequest {
		b := body
		if len(b) > 300 {
			b = b[:300]
		}
		return fmt.Errorf("registrasi gagal (status %d): %s", resp.StatusCode, b)
	}

	// 1. Cek pesan error domain temporary
	if strings.Contains(body, "Temporary emails are not supported") {
		return fmt.Errorf("domain email diblokir Seed4Me (Temporary emails not supported)")
	}

	// 2. Cek rate limit IP
	if strings.Contains(body, "You can not register now") || strings.Contains(body, "install App to your phone") {
		return fmt.Errorf("IP terkena rate limit Seed4Me")
	}

	// 3. Cek sukses
	if successRegex.MatchString(body) {
		return nil
	}

	// 4. Ekstrak pesan alert dari HTML jika gagal
	if matches := alertRegex.FindStringSubmatch(body); len(matches) > 1 {
		errText := strings.TrimSpace(tagRegex.ReplaceAllString(matches[1], " "))
		if errText != "" && !strings.Contains(errText, "display:none") {
			return fmt.Errorf("Seed4Me: %s", errText)
		}
	}

	return fmt.Errorf("registrasi gagal / ditolak oleh Seed4Me (periksa domain email & proxy)")
}

func ConfirmSeed4Me(token, proxyURL string) error {
	token = strings.TrimSpace(token)
	targetURL := token
	if !strings.HasPrefix(token, "http") {
		targetURL = fmt.Sprintf("%s/users/confirmEmail/%s", Seed4MeBaseURL, token)
	}

	client := newHTTP(proxyURL, 10*time.Second)
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", seed4MeUA)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	// Status non-2xx/3xx = aktivasi gagal; proxy gagal = return error,
	// jangan fallback koneksi langsung (bocor IP asli & batal rotasi).
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("aktivasi gagal (status %d)", resp.StatusCode)
	}
	return nil
}
