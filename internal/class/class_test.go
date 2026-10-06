package class

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
		`CREATE TABLE classes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			course_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			semester TEXT,
			academic_year TEXT,
			teacher_id INTEGER NOT NULL,
			status TEXT DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(course_id) REFERENCES courses(id),
			FOREIGN KEY(teacher_id) REFERENCES users(id)
		);`,
		`CREATE TABLE enrollments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			class_id INTEGER NOT NULL,
			student_id INTEGER NOT NULL,
			enrolled_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(class_id, student_id),
			FOREIGN KEY(class_id) REFERENCES classes(id),
			FOREIGN KEY(student_id) REFERENCES users(id)
		);`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("Lỗi tạo bảng test: %v", err)
		}
	}
	return db
}

func TestClassAndEnrollment(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Seed teacher, student, course
	_, _ = db.Exec(`INSERT INTO users (id, username, password_hash, full_name, role) VALUES (1, 'teacher', 'hash', 'Giảng viên', 'teacher')`)
	_, _ = db.Exec(`INSERT INTO users (id, username, password_hash, full_name, role) VALUES (2, 'student1', 'hash', 'Sinh viên 1', 'student')`)
	_, _ = db.Exec(`INSERT INTO users (id, username, password_hash, full_name, role) VALUES (3, 'student2', 'hash', 'Sinh viên 2', 'student')`)
	_, _ = db.Exec(`INSERT INTO courses (id, code, name) VALUES (1, 'DSA301', 'Cấu trúc dữ liệu và giải thuật')`)

	repo := NewRepository(db)
	service := NewService(repo)

	// 1. Tạo lớp học mới
	cl, err := service.CreateClass(1, "23CNTT1", "HK1", "2026-2027", 1)
	if err != nil {
		t.Fatalf("CreateClass lỗi: %v", err)
	}
	if cl.ID <= 0 {
		t.Errorf("ID lớp học không hợp lệ: %d", cl.ID)
	}

	// 2. Ghi danh sinh viên theo ID
	if err := service.EnrollStudent(cl.ID, 2); err != nil {
		t.Fatalf("EnrollStudent lỗi: %v", err)
	}

	// 3. Ghi danh trùng lặp -> báo lỗi
	if err := service.EnrollStudent(cl.ID, 2); err == nil {
		t.Errorf("Kỳ vọng lỗi khi ghi danh trùng lặp")
	}

	// 4. Ghi danh theo username
	if err := service.EnrollStudentByUsername(cl.ID, "student2"); err != nil {
		t.Fatalf("EnrollStudentByUsername lỗi: %v", err)
	}

	// 5. Kiểm tra chi tiết lớp học
	foundClass, students, err := service.GetClassDetail(cl.ID)
	if err != nil {
		t.Fatalf("GetClassDetail lỗi: %v", err)
	}
	if foundClass.Name != "23CNTT1" {
		t.Errorf("Tên lớp học không khớp: %s", foundClass.Name)
	}
	if len(students) != 2 {
		t.Errorf("Kỳ vọng 2 sinh viên trong lớp, nhận: %d", len(students))
	}

	// 6. Kiểm tra danh sách lớp học của sinh viên
	studentClasses, err := service.ListClassesForStudent(2)
	if err != nil {
		t.Fatalf("ListClassesForStudent lỗi: %v", err)
	}
	if len(studentClasses) != 1 {
		t.Errorf("Kỳ vọng sinh viên 1 thấy 1 lớp, nhận: %d", len(studentClasses))
	}

	// 7. Xoá sinh viên khỏi lớp
	if err := service.RemoveStudent(cl.ID, 2); err != nil {
		t.Fatalf("RemoveStudent lỗi: %v", err)
	}

	_, updatedStudents, err := service.GetClassDetail(cl.ID)
	if err != nil {
		t.Fatalf("GetClassDetail sau khi xoá lỗi: %v", err)
	}
	if len(updatedStudents) != 1 {
		t.Errorf("Kỳ vọng 1 sinh viên sau khi xoá, nhận: %d", len(updatedStudents))
	}
}
