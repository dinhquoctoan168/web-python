package teacher

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"

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

// HandleTeacherDashboard xử lý GET /teacher
func (h *Handler) HandleTeacherDashboard(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data, err := h.service.GetDashboardPageData(user)
	if err != nil {
		http.Error(w, "Lỗi nạp dữ liệu dashboard giảng viên: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Accept") == "application/json" || r.URL.Query().Get("format") == "json" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(data)
		return
	}

	h.renderTemplate(w, r, filepath.Join("web", "templates", "teacher", "dashboard.html"), data)
}

// HandleClassAnalytics xử lý GET /teacher/class/analytics?class_id=
func (h *Handler) HandleClassAnalytics(w http.ResponseWriter, r *http.Request) {
	classID, err := strconv.Atoi(r.URL.Query().Get("class_id"))
	if err != nil || classID <= 0 {
		http.Error(w, "ID lớp học không hợp lệ", http.StatusBadRequest)
		return
	}

	data, err := h.service.GetClassAnalytics(classID)
	if err != nil {
		http.Error(w, "Lỗi nạp phân tích lớp học: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Accept") == "application/json" || r.URL.Query().Get("format") == "json" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(data)
		return
	}

	user := auth.GetUser(r.Context())
	breadcrumbs := []frontend.Breadcrumb{
		{Label: "Lớp học", URL: "/teacher/classes"},
		{Label: data.ClassName, URL: "/teacher/class?id=" + strconv.Itoa(data.ClassID)},
		{Label: "Phân tích lớp học", URL: ""},
	}
	csrf := security.GetTokenFromContext(r.Context())
	nav := frontend.BuildNavigationData(user, "classes", breadcrumbs, "/teacher/class?id="+strconv.Itoa(data.ClassID), "Chi tiết lớp học", csrf)

	pageData := map[string]any{
		"Title":     "Phân tích lớp học: " + data.ClassName,
		"User":      user,
		"Nav":       nav,
		"ActiveNav": "classes",
		"Analytics": data,
	}

	h.renderTemplate(w, r, filepath.Join("web", "templates", "teacher", "class_analytics.html"), pageData)
}

// HandleStudentDetail xử lý GET /teacher/student/detail?student_id=&class_id=
func (h *Handler) HandleStudentDetail(w http.ResponseWriter, r *http.Request) {
	studentID, err1 := strconv.Atoi(r.URL.Query().Get("student_id"))
	classID, err2 := strconv.Atoi(r.URL.Query().Get("class_id"))
	if err1 != nil || err2 != nil || studentID <= 0 || classID <= 0 {
		http.Error(w, "Tham số không hợp lệ", http.StatusBadRequest)
		return
	}

	user := auth.GetUser(r.Context())
	data, err := h.service.GetStudentDiagnostics(studentID, classID, user)
	if err != nil {
		http.Error(w, "Lỗi chẩn đoán tiến độ học viên: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Accept") == "application/json" || r.URL.Query().Get("format") == "json" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(data)
		return
	}

	breadcrumbs := []frontend.Breadcrumb{
		{Label: "Lớp học", URL: "/teacher/classes"},
		{Label: "Lớp " + data.ClassName, URL: "/teacher/class/analytics?class_id=" + strconv.Itoa(data.ClassID)},
		{Label: data.Student.FullName, URL: ""},
	}
	csrf := security.GetTokenFromContext(r.Context())
	data.CSRFToken = csrf
	data.Nav = frontend.BuildNavigationData(user, "classes", breadcrumbs, "/teacher/class/analytics?class_id="+strconv.Itoa(data.ClassID), "Phân tích lớp học", csrf)

	h.renderTemplate(w, r, filepath.Join("web", "templates", "teacher", "student_detail.html"), data)
}

func (h *Handler) renderTemplate(w http.ResponseWriter, r *http.Request, tmplPath string, data any) {
	if m, ok := data.(map[string]any); ok {
		frontend.InjectCSRFToMap(r, m)
	} else if d, ok := data.(*TeacherDashboardPageData); ok {
		d.CSRFToken = security.GetTokenFromContext(r.Context())
		d.Nav = frontend.BuildNavigationData(d.User, "dashboard", nil, "", "", d.CSRFToken)
	} else if d, ok := data.(*StudentDetailData); ok {
		d.CSRFToken = security.GetTokenFromContext(r.Context())
		if d.Nav.Role == "" {
			d.Nav = frontend.BuildNavigationData(d.User, "classes", nil, "", "", d.CSRFToken)
		}
	}
	tmpl, err := frontend.ParseFilesWithShared(tmplPath)
	if err != nil {
		http.Error(w, "Lỗi tải giao diện: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}
