package email

import "testing"

func TestSharedTokenRegexes(t *testing.T) {
	html := `<a href="https://seed4.me/users/confirmEmail/AbC123_xYz">confirm</a>`

	if m := tokenRegex.FindStringSubmatch(html); len(m) != 2 || m[1] != "AbC123_xYz" {
		t.Fatalf("tokenRegex gagal: %v", m)
	}
	if m := linkRegex.FindString(html); m == "" {
		t.Fatal("linkRegex tidak menemukan link konfirmasi")
	}
	if m := uuidPattern.FindString(`<p>kode: 550e8400-e29b-41d4-a716-446655440000</p>`); m == "" {
		t.Fatal("uuidPattern tidak menemukan UUID")
	}
}