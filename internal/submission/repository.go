package submission

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

// GetOrCreateAttempt lấy phiên làm việc hiện tại hoặc tạo mới nếu chưa có
func (r *Repository) GetOrCreateAttempt(studentID, exerciseID int) (*ExerciseAttempt, error) {
	query := `SELECT id, student_id, exercise_id, started_at, ended_at, run_count, test_count, hint_count 
		FROM exercise_attempts 
		WHERE student_id = ? AND exercise_id = ? 
		ORDER BY id DESC LIMIT 1`

	var a ExerciseAttempt
	var ended sql.NullTime
	err := r.db.QueryRow(query, studentID, exerciseID).Scan(
		&a.ID, &a.StudentID, &a.ExerciseID, &a.StartedAt, &ended,
		&a.RunCount, &a.TestCount, &a.HintCount,
	)
	if err == nil {
		if ended.Valid {
			a.EndedAt = &ended.Time
		}
		return &a, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}

	// Tạo mới attempt
	now := time.Now()
	res, err := r.db.Exec(`INSERT INTO exercise_attempts 
		(student_id, exercise_id, started_at, run_count, test_count, hint_count) 
		VALUES (?, ?, ?, 0, 0, 0)`, studentID, exerciseID, now)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &ExerciseAttempt{
		ID:         int(id),
		StudentID:  studentID,
		ExerciseID: exerciseID,
		StartedAt:  now,
		RunCount:   0,
		TestCount:  0,
		HintCount:  0,
	}, nil
}

// IncrementActionMetric tăng bộ đếm hành vi (run, test, hint)
func (r *Repository) IncrementActionMetric(studentID, exerciseID int, action string) (*ExerciseAttempt, error) {
	attempt, err := r.GetOrCreateAttempt(studentID, exerciseID)
	if err != nil {
		return nil, err
	}

	var updateSQL string
	switch action {
	case "run":
		updateSQL = "UPDATE exercise_attempts SET run_count = run_count + 1 WHERE id = ?"
		attempt.RunCount++
	case "test":
		updateSQL = "UPDATE exercise_attempts SET test_count = test_count + 1 WHERE id = ?"
		attempt.TestCount++
	case "hint":
		updateSQL = "UPDATE exercise_attempts SET hint_count = hint_count + 1 WHERE id = ?"
		attempt.HintCount++
	default:
		return attempt, nil
	}

	_, err = r.db.Exec(updateSQL, attempt.ID)
	if err != nil {
		return nil, err
	}
	return attempt, nil
}

// CreateSubmission lưu một lần nộp bài thực tế
func (r *Repository) CreateSubmission(sub *Submission) (*Submission, error) {
	now := time.Now()
	query := `INSERT INTO submissions 
		(student_id, exercise_id, source_code, score, passed_tests, total_tests, status, submitted_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query, sub.StudentID, sub.ExerciseID, sub.SourceCode, sub.Score, sub.PassedTests, sub.TotalTests, sub.Status, now)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	sub.ID = int(id)
	sub.SubmittedAt = now
	return sub, nil
}

// ListSubmissionsByStudent lấy danh sách các lần nộp bài của một học viên
func (r *Repository) ListSubmissionsByStudent(studentID, exerciseID int) ([]Submission, error) {
	var query string
	var rows *sql.Rows
	var err error

	if exerciseID > 0 {
		query = `SELECT s.id, s.student_id, s.exercise_id, COALESCE(e.title, ''), s.source_code, 
			s.score, s.passed_tests, s.total_tests, s.status, s.submitted_at 
			FROM submissions s 
			LEFT JOIN exercises e ON s.exercise_id = e.id 
			WHERE s.student_id = ? AND s.exercise_id = ? 
			ORDER BY s.submitted_at DESC`
		rows, err = r.db.Query(query, studentID, exerciseID)
	} else {
		query = `SELECT s.id, s.student_id, s.exercise_id, COALESCE(e.title, ''), s.source_code, 
			s.score, s.passed_tests, s.total_tests, s.status, s.submitted_at 
			FROM submissions s 
			LEFT JOIN exercises e ON s.exercise_id = e.id 
			WHERE s.student_id = ? 
			ORDER BY s.submitted_at DESC`
		rows, err = r.db.Query(query, studentID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Submission
	for rows.Next() {
		var s Submission
		if err := rows.Scan(&s.ID, &s.StudentID, &s.ExerciseID, &s.ExerciseTitle, &s.SourceCode,
			&s.Score, &s.PassedTests, &s.TotalTests, &s.Status, &s.SubmittedAt); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

// ListSubmissionsForTeacher lấy danh sách bài nộp phục vụ giảng viên theo dõi
func (r *Repository) ListSubmissionsForTeacher(exerciseID int) ([]Submission, error) {
	var query string
	var rows *sql.Rows
	var err error

	if exerciseID > 0 {
		query = `SELECT s.id, s.student_id, u.full_name, s.exercise_id, e.title, s.source_code, 
			s.score, s.passed_tests, s.total_tests, s.status, s.submitted_at 
			FROM submissions s 
			JOIN users u ON s.student_id = u.id 
			JOIN exercises e ON s.exercise_id = e.id 
			WHERE s.exercise_id = ? 
			ORDER BY s.submitted_at DESC LIMIT 100`
		rows, err = r.db.Query(query, exerciseID)
	} else {
		query = `SELECT s.id, s.student_id, u.full_name, s.exercise_id, e.title, s.source_code, 
			s.score, s.passed_tests, s.total_tests, s.status, s.submitted_at 
			FROM submissions s 
			JOIN users u ON s.student_id = u.id 
			JOIN exercises e ON s.exercise_id = e.id 
			ORDER BY s.submitted_at DESC LIMIT 100`
		rows, err = r.db.Query(query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Submission
	for rows.Next() {
		var s Submission
		if err := rows.Scan(&s.ID, &s.StudentID, &s.StudentName, &s.ExerciseID, &s.ExerciseTitle, &s.SourceCode,
			&s.Score, &s.PassedTests, &s.TotalTests, &s.Status, &s.SubmittedAt); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

// GetSubmissionByID lấy chi tiết 1 bài nộp
func (r *Repository) GetSubmissionByID(id int) (*Submission, error) {
	query := `SELECT s.id, s.student_id, u.full_name, s.exercise_id, e.title, s.source_code, 
		s.score, s.passed_tests, s.total_tests, s.status, s.submitted_at 
		FROM submissions s 
		JOIN users u ON s.student_id = u.id 
		JOIN exercises e ON s.exercise_id = e.id 
		WHERE s.id = ?`

	var s Submission
	err := r.db.QueryRow(query, id).Scan(
		&s.ID, &s.StudentID, &s.StudentName, &s.ExerciseID, &s.ExerciseTitle, &s.SourceCode,
		&s.Score, &s.PassedTests, &s.TotalTests, &s.Status, &s.SubmittedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// GetStudentHistoryView lấy toàn bộ lịch sử nộp bài và chỉ số nỗ lực của sinh viên
func (r *Repository) GetStudentHistoryView(studentID, exerciseID int) (*StudentHistoryView, error) {
	view := &StudentHistoryView{
		StudentID:  studentID,
		ExerciseID: exerciseID,
	}

	// Lấy thông tin sinh viên
	_ = r.db.QueryRow(`SELECT username, full_name FROM users WHERE id = ?`, studentID).Scan(&view.Username, &view.StudentName)

	// Lấy thông tin bài tập
	_ = r.db.QueryRow(`SELECT title FROM exercises WHERE id = ?`, exerciseID).Scan(&view.ExerciseTitle)

	// Lấy thông tin Attempt (run, test, hint)
	attempt, err := r.GetOrCreateAttempt(studentID, exerciseID)
	if err == nil {
		view.Attempt = attempt
	}

	// Lấy danh sách Submissions
	subs, err := r.ListSubmissionsByStudent(studentID, exerciseID)
	if err == nil {
		view.Submissions = subs
	}

	return view, nil
}

// UpdateScore cập nhật điểm số cho bài nộp
func (r *Repository) UpdateScore(id int, newScore float64) error {
	_, err := r.db.Exec("UPDATE submissions SET score = ? WHERE id = ?", newScore, id)
	return err
}
