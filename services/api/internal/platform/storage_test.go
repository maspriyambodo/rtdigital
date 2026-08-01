package platform_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
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

func TestStorageIntegration(t *testing.T) {
	cfg := config.R2Config{
		Endpoint:        os.Getenv("R2_ENDPOINT"),
		Bucket:          os.Getenv("R2_BUCKET"),
		AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		UsePathStyle:    os.Getenv("R2_USE_PATH_STYLE") == "true",
	}
	if cfg.Endpoint == "" || cfg.Bucket == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		t.Skip("R2 integration environment is not configured")
	}

	storage, err := platform.NewStorage(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewStorage() error = %v", err)
	}

	key := "integration-tests/storage-" + time.Now().UTC().Format("20060102150405.000000000") + ".txt"
	body := []byte("RT Digital storage integration")

	upload, err := storage.PresignUpload(
		context.Background(),
		key,
		"text/plain",
		time.Minute,
	)
	if err != nil {
		t.Fatalf("PresignUpload() error = %v", err)
	}

	request, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPut,
		upload.URL,
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	request.Header.Set("Content-Type", upload.Headers["Content-Type"])

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("upload request error = %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("upload status = %d, want %d", response.StatusCode, http.StatusOK)
	}

	downloadURL, err := storage.PresignDownload(context.Background(), key, time.Minute)
	if err != nil {
		t.Fatalf("PresignDownload() error = %v", err)
	}

	response, err = http.Get(downloadURL)
	if err != nil {
		t.Fatalf("download request error = %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("download status = %d, want %d", response.StatusCode, http.StatusOK)
	}

	downloaded, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("io.ReadAll() error = %v", err)
	}
	if !bytes.Equal(downloaded, body) {
		t.Errorf("downloaded body = %q, want %q", downloaded, body)
	}
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
