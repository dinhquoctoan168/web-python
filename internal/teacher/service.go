package teacher

import (
	"database/sql"
	"errors"
	"web_python/internal/auth"
)

var (
	ErrUnauthorizedTeacher = errors.New("giảng viên không hợp lệ")
	ErrClassNotFound       = errors.New("không tìm thấy lớp học")
	ErrStudentNotFound     = errors.New("không tìm thấy sinh viên")
)

type Service struct {
	repo *Repository
	db   *sql.DB
}

func NewService(repo *Repository, db *sql.DB) *Service {
	return &Service{repo: repo, db: db}
}

// GetDashboardPageData tổng hợp thông tin KPI và danh sách lớp cho trang chủ giảng viên
func (s *Service) GetDashboardPageData(teacher *auth.User) (*TeacherDashboardPageData, error) {
	if teacher == nil {
		return nil, ErrUnauthorizedTeacher
	}
	isAdmin := teacher.Role == auth.RoleAdmin

	summary, err := s.repo.GetDashboardSummary(teacher.ID, isAdmin)
	if err != nil {
		return nil, err
	}

	classes, err := s.repo.ListClassesSummary(teacher.ID, isAdmin)
	if err != nil {
		return nil, err
	}

	return &TeacherDashboardPageData{
		Title:   "Bảng điều khiển Giảng viên",
		User:    teacher,
		Summary: summary,
		Classes: classes,
	}, nil
}

// GetClassAnalytics lấy thông tin phân tích tiến độ của cả lớp
func (s *Service) GetClassAnalytics(classID int) (*ClassAnalyticsData, error) {
	if classID <= 0 {
		return nil, ErrClassNotFound
	}
	return s.repo.GetClassAnalytics(classID)
}

// GetStudentDiagnostics lấy chẩn đoán chi tiết từng bài tập của sinh viên trong một lớp
func (s *Service) GetStudentDiagnostics(studentID, classID int, teacher *auth.User) (*StudentDetailData, error) {
	if studentID <= 0 {
		return nil, ErrStudentNotFound
	}

	// 1. Lấy thông tin lớp học và course_id
	var className string
	var courseID int
	err := s.db.QueryRow(`SELECT name, course_id FROM classes WHERE id = ?`, classID).Scan(&className, &courseID)
	if err != nil {
		return nil, ErrClassNotFound
	}

	// 2. Lấy thông tin sinh viên
	var st auth.User
	err = s.db.QueryRow(`SELECT id, username, full_name, role FROM users WHERE id = ?`, studentID).Scan(&st.ID, &st.Username, &st.FullName, &st.Role)
	if err != nil {
		return nil, ErrStudentNotFound
	}

	// 3. Lấy chẩn đoán các bài tập
	diagnostics, err := s.repo.GetStudentExerciseDiagnostics(studentID, courseID)
	if err != nil {
		return nil, err
	}

	return &StudentDetailData{
		Title:       "Chi tiết tiến độ học tập: " + st.FullName,
		User:        teacher,
		ClassID:     classID,
		ClassName:   className,
		Student:     &st,
		Diagnostics: diagnostics,
	}, nil
}
