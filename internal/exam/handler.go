package exam

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"web_python/internal/auth"
	"web_python/internal/class"
	"web_python/internal/exercise"
)

type Handler struct {
	examService     *Service
	classService    *class.Service
	exerciseService *exercise.Service
}

func NewHandler(es *Service, cs *class.Service, exs *exercise.Service) *Handler {
	return &Handler{
		examService:     es,
		classService:    cs,
		exerciseService: exs,
	}
}

// HandleTeacherListExams danh sách bài kiểm tra dành cho giảng viên
func (h *Handler) HandleTeacherListExams(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	exams, err := h.examService.ListTeacherExams(user.ID)
	if err != nil {
		http.Error(w, "Lỗi lấy danh sách bài thi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.renderTemplate(w, filepath.Join("web", "templates", "teacher", "exam_list.html"), map[string]any{
		"Title": "Quản lý Bài thi & Kiểm tra",
		"Exams": exams,
		"User":  user,
	})
}

// HandleTeacherNewExamForm form tạo đề thi mới
func (h *Handler) HandleTeacherNewExamForm(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	classes, err := h.classService.ListClassesForTeacher(user.ID, user.Role == auth.RoleAdmin)
	if err != nil {
		http.Error(w, "Lỗi tải lớp học: "+err.Error(), http.StatusInternalServerError)
		return
	}

	exercises, err := h.exerciseService.ListAll()
	if err != nil {
		http.Error(w, "Lỗi tải ngân hàng câu hỏi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.renderTemplate(w, filepath.Join("web", "templates", "teacher", "exam_form.html"), map[string]any{
		"Title":     "Tạo đề thi mới",
		"Classes":   classes,
		"Exercises": exercises,
		"User":      user,
	})
}

// HandleTeacherCreateExam xử lý lưu đề thi mới
func (h *Handler) HandleTeacherCreateExam(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
		return
	}

	classID, _ := strconv.Atoi(r.FormValue("class_id"))
	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	duration, _ := strconv.Atoi(r.FormValue("duration_minutes"))
	startAtStr := strings.TrimSpace(r.FormValue("start_at"))
	endAtStr := strings.TrimSpace(r.FormValue("end_at"))

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

	_, err := h.examService.CreateExam(user.ID, classID, title, description, duration, startAtStr, endAtStr, exerciseIDs, points)
	if err != nil {
		http.Error(w, "Không thể tạo đề thi: "+err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/teacher/exams", http.StatusSeeOther)
}

// HandleTeacherPublishExam công bố đề thi
func (h *Handler) HandleTeacherPublishExam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, "ID không hợp lệ", http.StatusBadRequest)
		return
	}

	if err := h.examService.PublishExam(id); err != nil {
		http.Error(w, "Không thể phát hành bài thi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/teacher/exams", http.StatusSeeOther)
}

// HandleStudentMyExams danh sách các bài thi của sinh viên
func (h *Handler) HandleStudentMyExams(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	exams, err := h.examService.ListStudentExams(user.ID)
	if err != nil {
		http.Error(w, "Lỗi lấy danh sách bài thi: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.renderTemplate(w, filepath.Join("web", "templates", "student", "exam_list.html"), map[string]any{
		"Title": "Bài kiểm tra & Thi cử",
		"Exams": exams,
		"User":  user,
	})
}

// HandleStudentTakeExam phòng thi trực tuyến của sinh viên
func (h *Handler) HandleStudentTakeExam(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUser(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	examID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || examID <= 0 {
		http.Error(w, "Mã bài thi không hợp lệ", http.StatusBadRequest)
		return
	}

	session, exam, answers, err := h.examService.StartOrResumeSession(examID, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	answersJSON, _ := json.Marshal(answers)

	h.renderTemplate(w, filepath.Join("web", "templates", "student", "exam_runner.html"), map[string]any{
		"Title":       exam.Title,
		"Exam":        exam,
		"Session":     session,
		"AnswersJSON": string(answersJSON),
		"User":        user,
	})
}

// HandleAPISaveAnswer lưu tạm đáp án câu hỏi trong phòng thi
func (h *Handler) HandleAPISaveAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Chưa đăng nhập", http.StatusUnauthorized)
		return
	}

	type reqBody struct {
		SessionID  int    `json:"session_id"`
		ExerciseID int    `json:"exercise_id"`
		SourceCode string `json:"source_code"`
	}

	var req reqBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON không hợp lệ", http.StatusBadRequest)
		return
	}

	if err := h.examService.SaveAnswerDraft(req.SessionID, req.ExerciseID, req.SourceCode); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// HandleAPISubmitExam nộp bài thi chính thức
func (h *Handler) HandleAPISubmitExam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
		return
	}

	user := auth.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Chưa đăng nhập", http.StatusUnauthorized)
		return
	}

	type reqBody struct {
		SessionID int `json:"session_id"`
	}

	var req reqBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON không hợp lệ", http.StatusBadRequest)
		return
	}

	finalScore, err := h.examService.SubmitExam(req.SessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]any{
		"ok":          true,
		"final_score": finalScore,
	})
}

func (h *Handler) renderTemplate(w http.ResponseWriter, tmplPath string, data any) {
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Lỗi nạp giao diện: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}
