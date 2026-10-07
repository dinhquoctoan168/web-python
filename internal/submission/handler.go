package submission

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"web_python/internal/auth"
	"web_python/internal/security"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// HandleRecordAction xử lý POST /api/attempt/action (run, test, hint)
func (h *Handler) HandleRecordAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Chưa đăng nhập", http.StatusUnauthorized)
		return
	}

	var req ActionMetricRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dữ liệu JSON không hợp lệ", http.StatusBadRequest)
		return
	}

	attempt, err := h.service.RecordAction(user.ID, req.ExerciseID, req.Action)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(attempt)
}

// HandleCreateSubmission xử lý POST /api/submissions
func (h *Handler) HandleCreateSubmission(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Chưa đăng nhập", http.StatusUnauthorized)
		return
	}

	var req CreateSubmissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dữ liệu JSON không hợp lệ", http.StatusBadRequest)
		return
	}

	// Nếu backend có Judge Service: chấm bằng Server-Side Judge chính thức (Phase 10)
	if h.service.judgeService != nil {
		sub, judgeRes, err := h.service.JudgeAndSubmit(user.ID, req.ExerciseID, req.SourceCode)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]any{
			"submission": sub,
			"judge":      judgeRes,
		})
		return
	}

	sub, err := h.service.SubmitCode(user.ID, req.ExerciseID, req.SourceCode, req.Score, req.PassedTests, req.TotalTests)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(sub)
}

// HandleGetMySubmissions xử lý GET /api/submissions/my
func (h *Handler) HandleGetMySubmissions(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Chưa đăng nhập", http.StatusUnauthorized)
		return
	}

	exID := 0
	if exIDStr := r.URL.Query().Get("exercise_id"); exIDStr != "" {
		exID, _ = strconv.Atoi(exIDStr)
	}

	subs, err := h.service.GetMySubmissions(user.ID, exID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(subs)
}

// TeacherSubmissionsPageData chứa dữ liệu truyền vào template giảng viên
type TeacherSubmissionsPageData struct {
	User        *auth.User
	CSRFToken   string
	Submissions []Submission
	History     *StudentHistoryView
	Submission  *Submission
}

// HandleTeacherSubmissions xử lý GET /teacher/submissions
func (h *Handler) HandleTeacherSubmissions(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	studentIDStr := r.URL.Query().Get("student_id")
	exerciseIDStr := r.URL.Query().Get("exercise_id")

	var data TeacherSubmissionsPageData
	data.User = user
	data.CSRFToken = security.GetTokenFromContext(r.Context())

	if studentIDStr != "" && exerciseIDStr != "" {
		studentID, _ := strconv.Atoi(studentIDStr)
		exerciseID, _ := strconv.Atoi(exerciseIDStr)
		history, err := h.service.GetStudentHistory(studentID, exerciseID)
		if err == nil {
			data.History = history
		}
	} else {
		exerciseID := 0
		if exerciseIDStr != "" {
			exerciseID, _ = strconv.Atoi(exerciseIDStr)
		}
		subs, err := h.service.GetTeacherSubmissions(exerciseID)
		if err == nil {
			data.Submissions = subs
		}
	}

	tmpl, err := template.ParseFiles("web/templates/teacher/submissions.html")
	if err != nil {
		http.Error(w, "Lỗi nạp template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}

// HandleTeacherSubmissionView xử lý GET /teacher/submission/view
func (h *Handler) HandleTeacherSubmissionView(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID không hợp lệ", http.StatusBadRequest)
		return
	}

	sub, err := h.service.GetSubmissionDetail(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	data := TeacherSubmissionsPageData{
		User:       user,
		CSRFToken:  security.GetTokenFromContext(r.Context()),
		Submission: sub,
	}

	tmpl, err := template.ParseFiles("web/templates/teacher/submissions.html")
	if err != nil {
		http.Error(w, "Lỗi nạp template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}
