package progress

import "fmt"

// Service cung cấp các phương thức tính toán tiến độ học tập
type Service struct {
	repo *Repository
}

// NewService khởi tạo Service tiến độ
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetLessonProgress lấy tiến độ của bài học cho học viên
func (s *Service) GetLessonProgress(studentID, lessonID int) (*LessonProgress, error) {
	if studentID <= 0 {
		return nil, fmt.Errorf("student_id không hợp lệ")
	}
	if lessonID <= 0 {
		return nil, fmt.Errorf("lesson_id không hợp lệ")
	}
	return s.repo.GetLessonProgress(studentID, lessonID)
}

// GetChapterProgress lấy tiến độ của chương cho học viên
func (s *Service) GetChapterProgress(studentID, chapterID int) (*ChapterProgress, error) {
	if studentID <= 0 {
		return nil, fmt.Errorf("student_id không hợp lệ")
	}
	if chapterID <= 0 {
		return nil, fmt.Errorf("chapter_id không hợp lệ")
	}
	return s.repo.GetChapterProgress(studentID, chapterID)
}

// GetCourseProgress lấy tiến độ của môn học cho học viên
func (s *Service) GetCourseProgress(studentID, courseID int) (*CourseProgress, error) {
	if studentID <= 0 {
		return nil, fmt.Errorf("student_id không hợp lệ")
	}
	if courseID <= 0 {
		return nil, fmt.Errorf("course_id không hợp lệ")
	}
	return s.repo.GetCourseProgress(studentID, courseID)
}
