package progress

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"web_python/internal/auth"
)

// Handler quản lý các API tiến độ
type Handler struct {
	service *Service
}

// NewHandler khởi tạo Handler tiến độ
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// resolveStudentID xác định studentID dựa trên session và quyền của user
func (h *Handler) resolveStudentID(r *http.Request, user *auth.User) int {
	if (user.Role == auth.RoleTeacher || user.Role == auth.RoleAdmin) && r.URL.Query().Get("student_id") != "" {
		if sid, err := strconv.Atoi(r.URL.Query().Get("student_id")); err == nil && sid > 0 {
			return sid
		}
	}
	return user.ID
}

// HandleGetCourseProgress xử lý GET /api/progress/course?id=
func (h *Handler) HandleGetCourseProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Chưa đăng nhập", http.StatusUnauthorized)
		return
	}

	courseID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || courseID <= 0 {
		http.Error(w, "Thiếu hoặc sai tham số id môn học", http.StatusBadRequest)
		return
	}

	studentID := h.resolveStudentID(r, user)

	progress, err := h.service.GetCourseProgress(studentID, courseID)
	if err != nil {
		if errors.Is(err, ErrCourseNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progress)
}

// HandleGetChapterProgress xử lý GET /api/progress/chapter?id=
func (h *Handler) HandleGetChapterProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Chưa đăng nhập", http.StatusUnauthorized)
		return
	}

	chapterID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || chapterID <= 0 {
		http.Error(w, "Thiếu hoặc sai tham số id chương", http.StatusBadRequest)
		return
	}

	studentID := h.resolveStudentID(r, user)

	progress, err := h.service.GetChapterProgress(studentID, chapterID)
	if err != nil {
		if errors.Is(err, ErrChapterNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progress)
}

// HandleGetLessonProgress xử lý GET /api/progress/lesson?id=
func (h *Handler) HandleGetLessonProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Chưa đăng nhập", http.StatusUnauthorized)
		return
	}

	lessonID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || lessonID <= 0 {
		http.Error(w, "Thiếu hoặc sai tham số id bài học", http.StatusBadRequest)
		return
	}

	studentID := h.resolveStudentID(r, user)

	progress, err := h.service.GetLessonProgress(studentID, lessonID)
	if err != nil {
		if errors.Is(err, ErrLessonNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progress)
}
