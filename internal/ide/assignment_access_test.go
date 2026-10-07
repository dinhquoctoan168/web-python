package ide

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"web_python/internal/assignment"
	"web_python/internal/auth"
	"web_python/internal/class"
	"web_python/internal/course"
	"web_python/internal/database"
	"web_python/internal/exercise"
	"web_python/internal/lesson"

	_ "modernc.org/sqlite"
)

func setupAssignmentTestDB(t *testing.T) *sql.DB {
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB error: %v", err)
	}

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations thất bại: %v", err)
	}

	// 1. Tạo users (teacher, studentA, studentB)
	_, _ = db.Exec(`INSERT INTO users (id, username, password_hash, full_name, role) VALUES 
		(2, 'teacher1', 'hash', 'Thay Giao', 'teacher'),
		(3, 'studentA', 'hash', 'Sinh Vien A', 'student'),
		(4, 'studentB', 'hash', 'Sinh Vien B', 'student')`)

	// 2. Tạo môn học và lớp học
	_, _ = db.Exec(`INSERT INTO courses (id, code, name, description, status) VALUES 
		(1, 'PY101', 'Python Co Ban', 'Mo ta', 'active')`)
	_, _ = db.Exec(`INSERT INTO classes (id, course_id, name, semester, academic_year, teacher_id) VALUES 
		(1, 1, 'Lop Python 01', 'HK1', '2026-2027', 2)`)

	// 3. Ghi danh studentA (id 3) vào lớp 1. studentB (id 4) không ghi danh
	_, _ = db.Exec(`INSERT INTO enrollments (student_id, class_id) VALUES (3, 1)`)

	// 4. Tạo bài tập
	_, _ = db.Exec(`INSERT INTO exercises (id, course_id, topic_id, title, difficulty, description, initial_code) VALUES 
		(101, 1, 1, 'Bài 101', 'Dễ', 'Mô tả bài 101', 'print(101)'),
		(102, 1, 1, 'Bài 102', 'Dễ', 'Mô tả bài 102', 'print(102)'),
		(99, 1, 1, 'Bài 99', 'Dễ', 'Mô tả bài 99', 'print(99)')`)

	return db
}

func TestAssignmentIDEAuthorization(t *testing.T) {
	origDir, _ := os.Getwd()
	if strings.HasSuffix(origDir, "ide") {
		_ = os.Chdir("../../")
		defer os.Chdir(origDir)
	}

	db := setupAssignmentTestDB(t)

	cRepo := course.NewRepository(db)
	cService := course.NewService(cRepo)
	lRepo := lesson.NewRepository(db)
	lService := lesson.NewService(lRepo)
	eRepo := exercise.NewRepository(db)
	eService := exercise.NewService(eRepo)
	aRepo := assignment.NewRepository(db)
	aService := assignment.NewService(aRepo)
	clRepo := class.NewRepository(db)
	clService := class.NewService(clRepo)
	aService.SetClassService(clService)

	handler := NewHandler(cService, lService, eService, aService)

	// Tạo assignment A thuộc Class 1, gồm exercise 101, 102 (trạng thái ban đầu là DRAFT)
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02T15:04")
	asgn, err := aService.CreateAssignment(2, 1, "Assignment A", "Mo ta", "", tomorrow, []int{101, 102}, []float64{5.0, 5.0})
	if err != nil {
		t.Fatalf("CreateAssignment thất bại: %v", err)
	}

	// -------------------------------------------------------------
	// Test B: Assignment đang ở trạng thái DRAFT
	// Student A (thuộc lớp) cố truy cập -> phải nhận 403 Forbidden
	// -------------------------------------------------------------
	t.Run("Test B - Assignment Draft", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ide?mode=assignment&assignment_id="+strconv.Itoa(asgn.ID)+"&id=101", nil)
		req = req.WithContext(auth.WithUser(req.Context(), &auth.User{ID: 3, Role: auth.RoleStudent}))
		rec := httptest.NewRecorder()

		handler.HandleIDE(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Kỳ vọng 403 Forbidden khi assignment là draft, nhận %d", rec.Code)
		}
	})

	// Giảng viên publish assignment
	if err := aService.PublishAssignment(asgn.ID); err != nil {
		t.Fatalf("PublishAssignment thất bại: %v", err)
	}

	// -------------------------------------------------------------
	// Test A: Student B (không thuộc lớp) truy cập assignment đã published
	// -> phải nhận 403 Forbidden
	// -------------------------------------------------------------
	t.Run("Test A - Student Outside Class", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ide?mode=assignment&assignment_id="+strconv.Itoa(asgn.ID)+"&id=101", nil)
		req = req.WithContext(auth.WithUser(req.Context(), &auth.User{ID: 4, Role: auth.RoleStudent}))
		rec := httptest.NewRecorder()

		handler.HandleIDE(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Kỳ vọng 403 Forbidden khi student không thuộc lớp, nhận %d", rec.Code)
		}
	})

	// -------------------------------------------------------------
	// Test C: Exercise 99 không thuộc assignment
	// Student A (thuộc lớp) yêu cầu assignment_id=A&id=99
	// -> phải nhận 403 hoặc 404
	// -------------------------------------------------------------
	t.Run("Test C - Exercise Not In Assignment", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ide?mode=assignment&assignment_id="+strconv.Itoa(asgn.ID)+"&id=99", nil)
		req = req.WithContext(auth.WithUser(req.Context(), &auth.User{ID: 3, Role: auth.RoleStudent}))
		rec := httptest.NewRecorder()

		handler.HandleIDE(rec, req)

		if rec.Code != http.StatusForbidden && rec.Code != http.StatusNotFound {
			t.Errorf("Kỳ vọng 403 hoặc 404 khi exercise không thuộc assignment, nhận %d", rec.Code)
		}
	})

	// -------------------------------------------------------------
	// Test D: Student A hợp lệ (thuộc lớp, assignment published, exercise 101 thuộc assignment)
	// -> phải nhận 200 OK
	// -------------------------------------------------------------
	t.Run("Test D - Valid Student Access", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ide?mode=assignment&assignment_id="+strconv.Itoa(asgn.ID)+"&id=101", nil)
		req = req.WithContext(auth.WithUser(req.Context(), &auth.User{ID: 3, Role: auth.RoleStudent}))
		rec := httptest.NewRecorder()

		handler.HandleIDE(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("Kỳ vọng 200 OK cho sinh viên hợp lệ, nhận %d", rec.Code)
		}
	})

	// -------------------------------------------------------------
	// Test E: Người dùng chưa đăng nhập (anonymous)
	// -> phải nhận 403 Forbidden
	// -------------------------------------------------------------
	t.Run("Test E - Anonymous Access", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ide?mode=assignment&assignment_id="+strconv.Itoa(asgn.ID)+"&id=101", nil)
		rec := httptest.NewRecorder()

		handler.HandleIDE(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Kỳ vọng 403 Forbidden cho người dùng chưa đăng nhập, nhận %d", rec.Code)
		}
	})
}

