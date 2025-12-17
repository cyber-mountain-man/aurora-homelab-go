package handlers

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sort"
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

	JustChecked bool

	// dependency correlation
	UpstreamIssue bool
	UpstreamNote  string

	// semantic reason classification
	ReasonClass string // e.g. "TIMEOUT", "DNS", "CONN", "PERMISSION"
	ReasonLabel string // short human label
	ReasonColor string // Bulma tag class, e.g. "is-warning"

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
	serviceTile := filepath.Join(templatesDir, "service_tile.html")

	tmpl, err := template.New("layout.html").
		Funcs(template.FuncMap{
			"safeid": safeID,
		}).
		ParseFiles(layout, dashboard, serviceTile)
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
	indexByName := make(map[string]int, len(h.services))

	// Pass 1: build tiles from results
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

			UpstreamIssue: false,
			UpstreamNote:  "",
		}

		if res, ok := results[svc.Name]; ok {
			v.Status = strings.TrimSpace(string(res.Status))
			v.StatusClass = bulmaClassForStatus(res.Status)
			v.LatencyMs = res.Latency.Milliseconds()
			v.LastChecked = res.CheckedAt
			v.LastError = res.Error

			// Semantic reason classification for errors
			if v.LastError != "" {
				rc := health.ClassifyError(v.LastError)
				label, color := health.ReasonPresentation(rc)
				v.ReasonClass = string(rc)
				v.ReasonLabel = label
				v.ReasonColor = color
			}

			// Stale detection
			if !res.CheckedAt.IsZero() && time.Since(res.CheckedAt) > staleAfter {
				v.IsStale = true
				v.StaleClass = "is-warning"
				v.StaleLabel = "STALE"

				// If stale but no error, attach a default "reason"
				if v.LastError == "" {
					v.LastError = "stale: no recent health result"

					rc := health.ReasonTimeout // or create ReasonStale later
					label, color := health.ReasonPresentation(rc)
					v.ReasonClass = string(rc)
					v.ReasonLabel = label
					v.ReasonColor = color
				}
			}
		}

		views = append(views, v)
		indexByName[v.Name] = len(views) - 1
	}

	// Pass 2: dependency correlation (depends_on)
	for i, svc := range h.services {
		if len(svc.DependsOn) == 0 {
			continue
		}

		worstDepName := ""
		worstDepStatus := ""
		worstRank := 999

		for _, depName := range svc.DependsOn {
			depIdx, ok := indexByName[depName]
			if !ok {
				// dependency name not found in config
				if worstRank > 3 {
					worstRank = 3
					worstDepName = depName
					worstDepStatus = "MISSING"
				}
				continue
			}

			dep := views[depIdx]

			// rank: DOWN(0), STALE(1), UNKNOWN(2), UP(3)
			r := 3
			if dep.Status == string(health.StatusDown) {
				r = 0
			} else if dep.IsStale {
				r = 1
			} else if dep.Status == string(health.StatusUnknown) {
				r = 2
			}

			if r < worstRank {
				worstRank = r
				worstDepName = dep.Name
				if dep.IsStale {
					worstDepStatus = "STALE"
				} else {
					worstDepStatus = dep.Status
				}
			}
		}

		// Only show upstream hint when it helps (service not clearly UP)
		if worstDepName != "" && worstRank <= 2 {
			if views[i].Status != string(health.StatusUp) || views[i].IsStale {
				views[i].UpstreamIssue = true
				views[i].UpstreamNote = "Upstream: " + worstDepName + " is " + worstDepStatus
			}
		}
	}

	sort.SliceStable(views, func(i, j int) bool {
		ri, rj := severityRank(views[i]), severityRank(views[j])
		if ri != rj {
			return ri < rj
		}

		// Secondary: Category (empty goes last)
		ci, cj := views[i].Category, views[j].Category
		if ci == "" && cj != "" {
			return false
		}
		if ci != "" && cj == "" {
			return true
		}
		if ci != cj {
			return ci < cj
		}

		// Tertiary: Name
		return views[i].Name < views[j].Name
	})

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

func severityRank(v ServiceView) int {
	// Lower number = higher priority (shown first)
	if v.Status == string(health.StatusDown) {
		return 0
	}
	if v.IsStale {
		return 1
	}
	if v.Status == string(health.StatusUnknown) {
		return 2
	}
	// UP (or anything else) last
	return 3
}

func (h *DashboardHandler) RecheckService(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "missing name", http.StatusBadRequest)
		return
	}

	if ok := h.checker.CheckNow(name); !ok {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}

	data := h.buildViewData()

	var tile *ServiceView
	for i := range data.Services {
		if data.Services[i].Name == name {
			tile = &data.Services[i]
			break
		}
	}
	if tile == nil {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}

	tile.JustChecked = true

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := h.tmpl.ExecuteTemplate(w, "service_tile", tile); err != nil {
		log.Printf("error rendering service tile: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func safeID(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))

	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('-') // replace spaces/symbols with -
		}
	}
	return b.String()
}
