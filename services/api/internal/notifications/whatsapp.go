package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrWhatsAppNotConfigured = errors.New("WhatsApp sender is not configured")
	ErrWhatsAppFailed        = errors.New("WhatsApp sending failed")
)

type WhatsAppSender interface {
	SendMessage(context.Context, string, string) error
}

type NoopWhatsAppSender struct{}

func (NoopWhatsAppSender) SendMessage(context.Context, string, string) error {
	return nil
}

type SaungWAClient struct {
	endpoint   string
	appKey     string
	authKey    string
	sandbox    bool
	httpClient *http.Client
}

// NewSaungWAClient accepts SAUNGWA_API_KEY as "<appkey>:<authkey>".
func NewSaungWAClient(apiKey, endpoint string, sandbox bool) (*SaungWAClient, error) {
	parts := strings.SplitN(strings.TrimSpace(apiKey), ":", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return nil, fmt.Errorf("%w: SAUNGWA_API_KEY must use <appkey>:<authkey>", ErrWhatsAppNotConfigured)
	}
	if endpoint == "" {
		endpoint = "https://app.saungwa.com/api/create-message"
	}

	return &SaungWAClient{
		endpoint:   endpoint,
		appKey:     strings.TrimSpace(parts[0]),
		authKey:    strings.TrimSpace(parts[1]),
		sandbox:    sandbox,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *SaungWAClient) SendMessage(ctx context.Context, toPhone, message string) error {
	phone := normalizeWAPhone(toPhone)
	if phone == "" || strings.TrimSpace(message) == "" {
		return ErrValidation
	}

	payload := url.Values{
		"appkey":  {c.appKey},
		"authkey": {c.authKey},
		"to":      {phone},
		"message": {strings.TrimSpace(message)},
		"sandbox": {fmt.Sprintf("%t", c.sandbox)},
	}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.endpoint,
		bytes.NewBufferString(payload.Encode()),
	)
	if err != nil {
		return fmt.Errorf("create SaungWA request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrWhatsAppFailed, err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 512))
		return fmt.Errorf("%w: HTTP %d: %s", ErrWhatsAppFailed, response.StatusCode, strings.TrimSpace(string(body)))
	}

	var result struct {
		MessageStatus string `json:"message_status"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return fmt.Errorf("%w: decode response: %v", ErrWhatsAppFailed, err)
	}
	if !strings.EqualFold(result.MessageStatus, "success") {
		return fmt.Errorf("%w: message_status=%q", ErrWhatsAppFailed, result.MessageStatus)
	}

	return nil
}

func normalizeWAPhone(input string) string {
	var digits strings.Builder
	for _, char := range input {
		if char >= '0' && char <= '9' {
			digits.WriteRune(char)
		}
	}
	value := digits.String()
	switch {
	case strings.HasPrefix(value, "0"):
		return "62" + value[1:]
	case strings.HasPrefix(value, "8"):
		return "62" + value
	default:
		return value
	}
}