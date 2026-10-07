package exercise

import "time"

// TestCase đại diện cho test case lưu trữ trong cơ sở dữ liệu
type TestCase struct {
	ID             int     `json:"id"`
	ExerciseID     int     `json:"exercise_id"`
	InputData      string  `json:"input_data"`
	CallExpression string  `json:"call_expression"`
	ExpectedOutput string  `json:"expected_output"`
	IsHidden       bool    `json:"is_hidden"`
	Weight         float64 `json:"weight"`
	OrderNum       int     `json:"order_num"`
}

// PublicTestCase đại diện cho test case công khai gửi về client (is_hidden = 0)
// Hỗ trợ trường call, input, expected cho Skulpt runner
type PublicTestCase struct {
	ID             int     `json:"id,omitempty"`
	InputData      string  `json:"input,omitempty"`
	CallExpression string  `json:"call,omitempty"`
	ExpectedOutput string  `json:"expected"`
	Weight         float64 `json:"weight,omitempty"`
	OrderNum       int     `json:"order_num,omitempty"`
}

// Exercise đại diện cho đầy đủ thông tin bài tập (dùng trong server judge, giáo viên)
type Exercise struct {
	ID               int        `json:"id"`
	CourseID         int        `json:"course_id"`
	LessonID         *int       `json:"lesson_id,omitempty"`
	TopicID          *int       `json:"topic_id,omitempty"`
	TopicName        string     `json:"topic_name,omitempty"`
	Title            string     `json:"title"`
	ExerciseType     string     `json:"exercise_type"`
	Difficulty       string     `json:"difficulty"`
	Description      string     `json:"description"`
	InitialCode      string     `json:"initial_code"`
	SolutionCode     string     `json:"solution_code,omitempty"`
	SolutionHint     string     `json:"solution_hint,omitempty"`
	AllowedFunctions  []string   `json:"allowed_functions"`
	TimeLimitMS       int        `json:"time_limit_ms"`
	VisualizationType string     `json:"visualization_type,omitempty"`
	Status            string     `json:"status"`
	CreatedBy         *int       `json:"created_by,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	TestCases         []TestCase `json:"test_cases,omitempty"`
}

// ClientExerciseDetail là định dạng trả về an toàn cho client qua /api/exercise hoặc IDE
// Tuyệt đối không chứa SolutionCode hoặc Hidden Test Cases
type ClientExerciseDetail struct {
	ID                int              `json:"id"`
	CourseID          int              `json:"course_id"`
	LessonID          *int             `json:"lesson_id,omitempty"`
	TopicID           *int             `json:"topic_id,omitempty"`
	TopicName         string           `json:"topic_name,omitempty"`
	Title             string           `json:"title"`
	ExerciseType      string           `json:"exercise_type"`
	Difficulty        string           `json:"difficulty"`
	Description       string           `json:"description"`
	InitialCode       string           `json:"initial_code"`
	SolutionHint      string           `json:"solution_hint,omitempty"`
	AllowedFunctions  []string         `json:"allowed_functions"`
	TimeLimitMS       int              `json:"time_limit_ms"`
	VisualizationType string           `json:"visualization_type,omitempty"`
	TestCases         []PublicTestCase `json:"test_cases"`
	TestCasesJSON     string           `json:"test_cases_json"`
}

// ChapterWithExercises chứa thông tin chương mục kèm danh sách bài tập (an toàn cho client)
type ChapterWithExercises struct {
	ID          int                    `json:"id"`
	CourseID    int                    `json:"course_id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	OrderNum    int                    `json:"order_num"`
	Exercises   []ClientExerciseDetail `json:"exercises"`
}
