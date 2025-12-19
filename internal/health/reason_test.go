package health

import "testing"

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected ReasonClass
	}{
		{
			name:     "empty error",
			errMsg:   "",
			expected: ReasonNone,
		},
		{
			name:     "timeout context deadline",
			errMsg:   "context deadline exceeded",
			expected: ReasonTimeout,
		},
		{
			name:     "io timeout",
			errMsg:   "dial tcp: i/o timeout",
			expected: ReasonTimeout,
		},
		{
			name:     "dns no such host",
			errMsg:   "lookup proxmox.local: no such host",
			expected: ReasonDNS,
		},
		{
			name:     "dns server misbehaving",
			errMsg:   "server misbehaving",
			expected: ReasonDNS,
		},
		{
			name:     "connection refused",
			errMsg:   "connect: connection refused",
			expected: ReasonConn,
		},
		{
			name:     "network unreachable",
			errMsg:   "network is unreachable",
			expected: ReasonConn,
		},
		{
			name:     "permission denied",
			errMsg:   "socket: permission denied",
			expected: ReasonPermission,
		},
		{
			name:     "operation not permitted",
			errMsg:   "operation not permitted",
			expected: ReasonPermission,
		},
		{
			name:     "tls x509 error",
			errMsg:   "x509: certificate signed by unknown authority",
			expected: ReasonTLS,
		},
		{
			name:     "http status error",
			errMsg:   "unexpected status code 500",
			expected: ReasonHTTP,
		},
		{
			name:     "unknown error",
			errMsg:   "something completely unexpected happened",
			expected: ReasonOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyError(tt.errMsg)
			if got != tt.expected {
				t.Fatalf("ClassifyError(%q) = %q, want %q",
					tt.errMsg, got, tt.expected)
			}
		})
	}
}

func TestReasonPresentation(t *testing.T) {
	tests := []struct {
		rc        ReasonClass
		wantLabel string
		wantColor string
	}{
		{ReasonTimeout, "Timeout", "is-warning"},
		{ReasonDNS, "DNS", "is-warning"},
		{ReasonConn, "Connect", "is-danger"},
		{ReasonPermission, "Permission", "is-warning"},
		{ReasonTLS, "TLS", "is-warning"},
		{ReasonHTTP, "HTTP", "is-warning"},
		{ReasonUnknown, "Unknown", "is-warning"},
		{ReasonOther, "Other", "is-warning"},
		{ReasonNone, "", ""},
	}

	for _, tt := range tests {
		t.Run(string(tt.rc), func(t *testing.T) {
			label, color := ReasonPresentation(tt.rc)
			if label != tt.wantLabel || color != tt.wantColor {
				t.Fatalf(
					"ReasonPresentation(%q) = (%q, %q), want (%q, %q)",
					tt.rc, label, color, tt.wantLabel, tt.wantColor,
				)
			}
		})
	}
}
