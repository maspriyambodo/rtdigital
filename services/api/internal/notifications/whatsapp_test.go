package notifications

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewSaungWAClient(t *testing.T) {
	_, err := NewSaungWAClient("app:auth", "http://example.test", true)
	if err != nil {
		t.Fatalf("valid configuration: %v", err)
	}

	_, err = NewSaungWAClient("app:", "", false)
	if !errors.Is(err, ErrWhatsAppNotConfigured) {
		t.Fatalf("invalid configuration error = %v, want ErrWhatsAppNotConfigured", err)
	}
}

func TestSaungWAClientSendMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("content type = %q", r.Header.Get("Content-Type"))
		}
		if r.FormValue("appkey") != "test-app" || r.FormValue("authkey") != "test-auth" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.FormValue("to") == "628123" && r.FormValue("message") == "ok" && r.FormValue("sandbox") == "true" {
			_, _ = w.Write([]byte(`{"message_status":"Success"}`))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message_status":"Failed"}`))
	}))
	defer server.Close()

	client, err := NewSaungWAClient(" test-app : test-auth ", server.URL, true)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if err := client.SendMessage(context.Background(), "08123", "ok"); err != nil {
		t.Fatalf("send success: %v", err)
	}
	if err := client.SendMessage(context.Background(), "invalid", "fail"); !errors.Is(err, ErrValidation) {
		t.Fatalf("invalid phone error = %v, want ErrValidation", err)
	}
	if err := client.SendMessage(context.Background(), "62899", "fail"); !errors.Is(err, ErrWhatsAppFailed) {
		t.Fatalf("provider failure error = %v, want ErrWhatsAppFailed", err)
	}
}

func TestNormalizeWAPhone(t *testing.T) {
	tests := []struct{ input, want string }{
		{"0812345", "62812345"},
		{"812345", "62812345"},
		{"+62 812-345", "62812345"},
		{"62812345", "62812345"},
		{"invalid", ""},
	}
	for _, test := range tests {
		if got := normalizeWAPhone(test.input); got != test.want {
			t.Errorf("normalizeWAPhone(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}