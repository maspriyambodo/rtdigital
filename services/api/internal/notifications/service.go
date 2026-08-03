package notifications

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

type Service struct {
	db  *pgxpool.Pool
	now func() time.Time
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{
		db:  db,
		now: func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) List(ctx context.Context, principal *auth.Principal, filter Filter) ([]Notification, error) {
	if principal == nil {
		return nil, ErrValidation
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `
		SELECT id, type, title, body, reference_type, reference_id, read_at, created_at
		FROM notifications
		WHERE organization_id = $1
		  AND user_id = $2
	`
	if filter.UnreadOnly {
		query += ` AND read_at IS NULL`
	}
	query += ` ORDER BY created_at DESC LIMIT $3`

	rows, err := s.db.Query(ctx, query, principal.OrganizationID, principal.UserID, limit)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	items := make([]Notification, 0)
	for rows.Next() {
		var item Notification
		if err := rows.Scan(
			&item.ID,
			&item.Type,
			&item.Title,
			&item.Body,
			&item.ReferenceType,
			&item.ReferenceID,
			&item.ReadAt,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notifications: %w", err)
	}

	return items, nil
}

func (s *Service) MarkAsRead(ctx context.Context, principal *auth.Principal, notificationID string) (Notification, error) {
	if principal == nil || notificationID == "" {
		return Notification{}, ErrValidation
	}

	var item Notification
	err := s.db.QueryRow(ctx, `
		UPDATE notifications
		SET read_at = COALESCE(read_at, $1)
		WHERE id = $2
		  AND organization_id = $3
		  AND user_id = $4
		RETURNING id, type, title, body, reference_type, reference_id, read_at, created_at
	`, s.now(), notificationID, principal.OrganizationID, principal.UserID).Scan(
		&item.ID,
		&item.Type,
		&item.Title,
		&item.Body,
		&item.ReferenceType,
		&item.ReferenceID,
		&item.ReadAt,
		&item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Notification{}, ErrNotificationNotFound
	}
	if err != nil {
		return Notification{}, fmt.Errorf("mark notification as read: %w", err)
	}

	return item, nil
}

func (s *Service) MarkAllAsRead(ctx context.Context, principal *auth.Principal) (int64, error) {
	if principal == nil {
		return 0, ErrValidation
	}

	tag, err := s.db.Exec(ctx, `
		UPDATE notifications
		SET read_at = $1
		WHERE organization_id = $2
		  AND user_id = $3
		  AND read_at IS NULL
	`, s.now(), principal.OrganizationID, principal.UserID)
	if err != nil {
		return 0, fmt.Errorf("mark all notifications as read: %w", err)
	}

	return tag.RowsAffected(), nil
}

func (s *Service) Create(ctx context.Context, tx pgx.Tx, organizationID string, params CreateParams) (Notification, error) {
	if tx == nil || organizationID == "" || params.Validate() != nil {
		return Notification{}, ErrValidation
	}

	item := Notification{
		ID:            newUUID(),
		Type:          params.Type,
		Title:         params.Title,
		Body:          params.Body,
		ReferenceType: params.ReferenceType,
		ReferenceID:   params.ReferenceID,
		CreatedAt:     s.now(),
	}
	err := tx.QueryRow(ctx, `
		INSERT INTO notifications (
			id, organization_id, user_id, type, title, body, reference_type, reference_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING read_at
	`, item.ID, organizationID, params.UserID, item.Type, item.Title, item.Body,
		item.ReferenceType, item.ReferenceID, item.CreatedAt,
	).Scan(&item.ReadAt)
	if err != nil {
		return Notification{}, fmt.Errorf("insert notification: %w", err)
	}

	return item, nil
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