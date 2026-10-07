package lesson

import (
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrChapterNotFound = errors.New("không tìm thấy chương mục")
	ErrLessonNotFound  = errors.New("không tìm thấy bài học")
)

// Repository quản lý các truy vấn CSDL cho chương mục và bài học
type Repository struct {
	db *sql.DB
}

// NewRepository khởi tạo Repository bài học
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetCurriculum tải toàn bộ cây chương mục và bài học của môn học
func (r *Repository) GetCurriculum(courseID int, onlyPublished bool) (*CourseCurriculum, error) {
	row := r.db.QueryRow("SELECT id, code, name, COALESCE(description, '') FROM courses WHERE id = ?", courseID)
	var curr CourseCurriculum
	if err := row.Scan(&curr.CourseID, &curr.CourseCode, &curr.CourseName, &curr.Description); err != nil {
		return nil, fmt.Errorf("không tìm thấy môn học: %w", err)
	}

	chapters, err := r.ListChaptersByCourse(courseID)
	if err != nil {
		return nil, err
	}

	for i := range chapters {
		lessons, err := r.ListLessonsByChapter(chapters[i].ID, onlyPublished)
		if err != nil {
			return nil, err
		}
		chapters[i].Lessons = lessons
	}

	curr.Chapters = chapters
	return &curr, nil
}

// ListChaptersByCourse lấy danh sách chương của môn học
func (r *Repository) ListChaptersByCourse(courseID int) ([]Chapter, error) {
	rows, err := r.db.Query(`
		SELECT id, course_id, title, COALESCE(description, ''), order_num
		FROM chapters
		WHERE course_id = ?
		ORDER BY order_num ASC, id ASC
	`, courseID)
	if err != nil {
		return nil, fmt.Errorf("truy vấn chương mục thất bại: %w", err)
	}
	defer rows.Close()

	var list []Chapter
	for rows.Next() {
		var c Chapter
		if err := rows.Scan(&c.ID, &c.CourseID, &c.Title, &c.Description, &c.OrderNum); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

// FindChapterByID tìm chương theo ID
func (r *Repository) FindChapterByID(id int) (*Chapter, error) {
	row := r.db.QueryRow(`
		SELECT id, course_id, title, COALESCE(description, ''), order_num
		FROM chapters
		WHERE id = ?
	`, id)

	var c Chapter
	err := row.Scan(&c.ID, &c.CourseID, &c.Title, &c.Description, &c.OrderNum)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrChapterNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("truy vấn chương thất bại: %w", err)
	}
	return &c, nil
}

// CreateChapter tạo chương mới
func (r *Repository) CreateChapter(c *Chapter) error {
	res, err := r.db.Exec(`
		INSERT INTO chapters (course_id, title, description, order_num)
		VALUES (?, ?, ?, ?)
	`, c.CourseID, c.Title, c.Description, c.OrderNum)
	if err != nil {
		return fmt.Errorf("thêm chương thất bại: %w", err)
	}
	id, err := res.LastInsertId()
	if err == nil {
		c.ID = int(id)
	}
	return nil
}

// ListLessonsByChapter lấy danh sách bài học trong chương
func (r *Repository) ListLessonsByChapter(chapterID int, onlyPublished bool) ([]Lesson, error) {
	query := `
		SELECT l.id, l.chapter_id, ch.title, ch.course_id, c.code, c.name,
		       l.title, COALESCE(l.content_html, ''), COALESCE(l.visualization_type, ''), COALESCE(l.visualization_config, ''), l.order_num, l.is_published, l.created_at
		FROM lessons l
		JOIN chapters ch ON l.chapter_id = ch.id
		JOIN courses c ON ch.course_id = c.id
		WHERE l.chapter_id = ?
	`
	if onlyPublished {
		query += " AND l.is_published = 1"
	}
	query += " ORDER BY l.order_num ASC, l.id ASC"

	rows, err := r.db.Query(query, chapterID)
	if err != nil {
		return nil, fmt.Errorf("truy vấn bài học thất bại: %w", err)
	}
	defer rows.Close()

	var list []Lesson
	for rows.Next() {
		var l Lesson
		var isPubInt int
		if err := rows.Scan(&l.ID, &l.ChapterID, &l.ChapterTitle, &l.CourseID, &l.CourseCode, &l.CourseName,
			&l.Title, &l.ContentHTML, &l.VisualizationType, &l.VisualizationConfig, &l.OrderNum, &isPubInt, &l.CreatedAt); err != nil {
			return nil, err
		}
		l.IsPublished = isPubInt == 1
		list = append(list, l)
	}
	return list, rows.Err()
}

// FindLessonByID tìm chi tiết bài học theo ID
func (r *Repository) FindLessonByID(id int) (*Lesson, error) {
	row := r.db.QueryRow(`
		SELECT l.id, l.chapter_id, ch.title, ch.course_id, c.code, c.name,
		       l.title, COALESCE(l.content_html, ''), COALESCE(l.visualization_type, ''), COALESCE(l.visualization_config, ''), l.order_num, l.is_published, l.created_at
		FROM lessons l
		JOIN chapters ch ON l.chapter_id = ch.id
		JOIN courses c ON ch.course_id = c.id
		WHERE l.id = ?
	`, id)

	var l Lesson
	var isPubInt int
	err := row.Scan(&l.ID, &l.ChapterID, &l.ChapterTitle, &l.CourseID, &l.CourseCode, &l.CourseName,
		&l.Title, &l.ContentHTML, &l.VisualizationType, &l.VisualizationConfig, &l.OrderNum, &isPubInt, &l.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrLessonNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("truy vấn chi tiết bài học thất bại: %w", err)
	}
	l.IsPublished = isPubInt == 1

	return &l, nil
}

// CreateLesson tạo bài học mới
func (r *Repository) CreateLesson(l *Lesson) error {
	isPubInt := 0
	if l.IsPublished {
		isPubInt = 1
	}

	res, err := r.db.Exec(`
		INSERT INTO lessons (chapter_id, title, content_html, visualization_type, visualization_config, order_num, is_published)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, l.ChapterID, l.Title, l.ContentHTML, l.VisualizationType, l.VisualizationConfig, l.OrderNum, isPubInt)
	if err != nil {
		return fmt.Errorf("thêm bài học thất bại: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		l.ID = int(id)
	}
	return nil
}

// UpdateLesson cập nhật bài học
func (r *Repository) UpdateLesson(l *Lesson) error {
	isPubInt := 0
	if l.IsPublished {
		isPubInt = 1
	}

	res, err := r.db.Exec(`
		UPDATE lessons
		SET title = ?, content_html = ?, visualization_type = ?, visualization_config = ?, order_num = ?, is_published = ?
		WHERE id = ?
	`, l.Title, l.ContentHTML, l.VisualizationType, l.VisualizationConfig, l.OrderNum, isPubInt, l.ID)
	if err != nil {
		return fmt.Errorf("cập nhật bài học thất bại: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrLessonNotFound
	}
	return nil
}
