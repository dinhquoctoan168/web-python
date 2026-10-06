package practice

import (
	"database/sql"
	"testing"

	"web_python/internal/database"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Không thể mở in-memory sqlite: %v", err)
	}

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Chạy migrations thất bại: %v", err)
	}

	// Tạo user và exercise mẫu
	_, _ = db.Exec(`INSERT INTO users (id, username, password_hash, full_name, role) VALUES (1, 'student1', 'hash', 'Sinh Vien 1', 'student')`)
	_, _ = db.Exec(`INSERT INTO exercises (id, course_id, topic_id, title, description, initial_code) VALUES (1, 3, 1, 'Bài 1', 'Mô tả', 'def test(): pass')`)
	_, _ = db.Exec(`INSERT INTO exercises (id, course_id, topic_id, title, description, initial_code) VALUES (2, 3, 1, 'Bài 2', 'Mô tả', 'def test2(): pass')`)

	return db
}

func TestPracticeDraftAndSubmission(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	// 1. Kiểm tra trạng thái ban đầu khi chưa làm
	initProg, err := svc.GetState(1, 1)
	if err != nil {
		t.Fatalf("GetState thất bại: %v", err)
	}
	if initProg.Status != StatusNotStarted {
		t.Errorf("Kỳ vọng status ban đầu là not_started, nhận %s", initProg.Status)
	}
	if initProg.LastCode != "" {
		t.Errorf("Kỳ vọng last_code ban đầu rỗng, nhận %s", initProg.LastCode)
	}

	// 2. Kiểm tra Auto-save bản nháp (debounce)
	draftCode := "def test():\n    return 42"
	p1, err := svc.SaveDraft(1, 1, draftCode)
	if err != nil {
		t.Fatalf("SaveDraft thất bại: %v", err)
	}
	if p1.Status != StatusInProgress {
		t.Errorf("Kỳ vọng sau khi lưu nháp status = in_progress, nhận %s", p1.Status)
	}
	if p1.LastCode != draftCode {
		t.Errorf("Kỳ vọng last_code lưu đúng draftCode")
	}

	// 3. Kiểm tra lưu lại lần 2 (cập nhật code mới)
	draftCode2 := "def test():\n    return 100"
	p2, err := svc.SaveDraft(1, 1, draftCode2)
	if err != nil {
		t.Fatalf("SaveDraft lần 2 thất bại: %v", err)
	}
	if p2.LastCode != draftCode2 {
		t.Errorf("Kỳ vọng last_code cập nhật thành draftCode2")
	}

	// 4. Kiểm tra SubmitPractice không pass
	p3, err := svc.SubmitPractice(1, 1, draftCode2, 50.0, false)
	if err != nil {
		t.Fatalf("SubmitPractice thất bại: %v", err)
	}
	if p3.Attempts != 1 {
		t.Errorf("Kỳ vọng attempts = 1, nhận %d", p3.Attempts)
	}
	if p3.BestScore != 50.0 {
		t.Errorf("Kỳ vọng best_score = 50.0, nhận %f", p3.BestScore)
	}
	if p3.Status != StatusInProgress {
		t.Errorf("Kỳ vọng status vẫn là in_progress khi chưa pass, nhận %s", p3.Status)
	}

	// 5. Kiểm tra SubmitPractice thành công (passed = true)
	finalCode := "def test():\n    return 42 # solved"
	p4, err := svc.SubmitPractice(1, 1, finalCode, 100.0, true)
	if err != nil {
		t.Fatalf("SubmitPractice (passed) thất bại: %v", err)
	}
	if p4.Attempts != 2 {
		t.Errorf("Kỳ vọng attempts = 2, nhận %d", p4.Attempts)
	}
	if p4.BestScore != 100.0 {
		t.Errorf("Kỳ vọng best_score = 100.0, nhận %f", p4.BestScore)
	}
	if p4.Status != StatusCompleted {
		t.Errorf("Kỳ vọng status = completed sau khi pass, nhận %s", p4.Status)
	}
	if p4.CompletedAt == nil {
		t.Errorf("Kỳ vọng completed_at được ghi nhận")
	}

	// 6. Kiểm tra GetAllStates
	states, err := svc.GetAllStates(1)
	if err != nil {
		t.Fatalf("GetAllStates thất bại: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("Kỳ vọng 1 bài tập có trạng thái, nhận %d", len(states))
	}
	if states[0].ExerciseID != 1 || states[0].Status != StatusCompleted {
		t.Errorf("Dữ liệu states không khớp: %+v", states[0])
	}
}
