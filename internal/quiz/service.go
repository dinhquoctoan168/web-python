package quiz

import (
	"errors"
	"fmt"
	"strconv"

	"web_python/internal/practice"
	"web_python/internal/submission"
)

var (
	ErrInvalidExercise = errors.New("bài tập không hợp lệ")
	ErrInvalidOption   = errors.New("phương án chọn không hợp lệ")
)

type Service struct {
	repo              *Repository
	practiceService   *practice.Service
	submissionService *submission.Service
}

func NewService(repo *Repository, ps *practice.Service, ss *submission.Service) *Service {
	return &Service{
		repo:              repo,
		practiceService:   ps,
		submissionService: ss,
	}
}

// GetOptionsForStudent lấy danh sách phương án trắc nghiệm gửi về cho học viên (không lộ is_correct)
func (s *Service) GetOptionsForStudent(exerciseID int) ([]PublicQuizOption, error) {
	if exerciseID <= 0 {
		return nil, ErrInvalidExercise
	}
	return s.repo.ListPublicOptions(exerciseID)
}

// SubmitQuiz chấm điểm bài trắc nghiệm / code tracing và cập nhật tiến độ học tập
func (s *Service) SubmitQuiz(studentID, exerciseID, optionID int) (*QuizResult, error) {
	if studentID <= 0 || exerciseID <= 0 {
		return nil, ErrInvalidExercise
	}
	if optionID <= 0 {
		return nil, ErrInvalidOption
	}

	isCorrect, err := s.repo.CheckOption(exerciseID, optionID)
	if err != nil {
		return nil, fmt.Errorf("lỗi kiểm tra đáp án: %w", err)
	}

	score := 0.0
	message := "Chưa chính xác. Hãy quan sát kỹ mã nguồn và thử lại!"
	passedTests := 0
	if isCorrect {
		score = 100.0
		message = "Chính xác! Bạn đã chọn đáp án đúng."
		passedTests = 1
	}

	answerText := "Selected Option ID: " + strconv.Itoa(optionID)

	// Cập nhật tiến độ luyện tập
	if s.practiceService != nil {
		_, _ = s.practiceService.SubmitPractice(studentID, exerciseID, answerText, score, isCorrect)
	}

	// Ghi nhận bản ghi nộp bài
	if s.submissionService != nil {
		_, _ = s.submissionService.SubmitCode(studentID, exerciseID, answerText, score, passedTests, 1)
	}

	return &QuizResult{
		ExerciseID:       exerciseID,
		SelectedOptionID: optionID,
		IsCorrect:        isCorrect,
		Score:            score,
		Message:          message,
	}, nil
}
