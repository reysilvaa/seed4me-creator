package model

// Account merepresentasikan entitas akun Seed4Me yang berhasil dibuat
type Account struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}
