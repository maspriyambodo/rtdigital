package audit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("audit log not found")

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) List(ctx context.Context, organizationID string, filter Filter) (ListResult, error) {
	limit := filter.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT a.id, a.actor_user_id, r.full_name, a.actor_role_codes, a.action, a.entity_type, a.entity_id,
		       a.metadata, a.before_data, a.after_data, a.request_id, a.created_at
		FROM audit_logs a
		LEFT JOIN users u ON u.id = a.actor_user_id
		LEFT JOIN residents r ON r.id = u.resident_id
		WHERE a.organization_id = $1`
	args := []any{organizationID}
	placeholder := 2

	if filter.Action != "" {
		query += fmt.Sprintf(" AND a.action = $%d", placeholder)
		args = append(args, filter.Action)
		placeholder++
	}
	if filter.ActorUserID != "" {
		query += fmt.Sprintf(" AND a.actor_user_id = $%d", placeholder)
		args = append(args, filter.ActorUserID)
		placeholder++
	}
	if filter.EntityType != "" {
		query += fmt.Sprintf(" AND a.entity_type = $%d", placeholder)
		args = append(args, filter.EntityType)
		placeholder++
	}
	if filter.EntityID != "" {
		query += fmt.Sprintf(" AND a.entity_id = $%d", placeholder)
		args = append(args, filter.EntityID)
		placeholder++
	}
	if filter.Cursor > 0 {
		query += fmt.Sprintf(" AND a.id < $%d", placeholder)
		args = append(args, filter.Cursor)
		placeholder++
	}

	query += fmt.Sprintf(" ORDER BY a.id DESC LIMIT $%d", placeholder)
	args = append(args, limit+1)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return ListResult{Data: []LogItem{}}, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	items := make([]LogItem, 0, limit+1)
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return ListResult{Data: []LogItem{}}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return ListResult{Data: []LogItem{}}, fmt.Errorf("iterate audit logs: %w", err)
	}

	hasMore := len(items) > limit
	var nextCursor *int64
	if hasMore {
		items = items[:limit]
		cursor := items[len(items)-1].ID
		nextCursor = &cursor
	}

	return ListResult{
		Data: items,
		Meta: Meta{NextCursor: nextCursor, HasMore: hasMore},
	}, nil
}

func (s *Service) Get(ctx context.Context, organizationID, id string) (LogItem, error) {
	logID, err := strconv.ParseInt(id, 10, 64)
	if err != nil || logID < 1 {
		return LogItem{}, ErrNotFound
	}

	row := s.db.QueryRow(ctx, `
		SELECT a.id, a.actor_user_id, r.full_name, a.actor_role_codes, a.action, a.entity_type, a.entity_id,
		       a.metadata, a.before_data, a.after_data, a.request_id, a.created_at
		FROM audit_logs a
		LEFT JOIN users u ON u.id = a.actor_user_id
		LEFT JOIN residents r ON r.id = u.resident_id
		WHERE a.organization_id = $1 AND a.id = $2`, organizationID, logID)

	item, err := scanItem(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return LogItem{}, ErrNotFound
	}
	if err != nil {
		return LogItem{}, err
	}
	return item, nil
}

type scanner interface {
	Scan(...any) error
}

func scanItem(row scanner) (LogItem, error) {
	var item LogItem
	var roles, metadata, before, after []byte
	if err := row.Scan(
		&item.ID, &item.ActorUserID, &item.ActorName, &roles, &item.Action, &item.EntityType, &item.EntityID,
		&metadata, &before, &after, &item.RequestID, &item.CreatedAt,
	); err != nil {
		return LogItem{}, fmt.Errorf("scan audit log: %w", err)
	}

	item.ActorRoleCodes = []string{}
	_ = json.Unmarshal(roles, &item.ActorRoleCodes)
	item.Metadata = sanitizeJSON(metadata)
	item.BeforeData = sanitizeJSON(before)
	item.AfterData = sanitizeJSON(after)
	return item, nil
}

func sanitizeJSON(raw []byte) json.RawMessage {
	if len(raw) == 0 || !json.Valid(raw) {
		return json.RawMessage(`{}`)
	}

	var value any
	if json.Unmarshal(raw, &value) != nil {
		return json.RawMessage(`{}`)
	}
	sanitizeValue(value)

	clean, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return clean
}

func sanitizeValue(value any) {
	switch data := value.(type) {
	case map[string]any:
		for key, nested := range data {
			if isSensitiveKey(key) {
				data[key] = "[REDACTED]"
				continue
			}
			sanitizeValue(nested)
		}
	case []any:
		for _, nested := range data {
			sanitizeValue(nested)
		}
	}
}

func isSensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(key, "-", "_"))
	for _, sensitive := range []string{
		"token", "password", "password_hash", "secret", "mfa",
		"national_id", "family_card", "nik", "nomor_kk",
		"refresh_token", "authorization", "cookie", "proof",
	} {
		if normalized == sensitive ||
			strings.HasPrefix(normalized, sensitive+"_") ||
			strings.HasSuffix(normalized, "_"+sensitive) {
			return true
		}
	}
	return false
}