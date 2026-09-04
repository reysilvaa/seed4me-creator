package email

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	CatchMailBaseURL = "https://api.catchmail.io/api/v1"
	catchMailUA      = "Mozilla/5.0"
)

type CatchMailboxItem struct {
	ID      string `json:"id"`
	From    string `json:"from"`
	Subject string `json:"subject"`
}

type CatchMailboxResponse struct {
	Address  string             `json:"address"`
	Messages []CatchMailboxItem `json:"messages"`
}

type CatchMessageDetail struct {
	ID   string `json:"id"`
	Body struct {
		Text string `json:"text"`
		HTML string `json:"html"`
	} `json:"body"`
}

type CatchMailClient struct {
	Domain     string
	HTTPClient *http.Client
}

func NewCatchMailClient(domain string) *CatchMailClient {
	if domain == "" {
		domain = "catchmail.io"
	}
	return &CatchMailClient{
		Domain:     domain,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *CatchMailClient) GenerateEmail() (string, error) {
	firstNames := []string{"alex", "sarah", "david", "james", "michael", "jessica", "daniel", "emma", "oliver", "sophia", "lucas", "mia", "noah", "ethan"}
	lastNames := []string{"miller", "wilson", "taylor", "anderson", "thomas", "jackson", "white", "harris", "martin", "clark", "walker", "hall"}

	fnIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(firstNames))))
	lnIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(lastNames))))
	num, _ := rand.Int(rand.Reader, big.NewInt(9000))
	return fmt.Sprintf("%s.%s%d@%s", firstNames[fnIdx.Int64()], lastNames[lnIdx.Int64()], num.Int64()+1000, c.Domain), nil
}

func (c *CatchMailClient) PollToken(address string, timeout time.Duration) (string, error) {
	if timeout <= 0 {
		timeout = 45 * time.Second
	}

	deadline := time.Now().Add(timeout)
	seen := make(map[string]bool)
	for time.Now().Before(deadline) {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/mailbox?address=%s", CatchMailBaseURL, url.QueryEscape(address)), nil)
		if err == nil {
			req.Header.Set("User-Agent", catchMailUA)
			if resp, err := c.HTTPClient.Do(req); err == nil {
				var mb CatchMailboxResponse
				_ = json.NewDecoder(resp.Body).Decode(&mb)
				_ = resp.Body.Close()

				for _, m := range mb.Messages {
					if seen[m.ID] {
						continue
					}
					msgReq, err := http.NewRequest("GET", fmt.Sprintf("%s/message/%s?mailbox=%s", CatchMailBaseURL, m.ID, url.QueryEscape(address)), nil)
					if err != nil {
						continue
					}
					msgReq.Header.Set("User-Agent", catchMailUA)
					msgResp, err := c.HTTPClient.Do(msgReq)
					if err != nil {
						continue // jangan tandai seen — fetch gagal transient, coba lagi di poll berikutnya
					}
					var detail CatchMessageDetail
					decErr := json.NewDecoder(msgResp.Body).Decode(&detail)
					_ = msgResp.Body.Close()
					if decErr != nil {
						continue
					}
					seen[m.ID] = true

					content := detail.Body.Text + " " + detail.Body.HTML
					if match := tokenRegex.FindStringSubmatch(content); len(match) > 1 {
						return match[1], nil
					}
					if match := uuidPattern.FindString(content); match != "" && strings.Contains(strings.ToLower(content), "confirm") {
						return match, nil
					}
				}
			}
		}
		time.Sleep(2 * time.Second)
	}

	return "", fmt.Errorf("timeout menunggu verifikasi email di CatchMail (%s)", address)
}
