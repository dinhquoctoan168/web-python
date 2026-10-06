package practice

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

// HandleSaveDraft xử lý POST /api/practice/save (Auto-save debounce)
func (h *Handler) HandleSaveDraft(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Chưa đăng nhập", http.StatusUnauthorized)
		return
	}

	var req SaveDraftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dữ liệu JSON không hợp lệ", http.StatusBadRequest)
		return
	}

	progress, err := h.service.SaveDraft(user.ID, req.ExerciseID, req.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(progress)
}

// HandleSubmitPractice xử lý POST /api/practice/submit
func (h *Handler) HandleSubmitPractice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Chưa đăng nhập", http.StatusUnauthorized)
		return
	}

	var req SubmitPracticeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dữ liệu JSON không hợp lệ", http.StatusBadRequest)
		return
	}

	progress, err := h.service.SubmitPractice(user.ID, req.ExerciseID, req.Code, req.Score, req.Passed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(progress)
}

// HandleGetState xử lý GET /api/practice/state?exercise_id=
func (h *Handler) HandleGetState(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Chưa đăng nhập", http.StatusUnauthorized)
		return
	}

	exIDStr := r.URL.Query().Get("exercise_id")
	exID, err := strconv.Atoi(exIDStr)
	if err != nil {
		http.Error(w, "ID bài tập không hợp lệ", http.StatusBadRequest)
		return
	}

	progress, err := h.service.GetState(user.ID, exID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(progress)
}

// HandleGetAllStates xử lý GET /api/practice/all-states
func (h *Handler) HandleGetAllStates(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Chưa đăng nhập", http.StatusUnauthorized)
		return
	}

	states, err := h.service.GetAllStates(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(states)
}
