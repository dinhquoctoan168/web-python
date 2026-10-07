package submission

import (
	"database/sql"
	"testing"

	"web_python/internal/exercise"
	"web_python/internal/judge"
)

func TestJudgeProgressUpdate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Bài tập 1 có 2 test cases: 1 public (output=42), 1 hidden (output=100)
	_, _ = db.Exec(`INSERT INTO exercise_test_cases (exercise_id, call_expression, expected_output, is_hidden, weight, order_num) 
		VALUES 
		(1, 'test()', '42', 0, 1.0, 1),
		(1, 'hidden()', '100', 1, 1.0, 2)`)

	repo := NewRepository(db)
	svc := NewService(repo)

	exRepo := exercise.NewRepository(db)
	exSvc := exercise.NewService(exRepo)
	judgeSvc := judge.NewService(exSvc)
	svc.SetJudgeService(judgeSvc)

	studentID := 1
	exerciseID := 1

	// Đặt trạng thái ban đầu là passed_public
	_, _ = db.Exec(`INSERT INTO student_exercise_progress (student_id, exercise_id, status, best_score, attempts)
		VALUES (?, ?, 'passed_public', 50, 1)`, studentID, exerciseID)

	// Test D: Official Judge fail (chỉ đúng public test, sai hidden test)
	// Code chỉ define test() = 42, không có hidden()
	t.Run("Test D - Official Judge Fail Does Not Complete Progress", func(t *testing.T) {
		sub, judgeRes, err := svc.JudgeAndSubmit(studentID, exerciseID, "def test(): return 42")
		if err != nil {
			t.Fatalf("JudgeAndSubmit thất bại: %v", err)
		}

		if judgeRes.Status == "pass" {
			t.Fatalf("Kỳ vọng judgeRes.Status không phải 'pass', nhận: %s", judgeRes.Status)
		}
		if sub.Status == "pass" {
			t.Fatalf("Kỳ vọng sub.Status không phải 'pass'")
		}

		var currentStatus string
		err = db.QueryRow("SELECT status FROM student_exercise_progress WHERE student_id = ? AND exercise_id = ?", studentID, exerciseID).Scan(&currentStatus)
		if err != nil {
			t.Fatalf("Query status thất bại: %v", err)
		}

		if currentStatus == "completed" {
			t.Errorf("Khi official judge fail, progress KHÔNG ĐƯỢC là 'completed', nhưng nhận '%s'", currentStatus)
		}
	})

	// Test C: Official Judge pass (đúng cả public và hidden test)
	t.Run("Test C - Official Judge Pass Sets Progress To Completed", func(t *testing.T) {
		sub, judgeRes, err := svc.JudgeAndSubmit(studentID, exerciseID, "def test(): return 42\ndef hidden(): return 100")
		if err != nil {
			t.Fatalf("JudgeAndSubmit thất bại: %v", err)
		}

		if judgeRes.Status != "pass" {
			t.Fatalf("Kỳ vọng judgeRes.Status là 'pass', nhận: %s", judgeRes.Status)
		}
		if sub.Status != "pass" {
			t.Fatalf("Kỳ vọng sub.Status là 'pass', nhận: %s", sub.Status)
		}

		var currentStatus string
		var completedAt sql.NullTime
		err = db.QueryRow("SELECT status, completed_at FROM student_exercise_progress WHERE student_id = ? AND exercise_id = ?", studentID, exerciseID).Scan(&currentStatus, &completedAt)
		if err != nil {
			t.Fatalf("Query status thất bại: %v", err)
		}

		if currentStatus != "completed" {
			t.Errorf("Kỳ vọng progress là 'completed' sau khi official judge pass, nhận '%s'", currentStatus)
		}
		if !completedAt.Valid {
			t.Errorf("Kỳ vọng completed_at được cập nhật khi hoàn thành chính thức")
		}
	})
}
