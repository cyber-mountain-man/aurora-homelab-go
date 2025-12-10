package health

import (
	"io"
	"net/http"
	"time"

	"github.com/cyber-mountain-man/aurora-homelab-go/internal/models"
)

// httpBackend implements HTTP-based health checks.
type httpBackend struct {
	client *http.Client
}

func newHTTPBackend(timeout time.Duration) Backend {
	return &httpBackend{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (b *httpBackend) Check(svc models.Service) Result {
	start := time.Now()

	res := Result{
		ServiceName: svc.Name,
		URL:         svc.URL,
		Status:      StatusUnknown,
		CheckedAt:   time.Now(),
	}

	if svc.URL == "" {
		res.Status = StatusDown
		res.Error = "missing URL for HTTP check"
		return res
	}

	resp, err := b.client.Get(svc.URL)
	if err != nil {
		res.Status = StatusDown
		res.Error = err.Error()
		return res
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	res.Latency = time.Since(start)
	res.CheckedAt = time.Now()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		res.Status = StatusUp
	} else {
		res.Status = StatusDown
		res.Error = resp.Status
	}

	return res
}
