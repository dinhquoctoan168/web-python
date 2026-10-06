package class

import (
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrClassNotFound    = errors.New("không tìm thấy lớp học")
	ErrStudentNotFound  = errors.New("không tìm thấy sinh viên")
	ErrAlreadyEnrolled  = errors.New("sinh viên đã được ghi danh vào lớp này")
	ErrNotEnrolled      = errors.New("sinh viên chưa ghi danh vào lớp này")
)

// Repository quản lý các truy vấn CSDL cho lớp học và ghi danh
type Repository struct {
	db *sql.DB
}

// NewRepository khởi tạo Repository lớp học
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ListClasses lấy danh sách lớp học kèm thông tin môn học và số lượng sinh viên
func (r *Repository) ListClasses(teacherID int) ([]Class, error) {
	query := `
		SELECT cl.id, cl.course_id, c.code, c.name, cl.name, COALESCE(cl.semester, ''),
		       COALESCE(cl.academic_year, ''), cl.teacher_id, COALESCE(u.full_name, ''),
		       cl.status, cl.created_at,
		       (SELECT COUNT(*) FROM enrollments e WHERE e.class_id = cl.id) AS student_count
		FROM classes cl
		JOIN courses c ON cl.course_id = c.id
		LEFT JOIN users u ON cl.teacher_id = u.id
	`
	var rows *sql.Rows
	var err error

	if teacherID > 0 {
		query += " WHERE cl.teacher_id = ? ORDER BY cl.id DESC"
		rows, err = r.db.Query(query, teacherID)
	} else {
		query += " ORDER BY cl.id DESC"
		rows, err = r.db.Query(query)
	}

	if err != nil {
		return nil, fmt.Errorf("truy vấn danh sách lớp học thất bại: %w", err)
	}
	defer rows.Close()

	var list []Class
	for rows.Next() {
		var cl Class
		if err := rows.Scan(&cl.ID, &cl.CourseID, &cl.CourseCode, &cl.CourseName, &cl.Name,
			&cl.Semester, &cl.AcademicYear, &cl.TeacherID, &cl.TeacherName, &cl.Status,
			&cl.CreatedAt, &cl.StudentCount); err != nil {
			return nil, err
		}
		list = append(list, cl)
	}
	return list, rows.Err()
}

// ListClassesByStudent lấy danh sách các lớp mà sinh viên đã ghi danh
func (r *Repository) ListClassesByStudent(studentID int) ([]Class, error) {
	query := `
		SELECT cl.id, cl.course_id, c.code, c.name, cl.name, COALESCE(cl.semester, ''),
		       COALESCE(cl.academic_year, ''), cl.teacher_id, COALESCE(u.full_name, ''),
		       cl.status, cl.created_at,
		       (SELECT COUNT(*) FROM enrollments e2 WHERE e2.class_id = cl.id) AS student_count
		FROM classes cl
		JOIN courses c ON cl.course_id = c.id
		JOIN enrollments e ON cl.id = e.class_id
		LEFT JOIN users u ON cl.teacher_id = u.id
		WHERE e.student_id = ? AND cl.status = 'active'
		ORDER BY cl.id DESC
	`
	rows, err := r.db.Query(query, studentID)
	if err != nil {
		return nil, fmt.Errorf("truy vấn lớp học của sinh viên thất bại: %w", err)
	}
	defer rows.Close()

	var list []Class
	for rows.Next() {
		var cl Class
		if err := rows.Scan(&cl.ID, &cl.CourseID, &cl.CourseCode, &cl.CourseName, &cl.Name,
			&cl.Semester, &cl.AcademicYear, &cl.TeacherID, &cl.TeacherName, &cl.Status,
			&cl.CreatedAt, &cl.StudentCount); err != nil {
			return nil, err
		}
		list = append(list, cl)
	}
	return list, rows.Err()
}

// FindClassByID tìm thông tin chi tiết lớp học theo ID
func (r *Repository) FindClassByID(id int) (*Class, error) {
	row := r.db.QueryRow(`
		SELECT cl.id, cl.course_id, c.code, c.name, cl.name, COALESCE(cl.semester, ''),
		       COALESCE(cl.academic_year, ''), cl.teacher_id, COALESCE(u.full_name, ''),
		       cl.status, cl.created_at,
		       (SELECT COUNT(*) FROM enrollments e WHERE e.class_id = cl.id) AS student_count
		FROM classes cl
		JOIN courses c ON cl.course_id = c.id
		LEFT JOIN users u ON cl.teacher_id = u.id
		WHERE cl.id = ?
	`, id)

	var cl Class
	err := row.Scan(&cl.ID, &cl.CourseID, &cl.CourseCode, &cl.CourseName, &cl.Name,
		&cl.Semester, &cl.AcademicYear, &cl.TeacherID, &cl.TeacherName, &cl.Status,
		&cl.CreatedAt, &cl.StudentCount)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrClassNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("truy vấn lớp học thất bại: %w", err)
	}
	return &cl, nil
}

// CreateClass tạo một lớp học mới
func (r *Repository) CreateClass(c *Class) error {
	res, err := r.db.Exec(`
		INSERT INTO classes (course_id, name, semester, academic_year, teacher_id, status)
		VALUES (?, ?, ?, ?, ?, ?)
	`, c.CourseID, c.Name, c.Semester, c.AcademicYear, c.TeacherID, c.Status)
	if err != nil {
		return fmt.Errorf("thêm lớp học thất bại: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		c.ID = int(id)
	}
	return nil
}

// GetEnrolledStudents lấy danh sách sinh viên ghi danh trong lớp
func (r *Repository) GetEnrolledStudents(classID int) ([]EnrolledStudent, error) {
	rows, err := r.db.Query(`
		SELECT u.id, u.username, u.full_name, COALESCE(u.email, ''), e.enrolled_at
		FROM users u
		JOIN enrollments e ON u.id = e.student_id
		WHERE e.class_id = ?
		ORDER BY e.enrolled_at ASC
	`, classID)
	if err != nil {
		return nil, fmt.Errorf("truy vấn sinh viên trong lớp thất bại: %w", err)
	}
	defer rows.Close()

	var students []EnrolledStudent
	for rows.Next() {
		var s EnrolledStudent
		if err := rows.Scan(&s.ID, &s.Username, &s.FullName, &s.Email, &s.EnrolledAt); err != nil {
			return nil, err
		}
		students = append(students, s)
	}
	return students, rows.Err()
}

// EnrollStudent ghi danh sinh viên vào lớp
func (r *Repository) EnrollStudent(classID, studentID int) error {
	_, err := r.db.Exec(`
		INSERT INTO enrollments (class_id, student_id)
		VALUES (?, ?)
	`, classID, studentID)
	if err != nil {
		return fmt.Errorf("ghi danh sinh viên thất bại: %w", err)
	}
	return nil
}

// RemoveStudent huỷ ghi danh sinh viên khỏi lớp
func (r *Repository) RemoveStudent(classID, studentID int) error {
	res, err := r.db.Exec(`DELETE FROM enrollments WHERE class_id = ? AND student_id = ?`, classID, studentID)
	if err != nil {
		return fmt.Errorf("xoá sinh viên khỏi lớp thất bại: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotEnrolled
	}
	return nil
}

// IsStudentEnrolled kiểm tra xem sinh viên đã ghi danh vào lớp chưa
func (r *Repository) IsStudentEnrolled(classID, studentID int) (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM enrollments WHERE class_id = ? AND student_id = ?`,
		classID, studentID).Scan(&count)
	return count > 0, err
}

// FindStudentByUsername tìm user sinh viên theo username
func (r *Repository) FindStudentByUsername(username string) (int, string, error) {
	var id int
	var fullName string
	err := r.db.QueryRow(`SELECT id, full_name FROM users WHERE username = ? AND role = 'student'`,
		username).Scan(&id, &fullName)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", ErrStudentNotFound
	}
	return id, fullName, err
}
