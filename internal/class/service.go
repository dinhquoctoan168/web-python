package class

import (
	"errors"
	"strings"
)

var (
	ErrInvalidClassData = errors.New("tên lớp học và môn học không được để trống")
)

// Service cung cấp logic nghiệp vụ cho lớp học và ghi danh
type Service struct {
	repo *Repository
}

// NewService khởi tạo Service lớp học
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ListClassesForTeacher lấy danh sách lớp học của giảng viên (hoặc tất cả nếu là admin)
func (s *Service) ListClassesForTeacher(teacherID int, isAdmin bool) ([]Class, error) {
	if isAdmin {
		return s.repo.ListClasses(0)
	}
	return s.repo.ListClasses(teacherID)
}

// ListClassesForStudent lấy danh sách các lớp học sinh viên đã tham gia
func (s *Service) ListClassesForStudent(studentID int) ([]Class, error) {
	if studentID <= 0 {
		return nil, ErrStudentNotFound
	}
	return s.repo.ListClassesByStudent(studentID)
}

// GetClassDetail lấy thông tin chi tiết lớp học và danh sách sinh viên ghi danh
func (s *Service) GetClassDetail(id int) (*Class, []EnrolledStudent, error) {
	if id <= 0 {
		return nil, nil, ErrClassNotFound
	}

	cl, err := s.repo.FindClassByID(id)
	if err != nil {
		return nil, nil, err
	}

	students, err := s.repo.GetEnrolledStudents(id)
	if err != nil {
		return nil, nil, err
	}

	return cl, students, nil
}

// CreateClass tạo một lớp học mới
func (s *Service) CreateClass(courseID int, name, semester, academicYear string, teacherID int) (*Class, error) {
	name = strings.TrimSpace(name)
	semester = strings.TrimSpace(semester)
	academicYear = strings.TrimSpace(academicYear)

	if courseID <= 0 || name == "" {
		return nil, ErrInvalidClassData
	}

	cl := &Class{
		CourseID:     courseID,
		Name:         name,
		Semester:     semester,
		AcademicYear: academicYear,
		TeacherID:    teacherID,
		Status:       "active",
	}

	if err := s.repo.CreateClass(cl); err != nil {
		return nil, err
	}
	return cl, nil
}

// EnrollStudentByUsername ghi danh sinh viên vào lớp theo username
func (s *Service) EnrollStudentByUsername(classID int, username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return errors.New("tên đăng nhập sinh viên không được để trống")
	}

	studentID, _, err := s.repo.FindStudentByUsername(username)
	if err != nil {
		return err
	}

	return s.EnrollStudent(classID, studentID)
}

// EnrollStudent ghi danh sinh viên theo ID
func (s *Service) EnrollStudent(classID, studentID int) error {
	if classID <= 0 || studentID <= 0 {
		return errors.New("thông tin lớp học hoặc sinh viên không hợp lệ")
	}

	enrolled, err := s.repo.IsStudentEnrolled(classID, studentID)
	if err != nil {
		return err
	}
	if enrolled {
		return ErrAlreadyEnrolled
	}

	return s.repo.EnrollStudent(classID, studentID)
}

// RemoveStudent huỷ ghi danh sinh viên khỏi lớp
func (s *Service) RemoveStudent(classID, studentID int) error {
	if classID <= 0 || studentID <= 0 {
		return errors.New("thông tin lớp học hoặc sinh viên không hợp lệ")
	}
	return s.repo.RemoveStudent(classID, studentID)
}
