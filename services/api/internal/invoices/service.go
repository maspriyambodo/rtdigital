package invoices

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

var (
	ErrValidation           = errors.New("validation failed")
	ErrDueTypeNotFound      = errors.New("due type not found")
	ErrInvoiceNotFound      = errors.New("invoice not found")
	ErrHouseholdNotFound    = errors.New("household not found")
	ErrDuplicateData        = errors.New("duplicate data")
	ErrConstraint           = errors.New("business constraint violated")
	ErrInvalidInvoiceStatus = errors.New("invalid invoice status for this operation")
)

type Service struct {
	db  *pgxpool.Pool
	now func() time.Time
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) audit(ctx context.Context, tx pgx.Tx, principal *auth.Principal, action, entityType, entityID string) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, entity_id)
		VALUES ($1, $2, $3, $4, $5)`,
		principal.OrganizationID, principal.UserID, action, entityType, entityID,
	); err != nil {
		return fmt.Errorf("audit %s: %w", action, err)
	}
	return nil
}

func validFrequency(value string) bool {
	switch value {
	case "once", "monthly", "quarterly", "yearly":
		return true
	}
	return false
}

func validDate(value string) bool {
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}

func nullableTrim(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func mapDatabaseError(err error, operation string) error {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "unique"):
		return ErrDuplicateData
	case strings.Contains(message, "check"), strings.Contains(message, "foreign key"):
		return ErrConstraint
	default:
		return fmt.Errorf("%s: %w", operation, err)
	}
}

func newUUID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		panic("secure random source unavailable")
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}
