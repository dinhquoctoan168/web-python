package progress

import (
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrLessonNotFound  = errors.New("không tìm thấy bài học")
	ErrChapterNotFound = errors.New("không tìm thấy chương mục")
	ErrCourseNotFound  = errors.New("không tìm thấy môn học")
)

// Repository quản lý truy vấn tiến độ học tập từ CSDL
type Repository struct {
	db *sql.DB
}

// NewRepository khởi tạo Repository tiến độ
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetLessonProgress tính toán tiến độ một bài học cho học viên
func (r *Repository) GetLessonProgress(studentID, lessonID int) (*LessonProgress, error) {
	var lp LessonProgress
	err := r.db.QueryRow(`
		SELECT id, chapter_id, title 
		FROM lessons 
		WHERE id = ?`, lessonID).Scan(&lp.LessonID, &lp.ChapterID, &lp.Title)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLessonNotFound
		}
		return nil, fmt.Errorf("lỗi truy vấn bài học: %w", err)
	}

	query := `
		SELECT e.id, e.title, e.exercise_type, 
		       COALESCE(sep.status, 'not_started'), 
		       COALESCE(sep.best_score, 0)
		FROM exercises e
		LEFT JOIN student_exercise_progress sep 
		       ON e.id = sep.exercise_id AND sep.student_id = ?
		WHERE e.lesson_id = ? AND e.status = 'active'
		ORDER BY e.id ASC`

	rows, err := r.db.Query(query, studentID, lessonID)
	if err != nil {
		return nil, fmt.Errorf("lỗi truy vấn bài tập bài học: %w", err)
	}

	var items []ExerciseProgressItem
	for rows.Next() {
		var item ExerciseProgressItem
		if err := rows.Scan(&item.ID, &item.Title, &item.Type, &item.Status, &item.BestScore); err != nil {
			rows.Close()
			return nil, fmt.Errorf("lỗi scan bài tập bài học: %w", err)
		}
		items = append(items, item)
	}
	rows.Close()

	lp.Exercises = items
	lp.TotalExercises = len(items)

	completedCount := 0
	for _, item := range items {
		if item.Status == "completed" {
			completedCount++
		}
	}
	lp.CompletedExercises = completedCount

	if lp.TotalExercises > 0 {
		lp.ProgressPercent = (completedCount * 100) / lp.TotalExercises
		if lp.ProgressPercent > 100 {
			lp.ProgressPercent = 100
		}
		lp.IsCompleted = (completedCount == lp.TotalExercises)
	} else {
		// Bài học không có bài tập bắt buộc -> tính là hoàn thành
		lp.ProgressPercent = 100
		lp.IsCompleted = true
	}

	return &lp, nil
}

// GetChapterProgress tính toán tiến độ một chương cho học viên
func (r *Repository) GetChapterProgress(studentID, chapterID int) (*ChapterProgress, error) {
	var cp ChapterProgress
	err := r.db.QueryRow(`
		SELECT id, course_id, title 
		FROM chapters 
		WHERE id = ?`, chapterID).Scan(&cp.ChapterID, &cp.CourseID, &cp.Title)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChapterNotFound
		}
		return nil, fmt.Errorf("lỗi truy vấn chương mục: %w", err)
	}

	rows, err := r.db.Query(`
		SELECT id 
		FROM lessons 
		WHERE chapter_id = ? AND is_published = 1 
		ORDER BY order_num ASC, id ASC`, chapterID)
	if err != nil {
		return nil, fmt.Errorf("lỗi truy vấn danh sách bài học của chương: %w", err)
	}

	var lessonIDs []int
	for rows.Next() {
		var lid int
		if err := rows.Scan(&lid); err != nil {
			rows.Close()
			return nil, fmt.Errorf("lỗi scan bài học id: %w", err)
		}
		lessonIDs = append(lessonIDs, lid)
	}
	rows.Close()

	cp.TotalLessons = len(lessonIDs)
	cp.Lessons = make([]LessonProgress, 0, len(lessonIDs))

	completedLessons := 0
	totalExercises := 0
	completedExercises := 0

	for _, lid := range lessonIDs {
		lp, err := r.GetLessonProgress(studentID, lid)
		if err != nil {
			return nil, err
		}
		cp.Lessons = append(cp.Lessons, *lp)
		if lp.IsCompleted {
			completedLessons++
		}
		totalExercises += lp.TotalExercises
		completedExercises += lp.CompletedExercises
	}

	cp.CompletedLessons = completedLessons
	cp.TotalExercises = totalExercises
	cp.CompletedExercises = completedExercises

	// Chapter progress: completed lessons / total lessons
	if cp.TotalLessons > 0 {
		cp.ProgressPercent = (completedLessons * 100) / cp.TotalLessons
		if cp.ProgressPercent > 100 {
			cp.ProgressPercent = 100
		}
		cp.IsCompleted = (completedLessons == cp.TotalLessons)
	} else {
		cp.ProgressPercent = 0
		cp.IsCompleted = false
	}

	return &cp, nil
}

// GetCourseProgress tính toán tiến độ tổng thể môn học cho học viên
func (r *Repository) GetCourseProgress(studentID, courseID int) (*CourseProgress, error) {
	var cp CourseProgress
	err := r.db.QueryRow(`
		SELECT id, name 
		FROM courses 
		WHERE id = ?`, courseID).Scan(&cp.CourseID, &cp.CourseName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCourseNotFound
		}
		return nil, fmt.Errorf("lỗi truy vấn môn học: %w", err)
	}

	rows, err := r.db.Query(`
		SELECT id 
		FROM chapters 
		WHERE course_id = ? 
		ORDER BY order_num ASC, id ASC`, courseID)
	if err != nil {
		return nil, fmt.Errorf("lỗi truy vấn danh sách chương môn học: %w", err)
	}

	var chapterIDs []int
	for rows.Next() {
		var cid int
		if err := rows.Scan(&cid); err != nil {
			rows.Close()
			return nil, fmt.Errorf("lỗi scan chapter id: %w", err)
		}
		chapterIDs = append(chapterIDs, cid)
	}
	rows.Close()

	cp.Chapters = make([]ChapterProgress, 0, len(chapterIDs))

	totalLessons := 0
	completedLessons := 0
	totalExercisesInChapters := 0
	completedExercisesInChapters := 0

	for _, cid := range chapterIDs {
		chp, err := r.GetChapterProgress(studentID, cid)
		if err != nil {
			return nil, err
		}
		cp.Chapters = append(cp.Chapters, *chp)
		totalLessons += chp.TotalLessons
		completedLessons += chp.CompletedLessons
		totalExercisesInChapters += chp.TotalExercises
		completedExercisesInChapters += chp.CompletedExercises
	}

	cp.TotalLessons = totalLessons
	cp.CompletedLessons = completedLessons

	// Đếm tổng số bài tập hoạt động trong course (bao gồm cả bài tập không thuộc chapter/lesson nếu có)
	var totalActivities int
	err = r.db.QueryRow(`
		SELECT COUNT(*) 
		FROM exercises 
		WHERE course_id = ? AND status = 'active'`, courseID).Scan(&totalActivities)
	if err != nil {
		totalActivities = totalExercisesInChapters
	}

	var completedActivities int
	err = r.db.QueryRow(`
		SELECT COUNT(DISTINCT sep.exercise_id)
		FROM student_exercise_progress sep
		JOIN exercises e ON sep.exercise_id = e.id
		WHERE sep.student_id = ? AND e.course_id = ? AND e.status = 'active' AND sep.status = 'completed'`,
		studentID, courseID).Scan(&completedActivities)
	if err != nil {
		completedActivities = completedExercisesInChapters
	}

	cp.TotalActivities = totalActivities
	cp.CompletedActivities = completedActivities

	// Course rule: completed required activities / total required activities
	if totalActivities > 0 {
		cp.ProgressPercent = (completedActivities * 100) / totalActivities
		if cp.ProgressPercent > 100 {
			cp.ProgressPercent = 100
		}
		cp.IsCompleted = (completedActivities == totalActivities)
	} else if totalLessons > 0 {
		cp.ProgressPercent = (completedLessons * 100) / totalLessons
		if cp.ProgressPercent > 100 {
			cp.ProgressPercent = 100
		}
		cp.IsCompleted = (completedLessons == totalLessons)
	} else {
		cp.ProgressPercent = 0
		cp.IsCompleted = false
	}

	return &cp, nil
}
