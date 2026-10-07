package course

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

	"web_python/internal/auth"
	"web_python/internal/frontend"
)

// Handler quản lý các HTTP endpoints của môn học
type Handler struct {
	service           *Service
	curriculumFetcher func(courseID int, isTeacher bool) (any, error)
}

// NewHandler khởi tạo Handler môn học
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// SetCurriculumFetcher cài đặt hàm lấy cấu trúc bài học cho môn học
func (h *Handler) SetCurriculumFetcher(fn func(courseID int, isTeacher bool) (any, error)) {
	h.curriculumFetcher = fn
}

// HandleListCourses hiển thị danh sách môn học cho sinh viên (GET /courses)
func (h *Handler) HandleListCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := h.service.ListCoursesForStudent()
	if err != nil {
		http.Error(w, "Lỗi lấy danh sách môn học: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":   true,
			"data": courses,
		})
		return
	}

	user := auth.GetUser(r.Context())
	h.renderTemplate(w, r, filepath.Join("web", "templates", "course", "list.html"), map[string]any{
		"Title":   "Danh sách môn học",
		"Courses": courses,
		"User":    user,
	})
}

// HandleCourseDetail hiển thị chi tiết môn học (GET /course?id=)
func (h *Handler) HandleCourseDetail(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "ID môn học không hợp lệ", http.StatusBadRequest)
		return
	}

	c, err := h.service.GetCourseByID(id)
	if err != nil {
		http.Error(w, "Không tìm thấy môn học", http.StatusNotFound)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":   true,
			"data": c,
		})
		return
	}

	user := auth.GetUser(r.Context())
	isTeacher := user != nil && (user.Role == auth.RoleTeacher || user.Role == auth.RoleAdmin)

	var curr any
	if h.curriculumFetcher != nil {
		curr, _ = h.curriculumFetcher(id, isTeacher)
	}

	h.renderTemplate(w, r, filepath.Join("web", "templates", "course", "detail.html"), map[string]any{
		"Title":      c.Name + " (" + c.Code + ")",
		"Course":     c,
		"Curriculum": curr,
		"User":       user,
	})
}

// HandleTeacherListCourses quản lý danh sách môn học cho giảng viên (GET /teacher/courses)
func (h *Handler) HandleTeacherListCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := h.service.ListCoursesForTeacher()
	if err != nil {
		http.Error(w, "Lỗi lấy danh sách môn học: "+err.Error(), http.StatusInternalServerError)
		return
	}

	user := auth.GetUser(r.Context())
	h.renderTemplate(w, r, filepath.Join("web", "templates", "teacher", "courses.html"), map[string]any{
		"Title":   "Quản lý Môn học",
		"Courses": courses,
		"User":    user,
	})
}

// HandleTeacherNewCourseForm hiển thị form thêm môn học (GET /teacher/course/new)
func (h *Handler) HandleTeacherNewCourseForm(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	h.renderTemplate(w, r, filepath.Join("web", "templates", "teacher", "course_form.html"), map[string]any{
		"Title": "Thêm môn học mới",
		"IsNew": true,
		"User":  user,
	})
}

// HandleTeacherCreateCourse xử lý tạo môn học mới (POST /teacher/course/create)
func (h *Handler) HandleTeacherCreateCourse(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Dữ liệu form không hợp lệ", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	name := r.FormValue("name")
	desc := r.FormValue("description")

	user := auth.GetUser(r.Context())
	userID := 0
	if user != nil {
		userID = user.ID
	}

	_, err := h.service.CreateCourse(code, name, desc, userID)
	if err != nil {
		h.renderTemplate(w, r, filepath.Join("web", "templates", "teacher", "course_form.html"), map[string]any{
			"Title": "Thêm môn học mới",
			"IsNew": true,
			"Error": err.Error(),
			"Course": map[string]string{
				"Code":        code,
				"Name":        name,
				"Description": desc,
			},
			"User": user,
		})
		return
	}

	http.Redirect(w, r, "/teacher/courses", http.StatusSeeOther)
}

// HandleTeacherEditCourseForm hiển thị form sửa môn học (GET /teacher/course/edit?id=)
func (h *Handler) HandleTeacherEditCourseForm(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "ID môn học không hợp lệ", http.StatusBadRequest)
		return
	}

	c, err := h.service.GetCourseByID(id)
	if err != nil {
		http.Error(w, "Không tìm thấy môn học", http.StatusNotFound)
		return
	}

	user := auth.GetUser(r.Context())
	h.renderTemplate(w, r, filepath.Join("web", "templates", "teacher", "course_form.html"), map[string]any{
		"Title":  "Chỉnh sửa môn học: " + c.Code,
		"IsNew":  false,
		"Course": c,
		"User":   user,
	})
}

// HandleTeacherUpdateCourse xử lý cập nhật môn học (POST /teacher/course/update)
func (h *Handler) HandleTeacherUpdateCourse(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Dữ liệu form không hợp lệ", http.StatusBadRequest)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "ID môn học không hợp lệ", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	name := r.FormValue("name")
	desc := r.FormValue("description")
	status := r.FormValue("status")

	user := auth.GetUser(r.Context())
	_, err = h.service.UpdateCourse(id, code, name, desc, status)
	if err != nil {
		h.renderTemplate(w, r, filepath.Join("web", "templates", "teacher", "course_form.html"), map[string]any{
			"Title": "Chỉnh sửa môn học",
			"IsNew": false,
			"Error": err.Error(),
			"Course": map[string]any{
				"ID":          id,
				"Code":        code,
				"Name":        name,
				"Description": desc,
				"Status":      status,
			},
			"User": user,
		})
		return
	}

	http.Redirect(w, r, "/teacher/courses", http.StatusSeeOther)
}

func (h *Handler) renderTemplate(w http.ResponseWriter, r *http.Request, tmplPath string, data any) {
	if m, ok := data.(map[string]any); ok {
		frontend.InjectCSRFToMap(r, m)
	}
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Không tìm thấy giao diện: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}
