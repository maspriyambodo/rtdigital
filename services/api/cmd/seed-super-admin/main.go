// Development-only super-admin bootstrap.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

const (
	superAdminUserID = "a0000000-0000-4000-8000-000000000001"
	superAdminOrgID  = "11111111-1111-1111-1111-111111111111"
	superAdminEmail  = "superadmin@rtdigital.local"
	superAdminPass   = "SuperAdmin123!"
	superAdminMFA    = "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP"
)

func main() {
	if os.Getenv("APP_ENV") != "development" {
		log.Fatal("super-admin seeder only runs when APP_ENV=development")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	encryptionKey := os.Getenv("DATA_ENCRYPTION_KEY")
	if len(encryptionKey) != 32 {
		log.Fatal("DATA_ENCRYPTION_KEY must be exactly 32 bytes")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer conn.Close(ctx)

	passwordHash, err := auth.HashPassword(superAdminPass)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	crypter, err := auth.NewAESCrypter([]byte(encryptionKey))
	if err != nil {
		log.Fatalf("initialize MFA encryption: %v", err)
	}

	encryptedMFA, err := crypter.Encrypt(superAdminMFA)
	if err != nil {
		log.Fatalf("encrypt MFA secret: %v", err)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatalf("begin transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		INSERT INTO organizations (id, name, rt_number, rw_number, address, timezone, status)
		VALUES ($1, 'RT Development', '01', '01', 'Data lokal development', 'Asia/Jakarta', 'active')
		ON CONFLICT (id) DO NOTHING`,
		superAdminOrgID,
	); err != nil {
		log.Fatalf("seed organization: %v", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO users (
			id, organization_id, email, password_hash, status,
			failed_login_count, locked_until, mfa_secret_encrypted, mfa_enabled_at
		)
		VALUES ($1, $2, $3, $4, 'active', 0, NULL, $5, now())
		ON CONFLICT (organization_id, lower(email)) WHERE email IS NOT NULL
		DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			status = 'active',
			failed_login_count = 0,
			locked_until = NULL,
			mfa_secret_encrypted = EXCLUDED.mfa_secret_encrypted,
			mfa_enabled_at = now()`,
		superAdminUserID, superAdminOrgID, superAdminEmail, passwordHash, encryptedMFA,
	); err != nil {
		log.Fatalf("seed super admin: %v", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		SELECT $1, id
		FROM roles
		WHERE code = 'super_admin' AND organization_id IS NULL
		ON CONFLICT DO NOTHING`,
		superAdminUserID,
	); err != nil {
		log.Fatalf("assign super admin role: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("commit seed: %v", err)
	}

	fmt.Printf("Super admin seeded: %s\n", superAdminEmail)
}
