package platform

import "github.com/maspriyambodo/rtdigital/services/api/internal/auth"

// NewSystemPrincipal represents a non-user actor for tenant-scoped automation.
// Its empty UserID must be persisted as NULL by audit writers.
func NewSystemPrincipal(organizationID string) *auth.Principal {
	return &auth.Principal{
		OrganizationID: organizationID,
		RoleCodes:      []string{"system"},
		Permissions: map[string]struct{}{
			"invoice.create": {},
			"invoice.read":   {},
		},
	}
}
