package notifications

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func TestDispatcherValidation(t *testing.T) {
	dispatcher := NewDispatcher(nil, nil, nil, nil, slog.Default())

	dispatcher.Dispatch(DispatchJob{})
	dispatcher.Dispatch(DispatchJob{OrganizationID: "org"})
	dispatcher.Dispatch(DispatchJob{
		OrganizationID:  "org",
		RecipientUserID: "user",
		Type:            "test",
	})
}

func TestNotificationAuthorizationRequiresPrincipal(t *testing.T) {
	service := &Service{}

	if _, err := service.List(context.Background(), nil, Filter{}); !errors.Is(err, ErrValidation) {
		t.Fatalf("List error = %v, want ErrValidation", err)
	}
	if _, err := service.MarkAsRead(context.Background(), nil, "notification"); !errors.Is(err, ErrValidation) {
		t.Fatalf("MarkAsRead error = %v, want ErrValidation", err)
	}
	if _, err := service.MarkAllAsRead(context.Background(), nil); !errors.Is(err, ErrValidation) {
		t.Fatalf("MarkAllAsRead error = %v, want ErrValidation", err)
	}
}

func TestEmailHTML(t *testing.T) {
	got := emailHTML("Title <script>", "Line 1\nLine 2 <script>alert(1)</script>")
	want := "<h2>Title \x26lt;script\x26gt;</h2><p>Line 1<br>Line 2 \x26lt;script\x26gt;alert(1)\x26lt;/script\x26gt;</p>"
	if got != want {
		t.Errorf("emailHTML() = %q, want %q", got, want)
	}
}

func TestWhatsAppText(t *testing.T) {
	if got, want := whatsappText("Title", "Body"), "*Title*\n\nBody"; got != want {
		t.Errorf("whatsappText() = %q, want %q", got, want)
	}
	if got, want := whatsappText("Title", ""), "*Title*"; got != want {
		t.Errorf("whatsappText() = %q, want %q", got, want)
	}
}

func TestNoopSenders(t *testing.T) {
	if err := (NoopEmailSender{}).SendEmail(context.Background(), "to", "title", "body"); err != nil {
		t.Errorf("NoopEmailSender: %v", err)
	}
	if err := (NoopWhatsAppSender{}).SendMessage(context.Background(), "to", "message"); err != nil {
		t.Errorf("NoopWhatsAppSender: %v", err)
	}
}

func TestCreateParamsValidation(t *testing.T) {
	for _, test := range []struct {
		name  string
		params CreateParams
		valid bool
	}{
		{"valid", CreateParams{UserID: "u", Type: "t", Title: "title"}, true},
		{"missing user", CreateParams{Type: "t", Title: "title"}, false},
		{"missing type", CreateParams{UserID: "u", Title: "title"}, false},
		{"missing title", CreateParams{UserID: "u", Type: "t"}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := test.params.Validate() == nil; got != test.valid {
				t.Errorf("valid = %t, want %t", got, test.valid)
			}
		})
	}
}

func TestMessageFormattingNeverLeavesRawHTML(t *testing.T) {
	message := emailHTML("<b>Title</b>", "<img src=x>")
	if strings.Contains(message, "<b>") || strings.Contains(message, "<img") {
		t.Fatalf("unsafe message HTML: %s", message)
	}
}