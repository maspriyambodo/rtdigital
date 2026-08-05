package letters

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// VerifyPublicLetter exposes no personal data or private document URL.
func (s *Service) VerifyPublicLetter(ctx context.Context, code string) (PublicLetterVerification, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return PublicLetterVerification{}, ErrValidation
	}

	var item PublicLetterVerification
	err := s.db.QueryRow(ctx, `
		SELECT lr.letter_number, lt.name, lr.issued_at, lr.status
		FROM letter_requests lr
		JOIN letter_types lt
		  ON lt.organization_id = lr.organization_id
		 AND lt.id = lr.letter_type_id
		WHERE lr.public_verification_code = $1
		  AND lr.status IN ('issued', 'cancelled')`,
		code,
	).Scan(&item.LetterNumber, &item.LetterType, &item.IssuedAt, &item.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return PublicLetterVerification{}, ErrLetterRequestNotFound
	}
	if err != nil {
		return PublicLetterVerification{}, fmt.Errorf("verify public letter: %w", err)
	}
	return item, nil
}
