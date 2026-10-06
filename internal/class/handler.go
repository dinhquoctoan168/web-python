package class

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

	"web_python/internal/auth"
	"web_python/internal/course"
)

// Handler quản lý các HTTP endpoints của lớp học
type Handler struct {
	service       *Service
	courseService *course.Service
}

// NewHandler khởi tạo Handler lớp học
func NewHandler(service *Service, courseService *course.Service) *Handler {
	return &Handler{
		service:       service,
		courseService: courseService,
	}
}

// HandleTeacherListClasses hiển thị danh sách lớp học cho giảng viên (GET /teacher/classes)
func (h *Handler) HandleTeacherListClasses(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	isAdmin := user != nil && user.Role == auth.RoleAdmin
	teacherID := 0
	if user != nil {
		teacherID = user.ID
	}

	classes, err := h.service.ListClassesForTeacher(teacherID, isAdmin)
	if err != nil {
		http.Error(w, "Lỗi lấy danh sách lớp học: "+err.Error(), http.StatusInternalServerError)
		return
	}

	courses, _ := h.courseService.ListCoursesForStudent()

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":   true,
			"data": classes,
		})
		return
	}

	h.renderTemplate(w, filepath.Join("web", "templates", "teacher", "classes.html"), map[string]any{
		"Title":   "Quản lý Lớp học",
		"Classes": classes,
		"Courses": courses,
		"User":    user,
	})
}

// HandleTeacherClassDetail xem chi tiết lớp học và danh sách sinh viên (GET /teacher/class?id=)
func (h *Handler) HandleTeacherClassDetail(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "ID lớp học không hợp lệ", http.StatusBadRequest)
		return
	}

	cl, students, err := h.service.GetClassDetail(id)
	if err != nil {
		http.Error(w, "Không tìm thấy lớp học: "+err.Error(), http.StatusNotFound)
		return
	}

	user := auth.GetUser(r.Context())
	msg := r.URL.Query().Get("msg")
	errMsg := r.URL.Query().Get("error")

	h.renderTemplate(w, filepath.Join("web", "templates", "teacher", "class_detail.html"), map[string]any{
		"Title":    "Lớp: " + cl.Name,
		"Class":    cl,
		"Students": students,
		"User":     user,
		"Message":  msg,
		"Error":    errMsg,
	})
}

// HandleTeacherCreateClass tạo một lớp học mới (POST /teacher/class/create)
func (h *Handler) HandleTeacherCreateClass(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
		return
	}

	courseID, err := strconv.Atoi(r.FormValue("course_id"))
	if err != nil || courseID <= 0 {
		http.Error(w, "Môn học không hợp lệ", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	semester := r.FormValue("semester")
	academicYear := r.FormValue("academic_year")

	user := auth.GetUser(r.Context())
	teacherID := 0
	if user != nil {
		teacherID = user.ID
	}

	_, err = h.service.CreateClass(courseID, name, semester, academicYear, teacherID)
	if err != nil {
		http.Error(w, "Không thể tạo lớp học: "+err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/teacher/classes", http.StatusSeeOther)
}

// HandleTeacherEnrollStudent thêm sinh viên vào lớp (POST /teacher/class/enroll)
func (h *Handler) HandleTeacherEnrollStudent(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
		return
	}

	classID, err := strconv.Atoi(r.FormValue("class_id"))
	if err != nil || classID <= 0 {
		http.Error(w, "ID lớp học không hợp lệ", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	if err := h.service.EnrollStudentByUsername(classID, username); err != nil {
		http.Redirect(w, r, "/teacher/class?id="+strconv.Itoa(classID)+"&error="+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/teacher/class?id="+strconv.Itoa(classID)+"&msg=Đã thêm sinh viên thành công", http.StatusSeeOther)
}

// HandleTeacherRemoveStudent xoá sinh viên khỏi lớp (POST /teacher/class/remove-student)
func (h *Handler) HandleTeacherRemoveStudent(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
		return
	}

	classID, _ := strconv.Atoi(r.FormValue("class_id"))
	studentID, _ := strconv.Atoi(r.FormValue("student_id"))

	if classID <= 0 || studentID <= 0 {
		http.Error(w, "Dữ liệu lớp học hoặc sinh viên không hợp lệ", http.StatusBadRequest)
		return
	}

	if err := h.service.RemoveStudent(classID, studentID); err != nil {
		http.Redirect(w, r, "/teacher/class?id="+strconv.Itoa(classID)+"&error="+err.Error(), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/teacher/class?id="+strconv.Itoa(classID)+"&msg=Đã xoá sinh viên khỏi lớp", http.StatusSeeOther)
}

// HandleStudentMyClasses hiển thị danh sách lớp học của sinh viên (GET /my-classes)
func (h *Handler) HandleStudentMyClasses(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	classes, err := h.service.ListClassesForStudent(user.ID)
	if err != nil {
		http.Error(w, "Lỗi lấy danh sách lớp học: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":   true,
			"data": classes,
		})
		return
	}

	h.renderTemplate(w, filepath.Join("web", "templates", "student", "classes.html"), map[string]any{
		"Title":   "Lớp học của tôi",
		"Classes": classes,
		"User":    user,
	})
}

func (h *Handler) renderTemplate(w http.ResponseWriter, tmplPath string, data any) {
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Không tìm thấy giao diện: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}
