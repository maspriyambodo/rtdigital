package communication

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
	"github.com/maspriyambodo/rtdigital/services/api/internal/notifications"
)

type Service struct {
	db         *pgxpool.Pool
	dispatcher *notifications.Dispatcher
	now        func() time.Time
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) SetNotificationDispatcher(dispatcher *notifications.Dispatcher) {
	s.dispatcher = dispatcher
}

func (s *Service) notifyImportantAnnouncement(ctx context.Context, organizationID, announcementID, priority, title, content string) {
	if s.dispatcher == nil || (priority != "important" && priority != "urgent") {
		return
	}

	rows, err := s.db.Query(ctx, `
		SELECT DISTINCT u.id
		FROM users u
		WHERE u.organization_id = $1
		  AND u.status = 'active'
		  AND `+userMatchesAnnouncementSQL("u"),
		organizationID,
	)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var userID string
		if rows.Scan(&userID) == nil {
			s.dispatcher.Dispatch(notifications.DispatchJob{
				OrganizationID:  organizationID,
				RecipientUserID: userID,
				Type:            "announcement_published",
				Title:           "Pengumuman penting: " + title,
				Body:            content,
				ReferenceType:   "announcement",
				ReferenceID:     announcementID,
			})
		}
	}
}

func (s *Service) CreateAnnouncement(ctx context.Context, principal *auth.Principal, req CreateAnnouncementRequest) (AnnouncementItem, error) {
	if principal == nil {
		return AnnouncementItem{}, ErrForbidden
	}
	if err := req.Validate(); err != nil {
		return AnnouncementItem{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return AnnouncementItem{}, fmt.Errorf("begin announcement: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	id := newUUID()
	publishAt := s.resolvePublishAt(req.Status, req.PublishAt)
	if _, err := tx.Exec(ctx, `
		INSERT INTO announcements (
			id, organization_id, author_user_id, title, content, category, priority, publish_at, expire_at, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		id, principal.OrganizationID, principal.UserID, req.Title, req.Content, req.Category,
		req.Priority, publishAt, req.ExpireAt, req.Status,
	); err != nil {
		return AnnouncementItem{}, fmt.Errorf("insert announcement: %w", err)
	}
	if err := s.syncAnnouncementTargets(ctx, tx, principal.OrganizationID, id, req.Targets); err != nil {
		return AnnouncementItem{}, err
	}
	if err := s.syncAttachments(ctx, tx, principal, id, "announcement", req.AttachmentFileIDs); err != nil {
		return AnnouncementItem{}, err
	}
	if err := s.audit(ctx, tx, principal, "announcement.create", "announcement", id); err != nil {
		return AnnouncementItem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return AnnouncementItem{}, fmt.Errorf("commit announcement: %w", err)
	}
	if req.Status == "published" {
		s.notifyImportantAnnouncement(ctx, principal.OrganizationID, id, req.Priority, req.Title, req.Content)
	}
	return s.GetAnnouncement(ctx, principal, id)
}

func (s *Service) UpdateAnnouncement(ctx context.Context, principal *auth.Principal, id string, req UpdateAnnouncementRequest) (AnnouncementItem, error) {
	if principal == nil {
		return AnnouncementItem{}, ErrForbidden
	}
	if err := req.Validate(); err != nil {
		return AnnouncementItem{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return AnnouncementItem{}, fmt.Errorf("begin update announcement: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var currentStatus string
	var existingPublishAt *time.Time
	if err := tx.QueryRow(ctx, `
		SELECT status, publish_at FROM announcements
		WHERE organization_id = $1 AND id = $2 FOR UPDATE`,
		principal.OrganizationID, id,
	).Scan(&currentStatus, &existingPublishAt); errors.Is(err, pgx.ErrNoRows) {
		return AnnouncementItem{}, ErrAnnouncementNotFound
	} else if err != nil {
		return AnnouncementItem{}, fmt.Errorf("lock announcement: %w", err)
	}
	if currentStatus == "archived" {
		return AnnouncementItem{}, ErrInvalidState
	}

	publishAt := s.resolvePublishAt(req.Status, req.PublishAt)
	if req.Status == "published" && currentStatus == "published" {
		publishAt = existingPublishAt
	}
	if _, err := tx.Exec(ctx, `
		UPDATE announcements
		SET title = $1, content = $2, category = $3, priority = $4,
		    publish_at = $5, expire_at = $6, status = $7
		WHERE organization_id = $8 AND id = $9`,
		req.Title, req.Content, req.Category, req.Priority, publishAt, req.ExpireAt,
		req.Status, principal.OrganizationID, id,
	); err != nil {
		return AnnouncementItem{}, fmt.Errorf("update announcement: %w", err)
	}
	if err := s.syncAnnouncementTargets(ctx, tx, principal.OrganizationID, id, req.Targets); err != nil {
		return AnnouncementItem{}, err
	}
	if err := s.syncAttachments(ctx, tx, principal, id, "announcement", req.AttachmentFileIDs); err != nil {
		return AnnouncementItem{}, err
	}
	if err := s.audit(ctx, tx, principal, "announcement.update", "announcement", id); err != nil {
		return AnnouncementItem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return AnnouncementItem{}, fmt.Errorf("commit update announcement: %w", err)
	}
	if req.Status == "published" && currentStatus != "published" {
		s.notifyImportantAnnouncement(ctx, principal.OrganizationID, id, req.Priority, req.Title, req.Content)
	}
	return s.GetAnnouncement(ctx, principal, id)
}

func (s *Service) PublishAnnouncement(ctx context.Context, principal *auth.Principal, id string) (AnnouncementItem, error) {
	if principal == nil {
		return AnnouncementItem{}, ErrForbidden
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return AnnouncementItem{}, fmt.Errorf("begin publish announcement: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status string
	if err := tx.QueryRow(ctx, `
		SELECT status FROM announcements WHERE organization_id = $1 AND id = $2 FOR UPDATE`,
		principal.OrganizationID, id,
	).Scan(&status); errors.Is(err, pgx.ErrNoRows) {
		return AnnouncementItem{}, ErrAnnouncementNotFound
	} else if err != nil {
		return AnnouncementItem{}, fmt.Errorf("lock announcement: %w", err)
	}
	if status != "draft" && status != "scheduled" {
		return AnnouncementItem{}, ErrInvalidState
	}
	if _, err := tx.Exec(ctx, `
		UPDATE announcements SET status = 'published', publish_at = $1
		WHERE organization_id = $2 AND id = $3`,
		s.now(), principal.OrganizationID, id,
	); err != nil {
		return AnnouncementItem{}, fmt.Errorf("publish announcement: %w", err)
	}
	if err := s.audit(ctx, tx, principal, "announcement.publish", "announcement", id); err != nil {
		return AnnouncementItem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return AnnouncementItem{}, fmt.Errorf("commit publish announcement: %w", err)
	}
	item, err := s.GetAnnouncement(ctx, principal, id)
	if err == nil {
		s.notifyImportantAnnouncement(ctx, principal.OrganizationID, id, item.Priority, item.Title, item.Content)
	}
	return item, err
}

func (s *Service) ArchiveAnnouncement(ctx context.Context, principal *auth.Principal, id string) (AnnouncementItem, error) {
	if principal == nil {
		return AnnouncementItem{}, ErrForbidden
	}
	result, err := s.db.Exec(ctx, `
		UPDATE announcements SET status = 'archived'
		WHERE organization_id = $1 AND id = $2 AND status <> 'archived'`,
		principal.OrganizationID, id,
	)
	if err != nil {
		return AnnouncementItem{}, fmt.Errorf("archive announcement: %w", err)
	}
	if result.RowsAffected() == 0 {
		return AnnouncementItem{}, ErrAnnouncementNotFound
	}
	if _, err := s.db.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, entity_id)
		VALUES ($1, $2, 'announcement.archive', 'announcement', $3)`,
		principal.OrganizationID, principal.UserID, id,
	); err != nil {
		return AnnouncementItem{}, fmt.Errorf("audit archive announcement: %w", err)
	}
	return s.GetAnnouncement(ctx, principal, id)
}

func (s *Service) ListAnnouncements(ctx context.Context, principal *auth.Principal, filter AnnouncementFilter) ([]AnnouncementItem, error) {
	if principal == nil {
		return nil, ErrForbidden
	}
	isManager := principal.HasPermission("announcement.create")
	now := s.now()
	rows, err := s.db.Query(ctx, `
		SELECT a.id
		FROM announcements a
		WHERE a.organization_id = $1
		  AND ($2 OR (
		    a.status = 'published'
		    AND (a.publish_at IS NULL OR a.publish_at <= $3)
		    AND (a.expire_at IS NULL OR a.expire_at > $3)
		    AND `+announcementVisibleSQL("a")+`
		  ))
		  AND ($4 = '' OR a.status = $4)
		  AND ($5 = '' OR a.category = $5)
		  AND ($6 = '' OR a.priority = $6)
		  AND ($7 = '' OR a.title ILIKE '%' || $7 || '%' OR a.content ILIKE '%' || $7 || '%')
		ORDER BY COALESCE(a.publish_at, a.created_at) DESC
		LIMIT 100`,
		principal.OrganizationID, isManager, now, strings.TrimSpace(filter.Status),
		strings.TrimSpace(filter.Category), strings.TrimSpace(filter.Priority), strings.TrimSpace(filter.Search),
		principal.UserID,
	)
	if err != nil {
		return nil, fmt.Errorf("list announcements: %w", err)
	}
	defer rows.Close()

	items := make([]AnnouncementItem, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan announcement id: %w", err)
		}
		item, err := s.getAnnouncement(ctx, principal, id, isManager)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) GetAnnouncement(ctx context.Context, principal *auth.Principal, id string) (AnnouncementItem, error) {
	if principal == nil {
		return AnnouncementItem{}, ErrForbidden
	}
	item, err := s.getAnnouncement(ctx, principal, id, principal.HasPermission("announcement.create"))
	if err != nil {
		return AnnouncementItem{}, err
	}
	if !principal.HasPermission("announcement.create") {
		now := s.now()
		if item.Status != "published" || (item.PublishAt != nil && item.PublishAt.After(now)) ||
			(item.ExpireAt != nil && !item.ExpireAt.After(now)) {
			return AnnouncementItem{}, ErrAnnouncementNotFound
		}
		visible, err := s.isAnnouncementVisible(ctx, principal, id)
		if err != nil {
			return AnnouncementItem{}, err
		}
		if !visible {
			return AnnouncementItem{}, ErrAnnouncementNotFound
		}
		if _, err := s.db.Exec(ctx, `
			INSERT INTO announcement_read_receipts (organization_id, announcement_id, user_id)
			VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
			principal.OrganizationID, id, principal.UserID,
		); err != nil {
			return AnnouncementItem{}, fmt.Errorf("record announcement read: %w", err)
		}
		item.IsRead = true
	}
	return item, nil
}

func (s *Service) GetAnnouncementReadStats(ctx context.Context, principal *auth.Principal, id string) (ReadStats, error) {
	if principal == nil {
		return ReadStats{}, ErrForbidden
	}
	var stats ReadStats
	err := s.db.QueryRow(ctx, `
		SELECT
			(SELECT count(*)::int FROM users u
			 WHERE u.organization_id = a.organization_id AND u.status = 'active'
			   AND `+userMatchesAnnouncementSQL("u")+`),
			(SELECT count(*)::int FROM announcement_read_receipts rr
			 WHERE rr.organization_id = a.organization_id AND rr.announcement_id = a.id)
		FROM announcements a
		WHERE a.organization_id = $1 AND a.id = $2`,
		principal.OrganizationID, id,
	).Scan(&stats.TotalAudience, &stats.ReadCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return ReadStats{}, ErrAnnouncementNotFound
	}
	if err != nil {
		return ReadStats{}, fmt.Errorf("announcement read stats: %w", err)
	}
	return stats, nil
}

func (s *Service) CreateEvent(ctx context.Context, principal *auth.Principal, req CreateEventRequest) (EventItem, error) {
	if principal == nil {
		return EventItem{}, ErrForbidden
	}
	if err := req.Validate(); err != nil {
		return EventItem{}, err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return EventItem{}, fmt.Errorf("begin event: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	id := newUUID()
	if _, err := tx.Exec(ctx, `
		INSERT INTO events (id, organization_id, author_user_id, title, description, location, starts_at, ends_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		id, principal.OrganizationID, principal.UserID, req.Title, req.Description, req.Location,
		req.StartsAt, req.EndsAt, req.Status,
	); err != nil {
		return EventItem{}, fmt.Errorf("insert event: %w", err)
	}
	if err := s.syncAttachments(ctx, tx, principal, id, "event", req.AttachmentFileIDs); err != nil {
		return EventItem{}, err
	}
	if err := s.audit(ctx, tx, principal, "event.create", "event", id); err != nil {
		return EventItem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return EventItem{}, fmt.Errorf("commit event: %w", err)
	}
	return s.GetEvent(ctx, principal, id)
}

func (s *Service) UpdateEvent(ctx context.Context, principal *auth.Principal, id string, req UpdateEventRequest) (EventItem, error) {
	if principal == nil {
		return EventItem{}, ErrForbidden
	}
	if err := req.Validate(); err != nil {
		return EventItem{}, err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return EventItem{}, fmt.Errorf("begin update event: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var status string
	if err := tx.QueryRow(ctx, `
		SELECT status FROM events WHERE organization_id = $1 AND id = $2 FOR UPDATE`,
		principal.OrganizationID, id,
	).Scan(&status); errors.Is(err, pgx.ErrNoRows) {
		return EventItem{}, ErrEventNotFound
	} else if err != nil {
		return EventItem{}, fmt.Errorf("lock event: %w", err)
	}
	if status == "cancelled" || status == "completed" {
		return EventItem{}, ErrInvalidState
	}
	if _, err := tx.Exec(ctx, `
		UPDATE events SET title = $1, description = $2, location = $3, starts_at = $4, ends_at = $5, status = $6
		WHERE organization_id = $7 AND id = $8`,
		req.Title, req.Description, req.Location, req.StartsAt, req.EndsAt, req.Status,
		principal.OrganizationID, id,
	); err != nil {
		return EventItem{}, fmt.Errorf("update event: %w", err)
	}
	if err := s.syncAttachments(ctx, tx, principal, id, "event", req.AttachmentFileIDs); err != nil {
		return EventItem{}, err
	}
	if err := s.audit(ctx, tx, principal, "event.update", "event", id); err != nil {
		return EventItem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return EventItem{}, fmt.Errorf("commit update event: %w", err)
	}
	return s.GetEvent(ctx, principal, id)
}

func (s *Service) CancelEvent(ctx context.Context, principal *auth.Principal, id string) (EventItem, error) {
	if principal == nil {
		return EventItem{}, ErrForbidden
	}
	result, err := s.db.Exec(ctx, `
		UPDATE events SET status = 'cancelled'
		WHERE organization_id = $1 AND id = $2 AND status IN ('planned', 'ongoing')`,
		principal.OrganizationID, id,
	)
	if err != nil {
		return EventItem{}, fmt.Errorf("cancel event: %w", err)
	}
	if result.RowsAffected() == 0 {
		return EventItem{}, ErrEventNotFound
	}
	if _, err := s.db.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, entity_id)
		VALUES ($1, $2, 'event.cancel', 'event', $3)`,
		principal.OrganizationID, principal.UserID, id,
	); err != nil {
		return EventItem{}, fmt.Errorf("audit cancel event: %w", err)
	}
	return s.GetEvent(ctx, principal, id)
}

func (s *Service) ListEvents(ctx context.Context, principal *auth.Principal, filter EventFilter) ([]EventItem, error) {
	if principal == nil {
		return nil, ErrForbidden
	}
	rows, err := s.db.Query(ctx, `
		SELECT id FROM events
		WHERE organization_id = $1
		  AND ($2 = '' OR status = $2)
		  AND (NOT $3 OR (starts_at >= $4 AND status IN ('planned', 'ongoing')))
		  AND ($5 = '' OR title ILIKE '%' || $5 || '%' OR description ILIKE '%' || $5 || '%' OR location ILIKE '%' || $5 || '%')
		ORDER BY starts_at ASC
		LIMIT 100`,
		principal.OrganizationID, strings.TrimSpace(filter.Status), filter.Upcoming, s.now(), strings.TrimSpace(filter.Search),
	)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()
	items := make([]EventItem, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan event id: %w", err)
		}
		item, err := s.GetEvent(ctx, principal, id)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) GetEvent(ctx context.Context, principal *auth.Principal, id string) (EventItem, error) {
	if principal == nil {
		return EventItem{}, ErrForbidden
	}
	var item EventItem
	err := s.db.QueryRow(ctx, `
		SELECT e.id, e.title, e.description, e.location, e.starts_at, e.ends_at, e.status,
		       e.author_user_id, COALESCE(r.full_name, u.email, u.phone, 'Pengurus'), e.created_at, e.updated_at
		FROM events e
		JOIN users u ON u.organization_id = e.organization_id AND u.id = e.author_user_id
		LEFT JOIN residents r ON r.organization_id = u.organization_id AND r.id = u.resident_id
		WHERE e.organization_id = $1 AND e.id = $2`,
		principal.OrganizationID, id,
	).Scan(&item.ID, &item.Title, &item.Description, &item.Location, &item.StartsAt, &item.EndsAt,
		&item.Status, &item.AuthorUserID, &item.AuthorName, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return EventItem{}, ErrEventNotFound
	}
	if err != nil {
		return EventItem{}, fmt.Errorf("get event: %w", err)
	}
	item.Attachments, err = s.attachments(ctx, principal.OrganizationID, "event", id)
	return item, err
}

func (s *Service) getAnnouncement(ctx context.Context, principal *auth.Principal, id string, _ bool) (AnnouncementItem, error) {
	var item AnnouncementItem
	err := s.db.QueryRow(ctx, `
		SELECT a.id, a.title, a.content, a.category, a.priority, a.publish_at, a.expire_at, a.status,
		       a.author_user_id, COALESCE(r.full_name, u.email, u.phone, 'Pengurus'), a.created_at, a.updated_at,
		       EXISTS(SELECT 1 FROM announcement_read_receipts rr
		              WHERE rr.organization_id = a.organization_id AND rr.announcement_id = a.id AND rr.user_id = $3),
		       (SELECT count(*)::int FROM announcement_read_receipts rr
		        WHERE rr.organization_id = a.organization_id AND rr.announcement_id = a.id)
		FROM announcements a
		JOIN users u ON u.organization_id = a.organization_id AND u.id = a.author_user_id
		LEFT JOIN residents r ON r.organization_id = u.organization_id AND r.id = u.resident_id
		WHERE a.organization_id = $1 AND a.id = $2`,
		principal.OrganizationID, id, principal.UserID,
	).Scan(&item.ID, &item.Title, &item.Content, &item.Category, &item.Priority, &item.PublishAt, &item.ExpireAt,
		&item.Status, &item.AuthorUserID, &item.AuthorName, &item.CreatedAt, &item.UpdatedAt, &item.IsRead, &item.ReadCount)
	if errors.Is(err, pgx.ErrNoRows) {
		return AnnouncementItem{}, ErrAnnouncementNotFound
	}
	if err != nil {
		return AnnouncementItem{}, fmt.Errorf("get announcement: %w", err)
	}
	var targetsErr error
	item.Targets, targetsErr = s.targets(ctx, principal.OrganizationID, id)
	if targetsErr != nil {
		return AnnouncementItem{}, targetsErr
	}
	item.Attachments, err = s.attachments(ctx, principal.OrganizationID, "announcement", id)
	return item, err
}

func (s *Service) syncAnnouncementTargets(ctx context.Context, tx pgx.Tx, organizationID, announcementID string, targets []TargetInput) error {
	if _, err := tx.Exec(ctx, `DELETE FROM announcement_targets WHERE organization_id = $1 AND announcement_id = $2`, organizationID, announcementID); err != nil {
		return fmt.Errorf("clear announcement targets: %w", err)
	}
	for _, target := range targets {
		if err := validateTarget(ctx, tx, organizationID, target); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO announcement_targets (id, organization_id, announcement_id, target_type, target_id)
			VALUES ($1, $2, $3, $4, $5)`,
			newUUID(), organizationID, announcementID, target.TargetType, target.TargetID,
		); err != nil {
			return fmt.Errorf("insert announcement target: %w", err)
		}
	}
	return nil
}

func validateTarget(ctx context.Context, tx pgx.Tx, organizationID string, target TargetInput) error {
	if target.TargetType == "all" {
		return nil
	}
	var found bool
	var err error
	switch target.TargetType {
	case "role":
		err = tx.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM roles WHERE id = $1 AND (organization_id = $2 OR organization_id IS NULL))`,
			target.TargetID, organizationID).Scan(&found)
	case "household":
		err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM households WHERE id = $1 AND organization_id = $2)`, target.TargetID, organizationID).Scan(&found)
	case "house_unit":
		err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM house_units WHERE id = $1 AND organization_id = $2)`, target.TargetID, organizationID).Scan(&found)
	default:
		return ErrValidation
	}
	if err != nil {
		return fmt.Errorf("validate announcement target: %w", err)
	}
	if !found {
		return ErrValidation
	}
	return nil
}

func (s *Service) syncAttachments(ctx context.Context, tx pgx.Tx, principal *auth.Principal, entityID, entityType string, fileIDs []string) error {
	if _, err := tx.Exec(ctx, `
		DELETE FROM file_attachments
		WHERE organization_id = $1 AND entity_type = $2 AND entity_id = $3`,
		principal.OrganizationID, entityType, entityID,
	); err != nil {
		return fmt.Errorf("clear attachments: %w", err)
	}
	seen := make(map[string]struct{}, len(fileIDs))
	for _, fileID := range fileIDs {
		if _, ok := seen[fileID]; ok {
			return ErrValidation
		}
		seen[fileID] = struct{}{}
		var exists bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM file_objects
				WHERE id = $1 AND organization_id = $2 AND uploaded_by = $3
				  AND confirmed_at IS NOT NULL AND deleted_at IS NULL
			)`, fileID, principal.OrganizationID, principal.UserID).Scan(&exists); err != nil {
			return fmt.Errorf("validate attachment: %w", err)
		}
		if !exists {
			return ErrValidation
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO file_attachments (id, organization_id, file_id, entity_type, entity_id, purpose)
			VALUES ($1, $2, $3, $4, $5, 'attachment')`,
			newUUID(), principal.OrganizationID, fileID, entityType, entityID,
		); err != nil {
			return fmt.Errorf("attach file: %w", err)
		}
	}
	return nil
}

func (s *Service) targets(ctx context.Context, organizationID, announcementID string) ([]AnnouncementTargetInfo, error) {
	rows, err := s.db.Query(ctx, `
		SELECT target_type, target_id FROM announcement_targets
		WHERE organization_id = $1 AND announcement_id = $2 ORDER BY created_at`,
		organizationID, announcementID,
	)
	if err != nil {
		return nil, fmt.Errorf("announcement targets: %w", err)
	}
	defer rows.Close()
	items := make([]AnnouncementTargetInfo, 0)
	for rows.Next() {
		var target AnnouncementTargetInfo
		if err := rows.Scan(&target.TargetType, &target.TargetID); err != nil {
			return nil, fmt.Errorf("scan announcement target: %w", err)
		}
		items = append(items, target)
	}
	return items, rows.Err()
}

func (s *Service) attachments(ctx context.Context, organizationID, entityType, entityID string) ([]AttachmentInfo, error) {
	rows, err := s.db.Query(ctx, `
		SELECT a.id, a.file_id, f.original_name, f.mime_type, f.size_bytes, a.purpose
		FROM file_attachments a
		JOIN file_objects f ON f.organization_id = a.organization_id AND f.id = a.file_id
		WHERE a.organization_id = $1 AND a.entity_type = $2 AND a.entity_id = $3
		  AND f.deleted_at IS NULL AND f.confirmed_at IS NOT NULL
		ORDER BY a.created_at`,
		organizationID, entityType, entityID,
	)
	if err != nil {
		return nil, fmt.Errorf("attachments: %w", err)
	}
	defer rows.Close()
	items := make([]AttachmentInfo, 0)
	for rows.Next() {
		var item AttachmentInfo
		if err := rows.Scan(&item.AttachmentID, &item.FileID, &item.OriginalName, &item.MIMEType, &item.SizeBytes, &item.Purpose); err != nil {
			return nil, fmt.Errorf("scan attachment: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Service) isAnnouncementVisible(ctx context.Context, principal *auth.Principal, id string) (bool, error) {
	var visible bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM announcements a
			WHERE a.organization_id = $1 AND a.id = $2 AND (
				EXISTS (SELECT 1 FROM announcement_targets t
					WHERE t.organization_id = a.organization_id AND t.announcement_id = a.id AND t.target_type = 'all')
				OR EXISTS (SELECT 1 FROM announcement_targets t JOIN user_roles ur ON ur.role_id = t.target_id
					WHERE t.organization_id = a.organization_id AND t.announcement_id = a.id
					  AND t.target_type = 'role' AND ur.user_id = $3)
				OR EXISTS (SELECT 1 FROM announcement_targets t
					JOIN households h ON h.organization_id = t.organization_id AND h.id = t.target_id
					JOIN household_members hm ON hm.household_id = h.id AND hm.is_active
					JOIN users u ON u.organization_id = hm.organization_id AND u.resident_id = hm.resident_id
					WHERE t.organization_id = a.organization_id AND t.announcement_id = a.id
					  AND t.target_type = 'household' AND u.id = $3)
				OR EXISTS (SELECT 1 FROM announcement_targets t
					JOIN house_units hu ON hu.organization_id = t.organization_id AND hu.id = t.target_id
					JOIN households h ON h.organization_id = hu.organization_id AND h.house_unit_id = hu.id
					JOIN household_members hm ON hm.household_id = h.id AND hm.is_active
					JOIN users u ON u.organization_id = hm.organization_id AND u.resident_id = hm.resident_id
					WHERE t.organization_id = a.organization_id AND t.announcement_id = a.id
					  AND t.target_type = 'house_unit' AND u.id = $3)
			)
		)`,
		principal.OrganizationID, id, principal.UserID,
	).Scan(&visible)
	if err != nil {
		return false, fmt.Errorf("check announcement visibility: %w", err)
	}
	return visible, nil
}

func announcementVisibleSQL(alias string) string {
	return `(
		EXISTS (SELECT 1 FROM announcement_targets t
			WHERE t.organization_id = ` + alias + `.organization_id AND t.announcement_id = ` + alias + `.id AND t.target_type = 'all')
		OR EXISTS (SELECT 1 FROM announcement_targets t JOIN user_roles ur ON ur.role_id = t.target_id
			WHERE t.organization_id = ` + alias + `.organization_id AND t.announcement_id = ` + alias + `.id
			  AND t.target_type = 'role' AND ur.user_id = $8)
		OR EXISTS (SELECT 1 FROM announcement_targets t
			JOIN households h ON h.organization_id = t.organization_id AND h.id = t.target_id
			JOIN household_members hm ON hm.household_id = h.id AND hm.is_active
			JOIN users u ON u.organization_id = hm.organization_id AND u.resident_id = hm.resident_id
			WHERE t.organization_id = ` + alias + `.organization_id AND t.announcement_id = ` + alias + `.id
			  AND t.target_type = 'household' AND u.id = $8)
		OR EXISTS (SELECT 1 FROM announcement_targets t
			JOIN house_units hu ON hu.organization_id = t.organization_id AND hu.id = t.target_id
			JOIN households h ON h.organization_id = hu.organization_id AND h.house_unit_id = hu.id
			JOIN household_members hm ON hm.household_id = h.id AND hm.is_active
			JOIN users u ON u.organization_id = hm.organization_id AND u.resident_id = hm.resident_id
			WHERE t.organization_id = ` + alias + `.organization_id AND t.announcement_id = ` + alias + `.id
			  AND t.target_type = 'house_unit' AND u.id = $8)
	)`
}

func userMatchesAnnouncementSQL(alias string) string {
	return `EXISTS (
		SELECT 1 FROM announcement_targets t
		WHERE t.organization_id = ` + alias + `.organization_id AND t.announcement_id = a.id
		  AND (
		    t.target_type = 'all'
		    OR (t.target_type = 'role' AND EXISTS (
		        SELECT 1 FROM user_roles ur
		        WHERE ur.role_id = t.target_id AND ur.user_id = ` + alias + `.id
		    ))
		    OR (t.target_type = 'household' AND EXISTS (
		        SELECT 1 FROM household_members hm
		        WHERE hm.household_id = t.target_id AND hm.is_active AND hm.resident_id = ` + alias + `.resident_id
		    ))
		    OR (t.target_type = 'house_unit' AND EXISTS (
		        SELECT 1 FROM house_units hu
		        JOIN households h ON h.organization_id = hu.organization_id AND h.house_unit_id = hu.id
		        JOIN household_members hm ON hm.household_id = h.id AND hm.is_active
		        WHERE hu.organization_id = t.organization_id
		          AND hu.id = t.target_id
		          AND hm.resident_id = ` + alias + `.resident_id
		    ))
		  )
	)`
}

func (s *Service) resolvePublishAt(status string, requested *time.Time) *time.Time {
	if status == "published" {
		now := s.now()
		return &now
	}
	return requested
}

func (s *Service) audit(ctx context.Context, tx pgx.Tx, principal *auth.Principal, action, entityType, entityID string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (organization_id, actor_user_id, action, entity_type, entity_id)
		VALUES ($1, $2, $3, $4, $5)`,
		principal.OrganizationID, principal.UserID, action, entityType, entityID,
	)
	if err != nil {
		return fmt.Errorf("audit %s: %w", action, err)
	}
	return nil
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
