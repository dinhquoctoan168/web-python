package dashboard

import (
	"errors"
	"web_python/internal/auth"
)

var (
	ErrUnauthorizedStudent = errors.New("học viên không hợp lệ")
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetStudentDashboardData tổng hợp dữ liệu học tập cho dashboard của học viên
func (s *Service) GetStudentDashboardData(user *auth.User) (*StudentDashboardData, error) {
	if user == nil || user.ID <= 0 {
		return nil, ErrUnauthorizedStudent
	}

	// 1. Tiến độ các môn học
	courses, err := s.repo.GetEnrolledCoursesProgress(user.ID)
	if err != nil {
		return nil, err
	}

	// 2. Các việc sắp tới: Bài tập & Ca thi
	assignments, err := s.repo.GetUpcomingAssignments(user.ID)
	if err != nil {
		return nil, err
	}

	exams, err := s.repo.GetUpcomingExams(user.ID)
	if err != nil {
		return nil, err
	}

	var upcoming []UpcomingItem
	upcoming = append(upcoming, assignments...)
	upcoming = append(upcoming, exams...)

	// 3. Kết quả các bài làm gần đây
	recentSubmissions, err := s.repo.GetRecentSubmissions(user.ID, 5)
	if err != nil {
		return nil, err
	}

	data := &StudentDashboardData{
		Title:             "Tổng quan học tập",
		User:              user,
		Courses:           courses,
		Upcoming:          upcoming,
		RecentSubmissions: recentSubmissions,
	}

	return data, nil
}
