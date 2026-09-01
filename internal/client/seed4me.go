package client

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	Seed4MeBaseURL = "https://seed4.me"
	seed4MeUA      = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
)

var (
	successBannerRegex = regexp.MustCompile(`(?i)<h4[^>]*>\s*Success\s*</h4>`)
	failureBannerRegex = regexp.MustCompile(`(?i)<h4[^>]*>\s*Failure\s*</h4>[\s\S]*?<p[^>]*>([\s\S]*?)</p>`)
	tagStripRegex      = regexp.MustCompile(`<[^>]+>`)
	whiteSpaceRegex    = regexp.MustCompile(`\s+`)
)

type Seed4MeClient struct {
	http *http.Client
}

func NewSeed4MeClient(proxyURL string) *Seed4MeClient {
	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{}
	if proxyURL != "" {
		if u, err := url.Parse(proxyURL); err == nil {
			transport.Proxy = http.ProxyURL(u)
		}
	}
	return &Seed4MeClient{
		http: &http.Client{
			Jar:       jar,
			Transport: transport,
			Timeout:   15 * time.Second,
		},
	}
}

func GeneratePassword() string {
	const charset = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789!@#$"
	b := make([]byte, 12)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

func (c *Seed4MeClient) Register(email, password, promoCode string) error {
	reqGet, _ := http.NewRequest("GET", Seed4MeBaseURL+"/users/register", nil)
	reqGet.Header.Set("User-Agent", seed4MeUA)
	if resp, err := c.http.Do(reqGet); err == nil {
		_ = resp.Body.Close()
	}

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

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	body := string(bodyBytes)

	if successBannerRegex.MatchString(body) ||
		strings.Contains(strings.ToLower(body), "confirmation code has been sent") ||
		strings.Contains(strings.ToLower(body), "confirm email address") {
		return nil
	}

	if match := failureBannerRegex.FindStringSubmatch(body); len(match) > 1 {
		errText := tagStripRegex.ReplaceAllString(match[1], " ")
		errText = whiteSpaceRegex.ReplaceAllString(errText, " ")
		errText = strings.TrimSpace(errText)
		if strings.Contains(errText, "install App to your phone") || strings.Contains(errText, "promotional code") {
			return fmt.Errorf("IP terkena rate limit Seed4Me")
		}
		if errText != "" {
			return fmt.Errorf("%s", errText)
		}
	}

	if strings.Contains(body, "Temporary emails are not supported") {
		return fmt.Errorf("domain email diblokir Seed4Me")
	}
	if strings.Contains(body, "You can not register now") {
		return fmt.Errorf("IP terkena rate limit Seed4Me")
	}

	return nil
}

func (c *Seed4MeClient) ConfirmEmail(tokenOrURL string) error {
	tokenOrURL = strings.TrimSpace(tokenOrURL)
	targetURL := tokenOrURL
	if !strings.HasPrefix(tokenOrURL, "http") {
		targetURL = fmt.Sprintf("%s/users/confirmEmail/%s", Seed4MeBaseURL, tokenOrURL)
	}

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", seed4MeUA)

	// Try with current client (proxy)
	resp, err := c.http.Do(req)
	if err == nil {
		_ = resp.Body.Close()
		return nil
	}

	// Fallback to direct request if proxy fails during confirmation
	directClient := &http.Client{Timeout: 10 * time.Second}
	if directResp, dErr := directClient.Do(req); dErr == nil {
		_ = directResp.Body.Close()
		return nil
	}

	return err
}

func (c *Seed4MeClient) LoginAndCheck(email, password string) (string, error) {
	form := url.Values{
		"data[User][username]": {email},
		"data[User][password]": {password},
	}

	req, err := http.NewRequest("POST", Seed4MeBaseURL+"/users/login", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", seed4MeUA)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", Seed4MeBaseURL)
	req.Header.Set("Referer", Seed4MeBaseURL+"/users/login")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	body := string(bodyBytes)

	if strings.Contains(body, "Invalid username or password") {
		return "", fmt.Errorf("username atau password salah")
	}

	expRegex := regexp.MustCompile(`(?i)Expires:?\s*<b>([^<]+)</b>`)
	if match := expRegex.FindStringSubmatch(body); len(match) > 1 {
		return match[1], nil
	}

	return "Aktif (Free Trial 7 Hari)", nil
}
