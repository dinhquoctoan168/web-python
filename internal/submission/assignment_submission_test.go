package submission_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"web_python/internal/assignment"
	"web_python/internal/auth"
	"web_python/internal/class"
	"web_python/internal/database"
	"web_python/internal/submission"

	_ "modernc.org/sqlite"
)

type mockAuthorizer struct {
	allowed map[string]bool
}

func (m *mockAuthorizer) CanAccessAssignmentExercise(assignmentID, studentID, exerciseID int) (bool, error) {
	key := string(rune(assignmentID)) + "_" + string(rune(studentID)) + "_" + string(rune(exerciseID))
	return m.allowed[key], nil
}

func setupSubmissionAssignmentTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Không thể mở in-memory sqlite: %v", err)
	}

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Chạy migrations thất bại: %v", err)
	}

	_, _ = db.Exec(`INSERT INTO users (id, username, password_hash, full_name, role) VALUES 
		(2, 'teacher1', 'hash', 'Thầy Giáo', 'teacher'),
		(3, 'studentA', 'hash', 'Sinh Viên A', 'student'),
		(4, 'studentB', 'hash', 'Sinh Viên B', 'student')`)

	_, _ = db.Exec(`INSERT INTO courses (id, code, name, description, status) VALUES 
		(1, 'PY101', 'Python Cơ Bản', 'Mô tả', 'active')`)
	_, _ = db.Exec(`INSERT INTO classes (id, course_id, name, semester, academic_year, teacher_id) VALUES 
		(1, 1, 'Lớp Python 01', 'HK1', '2026-2027', 2)`)
	_, _ = db.Exec(`INSERT INTO enrollments (student_id, class_id) VALUES (3, 1)`)

	_, _ = db.Exec(`INSERT INTO exercises (id, course_id, topic_id, title, difficulty, description, initial_code) VALUES 
		(101, 1, 1, 'Bài 101', 'Dễ', 'Mô tả bài 101', 'print(101)'),
		(99, 1, 1, 'Bài 99', 'Dễ', 'Mô tả bài 99', 'print(99)')`)

	return db
}

func TestAssignmentSubmissionAuthorization(t *testing.T) {
	db := setupSubmissionAssignmentTestDB(t)
	defer db.Close()

	clRepo := class.NewRepository(db)
	clService := class.NewService(clRepo)

	aRepo := assignment.NewRepository(db)
	aService := assignment.NewService(aRepo)
	aService.SetClassService(clService)

	subRepo := submission.NewRepository(db)
	subService := submission.NewService(subRepo)
	handler := submission.NewHandler(subService)
	handler.SetAssignmentService(aService)

	// Tạo Assignment A thuộc Class 1 gồm exercise 101 (đã publish)
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02T15:04")
	asgn, err := aService.CreateAssignment(2, 1, "Assignment A", "Mô tả", "", tomorrow, []int{101}, []float64{10.0})
	if err != nil {
		t.Fatalf("CreateAssignment thất bại: %v", err)
	}
	if err := aService.PublishAssignment(asgn.ID); err != nil {
		t.Fatalf("PublishAssignment thất bại: %v", err)
	}

	// Case 1: Student B (ngoài lớp) nộp bài vào assignment -> 403 Forbidden
	t.Run("Student Outside Class Rejected", func(t *testing.T) {
		payload := submission.CreateSubmissionRequest{
			ExerciseID:   101,
			AssignmentID: asgn.ID,
			SourceCode:   "print('hack')",
			Score:        100,
			PassedTests:  1,
			TotalTests:   1,
		}
		bodyBytes, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/submissions", bytes.NewReader(bodyBytes))
		req = req.WithContext(auth.WithUser(req.Context(), &auth.User{ID: 4, Role: auth.RoleStudent}))
		rec := httptest.NewRecorder()

		handler.HandleCreateSubmission(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Kỳ vọng 403 Forbidden khi student ngoài lớp nộp bài assignment, nhận %d", rec.Code)
		}
	})

	// Case 2: Exercise 99 (ngoài assignment) nộp kèm assignment_id -> 403 Forbidden
	t.Run("Exercise Outside Assignment Rejected", func(t *testing.T) {
		payload := submission.CreateSubmissionRequest{
			ExerciseID:   99,
			AssignmentID: asgn.ID,
			SourceCode:   "print('hack')",
			Score:        100,
			PassedTests:  1,
			TotalTests:   1,
		}
		bodyBytes, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/submissions", bytes.NewReader(bodyBytes))
		req = req.WithContext(auth.WithUser(req.Context(), &auth.User{ID: 3, Role: auth.RoleStudent}))
		rec := httptest.NewRecorder()

		handler.HandleCreateSubmission(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("Kỳ vọng 403 Forbidden khi exercise ngoài assignment, nhận %d", rec.Code)
		}
	})

	// Case 3: Student A hợp lệ nộp bài -> 200 OK
	t.Run("Valid Student Submission Allowed", func(t *testing.T) {
		payload := submission.CreateSubmissionRequest{
			ExerciseID:   101,
			AssignmentID: asgn.ID,
			SourceCode:   "print('valid')",
			Score:        100,
			PassedTests:  1,
			TotalTests:   1,
		}
		bodyBytes, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/submissions", bytes.NewReader(bodyBytes))
		req = req.WithContext(auth.WithUser(req.Context(), &auth.User{ID: 3, Role: auth.RoleStudent}))
		rec := httptest.NewRecorder()

		handler.HandleCreateSubmission(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Kỳ vọng 200 OK khi student hợp lệ nộp bài, nhận %d: %s", rec.Code, rec.Body.String())
		}
	})
}
