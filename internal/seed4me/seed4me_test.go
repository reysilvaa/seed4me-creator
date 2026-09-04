package seed4me

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func withTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	old := BaseURL
	BaseURL = srv.URL
	t.Cleanup(func() { BaseURL = old })
	return srv
}

func TestRegisterUnknownResponseFails(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html><body>an unexpected page</body></html>"))
	})

	if err := Register("a@b.io", "pw123", "", ""); err == nil {
		t.Fatal("expected error on unknown response, got nil")
	}
}

func TestRegisterHTTP500Fails(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	if err := Register("a@b.io", "pw123", "", ""); err == nil {
		t.Fatal("expected error on HTTP 500, got nil")
	}
}

func TestRegisterSuccessReturnsNil(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html><body>Please confirm email address</body></html>"))
	})

	if err := Register("a@b.io", "pw123", "", ""); err != nil {
		t.Fatalf("expected nil on success page, got: %v", err)
	}
}

func TestConfirmSuccess(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><body>Email confirmed!</body></html>"))
	})

	if err := Confirm("testtoken123", ""); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestConfirmHTTP500Fails(t *testing.T) {
	withTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	if err := Confirm("testtoken123", ""); err == nil {
		t.Fatal("expected error on HTTP 500, got nil")
	}
}
