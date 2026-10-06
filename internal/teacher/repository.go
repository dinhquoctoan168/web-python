package teacher

import (
	"database/sql"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetDashboardSummary lấy các chỉ số KPI tổng quan cho giảng viên
func (r *Repository) GetDashboardSummary(teacherID int, isAdmin bool) (TeacherDashboardSummary, error) {
	var s TeacherDashboardSummary

	var err error
	if isAdmin {
		_ = r.db.QueryRow(`SELECT COUNT(*) FROM courses`).Scan(&s.TotalCourses)
		_ = r.db.QueryRow(`SELECT COUNT(*) FROM classes`).Scan(&s.TotalClasses)
		_ = r.db.QueryRow(`SELECT COUNT(DISTINCT student_id) FROM enrollments`).Scan(&s.TotalStudents)
		_ = r.db.QueryRow(`SELECT COUNT(*) FROM assignments`).Scan(&s.TotalAssignments)
		_ = r.db.QueryRow(`SELECT COUNT(*) FROM exams`).Scan(&s.TotalExams)
	} else {
		_ = r.db.QueryRow(`SELECT COUNT(DISTINCT course_id) FROM classes WHERE teacher_id = ?`, teacherID).Scan(&s.TotalCourses)
		_ = r.db.QueryRow(`SELECT COUNT(*) FROM classes WHERE teacher_id = ?`, teacherID).Scan(&s.TotalClasses)
		_ = r.db.QueryRow(`SELECT COUNT(DISTINCT e.student_id) FROM enrollments e JOIN classes c ON e.class_id = c.id WHERE c.teacher_id = ?`, teacherID).Scan(&s.TotalStudents)
		_ = r.db.QueryRow(`SELECT COUNT(DISTINCT a.id) FROM assignments a JOIN classes c ON a.class_id = c.id WHERE c.teacher_id = ?`, teacherID).Scan(&s.TotalAssignments)
		_ = r.db.QueryRow(`SELECT COUNT(DISTINCT e.id) FROM exams e JOIN classes c ON e.class_id = c.id WHERE c.teacher_id = ?`, teacherID).Scan(&s.TotalExams)
	}

	return s, err
}

// ListClassesSummary lấy danh sách lớp học và tỷ lệ hoàn thành trung bình
func (r *Repository) ListClassesSummary(teacherID int, isAdmin bool) ([]ClassSummaryItem, error) {
	query := `SELECT cl.id, cl.course_id, co.code, co.name, cl.name, 
		COALESCE(cl.semester, ''), COALESCE(cl.academic_year, ''),
		(SELECT COUNT(*) FROM enrollments WHERE class_id = cl.id) as student_count
		FROM classes cl
		JOIN courses co ON cl.course_id = co.id
		WHERE (? OR cl.teacher_id = ?)
		ORDER BY cl.id DESC`

	rows, err := r.db.Query(query, isAdmin, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ClassSummaryItem
	for rows.Next() {
		var item ClassSummaryItem
		if err := rows.Scan(&item.ID, &item.CourseID, &item.CourseCode, &item.CourseName,
			&item.Name, &item.Semester, &item.AcademicYear, &item.StudentCount); err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range list {
		var totalExercises int
		_ = r.db.QueryRow(`SELECT COUNT(*) FROM exercises WHERE course_id = ? AND status = 'active'`, list[i].CourseID).Scan(&totalExercises)

		if list[i].StudentCount > 0 && totalExercises > 0 {
			var totalCompleted int
			_ = r.db.QueryRow(`SELECT COUNT(*) 
				FROM student_exercise_progress sep 
				JOIN exercises ex ON sep.exercise_id = ex.id 
				JOIN enrollments en ON sep.student_id = en.student_id 
				WHERE en.class_id = ? AND ex.course_id = ? AND sep.status = 'completed'`, list[i].ID, list[i].CourseID).Scan(&totalCompleted)

			possibleTotal := list[i].StudentCount * totalExercises
			list[i].CompletionRate = (totalCompleted * 100) / possibleTotal
			if list[i].CompletionRate > 100 {
				list[i].CompletionRate = 100
			}
		} else {
			list[i].CompletionRate = 0
		}
	}

	return list, nil
}

// GetClassAnalytics truy vấn toàn bộ dữ liệu phân tích chi tiết của một lớp học
func (r *Repository) GetClassAnalytics(classID int) (*ClassAnalyticsData, error) {
	queryClass := `SELECT cl.id, cl.name, cl.course_id, co.code, co.name, 
		COALESCE(cl.semester, ''), COALESCE(cl.academic_year, ''),
		(SELECT COUNT(*) FROM enrollments WHERE class_id = cl.id) as student_count
		FROM classes cl
		JOIN courses co ON cl.course_id = co.id
		WHERE cl.id = ?`

	var d ClassAnalyticsData
	err := r.db.QueryRow(queryClass, classID).Scan(
		&d.ClassID, &d.ClassName, &d.CourseID, &d.CourseCode, &d.CourseName,
		&d.Semester, &d.AcademicYear, &d.StudentCount,
	)
	if err != nil {
		return nil, err
	}

	// 1. Phân tích theo từng chương (Chapters)
	queryChapters := `SELECT ch.id, ch.title,
		(SELECT COUNT(*) FROM exercises e JOIN lessons l ON e.lesson_id = l.id WHERE l.chapter_id = ch.id AND e.status = 'active') as ex_count
		FROM chapters ch
		WHERE ch.course_id = ?
		ORDER BY ch.order_num ASC`

	chRows, err := r.db.Query(queryChapters, d.CourseID)
	if err == nil {
		var rawChapters []ChapterCompletionStat
		for chRows.Next() {
			var ch ChapterCompletionStat
			if err := chRows.Scan(&ch.ChapterID, &ch.ChapterTitle, &ch.TotalExercises); err == nil {
				rawChapters = append(rawChapters, ch)
			}
		}
		chRows.Close()

		for _, ch := range rawChapters {
			if d.StudentCount > 0 && ch.TotalExercises > 0 {
				var chDone int
				_ = r.db.QueryRow(`SELECT COUNT(*) 
					FROM student_exercise_progress sep 
					JOIN exercises ex ON sep.exercise_id = ex.id 
					JOIN lessons ls ON ex.lesson_id = ls.id 
					JOIN enrollments en ON sep.student_id = en.student_id 
					WHERE en.class_id = ? AND ls.chapter_id = ? AND sep.status = 'completed'`, classID, ch.ChapterID).Scan(&chDone)

				ch.CompletionRate = (chDone * 100) / (d.StudentCount * ch.TotalExercises)
				if ch.CompletionRate > 100 {
					ch.CompletionRate = 100
				}
			}
			d.ChapterStats = append(d.ChapterStats, ch)
		}
	}

	// 2. Thống kê từng sinh viên trong lớp
	queryStudents := `SELECT u.id, u.username, u.full_name
		FROM users u
		JOIN enrollments en ON u.id = en.student_id
		WHERE en.class_id = ?
		ORDER BY u.full_name ASC`

	stRows, err := r.db.Query(queryStudents, classID)
	if err != nil {
		return nil, err
	}
	type stBasic struct {
		id       int
		username string
		fullName string
	}
	var rawStudents []stBasic
	for stRows.Next() {
		var s stBasic
		if err := stRows.Scan(&s.id, &s.username, &s.fullName); err == nil {
			rawStudents = append(rawStudents, s)
		}
	}
	stRows.Close()

	var totalCourseExercises int
	_ = r.db.QueryRow(`SELECT COUNT(*) FROM exercises WHERE course_id = ? AND status = 'active'`, d.CourseID).Scan(&totalCourseExercises)

	for _, s := range rawStudents {
		row := StudentProgressRow{
			StudentID: s.id,
			Username:  s.username,
			FullName:  s.fullName,
		}

		// Số bài hoàn thành
		var completedCount int
		_ = r.db.QueryRow(`SELECT COUNT(DISTINCT sep.exercise_id) 
			FROM student_exercise_progress sep 
			JOIN exercises ex ON sep.exercise_id = ex.id 
			WHERE sep.student_id = ? AND ex.course_id = ? AND sep.status = 'completed'`, row.StudentID, d.CourseID).Scan(&completedCount)

		if totalCourseExercises > 0 {
			row.ProgressPercent = (completedCount * 100) / totalCourseExercises
			row.IncompleteExercises = totalCourseExercises - completedCount
		}

		// Điểm trung bình từ submissions
		var avgScore sql.NullFloat64
		_ = r.db.QueryRow(`SELECT AVG(s.score) 
			FROM submissions s 
			JOIN exercises ex ON s.exercise_id = ex.id 
			WHERE s.student_id = ? AND ex.course_id = ?`, row.StudentID, d.CourseID).Scan(&avgScore)
		if avgScore.Valid {
			row.AverageScore = avgScore.Float64
		}

		// Hoạt động gần nhất
		var lastAct sql.NullString
		_ = r.db.QueryRow(`SELECT COALESCE(
				(SELECT strftime('%d/%m/%Y %H:%M', MAX(submitted_at)) FROM submissions WHERE student_id = ?),
				(SELECT strftime('%d/%m/%Y %H:%M', MAX(updated_at)) FROM student_exercise_progress WHERE student_id = ?)
			)`, row.StudentID, row.StudentID).Scan(&lastAct)
		if lastAct.Valid && lastAct.String != "" {
			row.LastActivity = lastAct.String
		} else {
			row.LastActivity = "Chưa có"
		}

		// Nhận diện sinh viên đang gặp khó khăn (attempts >= 3 hoặc hint >= 2 mà chưa xong)
		var struggleCount int
		_ = r.db.QueryRow(`SELECT COUNT(*) 
			FROM student_exercise_progress sep 
			JOIN exercises ex ON sep.exercise_id = ex.id 
			WHERE sep.student_id = ? AND ex.course_id = ? AND sep.status != 'completed' AND sep.attempts >= 3`, row.StudentID, d.CourseID).Scan(&struggleCount)
		row.HasStruggling = (struggleCount > 0)

		d.Students = append(d.Students, row)
	}

	// Tính tiến độ chung của cả lớp
	if len(d.Students) > 0 {
		sum := 0
		for _, s := range d.Students {
			sum += s.ProgressPercent
		}
		d.OverallProgress = sum / len(d.Students)
	}

	return &d, nil
}

// GetStudentExerciseDiagnostics lấy phân tích từng bài tập của sinh viên để phát hiện bài khó
func (r *Repository) GetStudentExerciseDiagnostics(studentID, courseID int) ([]StudentExerciseDiagnostic, error) {
	queryEx := `SELECT e.id, e.title, e.difficulty
		FROM exercises e
		WHERE e.course_id = ? AND e.status = 'active'
		ORDER BY e.id ASC`

	rows, err := r.db.Query(queryEx, courseID)
	if err != nil {
		return nil, err
	}
	type exBasic struct {
		id         int
		title      string
		difficulty string
	}
	var rawEx []exBasic
	for rows.Next() {
		var e exBasic
		if err := rows.Scan(&e.id, &e.title, &e.difficulty); err == nil {
			rawEx = append(rawEx, e)
		}
	}
	rows.Close()

	var list []StudentExerciseDiagnostic
	for _, e := range rawEx {
		item := StudentExerciseDiagnostic{
			ExerciseID:    e.id,
			ExerciseTitle: e.title,
			Difficulty:    e.difficulty,
		}

		// Lấy trạng thái từ student_exercise_progress
		var status, lastUp sql.NullString
		var score sql.NullFloat64
		var attempts int
		_ = r.db.QueryRow(`SELECT status, best_score, attempts, strftime('%d/%m/%Y %H:%M', updated_at)
			FROM student_exercise_progress 
			WHERE student_id = ? AND exercise_id = ?`, studentID, item.ExerciseID).Scan(&status, &score, &attempts, &lastUp)

		if status.Valid {
			item.Status = status.String
		} else {
			item.Status = "not_started"
		}
		if score.Valid {
			item.Score = score.Float64
		}
		item.Attempts = attempts
		if lastUp.Valid {
			item.LastActivity = lastUp.String
		}

		// Lấy metrics từ exercise_attempts (run, test, hint)
		var runCount, testCount, hintCount int
		_ = r.db.QueryRow(`SELECT CAST(COALESCE(SUM(run_count), 0) AS INTEGER), CAST(COALESCE(SUM(test_count), 0) AS INTEGER), CAST(COALESCE(SUM(hint_count), 0) AS INTEGER)
			FROM exercise_attempts
			WHERE student_id = ? AND exercise_id = ?`, studentID, item.ExerciseID).Scan(&runCount, &testCount, &hintCount)

		item.RunCount = runCount
		item.TestCount = testCount
		item.HintCount = hintCount

		// Đánh giá sinh viên có đang gặp khó khăn ở bài này không
		if item.Attempts >= 4 && item.Score < 100 {
			item.IsStruggling = true
			item.StruggleReason = fmt.Sprintf("Thử %d lần nhưng chưa đạt điểm tuyệt đối", item.Attempts)
		} else if item.HintCount >= 2 && item.Status != "completed" {
			item.IsStruggling = true
			item.StruggleReason = fmt.Sprintf("Đã xem gợi ý %d lần nhưng chưa giải được", item.HintCount)
		} else if item.RunCount >= 10 && item.Status != "completed" {
			item.IsStruggling = true
			item.StruggleReason = fmt.Sprintf("Chạy thử %d lần nhưng chưa vượt qua test cases", item.RunCount)
		}

		list = append(list, item)
	}

	return list, nil
}
