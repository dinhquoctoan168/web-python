package dashboard

import (
	"database/sql"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetEnrolledCoursesProgress lấy tiến độ của tất cả môn học mà học viên đã ghi danh
func (r *Repository) GetEnrolledCoursesProgress(studentID int) ([]CourseProgressItem, error) {
	queryCourses := `SELECT DISTINCT c.id, c.code, c.name 
		FROM courses c 
		JOIN classes cl ON c.id = cl.course_id 
		JOIN enrollments ce ON cl.id = ce.class_id 
		WHERE ce.student_id = ? 
		ORDER BY c.id ASC`

	rows, err := r.db.Query(queryCourses, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []CourseProgressItem
	for rows.Next() {
		var item CourseProgressItem
		if err := rows.Scan(&item.CourseID, &item.CourseCode, &item.CourseName); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range items {
		// Tổng số bài tập trong môn học
		var total int
		err := r.db.QueryRow(`SELECT COUNT(*) FROM exercises WHERE course_id = ? AND status = 'active'`, items[i].CourseID).Scan(&total)
		if err != nil {
			total = 0
		}
		items[i].TotalExercises = total

		// Số bài tập học viên đã hoàn thành (status = 'completed')
		var completed int
		err = r.db.QueryRow(`SELECT COUNT(DISTINCT p.exercise_id) 
			FROM student_exercise_progress p 
			JOIN exercises e ON p.exercise_id = e.id 
			WHERE p.student_id = ? AND e.course_id = ? AND p.status = 'completed'`, studentID, items[i].CourseID).Scan(&completed)
		if err != nil {
			completed = 0
		}
		items[i].CompletedExercises = completed

		if total > 0 {
			items[i].ProgressPercent = (completed * 100) / total
			if items[i].ProgressPercent > 100 {
				items[i].ProgressPercent = 100
			}
		} else {
			items[i].ProgressPercent = 0
		}
	}

	return items, nil
}

// GetUpcomingAssignments lấy các bài tập sắp đến hạn của các lớp học viên tham gia
func (r *Repository) GetUpcomingAssignments(studentID int) ([]UpcomingItem, error) {
	query := `SELECT a.id, a.title, c.name as course_name, cl.name as class_name, 
		COALESCE(strftime('%d/%m/%Y %H:%M', a.due_at), 'Không có hạn') as deadline
		FROM assignments a
		JOIN classes cl ON a.class_id = cl.id
		JOIN courses c ON cl.course_id = c.id
		JOIN enrollments ce ON cl.id = ce.class_id
		WHERE ce.student_id = ? AND a.status IN ('published', 'active')
		  AND (a.due_at IS NULL OR a.due_at >= datetime('now', 'localtime'))
		ORDER BY a.due_at ASC LIMIT 5`

	rows, err := r.db.Query(query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []UpcomingItem
	for rows.Next() {
		var item UpcomingItem
		item.Type = "assignment"
		item.Link = "/my-assignments"
		if err := rows.Scan(&item.ID, &item.Title, &item.CourseName, &item.ClassName, &item.Deadline); err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	return list, rows.Err()
}

// GetUpcomingExams lấy các ca thi đang mở hoặc sắp diễn ra của các lớp học viên tham gia
func (r *Repository) GetUpcomingExams(studentID int) ([]UpcomingItem, error) {
	query := `SELECT e.id, e.title, c.name as course_name, cl.name as class_name, 
		COALESCE(strftime('%d/%m/%Y %H:%M', e.start_at), 'Chưa xác định') as deadline
		FROM exams e
		JOIN classes cl ON e.class_id = cl.id
		JOIN courses c ON cl.course_id = c.id
		JOIN enrollments ce ON cl.id = ce.class_id
		WHERE ce.student_id = ? AND e.status IN ('published', 'ongoing')
		  AND (e.end_at IS NULL OR e.end_at >= datetime('now', 'localtime'))
		ORDER BY e.start_at ASC LIMIT 5`

	rows, err := r.db.Query(query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []UpcomingItem
	for rows.Next() {
		var item UpcomingItem
		item.Type = "exam"
		item.Link = "/student/exams"
		if err := rows.Scan(&item.ID, &item.Title, &item.CourseName, &item.ClassName, &item.Deadline); err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	return list, rows.Err()
}

// GetRecentSubmissions lấy danh sách các lần nộp bài gần đây nhất của học viên
func (r *Repository) GetRecentSubmissions(studentID, limit int) ([]RecentSubmissionItem, error) {
	if limit <= 0 {
		limit = 5
	}
	query := `SELECT s.id, s.exercise_id, e.title, s.score, s.status, 
		COALESCE(strftime('%d/%m/%Y %H:%M', s.submitted_at), '') as submitted_at
		FROM submissions s
		JOIN exercises e ON s.exercise_id = e.id
		WHERE s.student_id = ?
		ORDER BY s.id DESC LIMIT ?`

	rows, err := r.db.Query(query, studentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []RecentSubmissionItem
	for rows.Next() {
		var item RecentSubmissionItem
		if err := rows.Scan(&item.ID, &item.ExerciseID, &item.ExerciseTitle, &item.Score, &item.Status, &item.SubmittedAt); err != nil {
			return nil, err
		}
		list = append(list, item)
	}

	return list, rows.Err()
}
