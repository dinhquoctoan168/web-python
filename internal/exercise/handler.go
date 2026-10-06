package exercise

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
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
