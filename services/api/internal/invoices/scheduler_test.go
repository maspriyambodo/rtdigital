package invoices

import (
	"testing"
	"time"
)

func TestScheduledPeriod(t *testing.T) {
	t.Parallel()

	location := time.UTC
	cases := []struct {
		name      string
		now       time.Time
		frequency string
		start     string
		end       string
	}{
		{"monthly", time.Date(2026, time.February, 14, 9, 0, 0, 0, location), "monthly", "2026-02-01", "2026-02-28"},
		{"quarterly", time.Date(2026, time.May, 14, 9, 0, 0, 0, location), "quarterly", "2026-04-01", "2026-06-30"},
		{"yearly", time.Date(2026, time.November, 14, 9, 0, 0, 0, location), "yearly", "2026-01-01", "2026-12-31"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			start, end := scheduledPeriod(test.now, test.frequency)
			if got := start.Format("2006-01-02"); got != test.start {
				t.Fatalf("start = %s, want %s", got, test.start)
			}
			if got := end.Format("2006-01-02"); got != test.end {
				t.Fatalf("end = %s, want %s", got, test.end)
			}
		})
	}
}

func TestNextScheduledPeriod(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
	nextStart, nextEnd := nextScheduledPeriod(start, "quarterly")
	if got := nextStart.Format("2006-01-02"); got != "2027-01-01" {
		t.Fatalf("next quarter start = %s", got)
	}
	if got := nextEnd.Format("2006-01-02"); got != "2027-03-31" {
		t.Fatalf("next quarter end = %s", got)
	}
}

func TestDateWithClampedDay(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, -1)
	day := 31
	if got := dateWithClampedDay(start, end, &day).Format("2006-01-02"); got != "2026-02-28" {
		t.Fatalf("clamped due date = %s, want 2026-02-28", got)
	}
}
