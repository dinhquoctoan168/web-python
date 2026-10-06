package course

import (
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrCourseNotFound = errors.New("không tìm thấy môn học")
)

// Repository quản lý các thao tác CSDL cho môn học
type Repository struct {
	db *sql.DB
}

// NewRepository khởi tạo Repository môn học
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ListCourses lấy danh sách môn học, tuỳ chọn lọc chỉ môn đang kích hoạt
func (r *Repository) ListCourses(onlyActive bool) ([]Course, error) {
	query := `
		SELECT c.id, c.code, c.name, COALESCE(c.description, ''), c.status, COALESCE(c.created_by, 0),
		       COALESCE(u.full_name, ''), c.created_at
		FROM courses c
		LEFT JOIN users u ON c.created_by = u.id
	`
	if onlyActive {
		query += " WHERE c.status = 'active'"
	}
	query += " ORDER BY c.id ASC"

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("truy vấn danh sách môn học thất bại: %w", err)
	}
	defer rows.Close()

	var list []Course
	for rows.Next() {
		var c Course
		if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.Description, &c.Status, &c.CreatedBy, &c.TeacherName, &c.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

// FindCourseByID tìm môn học theo ID
func (r *Repository) FindCourseByID(id int) (*Course, error) {
	row := r.db.QueryRow(`
		SELECT c.id, c.code, c.name, COALESCE(c.description, ''), c.status, COALESCE(c.created_by, 0),
		       COALESCE(u.full_name, ''), c.created_at
		FROM courses c
		LEFT JOIN users u ON c.created_by = u.id
		WHERE c.id = ?
	`, id)

	var c Course
	err := row.Scan(&c.ID, &c.Code, &c.Name, &c.Description, &c.Status, &c.CreatedBy, &c.TeacherName, &c.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCourseNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("truy vấn môn học thất bại: %w", err)
	}
	return &c, nil
}

// FindCourseByCode tìm môn học theo mã môn học (code)
func (r *Repository) FindCourseByCode(code string) (*Course, error) {
	row := r.db.QueryRow(`
		SELECT c.id, c.code, c.name, COALESCE(c.description, ''), c.status, COALESCE(c.created_by, 0),
		       COALESCE(u.full_name, ''), c.created_at
		FROM courses c
		LEFT JOIN users u ON c.created_by = u.id
		WHERE c.code = ?
	`, code)

	var c Course
	err := row.Scan(&c.ID, &c.Code, &c.Name, &c.Description, &c.Status, &c.CreatedBy, &c.TeacherName, &c.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCourseNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("truy vấn môn học theo mã thất bại: %w", err)
	}
	return &c, nil
}

// CreateCourse thêm một môn học mới
func (r *Repository) CreateCourse(c *Course) error {
	res, err := r.db.Exec(`
		INSERT INTO courses (code, name, description, status, created_by)
		VALUES (?, ?, ?, ?, ?)
	`, c.Code, c.Name, c.Description, c.Status, c.CreatedBy)
	if err != nil {
		return fmt.Errorf("thêm môn học thất bại: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		c.ID = int(id)
	}
	return nil
}

// UpdateCourse cập nhật thông tin môn học
func (r *Repository) UpdateCourse(c *Course) error {
	res, err := r.db.Exec(`
		UPDATE courses
		SET code = ?, name = ?, description = ?, status = ?
		WHERE id = ?
	`, c.Code, c.Name, c.Description, c.Status, c.ID)
	if err != nil {
		return fmt.Errorf("cập nhật môn học thất bại: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCourseNotFound
	}
	return nil
}

// CountCourses đếm tổng số môn học
func (r *Repository) CountCourses() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM courses").Scan(&count)
	return count, err
}
