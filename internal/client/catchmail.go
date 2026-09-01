package client

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const CatchMailBaseURL = "https://api.catchmail.io/api/v1"

type CatchMailboxItem struct {
	ID      string `json:"id"`
	Mailbox string `json:"mailbox"`
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

// GenerateEmail creates a natural name email address with the specified domain
func GenerateEmail(domain string) string {
	if domain == "" {
		domain = "catchmail.io"
	}
	firstNames := []string{"alex", "sarah", "david", "james", "michael", "jessica", "daniel", "emma", "oliver", "sophia", "lucas", "mia", "noah", "ethan"}
	lastNames := []string{"miller", "wilson", "taylor", "anderson", "thomas", "jackson", "white", "harris", "martin", "clark", "walker", "hall"}

	fnIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(firstNames))))
	lnIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(lastNames))))
	num, _ := rand.Int(rand.Reader, big.NewInt(9000))
	return fmt.Sprintf("%s.%s%d@%s", firstNames[fnIdx.Int64()], lastNames[lnIdx.Int64()], num.Int64()+1000, domain)
}

// PollToken polls the CatchMail API for verification emails until confirmation token is found
func PollToken(address string, timeout time.Duration, logFn func(string)) (string, error) {
	if timeout <= 0 {
		timeout = 45 * time.Second
	}

	uuidPattern := regexp.MustCompile(`(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	confirmPattern := regexp.MustCompile(`(?i)confirmEmail/([a-zA-Z0-9_-]+)`)

	client := &http.Client{Timeout: 10 * time.Second}
	deadline := time.Now().Add(timeout)
	attempt := 1

	for time.Now().Before(deadline) {
		if logFn != nil {
			logFn(fmt.Sprintf("Cek inbox %s (percobaan #%d)...", address, attempt))
		}

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/mailbox?address=%s", CatchMailBaseURL, url.QueryEscape(address)), nil)
		if err == nil {
			req.Header.Set("User-Agent", "Mozilla/5.0")

			if resp, err := client.Do(req); err == nil {
				var mb CatchMailboxResponse
				_ = json.NewDecoder(resp.Body).Decode(&mb)
				_ = resp.Body.Close()

				for _, m := range mb.Messages {
					msgReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/message/%s?mailbox=%s", CatchMailBaseURL, m.ID, url.QueryEscape(address)), nil)
					if msgReq != nil {
						msgReq.Header.Set("User-Agent", "Mozilla/5.0")

						if msgResp, err := client.Do(msgReq); err == nil {
							var msg CatchMessageDetail
							_ = json.NewDecoder(msgResp.Body).Decode(&msg)
							_ = msgResp.Body.Close()

							content := msg.Body.Text + " " + msg.Body.HTML
							if match := confirmPattern.FindStringSubmatch(content); len(match) > 1 {
								return strings.TrimSpace(match[1]), nil
							}
							if match := uuidPattern.FindString(content); match != "" {
								return strings.TrimSpace(match), nil
							}
						}
					}
				}
			}
		}

		attempt++
		time.Sleep(3 * time.Second)
	}

	return "", fmt.Errorf("timeout menunggu email verifikasi di CatchMail (%s)", address)
}
