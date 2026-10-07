package course

import (
	"errors"
	"strings"
)

var (
	ErrInvalidCourseData = errors.New("mã môn học và tên môn học không được để trống")
	ErrDuplicateCode     = errors.New("mã môn học đã tồn tại trên hệ thống")
	ErrInvalidStatus     = errors.New("trạng thái môn học không hợp lệ")
)

// Service cung cấp logic nghiệp vụ cho quản lý môn học
type Service struct {
	repo *Repository
}

// NewService khởi tạo Course Service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ListCoursesForStudent lấy danh sách môn học đang kích hoạt cho sinh viên
func (s *Service) ListCoursesForStudent() ([]Course, error) {
	return s.repo.ListCourses(true)
}

// ListCoursesForTeacher lấy tất cả môn học (bao gồm cả môn đã ẩn) cho giảng viên
func (s *Service) ListCoursesForTeacher() ([]Course, error) {
	return s.repo.ListCourses(false)
}

// GetCourseByID lấy thông tin chi tiết một môn học theo ID
func (s *Service) GetCourseByID(id int) (*Course, error) {
	if id <= 0 {
		return nil, ErrCourseNotFound
	}
	return s.repo.FindCourseByID(id)
}

// GetCourseByCode lấy thông tin môn học theo mã code
func (s *Service) GetCourseByCode(code string) (*Course, error) {
	code = strings.TrimSpace(strings.ToUpper(code))
	if code == "" {
		return nil, ErrInvalidCourseData
	}
	return s.repo.FindCourseByCode(code)
}

// CreateCourse tạo mới một môn học
func (s *Service) CreateCourse(code, name, desc string, createdBy int) (*Course, error) {
	code = strings.TrimSpace(strings.ToUpper(code))
	name = strings.TrimSpace(name)
	desc = strings.TrimSpace(desc)

	if code == "" || name == "" {
		return nil, ErrInvalidCourseData
	}

	existing, err := s.repo.FindCourseByCode(code)
	if err == nil && existing != nil {
		return nil, ErrDuplicateCode
	} else if err != nil && !errors.Is(err, ErrCourseNotFound) {
		return nil, err
	}

	c := &Course{
		Code:        code,
		Name:        name,
		Description: desc,
		Status:      StatusActive,
		CreatedBy:   createdBy,
	}

	if err := s.repo.CreateCourse(c); err != nil {
		return nil, err
	}
	return c, nil
}

// UpdateCourse cập nhật thông tin môn học
func (s *Service) UpdateCourse(id int, code, name, desc, status string) (*Course, error) {
	code = strings.TrimSpace(strings.ToUpper(code))
	name = strings.TrimSpace(name)
	desc = strings.TrimSpace(desc)
	status = strings.TrimSpace(strings.ToLower(status))

	if id <= 0 {
		return nil, ErrCourseNotFound
	}
	if code == "" || name == "" {
		return nil, ErrInvalidCourseData
	}
	if status != StatusActive && status != StatusArchived {
		return nil, ErrInvalidStatus
	}

	existing, err := s.repo.FindCourseByID(id)
	if err != nil {
		return nil, err
	}

	// Kiểm tra trùng mã với môn học khác
	byCode, err := s.repo.FindCourseByCode(code)
	if err == nil && byCode != nil && byCode.ID != id {
		return nil, ErrDuplicateCode
	}

	existing.Code = code
	existing.Name = name
	existing.Description = desc
	existing.Status = status

	if err := s.repo.UpdateCourse(existing); err != nil {
		return nil, err
	}
	return existing, nil
}
