package assignment

import "time"

// Trạng thái assignment
const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusClosed    = "closed"
)

// Assignment đại diện cho một đợt bài tập được giao cho một lớp học
type Assignment struct {
	ID          int                  `json:"id"`
	ClassID     int                  `json:"class_id"`
	ClassName   string               `json:"class_name,omitempty"`
	CourseCode  string               `json:"course_code,omitempty"`
	CourseName  string               `json:"course_name,omitempty"`
	Title       string               `json:"title"`
	Description string               `json:"description"`
	StartAt     *time.Time           `json:"start_at,omitempty"`
	DueAt       *time.Time           `json:"due_at,omitempty"`
	Status      string               `json:"status"` // draft, published, closed
	CreatedBy   int                  `json:"created_by"`
	CreatedAt   time.Time            `json:"created_at"`
	Exercises   []AssignmentExercise `json:"exercises,omitempty"`
	TotalPoints float64              `json:"total_points,omitempty"`
}

// AssignmentExercise liên kết một câu hỏi từ ngân hàng bài tập vào assignment
type AssignmentExercise struct {
	ID            int     `json:"id"`
	AssignmentID  int     `json:"assignment_id"`
	ExerciseID    int     `json:"exercise_id"`
	ExerciseTitle string  `json:"exercise_title,omitempty"`
	Difficulty    string  `json:"difficulty,omitempty"`
	Points        float64 `json:"points"`
	OrderNum      int     `json:"order_num"`
}

// StudentAssignmentSummary tóm tắt bài tập giao cho sinh viên trong các lớp tham gia
type StudentAssignmentSummary struct {
	ID            int        `json:"id"`
	ClassID       int        `json:"class_id"`
	ClassName     string     `json:"class_name"`
	CourseName    string     `json:"course_name"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	DueAt         *time.Time `json:"due_at,omitempty"`
	Status        string     `json:"status"`
	ExerciseCount int        `json:"exercise_count"`
	TotalPoints   float64    `json:"total_points"`
	IsUrgent      bool       `json:"is_urgent"` // Hạn nộp trong vòng 3 ngày tới
}
