package health

import (
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
	StatusStale   Status = "STALE"
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

// Backend defines a pluggable health check implementation.
// Different backends can check HTTP, TCP, ICMP, etc.
type Backend interface {
	Check(svc models.Service) Result
}

// Checker periodically checks the health of configured services.
type Checker struct {
	mu       sync.RWMutex
	results  map[string]Result
	services []models.Service

	backends map[string]Backend

	interval time.Duration
}

// NewChecker creates a new Checker.
// interval: how often to run checks (e.g., 30s)
// httpTimeout: HTTP timeout per request (e.g., 3s)
// tcpTimeout: TCP dial timeout (e.g., 2s)
func NewChecker(services []models.Service, interval, httpTimeout, tcpTimeout time.Duration) *Checker {
	backends := map[string]Backend{
		"http": newHTTPBackend(httpTimeout),
		"tcp":  newTCPBackend(tcpTimeout),
		"dns":  newDNSBackend(httpTimeout), // reuse HTTP timeout for DNS
		"ping": newPingBackend(tcpTimeout), // reuse TCP timeout for ping
	}

	return &Checker{
		results:  make(map[string]Result),
		services: services,
		backends: backends,
		interval: interval,
	}
}

// Start begins periodic health checks in a background goroutine.
// It also runs one initial check immediately.
func (c *Checker) Start() {
	c.checkAll()

	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for range ticker.C {
			c.checkAll()
		}
	}()
}

// CheckNow triggers an immediate health check for a single service by name.
// It runs synchronously so the caller can return updated UI right away.
func (c *Checker) CheckNow(name string) bool {
	// Find the service
	for _, svc := range c.services {
		if svc.Name == name {
			backend := c.getBackend(svc.Type)
			res := backend.Check(svc)
			c.storeResult(res)
			return true
		}
	}
	return false
}

// checkAll launches a check for each service.
func (c *Checker) checkAll() {
	for _, svc := range c.services {
		// Launch each check in its own goroutine.
		go c.checkOne(svc)
	}
}

// getBackend returns the backend for a given service type.
// Defaults to HTTP if type is empty or unknown.
func (c *Checker) getBackend(svcType string) Backend {
	if svcType == "" {
		svcType = "http"
	}
	if b, ok := c.backends[svcType]; ok {
		return b
	}
	return c.backends["http"]
}

// checkOne performs a single health check using the appropriate backend.
func (c *Checker) checkOne(svc models.Service) {
	backend := c.getBackend(svc.Type)
	res := backend.Check(svc)
	c.storeResult(res)
}

// storeResult safely writes a Result into the map.
func (c *Checker) storeResult(res Result) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.results[res.ServiceName] = res
}

// Snapshot returns a copy of the last known results map.
func (c *Checker) Snapshot() map[string]Result {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make(map[string]Result, len(c.results))
	for k, v := range c.results {
		out[k] = v
	}
	return out
}

// interval returns the health check interval.
func (c *Checker) Interval() time.Duration {
	return c.interval
}
