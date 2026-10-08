package dashboard

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"web_python/internal/auth"
	"web_python/internal/frontend"
	"web_python/internal/security"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// HandleStudentDashboard xử lý GET /dashboard
func (h *Handler) HandleStudentDashboard(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data, err := h.service.GetStudentDashboardData(user)
	if err != nil {
		http.Error(w, "Lỗi nạp dữ liệu bảng điều khiển: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Accept") == "application/json" || r.URL.Query().Get("format") == "json" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(data)
		return
	}

	h.renderTemplate(w, r, filepath.Join("web", "templates", "student", "dashboard.html"), data)
}

func (h *Handler) renderTemplate(w http.ResponseWriter, r *http.Request, tmplPath string, data any) {
	if m, ok := data.(map[string]any); ok {
		frontend.InjectCSRFToMap(r, m)
	} else if d, ok := data.(*StudentDashboardData); ok {
		d.CSRFToken = security.GetTokenFromContext(r.Context())
		d.Nav = frontend.BuildNavigationData(d.User, "dashboard", nil, "", "", d.CSRFToken)
	}
	tmpl, err := frontend.ParseFilesWithShared(tmplPath)
	if err != nil {
		http.Error(w, "Lỗi tải giao diện: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}
