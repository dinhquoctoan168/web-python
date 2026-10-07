package course

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"web_python/internal/auth"

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

func TestTeacherCreateCourse_And_StudentCannotCreate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	protectedCreate := auth.RequireTeacher(handler.HandleTeacherCreateCourse)

	teacherUser := &auth.User{
		ID:       10,
		Username: "teacher10",
		Role:     auth.RoleTeacher,
		IsActive: true,
	}
	studentUser := &auth.User{
		ID:       20,
		Username: "student20",
		Role:     auth.RoleStudent,
		IsActive: true,
	}

	form := url.Values{}
	form.Set("code", "DSA301")
	form.Set("name", "Cấu trúc dữ liệu & Giải thuật")
	form.Set("description", "Học phần CTDL & GT")

	// 1. Sinh viên cố gắng tạo môn học -> Bị từ chối HTTP 403 Forbidden
	reqStudent := httptest.NewRequest("POST", "/teacher/course/create", strings.NewReader(form.Encode()))
	reqStudent.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqStudent = reqStudent.WithContext(auth.WithUser(reqStudent.Context(), studentUser))
	recStudent := httptest.NewRecorder()

	protectedCreate(recStudent, reqStudent)

	if recStudent.Code != http.StatusForbidden {
		t.Errorf("Sinh viên gọi tạo môn học kỳ vọng 403 Forbidden, nhận %d", recStudent.Code)
	}

	// Xác nhận trong DB môn DSA301 chưa được tạo
	cNotCreated, _ := service.GetCourseByCode("DSA301")
	if cNotCreated != nil {
		t.Fatalf("LỖI: Môn học đã bị tạo trái phép bởi sinh viên")
	}

	// 2. Giảng viên tạo môn học -> Thành công HTTP 303 Redirect sang /teacher/courses
	reqTeacher := httptest.NewRequest("POST", "/teacher/course/create", strings.NewReader(form.Encode()))
	reqTeacher.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqTeacher = reqTeacher.WithContext(auth.WithUser(reqTeacher.Context(), teacherUser))
	recTeacher := httptest.NewRecorder()

	protectedCreate(recTeacher, reqTeacher)

	if recTeacher.Code != http.StatusSeeOther {
		t.Errorf("Giảng viên tạo môn học kỳ vọng 303 Redirect, nhận %d", recTeacher.Code)
	}
	if loc := recTeacher.Header().Get("Location"); loc != "/teacher/courses" {
		t.Errorf("Kỳ vọng chuyển hướng đến /teacher/courses, nhận %s", loc)
	}

	// Xác nhận môn học đã được tạo trong DB với creator = 10
	cCreated, err := service.GetCourseByCode("DSA301")
	if err != nil || cCreated == nil {
		t.Fatalf("Không tìm thấy môn học vừa được giảng viên tạo: %v", err)
	}
	if cCreated.CreatedBy != 10 {
		t.Errorf("Kỳ vọng CreatedBy = 10, nhận %d", cCreated.CreatedBy)
	}
}
