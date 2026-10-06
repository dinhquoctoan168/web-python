package submission

import "time"

// Submission đại diện cho một lần nộp bài thực tế của học viên
type Submission struct {
	ID            int       `json:"id"`
	StudentID     int       `json:"student_id"`
	StudentName   string    `json:"student_name,omitempty"`
	ExerciseID    int       `json:"exercise_id"`
	ExerciseTitle string    `json:"exercise_title,omitempty"`
	SourceCode    string    `json:"source_code"`
	Score         float64   `json:"score"`
	PassedTests   int       `json:"passed_tests"`
	TotalTests    int       `json:"total_tests"`
	Status        string    `json:"status"` // pass, fail, partial
	SubmittedAt   time.Time `json:"submitted_at"`
}

// ExerciseAttempt đại diện cho phiên học và các chỉ số nỗ lực (run, test, hint)
type ExerciseAttempt struct {
	ID         int        `json:"id"`
	StudentID  int        `json:"student_id"`
	ExerciseID int        `json:"exercise_id"`
	StartedAt  time.Time  `json:"started_at"`
	EndedAt    *time.Time `json:"ended_at,omitempty"`
	RunCount   int        `json:"run_count"`
	TestCount  int        `json:"test_count"`
	HintCount  int        `json:"hint_count"`
}

// ActionMetricRequest DTO khi học viên thực hiện thao tác Run, Test, hoặc Hint
type ActionMetricRequest struct {
	ExerciseID int    `json:"exercise_id"`
	Action     string `json:"action"` // "run", "test", "hint"
}

// CreateSubmissionRequest DTO gửi từ client khi lưu lượt nộp bài
type CreateSubmissionRequest struct {
	ExerciseID  int     `json:"exercise_id"`
	SourceCode  string  `json:"source_code"`
	Score       float64 `json:"score"`
	PassedTests int     `json:"passed_tests"`
	TotalTests  int     `json:"total_tests"`
}

// StudentHistoryView DTO tổng hợp lịch sử bài nộp và thống kê nỗ lực cho giảng viên
type StudentHistoryView struct {
	StudentID     int              `json:"student_id"`
	StudentName   string           `json:"student_name"`
	Username      string           `json:"username"`
	ExerciseID    int              `json:"exercise_id"`
	ExerciseTitle string           `json:"exercise_title"`
	Attempt       *ExerciseAttempt `json:"attempt,omitempty"`
	Submissions   []Submission     `json:"submissions"`
}
