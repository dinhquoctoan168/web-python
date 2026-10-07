package lesson

import (
	"errors"
	"html/template"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var sanitizer = func() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.AllowAttrs("class").Matching(regexp.MustCompile(`^[\w\s-]+$`)).OnElements("code", "pre", "div", "span")
	return p
}()

// SanitizeHTML làm sạch mã HTML của bài học để chống tấn công XSS
func SanitizeHTML(input string) string {
	return sanitizer.Sanitize(input)
}

var (
	ErrInvalidChapterData = errors.New("tiêu đề chương không được để trống")
	ErrInvalidLessonData  = errors.New("tiêu đề bài học không được để trống")
)

// Service cung cấp logic nghiệp vụ cho chương mục và bài học
type Service struct {
	repo *Repository
}

// NewService khởi tạo Service bài học
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetCurriculum lấy toàn bộ cấu trúc bài giảng của môn học
func (s *Service) GetCurriculum(courseID int, isTeacher bool) (*CourseCurriculum, error) {
	if courseID <= 0 {
		return nil, errors.New("ID môn học không hợp lệ")
	}
	onlyPublished := !isTeacher
	return s.repo.GetCurriculum(courseID, onlyPublished)
}

// GetLessonDetail lấy chi tiết một bài học và cấu trúc môn học tương ứng (phục vụ sidebar)
func (s *Service) GetLessonDetail(lessonID int, isTeacher bool) (*Lesson, *CourseCurriculum, error) {
	if lessonID <= 0 {
		return nil, nil, ErrLessonNotFound
	}

	lesson, err := s.repo.FindLessonByID(lessonID)
	if err != nil {
		return nil, nil, err
	}

	if !isTeacher && !lesson.IsPublished {
		return nil, nil, ErrLessonNotFound
	}

	lesson.SafeHTML = template.HTML(lesson.ContentHTML)

	curriculum, err := s.repo.GetCurriculum(lesson.CourseID, !isTeacher)
	if err != nil {
		return nil, nil, err
	}

	return lesson, curriculum, nil
}

// GetLessonByID lấy thông tin bài học đơn lẻ
func (s *Service) GetLessonByID(lessonID int) (*Lesson, error) {
	if lessonID <= 0 {
		return nil, ErrLessonNotFound
	}
	return s.repo.FindLessonByID(lessonID)
}

// GetChapterByID lấy thông tin chương đơn lẻ
func (s *Service) GetChapterByID(chapterID int) (*Chapter, error) {
	if chapterID <= 0 {
		return nil, ErrChapterNotFound
	}
	return s.repo.FindChapterByID(chapterID)
}

// CreateChapter tạo chương mục mới
func (s *Service) CreateChapter(courseID int, title, desc string, orderNum int) (*Chapter, error) {
	title = strings.TrimSpace(title)
	desc = strings.TrimSpace(desc)

	if courseID <= 0 || title == "" {
		return nil, ErrInvalidChapterData
	}

	c := &Chapter{
		CourseID:    courseID,
		Title:       title,
		Description: desc,
		OrderNum:    orderNum,
	}

	if err := s.repo.CreateChapter(c); err != nil {
		return nil, err
	}
	return c, nil
}

// CreateLesson tạo bài học mới
func (s *Service) CreateLesson(chapterID int, title, contentHTML string, orderNum int, isPublished bool, vizType ...string) (*Lesson, error) {
	title = strings.TrimSpace(title)
	contentHTML = SanitizeHTML(strings.TrimSpace(contentHTML))

	if chapterID <= 0 || title == "" {
		return nil, ErrInvalidLessonData
	}

	vt := ""
	if len(vizType) > 0 {
		vt = strings.TrimSpace(vizType[0])
	}

	l := &Lesson{
		ChapterID:         chapterID,
		Title:             title,
		ContentHTML:       contentHTML,
		VisualizationType: vt,
		OrderNum:          orderNum,
		IsPublished:       isPublished,
	}

	if err := s.repo.CreateLesson(l); err != nil {
		return nil, err
	}
	return l, nil
}

// UpdateLesson cập nhật bài học
func (s *Service) UpdateLesson(id int, title, contentHTML string, orderNum int, isPublished bool, vizType ...string) (*Lesson, error) {
	title = strings.TrimSpace(title)
	contentHTML = SanitizeHTML(strings.TrimSpace(contentHTML))

	if id <= 0 || title == "" {
		return nil, ErrInvalidLessonData
	}

	l, err := s.repo.FindLessonByID(id)
	if err != nil {
		return nil, err
	}

	l.Title = title
	l.ContentHTML = contentHTML
	l.OrderNum = orderNum
	l.IsPublished = isPublished
	if len(vizType) > 0 {
		l.VisualizationType = strings.TrimSpace(vizType[0])
	}

	if err := s.repo.UpdateLesson(l); err != nil {
		return nil, err
	}
	return l, nil
}
