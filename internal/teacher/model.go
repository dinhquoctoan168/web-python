package teacher

import "web_python/internal/auth"

// TeacherDashboardSummary chứa các chỉ số tổng quan ở trang chủ giáo viên
type TeacherDashboardSummary struct {
	TotalCourses     int `json:"total_courses"`
	TotalClasses     int `json:"total_classes"`
	TotalStudents    int `json:"total_students"`
	TotalAssignments int `json:"total_assignments"`
	TotalExams       int `json:"total_exams"`
}

// ChapterCompletionStat tỷ lệ hoàn thành theo từng chương học
type ChapterCompletionStat struct {
	ChapterID      int    `json:"chapter_id"`
	ChapterTitle   string `json:"chapter_title"`
	TotalExercises int    `json:"total_exercises"`
	CompletionRate int    `json:"completion_rate"`
}

// ClassSummaryItem tổng hợp thông tin lớp học cho giảng viên
type ClassSummaryItem struct {
	ID             int    `json:"id"`
	CourseID       int    `json:"course_id"`
	CourseCode     string `json:"course_code"`
	CourseName     string `json:"course_name"`
	Name           string `json:"name"`
	Semester       string `json:"semester"`
	AcademicYear   string `json:"academic_year"`
	StudentCount   int    `json:"student_count"`
	CompletionRate int    `json:"completion_rate"`
}

// ClassAnalyticsData toàn bộ dữ liệu phân tích chi tiết của một lớp học
type ClassAnalyticsData struct {
	ClassID         int                     `json:"class_id"`
	ClassName       string                  `json:"class_name"`
	CourseID        int                     `json:"course_id"`
	CourseCode      string                  `json:"course_code"`
	CourseName      string                  `json:"course_name"`
	Semester        string                  `json:"semester"`
	AcademicYear    string                  `json:"academic_year"`
	StudentCount    int                     `json:"student_count"`
	OverallProgress int                     `json:"overall_progress"`
	ChapterStats    []ChapterCompletionStat `json:"chapter_stats"`
	Students        []StudentProgressRow    `json:"students"`
}

// StudentProgressRow một dòng sinh viên trong bảng theo dõi lớp
type StudentProgressRow struct {
	StudentID           int     `json:"student_id"`
	Username            string  `json:"username"`
	FullName            string  `json:"full_name"`
	ProgressPercent     int     `json:"progress_percent"`
	AverageScore        float64 `json:"average_score"`
	IncompleteExercises int     `json:"incomplete_exercises"`
	LastActivity        string  `json:"last_activity"`
	HasStruggling       bool    `json:"has_struggling"`
}

// StudentExerciseDiagnostic chi tiết bài tập của sinh viên để phát hiện bài gặp khó (Definition of Done)
type StudentExerciseDiagnostic struct {
	ExerciseID     int     `json:"exercise_id"`
	ExerciseTitle  string  `json:"exercise_title"`
	Difficulty     string  `json:"difficulty"`
	Status         string  `json:"status"` // "completed", "in_progress", "not_started"
	Attempts       int     `json:"attempts"`
	RunCount       int     `json:"run_count"`
	TestCount      int     `json:"test_count"`
	HintCount      int     `json:"hint_count"`
	Score          float64 `json:"score"`
	LastActivity   string  `json:"last_activity"`
	IsStruggling   bool    `json:"is_struggling"`
	StruggleReason string  `json:"struggle_reason"`
}

// StudentDetailData cấu trúc dữ liệu gửi về trang chi tiết học tập của sinh viên
type StudentDetailData struct {
	Title       string                      `json:"title"`
	User        *auth.User                  `json:"user"`
	CSRFToken   string                      `json:"csrf_token"`
	ClassID     int                         `json:"class_id"`
	ClassName   string                      `json:"class_name"`
	Student     *auth.User                  `json:"student"`
	Diagnostics []StudentExerciseDiagnostic `json:"diagnostics"`
}

// TeacherDashboardPageData dữ liệu trang chính /teacher
type TeacherDashboardPageData struct {
	Title     string
	User      *auth.User
	CSRFToken string
	Summary   TeacherDashboardSummary
	Classes   []ClassSummaryItem
}
