package health

import (
	"context"
	"net"
	"time"

	"github.com/cyber-mountain-man/aurora-homelab-go/internal/models"
)

// dnsBackend implements DNS resolution health checks.
type dnsBackend struct {
	timeout  time.Duration
	resolver *net.Resolver
}

func newDNSBackend(timeout time.Duration) Backend {
	return &dnsBackend{
		timeout:  timeout,
		resolver: net.DefaultResolver,
	}
}

func (b *dnsBackend) Check(svc models.Service) Result {
	res := Result{
		ServiceName: svc.Name,
		Status:      StatusUnknown,
		CheckedAt:   time.Now(),
	}

	if svc.Host == "" {
		res.Status = StatusDown
		res.Error = "missing host for DNS check"
		return res
	}

	ctx, cancel := context.WithTimeout(context.Background(), b.timeout)
	defer cancel()

	start := time.Now()
	ips, err := b.resolver.LookupHost(ctx, svc.Host)
	if err != nil {
		res.Status = StatusDown
		res.Error = err.Error()
		return res
	}

	if len(ips) == 0 {
		res.Status = StatusDown
		res.Error = "no DNS records returned"
		return res
	}

	res.Status = StatusUp
	res.Latency = time.Since(start)
	res.CheckedAt = time.Now()
	res.URL = svc.Host // show the hostname in the UI

	return res
}
