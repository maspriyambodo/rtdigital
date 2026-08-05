package invoices

import (
	"context"
	"fmt"
	"time"

	"github.com/maspriyambodo/rtdigital/services/api/internal/platform"
)

type ScheduledGenerationResult struct {
	Attempted int
	Created   int
	Skipped   int
	Failed    int
}

// RunScheduledGeneration evaluates active, automatic due types once.
// Repeated calls are safe: run keys are period-specific and idempotent.
func (s *Service) RunScheduledGeneration(ctx context.Context) (ScheduledGenerationResult, error) {
	now := s.now().UTC()
	rows, err := s.db.Query(ctx, `
		SELECT organization_id, id, frequency, due_day, generation_lead_days
		FROM due_types
		WHERE status = 'active'
		  AND automatic_generation_enabled
		  AND frequency IN ('monthly', 'quarterly', 'yearly')
		  AND amount IS NOT NULL
		  AND amount > 0
		ORDER BY organization_id, id`)
	if err != nil {
		return ScheduledGenerationResult{}, fmt.Errorf("list automatic due types: %w", err)
	}
	defer rows.Close()

	var result ScheduledGenerationResult
	for rows.Next() {
		var organizationID, dueTypeID, frequency string
		var dueDay *int
		var leadDays int
		if err := rows.Scan(&organizationID, &dueTypeID, &frequency, &dueDay, &leadDays); err != nil {
			return ScheduledGenerationResult{}, fmt.Errorf("scan automatic due type: %w", err)
		}

		periodStart, periodEnd := scheduledPeriod(now, frequency)
		nextPeriodStart, nextPeriodEnd := nextScheduledPeriod(periodStart, frequency)
		if nextPeriodStart.Sub(now) <= time.Duration(leadDays)*24*time.Hour {
			periodStart, periodEnd = nextPeriodStart, nextPeriodEnd
		}
		dueDate := dateWithClampedDay(periodStart, periodEnd, dueDay)
		runKey := fmt.Sprintf("routine:%s:%s:%s", dueTypeID, periodStart.Format("2006-01-02"), periodEnd.Format("2006-01-02"))

		result.Attempted++
		run, err := s.CreateInvoiceGenerationRun(ctx, platform.NewSystemPrincipal(organizationID), runKey, CreateInvoiceGenerationRunRequest{
			DueTypeID:   dueTypeID,
			PeriodStart: periodStart.Format("2006-01-02"),
			PeriodEnd:   periodEnd.Format("2006-01-02"),
			DueDate:     dueDate.Format("2006-01-02"),
		})
		if err != nil {
			result.Failed++
			continue
		}
		result.Created += run.TotalCreated
		result.Skipped += run.TotalSkipped
	}
	if err := rows.Err(); err != nil {
		return ScheduledGenerationResult{}, fmt.Errorf("iterate automatic due types: %w", err)
	}
	return result, nil
}

func scheduledPeriod(now time.Time, frequency string) (time.Time, time.Time) {
	year, month, _ := now.Date()
	location := now.Location()
	switch frequency {
	case "quarterly":
		month = time.Month(((int(month)-1)/3)*3 + 1)
		start := time.Date(year, month, 1, 0, 0, 0, 0, location)
		return start, start.AddDate(0, 3, -1)
	case "yearly":
		start := time.Date(year, time.January, 1, 0, 0, 0, 0, location)
		return start, start.AddDate(1, 0, -1)
	default:
		start := time.Date(year, month, 1, 0, 0, 0, 0, location)
		return start, start.AddDate(0, 1, -1)
	}
}

func nextScheduledPeriod(periodStart time.Time, frequency string) (time.Time, time.Time) {
	switch frequency {
	case "quarterly":
		next := periodStart.AddDate(0, 3, 0)
		return next, next.AddDate(0, 3, -1)
	case "yearly":
		next := periodStart.AddDate(1, 0, 0)
		return next, next.AddDate(1, 0, -1)
	default:
		next := periodStart.AddDate(0, 1, 0)
		return next, next.AddDate(0, 1, -1)
	}
}

func dateWithClampedDay(periodStart, periodEnd time.Time, configuredDay *int) time.Time {
	day := 1
	if configuredDay != nil {
		day = *configuredDay
	}
	if day > periodEnd.Day() {
		day = periodEnd.Day()
	}
	return time.Date(periodStart.Year(), periodStart.Month(), day, 0, 0, 0, 0, periodStart.Location())
}
