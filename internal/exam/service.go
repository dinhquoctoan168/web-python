package exam

import (
	"errors"
	"fmt"
	"time"

	"web_python/internal/audit"
	"web_python/internal/judge"
)

var (
	ErrInvalidExamData = errors.New("thông tin bài thi không hợp lệ")
	ErrExamNotFound    = errors.New("bài thi không tồn tại")
	ErrSessionClosed   = errors.New("phiên thi đã kết thúc hoặc quá thời gian nộp bài")
)

type Service struct {
	repo         *Repository
	judgeService *judge.Service
	auditService *audit.Service
}

func NewService(repo *Repository, js *judge.Service) *Service {
	return &Service{
		repo:         repo,
		judgeService: js,
	}
}

func (s *Service) SetAuditService(as *audit.Service) {
	s.auditService = as
}

// CreateExam tạo mới đề thi cho lớp học
func (s *Service) CreateExam(teacherID, classID int, title, description string, durationMinutes int, startAtStr, endAtStr string, exerciseIDs []int, points []float64) (*Exam, error) {
	if teacherID <= 0 || classID <= 0 {
		return nil, errors.New("thông tin giảng viên hoặc lớp học không hợp lệ")
	}
	if title == "" {
		return nil, errors.New("tiêu đề bài thi không được để trống")
	}
	if durationMinutes <= 0 {
		return nil, errors.New("thời lượng làm bài phải lớn hơn 0 phút")
	}
	if len(exerciseIDs) == 0 {
		return nil, errors.New("cần chọn ít nhất một câu hỏi cho đề thi")
	}

	e := &Exam{
		ClassID:         classID,
		Title:           title,
		Description:     description,
		DurationMinutes: durationMinutes,
		Status:          StatusDraft,
		CreatedBy:       teacherID,
	}

	if startAtStr != "" {
		if t, err := time.ParseInLocation("2006-01-02T15:04", startAtStr, time.Local); err == nil {
			e.StartAt = &t
		} else if t, err := time.ParseInLocation("2006-01-02", startAtStr, time.Local); err == nil {
			e.StartAt = &t
		}
	}
	if endAtStr != "" {
		if t, err := time.ParseInLocation("2006-01-02T15:04", endAtStr, time.Local); err == nil {
			e.EndAt = &t
		} else if t, err := time.ParseInLocation("2006-01-02", endAtStr, time.Local); err == nil {
			t = t.Add(23*time.Hour + 59*time.Minute)
			e.EndAt = &t
		}
	}

	created, err := s.repo.CreateExam(e, exerciseIDs, points)
	if err != nil {
		return nil, err
	}
	if s.auditService != nil && created != nil {
		_ = s.auditService.LogAction(&teacherID, "create_exam", "exam", &created.ID, map[string]any{
			"title":    created.Title,
			"class_id": created.ClassID,
		})
	}
	return created, nil
}

// GetExam lấy chi tiết bài thi
func (s *Service) GetExam(id int) (*Exam, error) {
	if id <= 0 {
		return nil, ErrExamNotFound
	}
	e, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("không thể tìm bài thi: %w", err)
	}
	if e == nil {
		return nil, ErrExamNotFound
	}
	return e, nil
}

// ListTeacherExams lấy danh sách bài thi do giảng viên tạo
func (s *Service) ListTeacherExams(teacherID int) ([]Exam, error) {
	if teacherID <= 0 {
		return nil, errors.New("giảng viên không hợp lệ")
	}
	return s.repo.ListByTeacher(teacherID)
}

// ListStudentExams lấy danh sách bài thi của sinh viên
func (s *Service) ListStudentExams(studentID int) ([]StudentExamSummary, error) {
	if studentID <= 0 {
		return nil, errors.New("sinh viên không hợp lệ")
	}
	return s.repo.ListForStudent(studentID)
}

// PublishExam công bố bài thi
func (s *Service) PublishExam(id int) error {
	if id <= 0 {
		return ErrExamNotFound
	}
	if err := s.repo.Publish(id); err != nil {
		return err
	}
	if s.auditService != nil {
		_ = s.auditService.LogAction(nil, "publish_exam", "exam", &id, map[string]any{
			"status": StatusPublished,
		})
	}
	return nil
}

// StartOrResumeSession bắt đầu hoặc tiếp tục phiên thi của thí sinh
// Tính thời gian còn lại chuẩn xác dựa trên đồng hồ server (anti-cheat)
func (s *Service) StartOrResumeSession(examID, studentID int) (*ExamSession, *Exam, map[int]string, error) {
	exam, err := s.GetExam(examID)
	if err != nil {
		return nil, nil, nil, err
	}
	if exam.Status != StatusPublished {
		return nil, nil, nil, errors.New("kỳ thi chưa được mở hoặc đã đóng")
	}

	now := time.Now()
	if exam.StartAt != nil && now.Before(*exam.StartAt) {
		return nil, nil, nil, errors.New("chưa đến giờ mở ca thi")
	}
	if exam.EndAt != nil && now.After(*exam.EndAt) {
		return nil, nil, nil, errors.New("ca thi đã kết thúc")
	}

	// Kiểm tra phân quyền server-side: Học viên phải thuộc danh sách lớp của bài thi
	enrolled, err := s.repo.IsStudentEnrolledInExam(examID, studentID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("lỗi kiểm tra ghi danh: %w", err)
	}
	if !enrolled {
		return nil, nil, nil, errors.New("học viên không thuộc danh sách lớp học của kỳ thi này")
	}

	session, err := s.repo.GetOrCreateSession(examID, studentID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("không thể khởi tạo phiên thi: %w", err)
	}

	// Tính thời hạn chót của phiên thi: started_at + duration_minutes
	deadline := session.StartedAt.Add(time.Duration(exam.DurationMinutes) * time.Minute)
	if exam.EndAt != nil && exam.EndAt.Before(deadline) {
		deadline = *exam.EndAt
	}

	remaining := int64(time.Until(deadline).Seconds())
	if remaining <= 0 {
		remaining = 0
		if session.Status == SessionInProgress {
			session.Status = SessionTimedOut
		}
	}
	session.RemainingSeconds = remaining

	answers, err := s.repo.GetSessionAnswers(session.ID)
	if err != nil {
		answers = make(map[int]string)
	}

	return session, exam, answers, nil
}

// SaveAnswerDraft lưu tạm code câu trả lời trong quá trình làm bài thi
func (s *Service) SaveAnswerDraft(sessionID, exerciseID int, sourceCode string) error {
	session, err := s.repo.GetSession(sessionID)
	if err != nil || session == nil {
		return errors.New("phiên thi không hợp lệ")
	}
	if session.Status != SessionInProgress {
		return ErrSessionClosed
	}

	return s.repo.SaveSessionAnswer(sessionID, exerciseID, sourceCode, 0)
}

// SubmitExam hoàn thành bài thi và chấm điểm tổng kết
func (s *Service) SubmitExam(sessionID int) (float64, error) {
	session, err := s.repo.GetSession(sessionID)
	if err != nil || session == nil {
		return 0, errors.New("phiên thi không tồn tại")
	}
	if session.Status == SessionSubmitted {
		return session.FinalScore, nil
	}

	exam, err := s.GetExam(session.ExamID)
	if err != nil {
		return 0, err
	}

	answers, err := s.repo.GetSessionAnswers(sessionID)
	if err != nil {
		answers = make(map[int]string)
	}

	var finalScore float64
	for _, q := range exam.Questions {
		code := answers[q.ExerciseID]
		if code == "" {
			continue
		}

		if s.judgeService != nil {
			judgeRes, err := s.judgeService.Evaluate(judge.JudgeRequest{
				StudentID:  session.StudentID,
				ExerciseID: q.ExerciseID,
				SourceCode: code,
			})
			if err == nil && judgeRes != nil {
				// Điểm câu = (judge.Score / 100) * điểm trọng số câu hỏi
				earned := (judgeRes.Score / 100.0) * q.Points
				finalScore += earned
				_ = s.repo.SaveSessionAnswer(sessionID, q.ExerciseID, code, earned)
			}
		}
	}

	// Làm tròn 2 số thập phân
	finalScore = float64(int(finalScore*100+0.5)) / 100

	if err := s.repo.SubmitSession(sessionID, finalScore); err != nil {
		return 0, fmt.Errorf("không thể cập nhật nộp bài: %w", err)
	}

	return finalScore, nil
}

// RecordSessionEvent ghi nhận một sự kiện bất thường từ phòng thi của thí sinh
func (s *Service) RecordSessionEvent(sessionID int, eventType, eventData string) error {
	if sessionID <= 0 {
		return errors.New("phiên thi không hợp lệ")
	}

	validEvents := map[string]bool{
		"tab_hidden":       true,
		"window_blur":      true,
		"fullscreen_exit":  true,
		"copy_attempt":     true,
		"paste_attempt":    true,
		"devtool_shortcut": true,
		"context_menu":     true,
	}
	if !validEvents[eventType] {
		return errors.New("loại sự kiện giám sát không hợp lệ")
	}

	return s.repo.RecordEvent(sessionID, eventType, eventData)
}

// GetMonitoringReport lấy báo cáo giám sát tổng hợp của kỳ thi
func (s *Service) GetMonitoringReport(examID int) (*Exam, []StudentMonitoringSummary, error) {
	exam, err := s.GetExam(examID)
	if err != nil {
		return nil, nil, err
	}
	summaries, err := s.repo.GetExamMonitoringReport(examID)
	if err != nil {
		return nil, nil, err
	}
	return exam, summaries, nil
}

