package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/cyber-mountain-man/aurora-homelab-go/internal/models"
)

// DashboardHandler holds compiled templates and the configured services.
type DashboardHandler struct {
	tmpl     *template.Template
	services []models.Service
}

// NewDashboardHandler parses the HTML templates and returns a handler.
func NewDashboardHandler(templatesDir string, services []models.Service) (*DashboardHandler, error) {
	layout := filepath.Join(templatesDir, "layout.html")
	dashboard := filepath.Join(templatesDir, "dashboard.html")

	tmpl, err := template.ParseFiles(layout, dashboard)
	if err != nil {
		return nil, err
	}

	return &DashboardHandler{
		tmpl:     tmpl,
		services: services,
	}, nil
}

// viewData is what we pass into the templates.
type viewData struct {
	Title    string
	Services []models.Service
}

// Dashboard renders the main dashboard page.
func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	data := viewData{
		Title:    "Aurora Homelab",
		Services: h.services,
	}

	if err := h.tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("error rendering dashboard: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
