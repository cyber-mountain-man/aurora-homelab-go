package handlers

import (
	"testing"

	"github.com/cyber-mountain-man/aurora-homelab-go/internal/health"
)

func TestBuildSummary_CountsAndTopReason(t *testing.T) {
	views := []ServiceView{
		{Status: string(health.StatusDown), ReasonLabel: "Timeout"},
		{Status: string(health.StatusDown), ReasonLabel: "Timeout"},
		{Status: string(health.StatusDown), ReasonLabel: "DNS"},
		{Status: string(health.StatusUp)},
		{Status: string(health.StatusUnknown)},
		{Status: string(health.StatusUp), IsStale: true}, // stale counts regardless of status
	}

	s := buildSummary(views)

	if s.DownCount != 3 {
		t.Fatalf("DownCount=%d, want 3", s.DownCount)
	}
	if s.StaleCount != 1 {
		t.Fatalf("StaleCount=%d, want 1", s.StaleCount)
	}
	if s.UnknownCount != 1 {
		t.Fatalf("UnknownCount=%d, want 1", s.UnknownCount)
	}

	if s.TopReasonLabel != "Timeout" || s.TopReasonCount != 2 {
		t.Fatalf("TopReason=(%q,%d), want (%q,%d)", s.TopReasonLabel, s.TopReasonCount, "Timeout", 2)
	}

	// Since DownCount > 0, danger banner demonstrate priority behavior.
	if s.SeverityClass != "is-danger" {
		t.Fatalf("SeverityClass=%q, want %q", s.SeverityClass, "is-danger")
	}
}

func TestBuildSummary_TopReasonTieBreak(t *testing.T) {
	views := []ServiceView{
		{Status: string(health.StatusDown), ReasonLabel: "DNS"},
		{Status: string(health.StatusDown), ReasonLabel: "Timeout"},
	}

	s := buildSummary(views)

	// tie: 1 vs 1 -> deterministic tie-break should pick alphabetical label ("DNS")
	if s.TopReasonLabel != "DNS" || s.TopReasonCount != 1 {
		t.Fatalf("TopReason=(%q,%d), want (%q,%d)", s.TopReasonLabel, s.TopReasonCount, "DNS", 1)
	}
}

func TestBuildSummary_SeverityPriority(t *testing.T) {
	tests := []struct {
		name     string
		views    []ServiceView
		wantSev  string
		wantDown int
		wantSt   int
		wantUnk  int
	}{
		{
			name:     "down dominates",
			views:    []ServiceView{{Status: string(health.StatusDown)}},
			wantSev:  "is-danger",
			wantDown: 1,
		},
		{
			name:    "stale when no down",
			views:   []ServiceView{{Status: string(health.StatusUp), IsStale: true}},
			wantSev: "is-warning",
			wantSt:  1,
		},
		{
			name:    "unknown when no down or stale",
			views:   []ServiceView{{Status: string(health.StatusUnknown)}},
			wantSev: "is-dark",
			wantUnk: 1,
		},
		{
			name:    "all up",
			views:   []ServiceView{{Status: string(health.StatusUp)}, {Status: string(health.StatusUp)}},
			wantSev: "is-success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := buildSummary(tt.views)

			if s.SeverityClass != tt.wantSev {
				t.Fatalf("SeverityClass=%q, want %q", s.SeverityClass, tt.wantSev)
			}
			if s.DownCount != tt.wantDown {
				t.Fatalf("DownCount=%d, want %d", s.DownCount, tt.wantDown)
			}
			if s.StaleCount != tt.wantSt {
				t.Fatalf("StaleCount=%d, want %d", s.StaleCount, tt.wantSt)
			}
			if s.UnknownCount != tt.wantUnk {
				t.Fatalf("UnknownCount=%d, want %d", s.UnknownCount, tt.wantUnk)
			}
		})
	}
}
