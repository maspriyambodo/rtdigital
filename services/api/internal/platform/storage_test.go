package platform_test

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/maspriyambodo/rtdigital/services/api/internal/config"
	"github.com/maspriyambodo/rtdigital/services/api/internal/platform"
)

func TestStoragePresignURLs(t *testing.T) {
	t.Parallel()

	storage, err := platform.NewStorage(context.Background(), config.R2Config{
		Endpoint:        "http://localhost:9000",
		Bucket:          "rtdigital-local",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		UsePathStyle:    true,
	})
	if err != nil {
		t.Fatalf("NewStorage() error = %v", err)
	}

	upload, err := storage.PresignUpload(
		context.Background(),
		"payments/test.jpg",
		"image/jpeg",
		15*time.Minute,
	)
	if err != nil {
		t.Fatalf("PresignUpload() error = %v", err)
	}
	assertPresignedURL(t, upload.URL, "/rtdigital-local/payments/test.jpg")
	if got := upload.Headers["Content-Type"]; got != "image/jpeg" {
		t.Errorf("upload Content-Type = %q, want %q", got, "image/jpeg")
	}

	downloadURL, err := storage.PresignDownload(
		context.Background(),
		"payments/test.jpg",
		15*time.Minute,
	)
	if err != nil {
		t.Fatalf("PresignDownload() error = %v", err)
	}
	assertPresignedURL(t, downloadURL, "/rtdigital-local/payments/test.jpg")
}

func assertPresignedURL(t *testing.T, rawURL, wantPath string) {
	t.Helper()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	if parsed.Host != "localhost:9000" {
		t.Errorf("host = %q, want %q", parsed.Host, "localhost:9000")
	}
	if parsed.Path != wantPath {
		t.Errorf("path = %q, want %q", parsed.Path, wantPath)
	}
	if !strings.Contains(parsed.RawQuery, "X-Amz-Signature=") {
		t.Errorf("missing signature in URL: %s", rawURL)
	}
}
