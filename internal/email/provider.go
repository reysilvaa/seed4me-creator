package email

import (
	"regexp"
	"time"

	"seed4me-creator/internal/config"
)

var (
	tokenRegex = regexp.MustCompile(`(?i)confirmEmail/([a-zA-Z0-9_-]+)`)
	linkRegex  = regexp.MustCompile(`https?://[^\s"'<>]+confirmEmail/[a-zA-Z0-9_-]+`)
	trackRegex = regexp.MustCompile(`https?://post\.spmailtechn\.com/f/a/[^\s"'<>]+`)
	uuidPattern = regexp.MustCompile(`(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
)
// Provider mendefinisikan interface umum untuk seluruh layanan email sementara
type Provider interface {
	GenerateEmail() (string, error)
	PollToken(email string, timeout time.Duration) (string, error)
}

// GetProvider mengembalikan instance Provider sesuai konfigurasi
func GetProvider(cfg config.Config) Provider {
	switch cfg.EmailService {
	case config.EmailServiceTempMailLol:
		return NewTempMailLolClient(cfg.TempMailLolKey)
	case config.EmailServiceCatchMail:
		return NewCatchMailClient(cfg.EmailDomain)
	default:
		return NewTempMailIngClient()
	}
}
