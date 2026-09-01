package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	TempMailIngBaseURL = "https://api.tempmail.ing/api"
	tempMailIngUA      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
)

type TempMailIngGenerateResp struct {
	Success bool   `json:"success"`
	Address string `json:"address"`
	Email   struct {
		Address string `json:"address"`
	} `json:"email"`
}

type TempMailIngEmailItem struct {
	ID          any    `json:"id"`
	Subject     string `json:"subject"`
	Content     string `json:"content"`
	HTML        string `json:"html"`
	HTMLBody    string `json:"html_body"`
	Text        string `json:"text"`
	TextBody    string `json:"text_body"`
	Body        string `json:"body"`
	FromAddress string `json:"from_address"`
}

type TempMailIngInboxResp struct {
	Success bool                   `json:"success"`
	Emails  []TempMailIngEmailItem `json:"emails"`
}

type TempMailIngClient struct {
	HTTPClient *http.Client
}

func NewTempMailIngClient() *TempMailIngClient {
	return &TempMailIngClient{
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *TempMailIngClient) GenerateEmail() (string, error) {
	for attempt := 1; attempt <= 3; attempt++ {
		payload, _ := json.Marshal(map[string]string{"duration": "1d"})
		req, err := http.NewRequest("POST", TempMailIngBaseURL+"/generate", bytes.NewReader(payload))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", tempMailIngUA)
		req.Header.Set("Origin", "https://tempmail.ing")
		req.Header.Set("Referer", "https://tempmail.ing/")

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			time.Sleep(1500 * time.Millisecond)
			continue
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		var result TempMailIngGenerateResp
		if err := json.Unmarshal(bodyBytes, &result); err == nil {
			if result.Email.Address != "" {
				return result.Email.Address, nil
			}
			if result.Address != "" {
				return result.Address, nil
			}
		}

		if strings.Contains(string(bodyBytes), "slow down") || strings.Contains(string(bodyBytes), "Too many requests") {
			time.Sleep(2500 * time.Millisecond)
			continue
		}

		return "", fmt.Errorf("gagal generate email dari TempMail.ing: %s", string(bodyBytes))
	}
	return "", fmt.Errorf("timeout generate email dari TempMail.ing (rate limit)")
}

func (c *TempMailIngClient) PollToken(email string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	reqURL := fmt.Sprintf("%s/emails/%s", TempMailIngBaseURL, url.PathEscape(email))

	for time.Now().Before(deadline) {
		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		req.Header.Set("User-Agent", tempMailIngUA)
		req.Header.Set("Origin", "https://tempmail.ing")
		req.Header.Set("Referer", "https://tempmail.ing/")

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		var inboxResp TempMailIngInboxResp
		_ = json.NewDecoder(resp.Body).Decode(&inboxResp)
		_ = resp.Body.Close()

		if len(inboxResp.Emails) > 0 {
			for _, item := range inboxResp.Emails {
				raw := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n%s", item.Subject, item.Content, item.Text, item.TextBody, item.HTML, item.HTMLBody, item.Body)
				if match := tokenRegex.FindStringSubmatch(raw); len(match) > 1 {
					return match[1], nil
				}
				if match := linkRegex.FindString(raw); match != "" {
					return match, nil
				}
				if match := trackRegex.FindString(raw); match != "" {
					return match, nil
				}
			}
		}
		time.Sleep(2500 * time.Millisecond)
	}

	return "", fmt.Errorf("timeout menunggu email konfirmasi di TempMail.ing (%s)", email)
}
