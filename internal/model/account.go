package model

// Account merepresentasikan entitas akun Seed4Me yang berhasil dibuat
type Account struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	Status    string `json:"status"`
	PSK       string `json:"psk"`
	CreatedAt string `json:"created_at"`
	Notes     string `json:"notes,omitempty"`
}

// MailboxItem merepresentasikan item email dari inbox CatchMail
type MailboxItem struct {
	ID      string `json:"id"`
	Mailbox string `json:"mailbox"`
	From    string `json:"from"`
	Subject string `json:"subject"`
}

// MailboxResponse adalah format response API CatchMail
type MailboxResponse struct {
	Address  string        `json:"address"`
	Messages []MailboxItem `json:"messages"`
}

// MessageDetail adalah konten pesan email lengkap
type MessageDetail struct {
	ID   string `json:"id"`
	Body struct {
		Text string `json:"text"`
		HTML string `json:"html"`
	} `json:"body"`
}
