package email

import (
	"fmt"
	"regexp"
	"time"

	"seed4me-creator/internal/config"
)

var (
	tokenRegex  = regexp.MustCompile(`(?i)confirmEmail/([a-zA-Z0-9_-]+)`)
	linkRegex   = regexp.MustCompile(`https?://[^\s"'<>]+confirmEmail/[a-zA-Z0-9_-]+`)
	trackRegex  = regexp.MustCompile(`https?://post\.spmailtechn\.com/f/a/[^\s"'<>]+`)
	uuidPattern = regexp.MustCompile(`(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
)

// Provider mendefinisikan interface umum untuk seluruh layanan email sementara.
type Provider interface {
	GenerateEmail() (string, error)
	PollToken(email string, timeout time.Duration) (string, error)
}

func extractToken(content string) (string, bool) {
	if match := tokenRegex.FindStringSubmatch(content); len(match) > 1 {
		return match[1], true
	}
	if match := linkRegex.FindString(content); match != "" {
		return match, true
	}
	if match := trackRegex.FindString(content); match != "" {
		return match, true
	}
	return "", false
}

// GetProvider mengembalikan instance Provider sesuai konfigurasi.
func GetProvider(cfg config.Config) (Provider, error) {
	switch cfg.EmailService {
	case config.EmailServiceTempMailLol:
		return NewTempMailLolClient(cfg.TempMailLolKey), nil
	case config.EmailServiceCatchMail:
		return NewCatchMailClient(cfg.EmailDomain), nil
	case config.EmailServiceTempMailIng:
		return NewTempMailIngClient(), nil
	default:
		return nil, fmt.Errorf("provider email tidak dikenal: %q", cfg.EmailService)
	}
}
