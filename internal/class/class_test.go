package class

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"web_python/internal/auth"
	"web_python/internal/course"

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

func TestClassTemplatesRender(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, _ = db.Exec(`INSERT INTO users (id, username, password_hash, full_name, role) VALUES (1, 'teacher1', 'hash', 'Thầy Giáo', 'teacher')`)
	_, _ = db.Exec(`INSERT INTO users (id, username, password_hash, full_name, role) VALUES (2, 'student1', 'hash', 'Sinh Viên 1', 'student')`)
	_, _ = db.Exec(`INSERT INTO courses (id, code, name, description, created_by) VALUES (1, 'PY101', 'Python cơ bản', 'Mô tả', 1)`)

	repo := NewRepository(db)
	service := NewService(repo)
	courseRepo := course.NewRepository(db)
	courseService := course.NewService(courseRepo)
	handler := NewHandler(service, courseService)

	cl, err := service.CreateClass(1, "23CNTT1", "HK1", "2024-2025", 1)
	if err != nil {
		t.Fatalf("CreateClass lỗi: %v", err)
	}
	_ = service.EnrollStudent(cl.ID, 2)

	teacherUser := &auth.User{
		ID:       1,
		Username: "teacher1",
		FullName: "Thầy Giáo",
		Role:     auth.RoleTeacher,
	}

	studentUser := &auth.User{
		ID:       2,
		Username: "student1",
		FullName: "Sinh Viên 1",
		Role:     auth.RoleStudent,
	}

	// 1. Giảng viên xem danh sách lớp học
	reqTeacher := httptest.NewRequest("GET", "/teacher/classes", nil)
	reqTeacher = reqTeacher.WithContext(auth.WithUser(reqTeacher.Context(), teacherUser))
	recTeacher := httptest.NewRecorder()
	handler.HandleTeacherListClasses(recTeacher, reqTeacher)
	if recTeacher.Code != http.StatusOK {
		t.Fatalf("HandleTeacherListClasses kỳ vọng 200 OK, nhận %d: %s", recTeacher.Code, recTeacher.Body.String())
	}
	if !strings.Contains(recTeacher.Body.String(), "breadcrumb") {
		t.Errorf("Kỳ vọng trang teacher classes chứa breadcrumb")
	}

	// 2. Giảng viên xem chi tiết lớp học
	reqDetail := httptest.NewRequest("GET", "/teacher/class?id="+strconv.Itoa(cl.ID), nil)
	reqDetail = reqDetail.WithContext(auth.WithUser(reqDetail.Context(), teacherUser))
	recDetail := httptest.NewRecorder()
	handler.HandleTeacherClassDetail(recDetail, reqDetail)
	if recDetail.Code != http.StatusOK {
		t.Fatalf("HandleTeacherClassDetail kỳ vọng 200 OK, nhận %d: %s", recDetail.Code, recDetail.Body.String())
	}
	if !strings.Contains(recDetail.Body.String(), "breadcrumb") {
		t.Errorf("Kỳ vọng trang class detail chứa breadcrumb")
	}

	// 3. Sinh viên xem danh sách lớp học của mình
	reqStudent := httptest.NewRequest("GET", "/my-classes", nil)
	reqStudent = reqStudent.WithContext(auth.WithUser(reqStudent.Context(), studentUser))
	recStudent := httptest.NewRecorder()
	handler.HandleStudentMyClasses(recStudent, reqStudent)
	if recStudent.Code != http.StatusOK {
		t.Fatalf("HandleStudentMyClasses kỳ vọng 200 OK, nhận %d: %s", recStudent.Code, recStudent.Body.String())
	}
	if !strings.Contains(recStudent.Body.String(), "appNavLinks") {
		t.Errorf("Kỳ vọng trang my-classes chứa appNavLinks từ shared app_nav")
	}
}

