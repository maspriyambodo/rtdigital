package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Mailer interface {
	SendEmail(context.Context, string, string, string) error
}

type NoopMailer struct{}

func (NoopMailer) SendEmail(context.Context, string, string, string) error {
	return nil
}

type ResendMailer struct {
	apiKey     string
	fromEmail  string
	httpClient *http.Client
}

func NewResendMailer(apiKey, fromEmail string) *ResendMailer {
	return &ResendMailer{
		apiKey:     apiKey,
		fromEmail:  fromEmail,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (m *ResendMailer) SendEmail(ctx context.Context, to, subject, html string) error {
	if m.apiKey == "" || m.fromEmail == "" {
		return fmt.Errorf("Resend mailer is not configured")
	}

	payload, err := json.Marshal(map[string]any{
		"from":    m.fromEmail,
		"to":      []string{to},
		"subject": subject,
		"html":    html,
	})
	if err != nil {
		return fmt.Errorf("marshal Resend payload: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.resend.com/emails",
		bytes.NewReader(payload),
	)
	if err != nil {
		return fmt.Errorf("create Resend request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+m.apiKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := m.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("send Resend email: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("Resend returned HTTP %d", response.StatusCode)
	}
	return nil
}
