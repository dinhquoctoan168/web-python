package exam

import "time"

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusClosed    = "closed"

	SessionInProgress = "in_progress"
	SessionSubmitted  = "submitted"
	SessionTimedOut   = "timed_out"
)

// Exam đại diện cho một kỳ thi hoặc bài kiểm tra
type Exam struct {
	ID              int            `json:"id"`
	ClassID         int            `json:"class_id"`
	ClassName       string         `json:"class_name,omitempty"`
	CourseCode      string         `json:"course_code,omitempty"`
	CourseName      string         `json:"course_name,omitempty"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	DurationMinutes int            `json:"duration_minutes"`
	StartAt         *time.Time     `json:"start_at,omitempty"`
	EndAt           *time.Time     `json:"end_at,omitempty"`
	Status          string         `json:"status"` // draft, published, closed
	CreatedBy       int            `json:"created_by"`
	CreatedAt       time.Time      `json:"created_at"`
	Questions       []ExamQuestion `json:"questions,omitempty"`
	TotalPoints     float64        `json:"total_points,omitempty"`
}

// ExamQuestion liên kết một câu hỏi từ ngân hàng câu hỏi vào bài thi
type ExamQuestion struct {
	ID                  int     `json:"id"`
	ExamID              int     `json:"exam_id"`
	ExerciseID          int     `json:"exercise_id"`
	ExerciseTitle       string  `json:"exercise_title,omitempty"`
	ExerciseDescription string  `json:"exercise_description,omitempty"`
	Difficulty          string  `json:"difficulty,omitempty"`
	Points              float64 `json:"points"`
	OrderNum            int     `json:"order_num"`
	InitialCode         string  `json:"initial_code,omitempty"`
}

// ExamSession phiên làm bài thi của một thí sinh
type ExamSession struct {
	ID               int        `json:"id"`
	ExamID           int        `json:"exam_id"`
	ExamTitle        string     `json:"exam_title,omitempty"`
	StudentID        int        `json:"student_id"`
	StudentName      string     `json:"student_name,omitempty"`
	StartedAt        time.Time  `json:"started_at"`
	SubmittedAt      *time.Time `json:"submitted_at,omitempty"`
	Status           string     `json:"status"` // in_progress, submitted, timed_out
	FinalScore       float64    `json:"final_score"`
	RemainingSeconds int64      `json:"remaining_seconds"` // Số giây còn lại dựa trên server time
}

// ExamSessionAnswer câu trả lời / mã nguồn sinh viên lưu tạm trong phòng thi
type ExamSessionAnswer struct {
	ID         int       `json:"id"`
	SessionID  int       `json:"session_id"`
	ExerciseID int       `json:"exercise_id"`
	SourceCode string    `json:"source_code"`
	Score      float64   `json:"score"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// StudentExamSummary tóm tắt bài kiểm tra dành cho sinh viên
type StudentExamSummary struct {
	ID              int        `json:"id"`
	ClassID         int        `json:"class_id"`
	ClassName       string     `json:"class_name"`
	CourseName      string     `json:"course_name"`
	Title           string     `json:"title"`
	DurationMinutes int        `json:"duration_minutes"`
	StartAt         *time.Time `json:"start_at,omitempty"`
	EndAt           *time.Time `json:"end_at,omitempty"`
	Status          string     `json:"status"`
	QuestionCount   int        `json:"question_count"`
	TotalPoints     float64    `json:"total_points"`
	SessionStatus   string     `json:"session_status"` // not_started, in_progress, submitted, timed_out
	FinalScore      *float64   `json:"final_score,omitempty"`
	IsOpen          bool       `json:"is_open"`
}

// ExamEvent lưu một sự kiện giám sát phòng thi
type ExamEvent struct {
	ID        int       `json:"id"`
	SessionID int       `json:"session_id"`
	EventType string    `json:"event_type"` // tab_hidden, window_blur, fullscreen_exit, copy_attempt, paste_attempt, devtool_shortcut
	EventData string    `json:"event_data,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// StudentMonitoringSummary tổng kết các chỉ số vi phạm giám sát của thí sinh
type StudentMonitoringSummary struct {
	SessionID           int         `json:"session_id"`
	StudentID           int         `json:"student_id"`
	StudentName         string      `json:"student_name"`
	SessionStatus       string      `json:"session_status"`
	FinalScore          float64     `json:"final_score"`
	TabHiddenCount      int         `json:"tab_hidden_count"`
	WindowBlurCount     int         `json:"window_blur_count"`
	FullscreenExitCount int         `json:"fullscreen_exit_count"`
	PasteAttemptCount   int         `json:"paste_attempt_count"`
	CopyAttemptCount    int         `json:"copy_attempt_count"`
	TotalWarnings       int         `json:"total_warnings"`
	Events              []ExamEvent `json:"events,omitempty"`
}

