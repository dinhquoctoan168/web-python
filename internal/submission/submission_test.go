package submission

import (
	"database/sql"
	"testing"

	"web_python/internal/database"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Không thể mở sqlite: %v", err)
	}

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Chạy migrations thất bại: %v", err)
	}

	// Tạo user và bài tập mẫu
	_, _ = db.Exec(`INSERT INTO users (id, username, password_hash, full_name, role) VALUES (1, 'student1', 'hash', 'Sinh Vien 1', 'student')`)
	_, _ = db.Exec(`INSERT INTO exercises (id, course_id, topic_id, title, description, initial_code) VALUES (1, 3, 1, 'Bài 1', 'Mô tả', 'def test(): pass')`)

	return db
}

func TestAttemptMetricsAndSubmissions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	// 1. Kiểm tra tăng Action Metrics (Run, Test, Hint)
	a1, err := svc.RecordAction(1, 1, "run")
	if err != nil {
		t.Fatalf("RecordAction run thất bại: %v", err)
	}
	if a1.RunCount != 1 || a1.TestCount != 0 || a1.HintCount != 0 {
		t.Errorf("Kỳ vọng RunCount = 1, nhận %+v", a1)
	}

	// Tăng lần 2
	_, _ = svc.RecordAction(1, 1, "run")
	a2, _ := svc.RecordAction(1, 1, "test")
	if a2.RunCount != 2 || a2.TestCount != 1 {
		t.Errorf("Kỳ vọng RunCount = 2, TestCount = 1, nhận %+v", a2)
	}

	a3, _ := svc.RecordAction(1, 1, "hint")
	if a3.HintCount != 1 {
		t.Errorf("Kỳ vọng HintCount = 1, nhận %+v", a3)
	}

	// 2. Kiểm tra lưu Submission 1 (Fail)
	sub1, err := svc.SubmitCode(1, 1, "def test(): return 0", 0.0, 0, 2)
	if err != nil {
		t.Fatalf("SubmitCode thất bại: %v", err)
	}
	if sub1.Status != "fail" || sub1.Score != 0.0 {
		t.Errorf("Kỳ vọng status = fail, nhận %s", sub1.Status)
	}

	// 3. Kiểm tra lưu Submission 2 (Pass)
	sub2, err := svc.SubmitCode(1, 1, "def test(): return 42", 100.0, 2, 2)
	if err != nil {
		t.Fatalf("SubmitCode pass thất bại: %v", err)
	}
	if sub2.Status != "pass" || sub2.Score != 100.0 {
		t.Errorf("Kỳ vọng status = pass, nhận %s", sub2.Status)
	}

	// 4. Kiểm tra GetStudentHistory cho giáo viên
	history, err := svc.GetStudentHistory(1, 1)
	if err != nil {
		t.Fatalf("GetStudentHistory thất bại: %v", err)
	}
	if history.Attempt.RunCount != 2 || history.Attempt.TestCount != 1 || history.Attempt.HintCount != 1 {
		t.Errorf("Chỉ số nỗ lực không khớp: %+v", history.Attempt)
	}
	if len(history.Submissions) != 2 {
		t.Errorf("Kỳ vọng 2 bài nộp, nhận %d", len(history.Submissions))
	}
}
