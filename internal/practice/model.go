package practice

import "time"

// Các trạng thái bài tập của học viên
const (
	StatusNotStarted = "not_started"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
)

// StudentExerciseProgress đại diện cho tiến độ luyện tập của học viên với một bài tập
type StudentExerciseProgress struct {
	ID             int        `json:"id"`
	StudentID      int        `json:"student_id"`
	ExerciseID     int        `json:"exercise_id"`
	Status         string     `json:"status"` // not_started, in_progress, completed
	LastCode       string     `json:"last_code"`
	BestScore      float64    `json:"best_score"`
	Attempts       int        `json:"attempts"`
	FirstStartedAt *time.Time `json:"first_started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// SaveDraftRequest DTO nhận từ client khi thực hiện auto-save (debounce)
type SaveDraftRequest struct {
	ExerciseID int    `json:"exercise_id"`
	Code       string `json:"code"`
}

// SubmitPracticeRequest DTO nhận từ client khi hoàn thành test cases
type SubmitPracticeRequest struct {
	ExerciseID int     `json:"exercise_id"`
	Code       string  `json:"code"`
	Score      float64 `json:"score"`
	Passed     bool    `json:"passed"`
}

// ProgressSummary DTO rút gọn trả về cho Sidebar hiển thị trạng thái
type ProgressSummary struct {
	ExerciseID int    `json:"exercise_id"`
	Status     string `json:"status"`
}
