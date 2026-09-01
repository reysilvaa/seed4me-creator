package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Arahkan Seed4MeBaseURL ke server test (biar tidak hit jaringan asli).
func withTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	old := Seed4MeBaseURL
	Seed4MeBaseURL = srv.URL
	t.Cleanup(func() { Seed4MeBaseURL = old })
	return srv
}

func TestRegisterSeed4MeUnknownResponseFails(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html><body>an unexpected page</body></html>"))
	})

	if err := RegisterSeed4Me("a@b.io", "pw123", "", ""); err == nil {
		t.Fatal("expected error on unknown response, got nil")
	}
}

func TestRegisterSeed4MeHTTP500Fails(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	if err := RegisterSeed4Me("a@b.io", "pw123", "", ""); err == nil {
		t.Fatal("expected error on HTTP 500, got nil")
	}
}

func TestRegisterSeed4MeSuccessReturnsNil(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<p>confirmation code has been sent to your email</p>"))
	})

	if err := RegisterSeed4Me("a@b.io", "pw123", "", ""); err != nil {
		t.Fatalf("expected nil on success page, got %v", err)
	}
}

func TestConfirmSeed4MeHTTP500Fails(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	if err := ConfirmSeed4Me("tok123", ""); err == nil {
		t.Fatal("expected error on HTTP 500, got nil")
	}
}

func TestConfirmSeed4MeOKReturnsNil(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if err := ConfirmSeed4Me("tok123", ""); err != nil {
		t.Fatalf("expected nil on 200, got %v", err)
	}
}
