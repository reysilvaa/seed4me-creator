package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	TempMailLolBaseURL = "https://api.tempmail.lol/v3"
	tempMailLolUA      = "Mozilla/5.0"
)

type TempMailLolInboxResp struct {
	Address string `json:"address"`
	Token   string `json:"token"`
}

type TempMailLolWaitResp struct {
	Emails []struct {
		ID      string `json:"id"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
		HTML    string `json:"html"`
	} `json:"emails"`
	TimedOut bool `json:"timed_out"`
}

type TempMailLolClient struct {
	APIKey     string
	HTTPClient *http.Client
	mu         sync.RWMutex
	Tokens     map[string]string
}

func NewTempMailLolClient(apiKey string) *TempMailLolClient {
	return &TempMailLolClient{
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 45 * time.Second},
		Tokens:     make(map[string]string),
	}
}

func (c *TempMailLolClient) GenerateEmail() (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("TEMPMAIL_LOL_KEY belum diset — isi tempmail_lol_key di config.json atau env TEMPMAIL_LOL_KEY")
	}
	payload, _ := json.Marshal(map[string]string{"prefix": "seed"})

	var lastErr error
	for attempt := 1; attempt <= 5; attempt++ {
		req, err := http.NewRequest("POST", TempMailLolBaseURL+"/inboxes", bytes.NewReader(payload))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
		req.Header.Set("User-Agent", tempMailLolUA)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			// DNS/net transient (mis. "no such host") — coba lagi sampai 10 detik
			lastErr = err
			time.Sleep(2 * time.Second)
			continue
		}

		var result TempMailLolInboxResp
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || result.Address == "" {
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("gagal membuat inbox TempMail.lol")
			time.Sleep(2 * time.Second)
			continue
		}
		_ = resp.Body.Close()

		c.mu.Lock()
		c.Tokens[result.Address] = result.Token
		c.mu.Unlock()

		return result.Address, nil
	}
	return "", fmt.Errorf("gagal membuat inbox TempMail.lol: %v", lastErr)
}

func (c *TempMailLolClient) PollToken(email string, timeout time.Duration) (string, error) {
	c.mu.Lock()
	token, ok := c.Tokens[email]
	delete(c.Tokens, email)
	c.mu.Unlock()

	if !ok {
		return "", fmt.Errorf("token inbox tidak ditemukan untuk %s", email)
	}

	timeoutSec := int(timeout.Seconds())
	if timeoutSec > 60 {
		timeoutSec = 60
	}
	if timeoutSec < 10 {
		timeoutSec = 30
	}

	waitURL := fmt.Sprintf("%s/inboxes/%s/wait?timeout=%d", TempMailLolBaseURL, url.PathEscape(token), timeoutSec)
	req, err := http.NewRequest("GET", waitURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("User-Agent", tempMailLolUA)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var wResp TempMailLolWaitResp
	if err := json.NewDecoder(resp.Body).Decode(&wResp); err != nil {
		return "", fmt.Errorf("gagal parsing respon wait TempMail.lol: %w", err)
	}

	if len(wResp.Emails) > 0 {
		for _, item := range wResp.Emails {
			content := fmt.Sprintf("%s\n%s\n%s", item.Subject, item.Body, item.HTML)
			if match := tokenRegex.FindStringSubmatch(content); len(match) > 1 {
				return match[1], nil
			}
			if match := linkRegex.FindString(content); match != "" {
				return match, nil
			}
			if match := trackRegex.FindString(content); match != "" {
				return match, nil
			}
		}
	}

	return "", fmt.Errorf("timeout menunggu email di TempMail.lol (%s)", email)
}
