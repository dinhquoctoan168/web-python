package judge

// JudgeRequest chứa thông tin yêu cầu chấm bài từ client hoặc service khác
type JudgeRequest struct {
	StudentID  int    `json:"student_id,omitempty"`
	ExerciseID int    `json:"exercise_id"`
	SourceCode string `json:"source_code"`
}

// TestResult kết quả kiểm thử cho từng test case
type TestResult struct {
	TestCaseID int     `json:"test_case_id"`
	OrderNum   int     `json:"order_num"`
	IsHidden   bool    `json:"is_hidden"`
	Passed     bool    `json:"passed"`
	Input      string  `json:"input,omitempty"`
	Call       string  `json:"call,omitempty"`
	Expected   string  `json:"expected,omitempty"`
	Actual     string  `json:"actual,omitempty"`
	Error      string  `json:"error,omitempty"`
	RuntimeMS  int64   `json:"runtime_ms"`
	Weight     float64 `json:"weight"`
}

// JudgeResult kết quả tổng hợp của toàn bộ bộ test
type JudgeResult struct {
	Score          float64      `json:"score"`           // Thang điểm 100
	PassedTests    int          `json:"passed_tests"`    // Số lượng test vượt qua
	TotalTests     int          `json:"total_tests"`     // Tổng số test (gồm cả public và hidden)
	Status         string       `json:"status"`          // "pass", "fail", "partial"
	Tests          []TestResult `json:"tests"`           // Chi tiết các test case (đã che giấu thông tin của hidden test)
	ExecutionError string       `json:"execution_error,omitempty"`
}
