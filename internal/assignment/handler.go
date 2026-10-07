package assignment

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"web_python/internal/auth"
	"web_python/internal/class"
	"web_python/internal/exercise"
	"web_python/internal/frontend"
)

type Handler struct {
	assignmentService *Service
	classService      *class.Service
	exerciseService   *exercise.Service
}

func NewHandler(as *Service, cs *class.Service, es *exercise.Service) *Handler {
	return &Handler{
		assignmentService: as,
		classService:      cs,
		exerciseService:   es,
	}
}

// HandleTeacherListAssignments danh sách bài tập do giảng viên tạo
func (h *Handler) HandleTeacherListAssignments(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	assignments, err := h.assignmentService.ListTeacherAssignments(user.ID)
	if err != nil {
		http.Error(w, "Lỗi lấy danh sách bài tập: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.renderTemplate(w, r, filepath.Join("web", "templates", "teacher", "assignment_list.html"), map[string]any{
		"Title":       "Quản lý Bài tập",
		"Assignments": assignments,
		"User":        user,
	})
}

// HandleTeacherNewAssignmentForm hiển thị form giao bài tập mới
func (h *Handler) HandleTeacherNewAssignmentForm(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	classes, err := h.classService.ListClassesForTeacher(user.ID, user.Role == auth.RoleAdmin)
	if err != nil {
		http.Error(w, "Lỗi lấy danh sách lớp học: "+err.Error(), http.StatusInternalServerError)
		return
	}

	exercises, err := h.exerciseService.ListAll()
	if err != nil {
		http.Error(w, "Lỗi lấy ngân hàng câu hỏi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.renderTemplate(w, r, filepath.Join("web", "templates", "teacher", "assignment_form.html"), map[string]any{
		"Title":     "Giao bài tập mới",
		"Classes":   classes,
		"Exercises": exercises,
		"User":      user,
	})
}

// HandleTeacherCreateAssignment xử lý lưu bài tập mới
func (h *Handler) HandleTeacherCreateAssignment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Dữ liệu gửi lên không hợp lệ", http.StatusBadRequest)
		return
	}

	classID, _ := strconv.Atoi(r.FormValue("class_id"))
	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	startAtStr := strings.TrimSpace(r.FormValue("start_at"))
	dueAtStr := strings.TrimSpace(r.FormValue("due_at"))

	exerciseIDsRaw := r.Form["exercise_ids"]
	var exerciseIDs []int
	var points []float64

	for _, idStr := range exerciseIDsRaw {
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			continue
		}
		exerciseIDs = append(exerciseIDs, id)

		pStr := r.FormValue("point_" + idStr)
		p, err := strconv.ParseFloat(pStr, 64)
		if err != nil || p <= 0 {
			p = 1.0
		}
		points = append(points, p)
	}

	_, err := h.assignmentService.CreateAssignment(user.ID, classID, title, description, startAtStr, dueAtStr, exerciseIDs, points)
	if err != nil {
		http.Error(w, "Không thể tạo bài tập: "+err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/teacher/assignments", http.StatusSeeOther)
}

// HandleTeacherAssignmentDetail xem chi tiết đợt bài tập
func (h *Handler) HandleTeacherAssignmentDetail(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Error(w, "ID bài tập không hợp lệ", http.StatusBadRequest)
		return
	}

	a, err := h.assignmentService.GetAssignment(id)
	if err != nil || a == nil {
		http.Error(w, "Không tìm thấy bài tập", http.StatusNotFound)
		return
	}

	if user.Role != auth.RoleAdmin && a.CreatedBy != user.ID {
		http.Error(w, "Bạn không có quyền quản lý bài tập này", http.StatusForbidden)
		return
	}

	h.renderTemplate(w, r, filepath.Join("web", "templates", "teacher", "assignment_detail.html"), map[string]any{
		"Title":      a.Title,
		"Assignment": a,
		"User":       user,
	})
}

// HandleTeacherPublishAssignment công bố bài tập cho sinh viên xem
func (h *Handler) HandleTeacherPublishAssignment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, "ID bài tập không hợp lệ", http.StatusBadRequest)
		return
	}

	a, err := h.assignmentService.GetAssignment(id)
	if err != nil || a == nil {
		http.Error(w, "Không tìm thấy bài tập", http.StatusNotFound)
		return
	}

	if user.Role != auth.RoleAdmin && a.CreatedBy != user.ID {
		http.Error(w, "Bạn không có quyền công bố bài tập này", http.StatusForbidden)
		return
	}

	if err := h.assignmentService.PublishAssignment(id); err != nil {
		http.Error(w, "Không thể công bố bài tập: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/teacher/assignment?id="+strconv.Itoa(id), http.StatusSeeOther)
}

// HandleStudentAssignmentDetail xem chi tiết bài tập được giao cho sinh viên
func (h *Handler) HandleStudentAssignmentDetail(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Redirect(w, r, "/my-assignments", http.StatusSeeOther)
		return
	}

	a, err := h.assignmentService.GetAssignment(id)
	if err != nil || a == nil {
		http.Error(w, "Không tìm thấy bài tập", http.StatusNotFound)
		return
	}

	if user.Role == auth.RoleStudent && a.Status != StatusPublished {
		http.Error(w, "Bài tập chưa được công bố", http.StatusForbidden)
		return
	}

	if user.Role == auth.RoleStudent && h.classService != nil {
		enrolled, err := h.classService.IsStudentEnrolled(a.ClassID, user.ID)
		if err != nil || !enrolled {
			http.Error(w, "Bạn không thuộc lớp học được giao bài tập này", http.StatusForbidden)
			return
		}
	}

	h.renderTemplate(w, r, filepath.Join("web", "templates", "student", "assignment_detail.html"), map[string]any{
		"Title":      a.Title,
		"Assignment": a,
		"User":       user,
	})
}


// HandleStudentMyAssignments danh sách bài tập được giao cho học viên
func (h *Handler) HandleStudentMyAssignments(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	assignments, err := h.assignmentService.ListStudentAssignments(user.ID)
	if err != nil {
		http.Error(w, "Lỗi lấy danh sách bài tập: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.renderTemplate(w, r, filepath.Join("web", "templates", "student", "assignment_list.html"), map[string]any{
		"Title":       "Bài tập được giao",
		"Assignments": assignments,
		"User":        user,
	})
}

func (h *Handler) renderTemplate(w http.ResponseWriter, r *http.Request, tmplPath string, data any) {
	if m, ok := data.(map[string]any); ok {
		frontend.InjectCSRFToMap(r, m)
	}
	funcMap := template.FuncMap{
		"formatDate": func(t any) string {
			if t == nil {
				return "Không giới hạn"
			}
			return ""
		},
	}
	resolvedPath := frontend.ResolveTemplatePath(tmplPath)
	tmpl, err := template.New(filepath.Base(resolvedPath)).Funcs(funcMap).ParseFiles(resolvedPath)
	if err != nil {
		http.Error(w, "Lỗi nạp giao diện: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}
