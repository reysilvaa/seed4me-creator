package client

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"seed4me-linux/internal/model"
)

const (
	CatchMailBaseURL = "https://api.catchmail.io/api/v1"
	defaultUA        = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36"
)

var (
	uuidPattern = regexp.MustCompile(`(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	firstNames  = []string{"alex", "sarah", "david", "james", "michael", "jessica", "daniel", "emma", "oliver", "sophia", "lucas", "mia", "noah", "ethan"}
	lastNames   = []string{"miller", "wilson", "taylor", "anderson", "thomas", "jackson", "white", "harris", "martin", "clark", "lewis", "walker", "hall"}
)

type CatchMailClient struct {
	http *http.Client
}

func NewCatchMailClient(proxyURL string) *CatchMailClient {
	transport := &http.Transport{}
	if proxyURL != "" {
		if u, err := url.Parse(proxyURL); err == nil {
			transport.Proxy = http.ProxyURL(u)
		}
	}
	return &CatchMailClient{
		http: &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
		},
	}
}

// GenerateEmail membuat username email dengan struktur nama natural untuk menghindari deteksi bot
func GenerateEmail() string {
	fnIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(firstNames))))
	lnIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(lastNames))))
	num, _ := rand.Int(rand.Reader, big.NewInt(900))
	return fmt.Sprintf("%s.%s%d@catchmail.io", firstNames[fnIdx.Int64()], lastNames[lnIdx.Int64()], num.Int64()+100)
}

func (c *CatchMailClient) GetMailbox(address string) (*model.MailboxResponse, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/mailbox?address=%s", CatchMailBaseURL, url.QueryEscape(address)), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", defaultUA)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status response: %d", resp.StatusCode)
	}

	var res model.MailboxResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *CatchMailClient) GetMessage(address, messageID string) (*model.MessageDetail, error) {
	reqURL := fmt.Sprintf("%s/message/%s?mailbox=%s", CatchMailBaseURL, messageID, url.QueryEscape(address))
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", defaultUA)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var msg model.MessageDetail
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (c *CatchMailClient) PollVerificationToken(address string, timeout time.Duration, logFn func(string)) (string, error) {
	deadline := time.Now().Add(timeout)
	attempt := 1

	for time.Now().Before(deadline) {
		if logFn != nil {
			logFn(fmt.Sprintf("Cek inbox %s (percobaan #%d)...", address, attempt))
		}

		mb, err := c.GetMailbox(address)
		if err == nil && len(mb.Messages) > 0 {
			for _, m := range mb.Messages {
				msg, err := c.GetMessage(address, m.ID)
				if err == nil {
					content := msg.Body.Text + " " + msg.Body.HTML
					if match := uuidPattern.FindString(content); match != "" {
						return match, nil
					}
				}
			}
		}

		attempt++
		time.Sleep(2500 * time.Millisecond)
	}

	return "", fmt.Errorf("timeout menunggu email verifikasi dari Seed4Me")
}
