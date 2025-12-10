package health

import (
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/cyber-mountain-man/aurora-homelab-go/internal/models"
)

// Status represents the health status of a service.
type Status string

const (
	StatusUnknown Status = "UNKNOWN"
	StatusUp      Status = "UP"
	StatusDown    Status = "DOWN"
)

// Result holds the outcome of a single health check.
type Result struct {
	ServiceName string
	URL         string
	Status      Status
	Latency     time.Duration
	CheckedAt   time.Time
	Error       string // optional: last error message
}

// Checker periodically checks the health of configured services.
type Checker struct {
	mu       sync.RWMutex
	results  map[string]Result
	services []models.Service
	client   *http.Client
	interval time.Duration
}

// NewChecker creates a new Checker.
// interval: how often to run checks (e.g., 30s)
// timeout: HTTP timeout per request (e.g., 3s)
func NewChecker(services []models.Service, interval, timeout time.Duration) *Checker {
	return &Checker{
		results:  make(map[string]Result),
		services: services,
		client: &http.Client{
			Timeout: timeout,
		},
		interval: interval,
	}
}

// Start begins periodic health checks in a background goroutine.
// It also runs one initial check immediately.
func (c *Checker) Start() {
	// Run an initial round so we don't start with all UNKNOWN.
	c.checkAll()

	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for range ticker.C {
			c.checkAll()
		}
	}()
}

// checkAll launches a check for each service.
func (c *Checker) checkAll() {
	for _, svc := range c.services {
		// Launch each check in its own goroutine.
		go c.checkOne(svc)
	}
}

// checkOne performs a single HTTP health check for a service.
func (c *Checker) checkOne(svc models.Service) {
	start := time.Now()

	res := Result{
		ServiceName: svc.Name,
		URL:         svc.URL,
		Status:      StatusUnknown,
		CheckedAt:   time.Now(),
	}

	resp, err := c.client.Get(svc.URL)
	if err != nil {
		res.Status = StatusDown
		res.Error = err.Error()
		c.storeResult(res)
		return
	}
	defer resp.Body.Close()
	// Drain the body so the connection can be reused.
	_, _ = io.Copy(io.Discard, resp.Body)

	res.Latency = time.Since(start)
	res.CheckedAt = time.Now()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		res.Status = StatusUp
	} else {
		res.Status = StatusDown
		res.Error = resp.Status
	}

	c.storeResult(res)
}

// storeResult safely writes a Result into the map.
func (c *Checker) storeResult(res Result) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.results[res.ServiceName] = res
}

// Snapshot returns a copy of the last known results map.
// The copy avoids data races in callers.
func (c *Checker) Snapshot() map[string]Result {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make(map[string]Result, len(c.results))
	for k, v := range c.results {
		out[k] = v
	}
	return out
}
