package exercise_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"web_python/internal/assignment"
	"web_python/internal/auth"
	"web_python/internal/class"
	"web_python/internal/database"
	"web_python/internal/exercise"

	_ "modernc.org/sqlite"
)

func setupExerciseAssignmentTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Không thể mở in-memory sqlite: %v", err)
	}

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Chạy migrations thất bại: %v", err)
	}

	// 1. Tạo users (teacher, studentA, studentB)
	_, _ = db.Exec(`INSERT INTO users (id, username, password_hash, full_name, role) VALUES 
		(2, 'teacher1', 'hash', 'Thầy Giáo', 'teacher'),
		(3, 'studentA', 'hash', 'Sinh Viên A', 'student'),
		(4, 'studentB', 'hash', 'Sinh Viên B', 'student')`)

	// 2. Tạo môn học và lớp học
	_, _ = db.Exec(`INSERT INTO courses (id, code, name, description, status) VALUES 
		(1, 'PY101', 'Python Cơ Bản', 'Mô tả', 'active')`)
	_, _ = db.Exec(`INSERT INTO classes (id, course_id, name, semester, academic_year, teacher_id) VALUES 
		(1, 1, 'Lớp Python 01', 'HK1', '2026-2027', 2)`)

	// 3. Ghi danh studentA (id 3) vào lớp 1. studentB (id 4) không ghi danh
	_, _ = db.Exec(`INSERT INTO enrollments (student_id, class_id) VALUES (3, 1)`)

	// 4. Tạo bài tập 101, 102 và 99
	_, _ = db.Exec(`INSERT INTO exercises (id, course_id, topic_id, title, difficulty, description, initial_code, solution_code) VALUES 
		(101, 1, 1, 'Bài 101', 'Dễ', 'Mô tả bài 101', 'print(101)', 'SECRET_SOLUTION_101'),
		(102, 1, 1, 'Bài 102', 'Dễ', 'Mô tả bài 102', 'print(102)', 'SECRET_SOLUTION_102'),
		(99, 1, 1, 'Bài 99 Ngoài Assignment', 'Dễ', 'Mô tả bài 99', 'print(99)', 'SECRET_SOLUTION_99')`)

	// Thêm 1 public test case và 1 hidden test case cho bài 101
	_, _ = db.Exec(`INSERT INTO exercise_test_cases (exercise_id, input_data, call_expression, expected_output, is_hidden, weight, order_num) 
		VALUES (101, '', 'main()', '101', 0, 1.0, 1),
		       (101, '', 'hidden_check()', 'SECRET_HIDDEN_OUTPUT', 1, 1.0, 2)`)

	return db
}

func TestExerciseAPIAssignmentContext(t *testing.T) {
	db := setupExerciseAssignmentTestDB(t)
	defer db.Close()

	eRepo := exercise.NewRepository(db)
	eService := exercise.NewService(eRepo)

	clRepo := class.NewRepository(db)
	clService := class.NewService(clRepo)

	aRepo := assignment.NewRepository(db)
	aService := assignment.NewService(aRepo)
	aService.SetClassService(clService)

	handler := exercise.NewHandler(eService)
	handler.SetAssignmentService(aService)

	// Tạo Assignment A thuộc Class 1 gồm exercises 101, 102 (trạng thái DRAFT)
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02T15:04")
	asgn, err := aService.CreateAssignment(2, 1, "Assignment A", "Mô tả", "", tomorrow, []int{101, 102}, []float64{5.0, 5.0})
	if err != nil {
		t.Fatalf("CreateAssignment thất bại: %v", err)
	}

	asgnIDStr := strconv.Itoa(asgn.ID)

	// -------------------------------------------------------------
	// Test B: Student A (thuộc lớp) nhưng Assignment là DRAFT -> 403
	// -------------------------------------------------------------
	t.Run("Test B - Assignment Draft", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/exercise?id=101&assignment_id="+asgnIDStr, nil)
		req = req.WithContext(auth.WithUser(req.Context(), &auth.User{ID: 3, Role: auth.RoleStudent}))
		rec := httptest.NewRecorder()

		handler.HandleAPIExercise(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Kỳ vọng 403 Forbidden khi assignment là draft, nhận %d", rec.Code)
		}
	})

	// Teacher publish assignment
	if err := aService.PublishAssignment(asgn.ID); err != nil {
		t.Fatalf("PublishAssignment thất bại: %v", err)
	}

	// -------------------------------------------------------------
	// Test A: Student B (ngoài lớp) truy cập assignment đã publish -> 403
	// -------------------------------------------------------------
	t.Run("Test A - Student Outside Class", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/exercise?id=101&assignment_id="+asgnIDStr, nil)
		req = req.WithContext(auth.WithUser(req.Context(), &auth.User{ID: 4, Role: auth.RoleStudent}))
		rec := httptest.NewRecorder()

		handler.HandleAPIExercise(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Kỳ vọng 403 Forbidden cho student ngoài lớp, nhận %d", rec.Code)
		}
	})

	// -------------------------------------------------------------
	// Test C: Exercise không thuộc assignment -> 403
	// -------------------------------------------------------------
	t.Run("Test C - Exercise Not In Assignment", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/exercise?id=99&assignment_id="+asgnIDStr, nil)
		req = req.WithContext(auth.WithUser(req.Context(), &auth.User{ID: 3, Role: auth.RoleStudent}))
		rec := httptest.NewRecorder()

		handler.HandleAPIExercise(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Kỳ vọng 403 Forbidden khi exercise ngoài assignment, nhận %d", rec.Code)
		}
	})

	// -------------------------------------------------------------
	// Test D: Student hợp lệ truy cập exercise thuộc assignment -> 200
	// -------------------------------------------------------------
	t.Run("Test D - Valid Student Access", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/exercise?id=101&assignment_id="+asgnIDStr, nil)
		req = req.WithContext(auth.WithUser(req.Context(), &auth.User{ID: 3, Role: auth.RoleStudent}))
		rec := httptest.NewRecorder()

		handler.HandleAPIExercise(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Kỳ vọng 200 OK cho student hợp lệ, nhận %d: %s", rec.Code, rec.Body.String())
		}

		body := rec.Body.String()
		if strings.Contains(body, "SECRET_SOLUTION_101") {
			t.Errorf("Response không được chứa solution_code")
		}
		if strings.Contains(body, "SECRET_HIDDEN_OUTPUT") {
			t.Errorf("Response không được chứa hidden test case")
		}

		var resp exercise.ClientExerciseDetail
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Unmarshal response thất bại: %v", err)
		}
		if resp.ID != 101 {
			t.Errorf("Kỳ vọng exercise ID 101, nhận %d", resp.ID)
		}
	})

	// -------------------------------------------------------------
	// Test E: Practice mode (không truyền assignment_id) -> 200 OK
	// -------------------------------------------------------------
	t.Run("Test E - Practice Mode Regression Check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/exercise?id=101", nil)
		req = req.WithContext(auth.WithUser(req.Context(), &auth.User{ID: 3, Role: auth.RoleStudent}))
		rec := httptest.NewRecorder()

		handler.HandleAPIExercise(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Kỳ vọng 200 OK cho practice mode không truyền assignment_id, nhận %d", rec.Code)
		}
	})
}
