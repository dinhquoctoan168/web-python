package submission

import (
	"errors"
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// RecordAction ghi nhận hành vi học tập (run, test, hint)
func (s *Service) RecordAction(studentID, exerciseID int, action string) (*ExerciseAttempt, error) {
	if studentID <= 0 || exerciseID <= 0 {
		return nil, errors.New("học viên hoặc bài tập không hợp lệ")
	}
	if action != "run" && action != "test" && action != "hint" {
		return nil, errors.New("hành động không được hỗ trợ")
	}

	return s.repo.IncrementActionMetric(studentID, exerciseID, action)
}

// SubmitCode lưu một lượt nộp bài chính thức
func (s *Service) SubmitCode(studentID, exerciseID int, code string, score float64, passedTests, totalTests int) (*Submission, error) {
	if studentID <= 0 || exerciseID <= 0 {
		return nil, errors.New("học viên hoặc bài tập không hợp lệ")
	}

	status := "fail"
	if totalTests > 0 && passedTests == totalTests {
		status = "pass"
	} else if passedTests > 0 {
		status = "partial"
	}

	sub := &Submission{
		StudentID:   studentID,
		ExerciseID:  exerciseID,
		SourceCode:  code,
		Score:       score,
		PassedTests: passedTests,
		TotalTests:  totalTests,
		Status:      status,
	}

	return s.repo.CreateSubmission(sub)
}

// GetMySubmissions lấy danh sách bài nộp của học viên
func (s *Service) GetMySubmissions(studentID, exerciseID int) ([]Submission, error) {
	if studentID <= 0 {
		return nil, errors.New("học viên không hợp lệ")
	}
	return s.repo.ListSubmissionsByStudent(studentID, exerciseID)
}

// GetTeacherSubmissions lấy danh sách nộp bài cho giáo viên theo dõi
func (s *Service) GetTeacherSubmissions(exerciseID int) ([]Submission, error) {
	return s.repo.ListSubmissionsForTeacher(exerciseID)
}

// GetStudentHistory lấy lịch sử làm bài chi tiết của sinh viên
func (s *Service) GetStudentHistory(studentID, exerciseID int) (*StudentHistoryView, error) {
	if studentID <= 0 || exerciseID <= 0 {
		return nil, errors.New("học viên hoặc bài tập không hợp lệ")
	}
	return s.repo.GetStudentHistoryView(studentID, exerciseID)
}

// GetSubmissionDetail lấy chi tiết một bài nộp
func (s *Service) GetSubmissionDetail(id int) (*Submission, error) {
	if id <= 0 {
		return nil, errors.New("ID bài nộp không hợp lệ")
	}
	sub, err := s.repo.GetSubmissionByID(id)
	if err != nil {
		return nil, fmt.Errorf("không thể tìm bài nộp: %w", err)
	}
	if sub == nil {
		return nil, errors.New("bài nộp không tồn tại")
	}
	return sub, nil
}
