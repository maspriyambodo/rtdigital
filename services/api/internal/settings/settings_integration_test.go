package settings_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
	"github.com/maspriyambodo/rtdigital/services/api/internal/platform"
	"github.com/maspriyambodo/rtdigital/services/api/internal/settings"
)

func TestSettingsIntegration(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := platform.NewDatabase(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	orgID := testUUID(t)
	userID := testUUID(t)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM organizations WHERE id = $1`, orgID)
	})

	if _, err := pool.Exec(ctx, `
		INSERT INTO organizations (id, name, rt_number, rw_number, status)
		VALUES ($1, 'RT Staging Test', '005', '005', 'active')`,
		orgID,
	); err != nil {
		t.Fatalf("insert organization: %v", err)
	}

	service := settings.NewService(pool)
	principal := &auth.Principal{
		UserID:         userID,
		OrganizationID: orgID,
		RoleCodes:      []string{"ketua_rt"},
	}

	t.Run("get settings default values", func(t *testing.T) {
		current, err := service.Get(ctx, orgID)
		if err != nil {
			t.Fatalf("Get(): %v", err)
		}
		if current.Name != "RT Staging Test" {
			t.Errorf("got settings name %q; want RT Staging Test", current.Name)
		}
		if current.MaxUploadSizeBytes != 10_485_760 {
			t.Errorf("got max upload %d; want 10485760", current.MaxUploadSizeBytes)
		}
		if current.DefaultLetterNumberPattern != "SK/{YEAR}/{MONTH}/{SEQ}" {
			t.Errorf("got letter pattern %q; want SK/{YEAR}/{MONTH}/{SEQ}", current.DefaultLetterNumberPattern)
		}
	})

	t.Run("update settings and assert audit", func(t *testing.T) {
		name := "RT Staging Baru"
		address := "Jl. Staging No. 10"
		timezone := "Asia/Jayapura"
		maxUpload := int64(20_971_520)
		pattern := "RT-05/{SEQ}/{YEAR}"

		updated, err := service.Update(ctx, principal, settings.UpdateOrganizationSettingsRequest{
			Name:                       &name,
			Address:                    &address,
			Timezone:                   &timezone,
			MaxUploadSizeBytes:         &maxUpload,
			DefaultLetterNumberPattern: &pattern,
		}, "test-req-id")
		if err != nil {
			t.Fatalf("Update(): %v", err)
		}
		if updated.Name != name || updated.Address == nil || *updated.Address != address ||
			updated.Timezone != timezone || updated.MaxUploadSizeBytes != maxUpload ||
			updated.DefaultLetterNumberPattern != pattern {
			t.Errorf("updated values mismatch: %+v", updated)
		}

		var before, after []byte
		var requestID string
		err = pool.QueryRow(ctx, `
			SELECT before_data, after_data, request_id
			FROM audit_logs
			WHERE organization_id = $1 AND action = 'organization.update'
			ORDER BY id DESC LIMIT 1`, orgID,
		).Scan(&before, &after, &requestID)
		if err != nil {
			t.Fatalf("query audit log: %v", err)
		}
		if requestID != "test-req-id" || len(before) == 0 || len(after) == 0 {
			t.Errorf("invalid audit data: request_id=%q before=%s after=%s", requestID, before, after)
		}
	})

	t.Run("audit log is immutable", func(t *testing.T) {
		var logID int64
		if err := pool.QueryRow(ctx, `SELECT id FROM audit_logs WHERE organization_id = $1 ORDER BY id DESC LIMIT 1`, orgID).Scan(&logID); err != nil {
			t.Fatalf("get audit log id: %v", err)
		}

		for _, query := range []string{
			`UPDATE audit_logs SET action = 'tampered' WHERE id = $1`,
			`DELETE FROM audit_logs WHERE id = $1`,
		} {
			if _, err := pool.Exec(ctx, query, logID); err == nil || !strings.Contains(err.Error(), "append-only") {
				t.Errorf("immutable query error = %v; want append-only rejection", err)
			}
		}
	})
}

func testUUID(t *testing.T) string {
	t.Helper()
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		t.Fatalf("generate uuid: %v", err)
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return hex.EncodeToString(bytes[0:4]) + "-" +
		hex.EncodeToString(bytes[4:6]) + "-" +
		hex.EncodeToString(bytes[6:8]) + "-" +
		hex.EncodeToString(bytes[8:10]) + "-" +
		hex.EncodeToString(bytes[10:16])
}