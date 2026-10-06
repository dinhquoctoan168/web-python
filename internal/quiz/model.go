package quiz

// QuizOption đại diện cho một phương án trắc nghiệm trong cơ sở dữ liệu
type QuizOption struct {
	ID         int    `json:"id"`
	ExerciseID int    `json:"exercise_id"`
	Content    string `json:"content"`
	IsCorrect  bool   `json:"is_correct"`
	OrderNum   int    `json:"order_num"`
}

// PublicQuizOption phương án trả về client (TUYỆT ĐỐI không chứa is_correct để chống lộ đề)
type PublicQuizOption struct {
	ID         int    `json:"id"`
	ExerciseID int    `json:"exercise_id"`
	Content    string `json:"content"`
	OrderNum   int    `json:"order_num"`
}

// SubmitQuizRequest dữ liệu học viên gửi lên khi chọn đáp án
type SubmitQuizRequest struct {
	ExerciseID int `json:"exercise_id"`
	OptionID   int `json:"option_id"`
}

// QuizResult kết quả chấm trắc nghiệm trả về cho học viên
type QuizResult struct {
	ExerciseID       int     `json:"exercise_id"`
	SelectedOptionID int     `json:"selected_option_id"`
	IsCorrect        bool    `json:"is_correct"`
	Score            float64 `json:"score"`
	Message          string  `json:"message"`
}
