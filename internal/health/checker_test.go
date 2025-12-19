package health

import (
	"testing"
	"time"

	"github.com/cyber-mountain-man/aurora-homelab-go/internal/models"
)

type stubBackend struct {
	res Result
}

func (s stubBackend) Check(svc models.Service) Result {
	r := s.res
	r.ServiceName = svc.Name
	r.CheckedAt = time.Now()
	return r
}

func TestCheckerCheckNowStoresResult(t *testing.T) {
	services := []models.Service{
		{Name: "Google", Type: "http", URL: "https://www.google.com"},
	}

	c := NewChecker(services, 30*time.Second, 3*time.Second, 2*time.Second)

	// Override backend with stub so we don't do real network calls in unit tests.
	c.backends["http"] = stubBackend{
		res: Result{
			Status:  StatusUp,
			Latency: 12 * time.Millisecond,
			Error:   "",
		},
	}

	ok := c.CheckNow("Google")
	if !ok {
		t.Fatalf("CheckNow returned false; expected true")
	}

	snap := c.Snapshot()
	got, exists := snap["Google"]
	if !exists {
		t.Fatalf("expected result for service Google in snapshot")
	}
	if got.Status != StatusUp {
		t.Fatalf("got status %q, want %q", got.Status, StatusUp)
	}
	if got.CheckedAt.IsZero() {
		t.Fatalf("expected CheckedAt to be set")
	}
}

func TestCheckerCheckNowMissingService(t *testing.T) {
	c := NewChecker(nil, 30*time.Second, 3*time.Second, 2*time.Second)
	if ok := c.CheckNow("does-not-exist"); ok {
		t.Fatalf("expected CheckNow to return false for missing service")
	}
}
