package assignment

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

// CreateAssignment tạo mới đợt bài tập và liên kết danh sách câu hỏi
func (r *Repository) CreateAssignment(a *Assignment, exerciseIDs []int, points []float64) (*Assignment, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now()
	query := `INSERT INTO assignments (class_id, title, description, start_at, due_at, status, created_by, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := tx.Exec(query, a.ClassID, a.Title, a.Description, a.StartAt, a.DueAt, a.Status, a.CreatedBy, now)
	if err != nil {
		return nil, err
	}

	assignmentID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	a.ID = int(assignmentID)
	a.CreatedAt = now

	insertExSQL := `INSERT INTO assignment_exercises (assignment_id, exercise_id, points, order_num) VALUES (?, ?, ?, ?)`
	for i, exID := range exerciseIDs {
		p := 1.0
		if i < len(points) && points[i] > 0 {
			p = points[i]
		}
		if _, err := tx.Exec(insertExSQL, a.ID, exID, p, i+1); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return a, nil
}

// FindByID lấy thông tin chi tiết assignment kèm danh sách câu hỏi
func (r *Repository) FindByID(id int) (*Assignment, error) {
	query := `SELECT a.id, a.class_id, c.name, co.code, co.name, a.title, a.description, 
		a.start_at, a.due_at, a.status, a.created_by, a.created_at 
		FROM assignments a 
		JOIN classes c ON a.class_id = c.id 
		JOIN courses co ON c.course_id = co.id 
		WHERE a.id = ?`

	var a Assignment
	var startAt, dueAt sql.NullTime
	err := r.db.QueryRow(query, id).Scan(
		&a.ID, &a.ClassID, &a.ClassName, &a.CourseCode, &a.CourseName,
		&a.Title, &a.Description, &startAt, &dueAt, &a.Status,
		&a.CreatedBy, &a.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if startAt.Valid {
		a.StartAt = &startAt.Time
	}
	if dueAt.Valid {
		a.DueAt = &dueAt.Time
	}

	// Lấy danh sách câu hỏi
	exQuery := `SELECT ae.id, ae.assignment_id, ae.exercise_id, e.title, e.difficulty, ae.points, ae.order_num 
		FROM assignment_exercises ae 
		JOIN exercises e ON ae.exercise_id = e.id 
		WHERE ae.assignment_id = ? 
		ORDER BY ae.order_num ASC, ae.id ASC`
	rows, err := r.db.Query(exQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []AssignmentExercise
	var totalPoints float64
	for rows.Next() {
		var ae AssignmentExercise
		if err := rows.Scan(&ae.ID, &ae.AssignmentID, &ae.ExerciseID, &ae.ExerciseTitle, &ae.Difficulty, &ae.Points, &ae.OrderNum); err != nil {
			return nil, err
		}
		totalPoints += ae.Points
		exercises = append(exercises, ae)
	}
	a.Exercises = exercises
	a.TotalPoints = totalPoints

	return &a, nil
}

// ListByTeacher lấy danh sách bài tập do giảng viên tạo
func (r *Repository) ListByTeacher(teacherID int) ([]Assignment, error) {
	query := `SELECT a.id, a.class_id, c.name, co.code, co.name, a.title, a.description, 
		a.start_at, a.due_at, a.status, a.created_by, a.created_at,
		COALESCE((SELECT SUM(points) FROM assignment_exercises WHERE assignment_id = a.id), 0) as total_points,
		COALESCE((SELECT COUNT(*) FROM assignment_exercises WHERE assignment_id = a.id), 0) as ex_count
		FROM assignments a 
		JOIN classes c ON a.class_id = c.id 
		JOIN courses co ON c.course_id = co.id 
		WHERE a.created_by = ? 
		ORDER BY a.created_at DESC`

	rows, err := r.db.Query(query, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Assignment
	for rows.Next() {
		var a Assignment
		var startAt, dueAt sql.NullTime
		var exCount int
		if err := rows.Scan(&a.ID, &a.ClassID, &a.ClassName, &a.CourseCode, &a.CourseName,
			&a.Title, &a.Description, &startAt, &dueAt, &a.Status,
			&a.CreatedBy, &a.CreatedAt, &a.TotalPoints, &exCount); err != nil {
			return nil, err
		}
		if startAt.Valid {
			a.StartAt = &startAt.Time
		}
		if dueAt.Valid {
			a.DueAt = &dueAt.Time
		}
		list = append(list, a)
	}
	return list, nil
}

// ListForStudent lấy danh sách bài tập đã phát hành cho học viên theo các lớp học đang tham gia
func (r *Repository) ListForStudent(studentID int) ([]StudentAssignmentSummary, error) {
	query := `SELECT a.id, a.class_id, c.name, co.name, a.title, a.description, a.due_at, a.status,
		COALESCE((SELECT COUNT(*) FROM assignment_exercises WHERE assignment_id = a.id), 0) as ex_count,
		COALESCE((SELECT SUM(points) FROM assignment_exercises WHERE assignment_id = a.id), 0) as total_points
		FROM assignments a
		JOIN enrollments e ON a.class_id = e.class_id
		JOIN classes c ON a.class_id = c.id
		JOIN courses co ON c.course_id = co.id
		WHERE e.student_id = ? AND a.status = 'published'
		ORDER BY a.due_at ASC, a.created_at DESC`

	rows, err := r.db.Query(query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	now := time.Now()
	var list []StudentAssignmentSummary
	for rows.Next() {
		var s StudentAssignmentSummary
		var dueAt sql.NullTime
		if err := rows.Scan(&s.ID, &s.ClassID, &s.ClassName, &s.CourseName, &s.Title, &s.Description,
			&dueAt, &s.Status, &s.ExerciseCount, &s.TotalPoints); err != nil {
			return nil, err
		}
		if dueAt.Valid {
			s.DueAt = &dueAt.Time
			// Gần đến hạn nếu còn ít hơn 3 ngày và chưa qua hạn
			diff := dueAt.Time.Sub(now)
			if diff > 0 && diff < 72*time.Hour {
				s.IsUrgent = true
			}
		}
		list = append(list, s)
	}
	return list, nil
}

// Publish công bố bài tập để sinh viên có thể xem
func (r *Repository) Publish(id int) error {
	_, err := r.db.Exec(`UPDATE assignments SET status = 'published' WHERE id = ?`, id)
	return err
}
