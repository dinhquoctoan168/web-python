package quiz

import (
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ListPublicOptions lấy danh sách phương án trắc nghiệm công khai (không lộ is_correct)
func (r *Repository) ListPublicOptions(exerciseID int) ([]PublicQuizOption, error) {
	query := `SELECT id, exercise_id, content, order_num 
		FROM exercise_options 
		WHERE exercise_id = ? 
		ORDER BY order_num ASC, id ASC`

	rows, err := r.db.Query(query, exerciseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []PublicQuizOption
	for rows.Next() {
		var opt PublicQuizOption
		if err := rows.Scan(&opt.ID, &opt.ExerciseID, &opt.Content, &opt.OrderNum); err != nil {
			return nil, err
		}
		options = append(options, opt)
	}
	return options, nil
}

// CheckOption kiểm tra phương án thí sinh chọn có chính xác hay không
func (r *Repository) CheckOption(exerciseID, optionID int) (bool, error) {
	query := `SELECT is_correct FROM exercise_options WHERE exercise_id = ? AND id = ?`

	var isCorrectInt int
	err := r.db.QueryRow(query, exerciseID, optionID).Scan(&isCorrectInt)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return isCorrectInt == 1, nil
}

// AddOption thêm phương án trắc nghiệm mới
func (r *Repository) AddOption(opt *QuizOption) error {
	query := `INSERT INTO exercise_options (exercise_id, content, is_correct, order_num) VALUES (?, ?, ?, ?)`
	isCorr := 0
	if opt.IsCorrect {
		isCorr = 1
	}
	res, err := r.db.Exec(query, opt.ExerciseID, opt.Content, isCorr, opt.OrderNum)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	opt.ID = int(id)
	return nil
}
