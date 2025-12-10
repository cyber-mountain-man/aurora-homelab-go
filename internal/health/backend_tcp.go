package health

import (
	"net"
	"strconv"
	"time"

	"github.com/cyber-mountain-man/aurora-homelab-go/internal/models"
)

type tcpBackend struct {
	timeout time.Duration
}

func newTCPBackend(timeout time.Duration) Backend {
	return &tcpBackend{
		timeout: timeout,
	}
}

func (b *tcpBackend) Check(svc models.Service) Result {
	res := Result{
		ServiceName: svc.Name,
		Status:      StatusUnknown,
		CheckedAt:   time.Now(),
	}

	if svc.Host == "" || svc.Port == 0 {
		res.Status = StatusDown
		res.Error = "missing host or port for TCP check"
		return res
	}

	// Correct IPv6-safe target formatting
	addr := net.JoinHostPort(svc.Host, strconv.Itoa(svc.Port))

	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, b.timeout)
	if err != nil {
		res.Status = StatusDown
		res.Error = err.Error()
		return res
	}
	_ = conn.Close()

	res.Status = StatusUp
	res.Latency = time.Since(start)
	res.CheckedAt = time.Now()
	res.URL = addr

	return res
}
