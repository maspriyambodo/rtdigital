package audit_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/maspriyambodo/rtdigital/services/api/internal/audit"
	"github.com/maspriyambodo/rtdigital/services/api/internal/platform"
)

func TestAuditIntegration(t *testing.T) {
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
	otherOrgID := testUUID(t)
	actorID := testUUID(t)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, actorID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM organizations WHERE id IN ($1, $2)`, orgID, otherOrgID)
	})

	for _, id := range []string{orgID, otherOrgID} {
		if _, err := pool.Exec(ctx, `
			INSERT INTO organizations (id, name, rt_number, rw_number, status)
			VALUES ($1, 'RT Audit Test', '005', '005', 'active')`, id,
		); err != nil {
			t.Fatalf("insert organization: %v", err)
		}
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, status)
		VALUES ($1, $2, 'audit-actor@example.test', 'hash', 'active')`,
		actorID, orgID,
	); err != nil {
		t.Fatalf("insert actor: %v", err)
	}

	for range 3 {
		if _, err := pool.Exec(ctx, `
			INSERT INTO audit_logs (
				organization_id, actor_user_id, actor_role_codes, action, entity_type,
				metadata, before_data, after_data, request_id
			)
			VALUES ($1, $2, '["ketua_rt"]'::jsonb, 'resident.read_sensitive', 'residents',
			        '{"reason":"bantuan","secret_token":"leak"}'::jsonb,
			        '{"national_id":"317123"}'::jsonb,
			        '{"nested":{"mfa_secret":"secret","safe":true}}'::jsonb,
			        'req-test')`,
			orgID, actorID,
		); err != nil {
			t.Fatalf("insert audit log: %v", err)
		}
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, action, entity_type)
		VALUES ($1, 'other.action', 'other')`, otherOrgID,
	); err != nil {
		t.Fatalf("insert other tenant log: %v", err)
	}

	service := audit.NewService(pool)

	t.Run("tenant isolation and keyset pagination", func(t *testing.T) {
		first, err := service.List(ctx, orgID, audit.Filter{Limit: 2})
		if err != nil {
			t.Fatalf("List(): %v", err)
		}
		if len(first.Data) != 2 || !first.Meta.HasMore || first.Meta.NextCursor == nil {
			t.Fatalf("unexpected first page: %+v", first.Meta)
		}

		second, err := service.List(ctx, orgID, audit.Filter{
			Limit:  2,
			Cursor: *first.Meta.NextCursor,
		})
		if err != nil {
			t.Fatalf("List(page 2): %v", err)
		}
		if len(second.Data) != 1 || second.Meta.HasMore {
			t.Fatalf("unexpected second page: data=%d meta=%+v", len(second.Data), second.Meta)
		}
		if second.Data[0].Action != "resident.read_sensitive" {
			t.Errorf("tenant isolation failed: action=%q", second.Data[0].Action)
		}
	})

	t.Run("sensitive fields are redacted recursively", func(t *testing.T) {
		result, err := service.List(ctx, orgID, audit.Filter{Limit: 1})
		if err != nil || len(result.Data) != 1 {
			t.Fatalf("List(): result=%+v err=%v", result.Meta, err)
		}

		item := result.Data[0]
		for _, leaked := range []string{"leak", "317123", `"secret"`} {
			if strings.Contains(string(item.Metadata), leaked) ||
				strings.Contains(string(item.BeforeData), leaked) ||
				strings.Contains(string(item.AfterData), leaked) {
				t.Errorf("sensitive value leaked: %q", leaked)
			}
		}
		for _, value := range []string{string(item.Metadata), string(item.BeforeData), string(item.AfterData)} {
			if !strings.Contains(value, "[REDACTED]") {
				t.Errorf("redaction marker missing: %s", value)
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