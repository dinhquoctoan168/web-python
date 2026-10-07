package exercise

import (
	"encoding/json"
	"net/http"
	"strconv"

	"web_python/internal/auth"
)

// AssignmentAuthorizer xác thực quyền truy cập bài tập trong context assignment
type AssignmentAuthorizer interface {
	CanAccessAssignmentExercise(assignmentID, studentID, exerciseID int) (bool, error)
}

type Handler struct {
	service              *Service
	assignmentAuthorizer AssignmentAuthorizer
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SetAssignmentService(authorizer AssignmentAuthorizer) {
	h.assignmentAuthorizer = authorizer
}

// HandleAPIExercise trả về bài tập cho client qua GET /api/exercise?id=
// Tuyệt đối không trả về hidden test cases hoặc solution_code
func (h *Handler) HandleAPIExercise(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID bài tập không hợp lệ", http.StatusBadRequest)
		return
	}

	// Nếu có ngữ cảnh assignment, kiểm tra phân quyền truy cập
	asgnIDStr := r.URL.Query().Get("assignment_id")
	if asgnIDStr != "" && h.assignmentAuthorizer != nil {
		asgnID, err := strconv.Atoi(asgnIDStr)
		if err == nil && asgnID > 0 {
			user := auth.GetUser(r.Context())
			studentID := 0
			if user != nil {
				studentID = user.ID
			}
			ok, err := h.assignmentAuthorizer.CanAccessAssignmentExercise(asgnID, studentID, id)
			if err != nil || !ok {
				http.Error(w, "Không có quyền truy cập bài tập trong bài tập lớn này", http.StatusForbidden)
				return
			}
		}
	}

	detail, err := h.service.GetExerciseForClient(id)
	if err != nil {
		http.Error(w, "Lỗi truy vấn bài tập", http.StatusInternalServerError)
		return
	}
	if detail == nil {
		http.Error(w, "Không tìm thấy bài tập", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(detail)
}
