package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/cyber-mountain-man/aurora-homelab-go/internal/health"
	"github.com/cyber-mountain-man/aurora-homelab-go/internal/models"
)

// ServiceView is what the template sees: service + health fields.
type ServiceView struct {
	Name        string
	Description string
	Category    string
	URL         string

	Status      string
	StatusClass string
	LatencyMs   int64

	Protocol      string
	ProtocolClass string

	LastError   string
	LastChecked time.Time

	IsStale    bool
	StaleClass string
	StaleLabel string
}

// DashboardHandler holds compiled templates, services, and the health checker.
type DashboardHandler struct {
	tmpl     *template.Template
	services []models.Service
	checker  *health.Checker
}

// NewDashboardHandler parses the HTML templates and returns a handler.
func NewDashboardHandler(templatesDir string, services []models.Service, checker *health.Checker) (*DashboardHandler, error) {
	layout := filepath.Join(templatesDir, "layout.html")
	dashboard := filepath.Join(templatesDir, "dashboard.html")

	tmpl, err := template.ParseFiles(layout, dashboard)
	if err != nil {
		return nil, err
	}

	return &DashboardHandler{
		tmpl:     tmpl,
		services: services,
		checker:  checker,
	}, nil
}

// viewData is what we pass into the templates.
type viewData struct {
	Title    string
	Services []ServiceView
}

// buildViewData creates the view model from services + health results.
func (h *DashboardHandler) buildViewData() viewData {
	results := h.checker.Snapshot()

	staleAfter := 2*h.checker.Interval() + 10*time.Second

	views := make([]ServiceView, 0, len(h.services))
	for _, svc := range h.services {
		protoLabel := protocolLabel(svc.Type)

		v := ServiceView{
			Name:        svc.Name,
			Description: svc.Description,
			Category:    svc.Category,
			URL:         svc.URL,

			Status:      string(health.StatusUnknown),
			StatusClass: "is-dark",
			LatencyMs:   0,

			Protocol:      protoLabel,
			ProtocolClass: protocolClass(protoLabel),

			LastError:   "",
			LastChecked: time.Time{},

			IsStale:    false,
			StaleClass: "",
			StaleLabel: "",
		}

		if res, ok := results[svc.Name]; ok {
			v.Status = string(res.Status)
			v.StatusClass = bulmaClassForStatus(res.Status)
			v.LatencyMs = res.Latency.Milliseconds()
			v.LastChecked = res.CheckedAt
			v.LastError = res.Error

			if !res.CheckedAt.IsZero() && time.Since(res.CheckedAt) > staleAfter {
				v.IsStale = true
				v.StaleClass = "is-warning"
				v.StaleLabel = "STALE"
			}
		}

		views = append(views, v)
	}

	return viewData{
		Title:    "Aurora Homelab",
		Services: views,
	}
}

// Dashboard renders the main dashboard page with the full layout.
func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	data := h.buildViewData()

	if err := h.tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("error rendering dashboard: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

// DashboardPartial renders ONLY the tiles (no layout) for HTMX polling.
func (h *DashboardHandler) DashboardPartial(w http.ResponseWriter, r *http.Request) {
	data := h.buildViewData()

	// Optional: explicitly set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := h.tmpl.ExecuteTemplate(w, "dashboard", data); err != nil {
		log.Printf("error rendering dashboard partial: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

// bulmaClassForStatus maps a health.Status to a Bulma tag color class.
func bulmaClassForStatus(s health.Status) string {
	switch s {
	case health.StatusUp:
		return "is-success"
	case health.StatusDown:
		return "is-danger"
	case health.StatusStale:
		return "is-warning"
	default:
		return "is-dark"
	}
}

// protocolLabel maps a service type string to a display label.
func protocolLabel(svcType string) string {
	switch strings.ToLower(svcType) {
	case "", "http":
		return "HTTP"
	case "tcp":
		return "TCP"
	case "dns":
		return "DNS"
	case "ping":
		return "PING"
	default:
		if svcType == "" {
			return ""
		}
		return strings.ToUpper(svcType)
	}
}

// protocolClass maps a protocol label to a Bulma color class.
func protocolClass(p string) string {
	switch strings.ToUpper(p) {
	case "HTTP":
		return "is-info"
	case "TCP":
		return "is-warning"
	case "DNS":
		return "is-primary"
	case "PING":
		return "is-success"
	default:
		return "is-dark"
	}
}
