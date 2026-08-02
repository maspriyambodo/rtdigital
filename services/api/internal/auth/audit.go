package auth

import (
	"context"
	"fmt"
)

func (s *Service) writeAudit(
	ctx context.Context,
	organizationID, actorUserID, action, entityType, entityID, metadata, ipAddress, userAgent string,
) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO audit_logs (
			organization_id, actor_user_id, action, entity_type, entity_id,
			metadata, ip_address, user_agent
		)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8)`,
		organizationID,
		actorUserID,
		action,
		entityType,
		entityID,
		metadata,
		validIP(ipAddress),
		userAgent,
	)
	if err != nil {
		return fmt.Errorf("write audit log: %w", err)
	}
	return nil
}
