package practice

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

// SaveDraft xử lý tự động lưu bản nháp code
func (s *Service) SaveDraft(studentID, exerciseID int, code string) (*StudentExerciseProgress, error) {
	if studentID <= 0 || exerciseID <= 0 {
		return nil, errors.New("thông tin học viên hoặc bài tập không hợp lệ")
	}

	progress, err := s.repo.UpsertDraft(studentID, exerciseID, code)
	if err != nil {
		return nil, fmt.Errorf("không thể lưu bản nháp: %w", err)
	}
	return progress, nil
}

// SubmitPractice xử lý ghi nhận kết quả kiểm thử bài tập
func (s *Service) SubmitPractice(studentID, exerciseID int, code string, score float64, passed bool) (*StudentExerciseProgress, error) {
	if studentID <= 0 || exerciseID <= 0 {
		return nil, errors.New("thông tin học viên hoặc bài tập không hợp lệ")
	}

	progress, err := s.repo.RecordSubmission(studentID, exerciseID, code, score, passed)
	if err != nil {
		return nil, fmt.Errorf("không thể ghi nhận kết quả: %w", err)
	}
	return progress, nil
}

// GetState lấy trạng thái luyện tập của học viên đối với bài tập
func (s *Service) GetState(studentID, exerciseID int) (*StudentExerciseProgress, error) {
	if studentID <= 0 || exerciseID <= 0 {
		return nil, errors.New("thông tin học viên hoặc bài tập không hợp lệ")
	}

	progress, err := s.repo.GetProgress(studentID, exerciseID)
	if err != nil {
		return nil, fmt.Errorf("lỗi lấy trạng thái: %w", err)
	}

	// Nếu chưa có tiến độ thì trả về trạng thái mặc định not_started
	if progress == nil {
		return &StudentExerciseProgress{
			StudentID:  studentID,
			ExerciseID: exerciseID,
			Status:     StatusNotStarted,
			LastCode:   "",
			BestScore:  0,
			Attempts:   0,
		}, nil
	}

	return progress, nil
}

// GetAllStates lấy danh sách trạng thái của toàn bộ bài tập cho sidebar
func (s *Service) GetAllStates(studentID int) ([]ProgressSummary, error) {
	if studentID <= 0 {
		return nil, errors.New("học viên không hợp lệ")
	}
	return s.repo.GetAllProgressForStudent(studentID)
}
