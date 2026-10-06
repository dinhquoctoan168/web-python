package exam

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

// CreateExam tạo bài thi mới và gắn danh sách câu hỏi
func (r *Repository) CreateExam(e *Exam, exerciseIDs []int, points []float64) (*Exam, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now()
	query := `INSERT INTO exams (class_id, title, description, duration_minutes, start_at, end_at, status, created_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := tx.Exec(query, e.ClassID, e.Title, e.Description, e.DurationMinutes, e.StartAt, e.EndAt, e.Status, e.CreatedBy, now)
	if err != nil {
		return nil, err
	}

	examID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	e.ID = int(examID)
	e.CreatedAt = now

	insertQ := `INSERT INTO exam_questions (exam_id, exercise_id, points, order_num) VALUES (?, ?, ?, ?)`
	for i, exID := range exerciseIDs {
		p := 1.0
		if i < len(points) && points[i] > 0 {
			p = points[i]
		}
		if _, err := tx.Exec(insertQ, e.ID, exID, p, i+1); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return e, nil
}

// FindByID lấy chi tiết kỳ thi kèm danh sách câu hỏi
func (r *Repository) FindByID(id int) (*Exam, error) {
	query := `SELECT e.id, e.class_id, c.name, co.code, co.name, e.title, COALESCE(e.description, ''),
		e.duration_minutes, e.start_at, e.end_at, e.status, e.created_by, e.created_at
		FROM exams e
		JOIN classes c ON e.class_id = c.id
		JOIN courses co ON c.course_id = co.id
		WHERE e.id = ?`

	var e Exam
	var startAt, endAt sql.NullTime
	err := r.db.QueryRow(query, id).Scan(
		&e.ID, &e.ClassID, &e.ClassName, &e.CourseCode, &e.CourseName,
		&e.Title, &e.Description, &e.DurationMinutes, &startAt, &endAt,
		&e.Status, &e.CreatedBy, &e.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if startAt.Valid {
		e.StartAt = &startAt.Time
	}
	if endAt.Valid {
		e.EndAt = &endAt.Time
	}

	qQuery := `SELECT eq.id, eq.exam_id, eq.exercise_id, ex.title, ex.description, ex.difficulty, 
		COALESCE(ex.initial_code, ''), eq.points, eq.order_num
		FROM exam_questions eq
		JOIN exercises ex ON eq.exercise_id = ex.id
		WHERE eq.exam_id = ?
		ORDER BY eq.order_num ASC, eq.id ASC`

	rows, err := r.db.Query(qQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []ExamQuestion
	var totalPoints float64
	for rows.Next() {
		var q ExamQuestion
		if err := rows.Scan(&q.ID, &q.ExamID, &q.ExerciseID, &q.ExerciseTitle, &q.ExerciseDescription,
			&q.Difficulty, &q.InitialCode, &q.Points, &q.OrderNum); err != nil {
			return nil, err
		}
		totalPoints += q.Points
		questions = append(questions, q)
	}
	e.Questions = questions
	e.TotalPoints = totalPoints

	return &e, nil
}

// ListByTeacher lấy danh sách bài kiểm tra do giảng viên tạo
func (r *Repository) ListByTeacher(teacherID int) ([]Exam, error) {
	query := `SELECT e.id, e.class_id, c.name, co.code, co.name, e.title, e.description,
		e.duration_minutes, e.start_at, e.end_at, e.status, e.created_by, e.created_at,
		COALESCE((SELECT SUM(points) FROM exam_questions WHERE exam_id = e.id), 0) as total_points,
		COALESCE((SELECT COUNT(*) FROM exam_questions WHERE exam_id = e.id), 0) as q_count
		FROM exams e
		JOIN classes c ON e.class_id = c.id
		JOIN courses co ON c.course_id = co.id
		WHERE e.created_by = ?
		ORDER BY e.created_at DESC`

	rows, err := r.db.Query(query, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Exam
	for rows.Next() {
		var e Exam
		var startAt, endAt sql.NullTime
		var qCount int
		if err := rows.Scan(&e.ID, &e.ClassID, &e.ClassName, &e.CourseCode, &e.CourseName,
			&e.Title, &e.Description, &e.DurationMinutes, &startAt, &endAt,
			&e.Status, &e.CreatedBy, &e.CreatedAt, &e.TotalPoints, &qCount); err != nil {
			return nil, err
		}
		if startAt.Valid {
			e.StartAt = &startAt.Time
		}
		if endAt.Valid {
			e.EndAt = &endAt.Time
		}
		list = append(list, e)
	}
	return list, nil
}

// ListForStudent lấy danh sách bài kiểm tra dành cho sinh viên
func (r *Repository) ListForStudent(studentID int) ([]StudentExamSummary, error) {
	query := `SELECT e.id, e.class_id, c.name, co.name, e.title, e.duration_minutes, e.start_at, e.end_at, e.status,
		COALESCE((SELECT COUNT(*) FROM exam_questions WHERE exam_id = e.id), 0) as q_count,
		COALESCE((SELECT SUM(points) FROM exam_questions WHERE exam_id = e.id), 0) as total_points,
		es.status as session_status,
		es.final_score
		FROM exams e
		JOIN enrollments en ON e.class_id = en.class_id
		JOIN classes c ON e.class_id = c.id
		JOIN courses co ON c.course_id = co.id
		LEFT JOIN exam_sessions es ON e.id = es.exam_id AND es.student_id = ?
		WHERE en.student_id = ? AND e.status = 'published'
		ORDER BY e.start_at ASC, e.created_at DESC`

	rows, err := r.db.Query(query, studentID, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	now := time.Now()
	var list []StudentExamSummary
	for rows.Next() {
		var s StudentExamSummary
		var startAt, endAt sql.NullTime
		var sessStatus sql.NullString
		var score sql.NullFloat64

		if err := rows.Scan(&s.ID, &s.ClassID, &s.ClassName, &s.CourseName, &s.Title,
			&s.DurationMinutes, &startAt, &endAt, &s.Status, &s.QuestionCount, &s.TotalPoints,
			&sessStatus, &score); err != nil {
			return nil, err
		}

		if startAt.Valid {
			s.StartAt = &startAt.Time
		}
		if endAt.Valid {
			s.EndAt = &endAt.Time
		}

		s.SessionStatus = "not_started"
		if sessStatus.Valid {
			s.SessionStatus = sessStatus.String
		}
		if score.Valid {
			sc := score.Float64
			s.FinalScore = &sc
		}

		// Kiểm tra thời gian mở ca thi
		isOpen := true
		if s.StartAt != nil && now.Before(*s.StartAt) {
			isOpen = false
		}
		if s.EndAt != nil && now.After(*s.EndAt) {
			isOpen = false
		}
		s.IsOpen = isOpen

		list = append(list, s)
	}
	return list, nil
}

// Publish phát hành bài kiểm tra
func (r *Repository) Publish(id int) error {
	_, err := r.db.Exec(`UPDATE exams SET status = 'published' WHERE id = ?`, id)
	return err
}

// GetOrCreateSession lấy phiên thi hiện tại hoặc tạo mới khi sinh viên bấm vào thi
func (r *Repository) GetOrCreateSession(examID, studentID int) (*ExamSession, error) {
	query := `SELECT id, exam_id, student_id, started_at, submitted_at, status, final_score 
		FROM exam_sessions WHERE exam_id = ? AND student_id = ?`

	var s ExamSession
	var submittedAt sql.NullTime
	err := r.db.QueryRow(query, examID, studentID).Scan(
		&s.ID, &s.ExamID, &s.StudentID, &s.StartedAt, &submittedAt, &s.Status, &s.FinalScore,
	)

	if err == nil {
		if submittedAt.Valid {
			s.SubmittedAt = &submittedAt.Time
		}
		return &s, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	// Tạo mới phiên thi
	now := time.Now()
	insertSQL := `INSERT INTO exam_sessions (exam_id, student_id, started_at, status, final_score) 
		VALUES (?, ?, ?, 'in_progress', 0)`
	res, err := r.db.Exec(insertSQL, examID, studentID, now)
	if err != nil {
		return nil, err
	}

	sid, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	s = ExamSession{
		ID:         int(sid),
		ExamID:     examID,
		StudentID:  studentID,
		StartedAt:  now,
		Status:     SessionInProgress,
		FinalScore: 0,
	}
	return &s, nil
}

// GetSession lấy phiên thi theo ID
func (r *Repository) GetSession(sessionID int) (*ExamSession, error) {
	query := `SELECT s.id, s.exam_id, e.title, s.student_id, u.full_name, s.started_at, s.submitted_at, s.status, s.final_score
		FROM exam_sessions s
		JOIN exams e ON s.exam_id = e.id
		JOIN users u ON s.student_id = u.id
		WHERE s.id = ?`

	var s ExamSession
	var submittedAt sql.NullTime
	err := r.db.QueryRow(query, sessionID).Scan(
		&s.ID, &s.ExamID, &s.ExamTitle, &s.StudentID, &s.StudentName,
		&s.StartedAt, &submittedAt, &s.Status, &s.FinalScore,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if submittedAt.Valid {
		s.SubmittedAt = &submittedAt.Time
	}
	return &s, nil
}

// SaveSessionAnswer lưu nháp code câu hỏi trong phòng thi
func (r *Repository) SaveSessionAnswer(sessionID, exerciseID int, sourceCode string, score float64) error {
	query := `INSERT INTO exam_session_answers (session_id, exercise_id, source_code, score, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(session_id, exercise_id) DO UPDATE SET 
		source_code = excluded.source_code,
		score = excluded.score,
		updated_at = CURRENT_TIMESTAMP`
	_, err := r.db.Exec(query, sessionID, exerciseID, sourceCode, score)
	return err
}

// GetSessionAnswers tải toàn bộ code các câu đã làm của phiên thi
func (r *Repository) GetSessionAnswers(sessionID int) (map[int]string, error) {
	query := `SELECT exercise_id, source_code FROM exam_session_answers WHERE session_id = ?`
	rows, err := r.db.Query(query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	answers := make(map[int]string)
	for rows.Next() {
		var exID int
		var code string
		if err := rows.Scan(&exID, &code); err != nil {
			return nil, err
		}
		answers[exID] = code
	}
	return answers, nil
}

// SubmitSession cập nhật nộp bài và khóa phiên thi
func (r *Repository) SubmitSession(sessionID int, finalScore float64) error {
	now := time.Now()
	query := `UPDATE exam_sessions SET status = 'submitted', submitted_at = ?, final_score = ? WHERE id = ?`
	_, err := r.db.Exec(query, now, finalScore, sessionID)
	return err
}

// RecordEvent ghi nhận sự kiện giám sát phòng thi
func (r *Repository) RecordEvent(sessionID int, eventType, eventData string) error {
	query := `INSERT INTO exam_events (session_id, event_type, event_data, created_at) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, sessionID, eventType, eventData, time.Now())
	return err
}

// GetExamMonitoringReport tổng hợp báo cáo giám sát tất cả thí sinh trong kỳ thi
func (r *Repository) GetExamMonitoringReport(examID int) ([]StudentMonitoringSummary, error) {
	query := `SELECT s.id, s.student_id, u.full_name, s.status, s.final_score
		FROM exam_sessions s
		JOIN users u ON s.student_id = u.id
		WHERE s.exam_id = ?
		ORDER BY s.started_at ASC`

	rows, err := r.db.Query(query, examID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []StudentMonitoringSummary
	for rows.Next() {
		var sum StudentMonitoringSummary
		if err := rows.Scan(&sum.SessionID, &sum.StudentID, &sum.StudentName, &sum.SessionStatus, &sum.FinalScore); err != nil {
			return nil, err
		}
		summaries = append(summaries, sum)
	}
	rows.Close()

	// Tổng hợp số lượng vi phạm cho từng phiên
	for i := range summaries {
		sID := summaries[i].SessionID

		countsQuery := `SELECT event_type, COUNT(*) FROM exam_events WHERE session_id = ? GROUP BY event_type`
		cRows, err := r.db.Query(countsQuery, sID)
		if err == nil {
			for cRows.Next() {
				var evType string
				var count int
				if err := cRows.Scan(&evType, &count); err == nil {
					switch evType {
					case "tab_hidden":
						summaries[i].TabHiddenCount = count
					case "window_blur":
						summaries[i].WindowBlurCount = count
					case "fullscreen_exit":
						summaries[i].FullscreenExitCount = count
					case "paste_attempt":
						summaries[i].PasteAttemptCount = count
					case "copy_attempt":
						summaries[i].CopyAttemptCount = count
					}
					summaries[i].TotalWarnings += count
				}
			}
			cRows.Close()
		}

		// Lấy tối đa 15 sự kiện gần nhất
		eventsQuery := `SELECT id, session_id, event_type, COALESCE(event_data, ''), created_at 
			FROM exam_events WHERE session_id = ? ORDER BY created_at DESC LIMIT 15`
		eRows, err := r.db.Query(eventsQuery, sID)
		if err == nil {
			var events []ExamEvent
			for eRows.Next() {
				var ev ExamEvent
				if err := eRows.Scan(&ev.ID, &ev.SessionID, &ev.EventType, &ev.EventData, &ev.CreatedAt); err == nil {
					events = append(events, ev)
				}
			}
			eRows.Close()
			summaries[i].Events = events
		}
	}

	return summaries, nil
}

// IsStudentEnrolledInExam kiểm tra xem học viên có thuộc lớp được phân công bài thi hay không
func (r *Repository) IsStudentEnrolledInExam(examID, studentID int) (bool, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) 
		FROM exams e
		JOIN enrollments en ON e.class_id = en.class_id
		WHERE e.id = ? AND en.student_id = ?`, examID, studentID).Scan(&count)
	return count > 0, err
}

