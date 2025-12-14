//lint:file-ignore SA1019 temporary for MVP; replace ICMP ping implementation later
package health

import (
	"time"

	"github.com/cyber-mountain-man/aurora-homelab-go/internal/models"

	// NOTE: github.com/go-ping/ping is deprecated (SA1019). Kept it temporarily for MVP ICMP support.
	// TODO(aurora): Replace ICMP ping with TCP-based reachability (no raw socket/capabilities) or x/net/icmp.
	ping "github.com/go-ping/ping"
)

// pingBackend implements ICMP reachability checks using github.com/go-ping/ping.
// It runs in unprivileged mode so the app does not need root.
type pingBackend struct {
	timeout time.Duration
}

func newPingBackend(timeout time.Duration) Backend {
	return &pingBackend{
		timeout: timeout,
	}
}

func (b *pingBackend) Check(svc models.Service) Result {
	res := Result{
		ServiceName: svc.Name,
		Status:      StatusUnknown,
		CheckedAt:   time.Now(),
	}

	if svc.Host == "" {
		res.Status = StatusDown
		res.Error = "missing host for ping check"
		return res
	}

	pinger, err := ping.NewPinger(svc.Host)
	if err != nil {
		res.Status = StatusDown
		res.Error = err.Error()
		return res
	}

	// Unprivileged mode: uses UDP fallback where supported.
	pinger.SetPrivileged(false)
	pinger.Count = 1
	pinger.Timeout = b.timeout

	start := time.Now()
	if err := pinger.Run(); err != nil {
		res.Status = StatusDown
		res.Error = err.Error()
		return res
	}
	stats := pinger.Statistics()

	if stats.PacketsRecv < 1 {
		res.Status = StatusDown
		res.Error = "no ping reply"
		return res
	}

	res.Status = StatusUp
	// You can use stats.AvgRtt, but Count=1 so it's effectively the same as this:
	res.Latency = time.Since(start)
	res.CheckedAt = time.Now()
	res.URL = svc.Host // just show the hostname in the UI

	return res
}
