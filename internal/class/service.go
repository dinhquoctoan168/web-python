package class

import (
	"errors"
	"strings"

	"web_python/internal/audit"
)

var (
	ErrInvalidClassData     = errors.New("tên lớp học và môn học không được để trống")
	ErrUnauthorizedTeacher = errors.New("giảng viên không có quyền quản lý lớp học này")
)

// Service cung cấp logic nghiệp vụ cho lớp học và ghi danh
type Service struct {
	repo         *Repository
	auditService *audit.Service
}

// NewService khởi tạo Service lớp học
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// SetAuditService thiết lập service kiểm toán
func (s *Service) SetAuditService(as *audit.Service) {
	s.auditService = as
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

	if err := s.repo.EnrollStudent(classID, studentID); err != nil {
		return err
	}

	if s.auditService != nil {
		_ = s.auditService.LogAction(nil, "enroll_student", "class", &classID, map[string]any{
			"student_id": studentID,
		})
	}
	return nil
}

// RemoveStudent huỷ ghi danh sinh viên khỏi lớp
func (s *Service) RemoveStudent(classID, studentID int) error {
	if classID <= 0 || studentID <= 0 {
		return errors.New("thông tin lớp học hoặc sinh viên không hợp lệ")
	}
	if err := s.repo.RemoveStudent(classID, studentID); err != nil {
		return err
	}

	if s.auditService != nil {
		_ = s.auditService.LogAction(nil, "delete_student", "class", &classID, map[string]any{
			"student_id": studentID,
		})
	}
	return nil
}

// VerifyTeacherOwnership kiểm tra giảng viên có quyền quản lý lớp học hay không (trừ admin)
func (s *Service) VerifyTeacherOwnership(classID, teacherID int, isAdmin bool) error {
	if isAdmin {
		return nil
	}
	if classID <= 0 || teacherID <= 0 {
		return ErrUnauthorizedTeacher
	}
	cl, err := s.repo.FindClassByID(classID)
	if err != nil {
		return err
	}
	if cl.TeacherID != teacherID {
		return ErrUnauthorizedTeacher
	}
	return nil
}

// IsStudentEnrolled kiểm tra sinh viên có trong lớp học hay không
func (s *Service) IsStudentEnrolled(classID, studentID int) (bool, error) {
	if classID <= 0 || studentID <= 0 {
		return false, nil
	}
	return s.repo.IsStudentEnrolled(classID, studentID)
}
