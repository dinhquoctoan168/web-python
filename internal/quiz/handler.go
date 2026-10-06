package quiz

import (
	"encoding/json"
	"net/http"
	"strconv"

	"web_python/internal/auth"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// HandleGetOptions xử lý GET /api/quiz?exercise_id=
// Trả về danh sách phương án trắc nghiệm công khai (tuyệt đối không kèm is_correct)
func (h *Handler) HandleGetOptions(w http.ResponseWriter, r *http.Request) {
	exID, err := strconv.Atoi(r.URL.Query().Get("exercise_id"))
	if err != nil || exID <= 0 {
		http.Error(w, "ID bài tập không hợp lệ", http.StatusBadRequest)
		return
	}

	options, err := h.service.GetOptionsForStudent(exID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"options": options,
	})
}

// HandleSubmitQuiz xử lý POST /api/quiz/submit
func (h *Handler) HandleSubmitQuiz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Chưa đăng nhập", http.StatusUnauthorized)
		return
	}

	var req SubmitQuizRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON không hợp lệ", http.StatusBadRequest)
		return
	}

	result, err := h.service.SubmitQuiz(user.ID, req.ExerciseID, req.OptionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(result)
}
