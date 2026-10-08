package lesson

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"

	"web_python/internal/auth"
	"web_python/internal/frontend"
	"web_python/internal/security"
)

// Handler xử lý các yêu cầu HTTP liên quan đến chương mục và bài học
type Handler struct {
	service *Service
}

// NewHandler khởi tạo Handler bài học
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// HandleStudentLesson xem bài học và chạy thử code ví dụ (GET /lesson?id=)
func (h *Handler) HandleStudentLesson(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "ID bài học không hợp lệ", http.StatusBadRequest)
		return
	}

	user := auth.GetUser(r.Context())
	isTeacher := user != nil && (user.Role == auth.RoleTeacher || user.Role == auth.RoleAdmin)

	lesson, curriculum, err := h.service.GetLessonDetail(id, isTeacher)
	if err != nil {
		http.Error(w, "Không tìm thấy bài học: "+err.Error(), http.StatusNotFound)
		return
	}

	breadcrumbs := []frontend.Breadcrumb{
		{Label: "Môn học", URL: "/courses"},
		{Label: lesson.CourseName, URL: "/course?id=" + strconv.Itoa(lesson.CourseID)},
		{Label: lesson.Title, URL: ""},
	}
	csrf := security.GetTokenFromContext(r.Context())
	nav := frontend.BuildNavigationData(user, "courses", breadcrumbs, "/course?id="+strconv.Itoa(lesson.CourseID), lesson.CourseName, csrf)

	h.renderTemplate(w, r, filepath.Join("web", "templates", "course", "lesson.html"), map[string]any{
		"Title":      lesson.Title + " - " + lesson.CourseName,
		"Lesson":     lesson,
		"Curriculum": curriculum,
		"User":       user,
		"Nav":        nav,
		"ActiveNav":  "courses",
	})
}

// HandleAPICourseCurriculum trả về cấu trúc môn học dạng JSON (GET /api/course/curriculum?id=)
func (h *Handler) HandleAPICourseCurriculum(w http.ResponseWriter, r *http.Request) {
	courseIDStr := r.URL.Query().Get("id")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil || courseID <= 0 {
		http.Error(w, "ID môn học không hợp lệ", http.StatusBadRequest)
		return
	}

	user := auth.GetUser(r.Context())
	isTeacher := user != nil && (user.Role == auth.RoleTeacher || user.Role == auth.RoleAdmin)

	curriculum, err := h.service.GetCurriculum(courseID, isTeacher)
	if err != nil {
		http.Error(w, "Lỗi lấy nội dung môn học: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":   true,
		"data": curriculum,
	})
}

// HandleTeacherCurriculum quản lý danh sách chương và bài học của môn (GET /teacher/curriculum?course_id=)
func (h *Handler) HandleTeacherCurriculum(w http.ResponseWriter, r *http.Request) {
	courseIDStr := r.URL.Query().Get("course_id")
	if courseIDStr == "" {
		http.Redirect(w, r, "/teacher/courses", http.StatusSeeOther)
		return
	}

	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil || courseID <= 0 {
		http.Redirect(w, r, "/teacher/courses", http.StatusSeeOther)
		return
	}

	curriculum, err := h.service.GetCurriculum(courseID, true)
	if err != nil {
		http.Error(w, "Lỗi nạp cấu trúc môn học: "+err.Error(), http.StatusNotFound)
		return
	}

	user := auth.GetUser(r.Context())
	breadcrumbs := []frontend.Breadcrumb{
		{Label: "Môn học", URL: "/teacher/courses"},
		{Label: curriculum.CourseName, URL: ""},
		{Label: "Nội dung đào tạo", URL: ""},
	}
	csrf := security.GetTokenFromContext(r.Context())
	nav := frontend.BuildNavigationData(user, "courses", breadcrumbs, "/teacher/courses", "Quản lý môn học", csrf)

	h.renderTemplate(w, r, filepath.Join("web", "templates", "teacher", "curriculum.html"), map[string]any{
		"Title":      "Quản lý Nội dung: " + curriculum.CourseName,
		"Curriculum": curriculum,
		"User":       user,
		"Nav":        nav,
		"ActiveNav":  "courses",
	})
}

// HandleTeacherCreateChapter tạo một chương mới (POST /teacher/chapter/create)
func (h *Handler) HandleTeacherCreateChapter(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
		return
	}

	courseID, _ := strconv.Atoi(r.FormValue("course_id"))
	title := r.FormValue("title")
	desc := r.FormValue("description")
	orderNum, _ := strconv.Atoi(r.FormValue("order_num"))

	if courseID <= 0 || title == "" {
		http.Error(w, "Môn học hoặc tiêu đề chương không hợp lệ", http.StatusBadRequest)
		return
	}

	if _, err := h.service.CreateChapter(courseID, title, desc, orderNum); err != nil {
		http.Error(w, "Lỗi tạo chương: "+err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/teacher/curriculum?course_id="+strconv.Itoa(courseID), http.StatusSeeOther)
}

// HandleTeacherNewLessonForm hiển thị form thêm bài học (GET /teacher/lesson/new?chapter_id=)
func (h *Handler) HandleTeacherNewLessonForm(w http.ResponseWriter, r *http.Request) {
	chapterIDStr := r.URL.Query().Get("chapter_id")
	chapterID, err := strconv.Atoi(chapterIDStr)
	if err != nil || chapterID <= 0 {
		http.Error(w, "ID chương không hợp lệ", http.StatusBadRequest)
		return
	}

	chapter, err := h.service.GetChapterByID(chapterID)
	if err != nil {
		http.Error(w, "Không tìm thấy chương", http.StatusNotFound)
		return
	}

	user := auth.GetUser(r.Context())
	breadcrumbs := []frontend.Breadcrumb{
		{Label: "Môn học", URL: "/teacher/courses"},
		{Label: "Nội dung đào tạo", URL: "/teacher/curriculum?course_id=" + strconv.Itoa(chapter.CourseID)},
		{Label: "Thêm bài học", URL: ""},
	}
	csrf := security.GetTokenFromContext(r.Context())
	nav := frontend.BuildNavigationData(user, "courses", breadcrumbs, "/teacher/curriculum?course_id="+strconv.Itoa(chapter.CourseID), "Quản lý nội dung", csrf)

	h.renderTemplate(w, r, filepath.Join("web", "templates", "teacher", "lesson_form.html"), map[string]any{
		"Title":     "Thêm bài học mới",
		"IsNew":     true,
		"Chapter":   chapter,
		"User":      user,
		"Nav":       nav,
		"ActiveNav": "courses",
	})
}

// HandleTeacherCreateLesson xử lý tạo bài học mới (POST /teacher/lesson/create)
func (h *Handler) HandleTeacherCreateLesson(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
		return
	}

	chapterID, _ := strconv.Atoi(r.FormValue("chapter_id"))
	title := r.FormValue("title")
	contentHTML := r.FormValue("content_html")
	vizType := r.FormValue("visualization_type")
	orderNum, _ := strconv.Atoi(r.FormValue("order_num"))
	isPublished := r.FormValue("is_published") == "1"

	if chapterID <= 0 || title == "" {
		http.Error(w, "Chương hoặc tiêu đề bài học không hợp lệ", http.StatusBadRequest)
		return
	}

	chapter, err := h.service.GetChapterByID(chapterID)
	if err != nil {
		http.Error(w, "Không tìm thấy chương: "+err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := h.service.CreateLesson(chapterID, title, contentHTML, orderNum, isPublished, vizType); err != nil {
		http.Error(w, "Lỗi tạo bài học: "+err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/teacher/curriculum?course_id="+strconv.Itoa(chapter.CourseID), http.StatusSeeOther)
}

// HandleTeacherEditLessonForm hiển thị form chỉnh sửa bài học (GET /teacher/lesson/edit?id=)
func (h *Handler) HandleTeacherEditLessonForm(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "ID bài học không hợp lệ", http.StatusBadRequest)
		return
	}

	lesson, err := h.service.GetLessonByID(id)
	if err != nil {
		http.Error(w, "Không tìm thấy bài học", http.StatusNotFound)
		return
	}

	chapter, err := h.service.GetChapterByID(lesson.ChapterID)
	if err != nil {
		http.Error(w, "Không tìm thấy chương", http.StatusNotFound)
		return
	}

	user := auth.GetUser(r.Context())
	breadcrumbs := []frontend.Breadcrumb{
		{Label: "Môn học", URL: "/teacher/courses"},
		{Label: "Nội dung đào tạo", URL: "/teacher/curriculum?course_id=" + strconv.Itoa(chapter.CourseID)},
		{Label: lesson.Title, URL: ""},
	}
	csrf := security.GetTokenFromContext(r.Context())
	nav := frontend.BuildNavigationData(user, "courses", breadcrumbs, "/teacher/curriculum?course_id="+strconv.Itoa(chapter.CourseID), "Quản lý nội dung", csrf)

	h.renderTemplate(w, r, filepath.Join("web", "templates", "teacher", "lesson_form.html"), map[string]any{
		"Title":     "Chỉnh sửa bài học: " + lesson.Title,
		"IsNew":     false,
		"Lesson":    lesson,
		"Chapter":   chapter,
		"User":      user,
		"Nav":       nav,
		"ActiveNav": "courses",
	})
}

// HandleTeacherUpdateLesson xử lý cập nhật bài học (POST /teacher/lesson/update)
func (h *Handler) HandleTeacherUpdateLesson(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
		return
	}

	id, _ := strconv.Atoi(r.FormValue("id"))
	title := r.FormValue("title")
	contentHTML := r.FormValue("content_html")
	vizType := r.FormValue("visualization_type")
	orderNum, _ := strconv.Atoi(r.FormValue("order_num"))
	isPublished := r.FormValue("is_published") == "1"

	if id <= 0 || title == "" {
		http.Error(w, "ID bài học hoặc tiêu đề không hợp lệ", http.StatusBadRequest)
		return
	}

	lesson, err := h.service.GetLessonByID(id)
	if err != nil {
		http.Error(w, "Không tìm thấy bài học", http.StatusBadRequest)
		return
	}

	chapter, _ := h.service.GetChapterByID(lesson.ChapterID)

	if _, err := h.service.UpdateLesson(id, title, contentHTML, orderNum, isPublished, vizType); err != nil {
		http.Error(w, "Lỗi cập nhật bài học: "+err.Error(), http.StatusBadRequest)
		return
	}

	courseID := 1
	if chapter != nil {
		courseID = chapter.CourseID
	}
	http.Redirect(w, r, "/teacher/curriculum?course_id="+strconv.Itoa(courseID), http.StatusSeeOther)
}

func (h *Handler) renderTemplate(w http.ResponseWriter, r *http.Request, tmplPath string, data any) {
	if m, ok := data.(map[string]any); ok {
		frontend.InjectCSRFToMap(r, m)
	}
	tmpl, err := frontend.ParseFilesWithShared(tmplPath)
	if err != nil {
		http.Error(w, "Không tìm thấy giao diện: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}
