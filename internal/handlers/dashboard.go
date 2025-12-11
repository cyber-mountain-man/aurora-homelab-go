package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

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
	Protocol    string
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

	views := make([]ServiceView, 0, len(h.services))
	for _, svc := range h.services {
		v := ServiceView{
			Name:        svc.Name,
			Description: svc.Description,
			Category:    svc.Category,
			URL:         svc.URL,
			Status:      string(health.StatusUnknown),
			StatusClass: "is-dark",
			LatencyMs:   0,
			Protocol:    protocolLabel(svc.Type),
		}

		if res, ok := results[svc.Name]; ok {
			v.Status = string(res.Status)
			v.LatencyMs = res.Latency.Milliseconds()
			v.StatusClass = bulmaClassForStatus(res.Status)
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
