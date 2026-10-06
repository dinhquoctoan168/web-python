package course

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Không thể mở sqlite :memory:: %v", err)
	}

	queries := []string{
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			full_name TEXT NOT NULL,
			email TEXT,
			role TEXT NOT NULL,
			is_active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE courses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			description TEXT,
			status TEXT DEFAULT 'active',
			created_by INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(created_by) REFERENCES users(id)
		);`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("Lỗi tạo bảng test: %v", err)
		}
	}
	return db
}

func TestCourseCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	service := NewService(repo)

	// 1. Tạo môn học mới
	c, err := service.CreateCourse("PY101", "Python cơ bản", "Học phần Python", 1)
	if err != nil {
		t.Fatalf("CreateCourse lỗi: %v", err)
	}
	if c.ID <= 0 {
		t.Errorf("ID môn học không hợp lệ: %d", c.ID)
	}
	if c.Code != "PY101" {
		t.Errorf("Mã môn học không khớp: %s", c.Code)
	}

	// 2. Không cho phép tạo trùng mã
	_, err = service.CreateCourse("py101", "Trùng mã", "Desc", 1)
	if err == nil {
		t.Errorf("Kỳ vọng lỗi khi tạo môn học trùng mã")
	}

	// 3. Không cho phép để trống mã hoặc tên
	_, err = service.CreateCourse("", "Tên", "Desc", 1)
	if err == nil {
		t.Errorf("Kỳ vọng lỗi khi mã môn học rỗng")
	}

	// 4. Lấy môn học theo ID
	found, err := service.GetCourseByID(c.ID)
	if err != nil {
		t.Fatalf("GetCourseByID lỗi: %v", err)
	}
	if found.Name != "Python cơ bản" {
		t.Errorf("Tên môn học không khớp: %s", found.Name)
	}

	// 5. Cập nhật môn học
	updated, err := service.UpdateCourse(c.ID, "PY101", "Python Cơ Bản Cập Nhật", "Mô tả mới", StatusArchived)
	if err != nil {
		t.Fatalf("UpdateCourse lỗi: %v", err)
	}
	if updated.Status != StatusArchived {
		t.Errorf("Trạng thái chưa được cập nhật thành archived: %s", updated.Status)
	}

	// 6. Kiểm tra lọc danh sách môn học cho sinh viên (chỉ active) vs giảng viên (tất cả)
	studentCourses, err := service.ListCoursesForStudent()
	if err != nil {
		t.Fatalf("ListCoursesForStudent lỗi: %v", err)
	}
	if len(studentCourses) != 0 {
		t.Errorf("Kỳ vọng sinh viên không thấy môn archived, nhận %d", len(studentCourses))
	}

	teacherCourses, err := service.ListCoursesForTeacher()
	if err != nil {
		t.Fatalf("ListCoursesForTeacher lỗi: %v", err)
	}
	if len(teacherCourses) != 1 {
		t.Errorf("Kỳ vọng giảng viên thấy 1 môn, nhận %d", len(teacherCourses))
	}
}
