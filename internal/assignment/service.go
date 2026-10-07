package assignment

import (
	"errors"
	"fmt"
	"time"

	"web_python/internal/audit"
	"web_python/internal/class"
)

type Service struct {
	repo         *Repository
	auditService *audit.Service
	classService *class.Service
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SetAuditService(as *audit.Service) {
	s.auditService = as
}

func (s *Service) SetClassService(cs *class.Service) {
	s.classService = cs
}

// GetAssignmentForStudent lấy bài tập cho học viên và kiểm tra tính hợp lệ
func (s *Service) GetAssignmentForStudent(assignmentID, studentID int) (*Assignment, error) {
	if assignmentID <= 0 || studentID <= 0 {
		return nil, errors.New("thông tin bài tập hoặc học viên không hợp lệ")
	}
	asgn, err := s.GetAssignment(assignmentID)
	if err != nil {
		return nil, err
	}
	if asgn.Status != StatusPublished {
		return nil, errors.New("bài tập chưa được công bố")
	}
	if s.classService != nil {
		enrolled, err := s.classService.IsStudentEnrolled(asgn.ClassID, studentID)
		if err != nil || !enrolled {
			return nil, errors.New("học viên không thuộc lớp học được giao bài tập")
		}
	}
	return asgn, nil
}

// AssignmentContainsExercise kiểm tra bài tập có chứa câu hỏi này không
func (s *Service) AssignmentContainsExercise(assignmentID, exerciseID int) (bool, error) {
	if assignmentID <= 0 || exerciseID <= 0 {
		return false, nil
	}
	asgn, err := s.GetAssignment(assignmentID)
	if err != nil {
		return false, err
	}
	for _, ex := range asgn.Exercises {
		if ex.ExerciseID == exerciseID {
			return true, nil
		}
	}
	return false, nil
}

// CreateAssignment xử lý tạo mới bài tập cho lớp
func (s *Service) CreateAssignment(teacherID, classID int, title, description, startAtStr, dueAtStr string, exerciseIDs []int, points []float64) (*Assignment, error) {
	if teacherID <= 0 || classID <= 0 {
		return nil, errors.New("thông tin giảng viên hoặc lớp học không hợp lệ")
	}
	if title == "" {
		return nil, errors.New("tiêu đề bài tập không được để trống")
	}
	if len(exerciseIDs) == 0 {
		return nil, errors.New("cần chọn ít nhất một câu hỏi từ ngân hàng bài tập")
	}

	a := &Assignment{
		ClassID:     classID,
		Title:       title,
		Description: description,
		Status:      StatusDraft,
		CreatedBy:   teacherID,
	}

	// Xử lý thời gian (hỗ trợ định dạng YYYY-MM-DD hoặc RFC3339)
	if startAtStr != "" {
		if t, err := time.Parse("2006-01-02T15:04", startAtStr); err == nil {
			a.StartAt = &t
		} else if t, err := time.Parse("2006-01-02", startAtStr); err == nil {
			a.StartAt = &t
		}
	}
	if dueAtStr != "" {
		if t, err := time.Parse("2006-01-02T15:04", dueAtStr); err == nil {
			a.DueAt = &t
		} else if t, err := time.Parse("2006-01-02", dueAtStr); err == nil {
			// Cuối ngày
			t = t.Add(23*time.Hour + 59*time.Minute)
			a.DueAt = &t
		}
	}

	created, err := s.repo.CreateAssignment(a, exerciseIDs, points)
	if err != nil {
		return nil, err
	}
	if s.auditService != nil && created != nil {
		_ = s.auditService.LogAction(&teacherID, "create_assignment", "assignment", &created.ID, map[string]any{
			"title":    created.Title,
			"class_id": created.ClassID,
		})
	}
	return created, nil
}

// GetAssignment lấy chi tiết bài tập
func (s *Service) GetAssignment(id int) (*Assignment, error) {
	if id <= 0 {
		return nil, errors.New("ID bài tập không hợp lệ")
	}
	a, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("không thể tìm bài tập: %w", err)
	}
	if a == nil {
		return nil, errors.New("bài tập không tồn tại")
	}
	return a, nil
}

// ListTeacherAssignments danh sách bài tập của giảng viên
func (s *Service) ListTeacherAssignments(teacherID int) ([]Assignment, error) {
	if teacherID <= 0 {
		return nil, errors.New("giảng viên không hợp lệ")
	}
	return s.repo.ListByTeacher(teacherID)
}

// ListStudentAssignments danh sách bài tập được giao cho học viên
func (s *Service) ListStudentAssignments(studentID int) ([]StudentAssignmentSummary, error) {
	if studentID <= 0 {
		return nil, errors.New("học viên không hợp lệ")
	}
	return s.repo.ListForStudent(studentID)
}

// PublishAssignment công bố bài tập
func (s *Service) PublishAssignment(id int) error {
	if id <= 0 {
		return errors.New("ID bài tập không hợp lệ")
	}
	if err := s.repo.Publish(id); err != nil {
		return err
	}
	if s.auditService != nil {
		_ = s.auditService.LogAction(nil, "publish_assignment", "assignment", &id, map[string]any{
			"status": StatusPublished,
		})
	}
	return nil
}
