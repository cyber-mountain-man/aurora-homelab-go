package health

import "strings"

type ReasonClass string

const (
	ReasonNone       ReasonClass = ""
	ReasonTimeout    ReasonClass = "TIMEOUT"
	ReasonDNS        ReasonClass = "DNS"
	ReasonConn       ReasonClass = "CONN"
	ReasonPermission ReasonClass = "PERMISSION"
	ReasonTLS        ReasonClass = "TLS"
	ReasonHTTP       ReasonClass = "HTTP"
	ReasonUnknown    ReasonClass = "UNKNOWN"
	ReasonOther      ReasonClass = "OTHER"
)

func ClassifyError(err string) ReasonClass {
	e := strings.ToLower(strings.TrimSpace(err))
	if e == "" {
		return ReasonNone
	}

	switch {
	case strings.Contains(e, "context deadline exceeded"),
		strings.Contains(e, "i/o timeout"),
		strings.Contains(e, "timeout"):
		return ReasonTimeout

	case strings.Contains(e, "no such host"),
		strings.Contains(e, "server misbehaving"),
		strings.Contains(e, "dns"):
		return ReasonDNS

	case strings.Contains(e, "connection refused"),
		strings.Contains(e, "network is unreachable"),
		strings.Contains(e, "no route to host"),
		strings.Contains(e, "connection reset"),
		strings.Contains(e, "broken pipe"):
		return ReasonConn

	case strings.Contains(e, "permission denied"),
		strings.Contains(e, "operation not permitted"):
		return ReasonPermission

	case strings.Contains(e, "x509"),
		strings.Contains(e, "tls"):
		return ReasonTLS

	case strings.Contains(e, "status code"),
		strings.Contains(e, "unexpected status"),
		strings.Contains(e, "http response"):
		return ReasonHTTP

	default:
		return ReasonOther
	}
}

func ReasonPresentation(rc ReasonClass) (label string, bulmaColor string) {
	switch rc {
	case ReasonTimeout:
		return "Timeout", "is-warning"
	case ReasonDNS:
		return "DNS", "is-warning"
	case ReasonConn:
		return "Connect", "is-danger"
	case ReasonPermission:
		return "Permission", "is-warning"
	case ReasonTLS:
		return "TLS", "is-warning"
	case ReasonHTTP:
		return "HTTP", "is-warning"
	case ReasonUnknown:
		return "Unknown", "is-warning"
	case ReasonOther:
		return "Other", "is-warning"
	default:
		return "", ""
	}
}
