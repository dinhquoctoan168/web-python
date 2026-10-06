package practice

import (
	"database/sql"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetProgress lấy tiến độ của một học viên đối với một bài tập
func (r *Repository) GetProgress(studentID, exerciseID int) (*StudentExerciseProgress, error) {
	query := `SELECT id, student_id, exercise_id, status, COALESCE(last_code, ''), 
		best_score, attempts, first_started_at, completed_at, updated_at
		FROM student_exercise_progress
		WHERE student_id = ? AND exercise_id = ?`

	var p StudentExerciseProgress
	var firstStarted, completed sql.NullTime

	err := r.db.QueryRow(query, studentID, exerciseID).Scan(
		&p.ID, &p.StudentID, &p.ExerciseID, &p.Status, &p.LastCode,
		&p.BestScore, &p.Attempts, &firstStarted, &completed, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if firstStarted.Valid {
		p.FirstStartedAt = &firstStarted.Time
	}
	if completed.Valid {
		p.CompletedAt = &completed.Time
	}

	return &p, nil
}

// GetAllProgressForStudent lấy toàn bộ trạng thái tiến độ các bài tập của học viên
func (r *Repository) GetAllProgressForStudent(studentID int) ([]ProgressSummary, error) {
	query := `SELECT exercise_id, status FROM student_exercise_progress WHERE student_id = ?`
	rows, err := r.db.Query(query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ProgressSummary
	for rows.Next() {
		var item ProgressSummary
		if err := rows.Scan(&item.ExerciseID, &item.Status); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

// UpsertDraft lưu bản nháp mã nguồn (Auto-save debounce)
func (r *Repository) UpsertDraft(studentID, exerciseID int, code string) (*StudentExerciseProgress, error) {
	existing, err := r.GetProgress(studentID, exerciseID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if existing == nil {
		query := `INSERT INTO student_exercise_progress 
			(student_id, exercise_id, status, last_code, best_score, attempts, first_started_at, updated_at) 
			VALUES (?, ?, ?, ?, 0, 0, ?, ?)`
		_, err := r.db.Exec(query, studentID, exerciseID, StatusInProgress, code, now, now)
		if err != nil {
			return nil, err
		}
	} else {
		// Giữ nguyên status nếu đã completed, ngược lại chuyển sang in_progress
		status := existing.Status
		if status == StatusNotStarted {
			status = StatusInProgress
		}

		query := `UPDATE student_exercise_progress 
			SET last_code = ?, status = ?, updated_at = ? 
			WHERE student_id = ? AND exercise_id = ?`
		_, err := r.db.Exec(query, code, status, now, studentID, exerciseID)
		if err != nil {
			return nil, err
		}
	}

	return r.GetProgress(studentID, exerciseID)
}

// RecordSubmission ghi nhận lượt nộp bài hoặc chạy thử hoàn tất các test cases
func (r *Repository) RecordSubmission(studentID, exerciseID int, code string, score float64, passed bool) (*StudentExerciseProgress, error) {
	existing, err := r.GetProgress(studentID, exerciseID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if existing == nil {
		status := StatusInProgress
		var completedAt any = nil
		if passed {
			status = StatusCompleted
			completedAt = now
		}
		query := `INSERT INTO student_exercise_progress 
			(student_id, exercise_id, status, last_code, best_score, attempts, first_started_at, completed_at, updated_at) 
			VALUES (?, ?, ?, ?, ?, 1, ?, ?, ?)`
		_, err := r.db.Exec(query, studentID, exerciseID, status, code, score, now, completedAt, now)
		if err != nil {
			return nil, err
		}
	} else {
		newAttempts := existing.Attempts + 1
		newBestScore := existing.BestScore
		if score > newBestScore {
			newBestScore = score
		}

		status := existing.Status
		if passed {
			status = StatusCompleted
		}

		var query string
		if passed && existing.CompletedAt == nil {
			query = `UPDATE student_exercise_progress 
				SET last_code = ?, best_score = ?, attempts = ?, status = ?, completed_at = ?, updated_at = ? 
				WHERE student_id = ? AND exercise_id = ?`
			_, err = r.db.Exec(query, code, newBestScore, newAttempts, status, now, now, studentID, exerciseID)
		} else {
			query = `UPDATE student_exercise_progress 
				SET last_code = ?, best_score = ?, attempts = ?, status = ?, updated_at = ? 
				WHERE student_id = ? AND exercise_id = ?`
			_, err = r.db.Exec(query, code, newBestScore, newAttempts, status, now, studentID, exerciseID)
		}
		if err != nil {
			return nil, err
		}
	}

	return r.GetProgress(studentID, exerciseID)
}
