package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

// DashboardHandler holds compiled templates and, later, config + health data.
type DashboardHandler struct {
	tmpl *template.Template
}

// NewDashboardHandler parses the HTML templates and returns a handler.
func NewDashboardHandler(templatesDir string) (*DashboardHandler, error) {
	layout := filepath.Join(templatesDir, "layout.html")
	dashboard := filepath.Join(templatesDir, "dashboard.html")

	tmpl, err := template.ParseFiles(layout, dashboard)
	if err != nil {
		return nil, err
	}

	return &DashboardHandler{tmpl: tmpl}, nil
}

// Dashboard renders the main dashboard page.
func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title string
	}{
		Title: "Aurora Homelab",
	}

	if err := h.tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("error rendering dashboard: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
