package lesson

import (
	"html/template"
	"time"
)

// Chapter đại diện cho một chương mục trong môn học
type Chapter struct {
	ID          int      `json:"id"`
	CourseID    int      `json:"course_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	OrderNum    int      `json:"order_num"`
	Lessons     []Lesson `json:"lessons,omitempty"`
}

// Lesson đại diện cho một bài học cụ thể trong chương
type Lesson struct {
	ID           int           `json:"id"`
	ChapterID    int           `json:"chapter_id"`
	ChapterTitle string        `json:"chapter_title,omitempty"`
	CourseID     int           `json:"course_id,omitempty"`
	CourseCode   string        `json:"course_code,omitempty"`
	CourseName   string        `json:"course_name,omitempty"`
	Title        string        `json:"title"`
	ContentHTML       string        `json:"content_html"`
	SafeHTML          template.HTML `json:"-"`
	VisualizationType string        `json:"visualization_type,omitempty"`
	OrderNum          int           `json:"order_num"`
	IsPublished       bool          `json:"is_published"`
	CreatedAt         time.Time     `json:"created_at"`
}

// CourseCurriculum đại diện cho toàn bộ cây chương mục và bài học của môn học
type CourseCurriculum struct {
	CourseID    int       `json:"course_id"`
	CourseCode  string    `json:"course_code"`
	CourseName  string    `json:"course_name"`
	Description string    `json:"description"`
	Chapters    []Chapter `json:"chapters"`
}
