package dashboard

import (
	"web_python/internal/auth"
	"web_python/internal/frontend"
)

// CourseProgressItem chứa thông tin tiến độ học tập một môn học của học viên
type CourseProgressItem struct {
	CourseID           int    `json:"course_id"`
	CourseCode         string `json:"course_code"`
	CourseName         string `json:"course_name"`
	CompletedExercises int    `json:"completed_exercises"`
	TotalExercises     int    `json:"total_exercises"`
	ProgressPercent    int    `json:"progress_percent"`
}

// UpcomingItem đại diện cho một bài tập hoặc ca thi sắp đến hạn
type UpcomingItem struct {
	Type       string `json:"type"` // "assignment" hoặc "exam"
	ID         int    `json:"id"`
	Title      string `json:"title"`
	CourseName string `json:"course_name"`
	ClassName  string `json:"class_name"`
	Deadline   string `json:"deadline"`
	Link       string `json:"link"`
}

// RecentSubmissionItem đại diện cho một bài nộp gần đây của học viên
type RecentSubmissionItem struct {
	ID            int     `json:"id"`
	ExerciseID    int     `json:"exercise_id"`
	ExerciseTitle string  `json:"exercise_title"`
	Score         float64 `json:"score"`
	Status        string  `json:"status"`
	SubmittedAt   string  `json:"submitted_at"`
}

// StudentDashboardData cấu trúc dữ liệu tổng hợp cho trang /dashboard
type StudentDashboardData struct {
	Title             string
	User              *auth.User
	CSRFToken         string
	Nav               frontend.NavigationData
	Courses           []CourseProgressItem
	Upcoming          []UpcomingItem
	RecentSubmissions []RecentSubmissionItem
}
