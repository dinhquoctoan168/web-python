package submission

import (
	"errors"
	"fmt"

	"web_python/internal/audit"
	"web_python/internal/judge"
)

type Service struct {
	repo         *Repository
	judgeService *judge.Service
	auditService *audit.Service
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SetJudgeService(js *judge.Service) {
	s.judgeService = js
}

func (s *Service) SetAuditService(as *audit.Service) {
	s.auditService = as
}

// JudgeAndSubmit chấm bài thông qua Server-Side Judge và lưu bản ghi nộp bài
func (s *Service) JudgeAndSubmit(studentID, exerciseID int, sourceCode string) (*Submission, *judge.JudgeResult, error) {
	if studentID <= 0 || exerciseID <= 0 {
		return nil, nil, errors.New("học viên hoặc bài tập không hợp lệ")
	}

	var judgeRes *judge.JudgeResult
	var err error

	if s.judgeService != nil {
		judgeRes, err = s.judgeService.Evaluate(judge.JudgeRequest{
			StudentID:  studentID,
			ExerciseID: exerciseID,
			SourceCode: sourceCode,
		})
		if err != nil {
			audit.LogJudgeError(0, exerciseID, err)
			return nil, nil, fmt.Errorf("lỗi chấm bài: %w", err)
		}
	} else {
		judgeRes = &judge.JudgeResult{
			Score:       0,
			PassedTests: 0,
			TotalTests:  0,
			Status:      "fail",
		}
	}

	sub := &Submission{
		StudentID:   studentID,
		ExerciseID:  exerciseID,
		SourceCode:  sourceCode,
		Score:       judgeRes.Score,
		PassedTests: judgeRes.PassedTests,
		TotalTests:  judgeRes.TotalTests,
		Status:      judgeRes.Status,
	}

	createdSub, err := s.repo.CreateSubmission(sub)
	if err != nil {
		audit.LogSubmissionFailure(studentID, exerciseID, err.Error())
		return nil, nil, err
	}

	_ = s.repo.UpdateExerciseProgress(studentID, exerciseID, sourceCode, sub.Score, sub.Status == "pass")

	return createdSub, judgeRes, nil
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

	created, err := s.repo.CreateSubmission(sub)
	if err != nil {
		audit.LogSubmissionFailure(studentID, exerciseID, err.Error())
		return nil, err
	}

	_ = s.repo.UpdateExerciseProgress(studentID, exerciseID, code, sub.Score, sub.Status == "pass")

	return created, nil
}

// EditScore cho phép giảng viên chỉnh sửa điểm bài nộp và lưu audit log
func (s *Service) EditScore(teacherID, submissionID int, newScore float64) error {
	if submissionID <= 0 {
		return errors.New("bài nộp không hợp lệ")
	}
	sub, err := s.repo.GetSubmissionByID(submissionID)
	if err != nil {
		return err
	}
	if sub == nil {
		return errors.New("không tìm thấy bài nộp")
	}
	oldScore := sub.Score
	if err := s.repo.UpdateScore(submissionID, newScore); err != nil {
		return err
	}
	if s.auditService != nil {
		_ = s.auditService.LogAction(&teacherID, "edit_score", "submission", &submissionID, map[string]any{
			"student_id": sub.StudentID,
			"old_score":  oldScore,
			"new_score":  newScore,
		})
	}
	return nil
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
