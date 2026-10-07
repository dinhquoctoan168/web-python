package practice

import (
	"testing"
)

func TestProgressPublicState(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	studentID := 1
	exerciseID := 1

	// Test A: Public tests pass (score=100, passed=true)
	// Trạng thái phải là "passed_public", TUYỆT ĐỐI không phải "completed"
	t.Run("Test A - Public Tests Pass Sets passed_public", func(t *testing.T) {
		prog, err := svc.SubmitPractice(studentID, exerciseID, "def test(): return 42", 100.0, true)
		if err != nil {
			t.Fatalf("SubmitPractice thất bại: %v", err)
		}

		if prog.Status != "passed_public" {
			t.Errorf("Kỳ vọng status là 'passed_public', nhận '%s' (không được đánh dấu 'completed')", prog.Status)
		}
		if prog.CompletedAt != nil {
			t.Errorf("Kỳ vọng CompletedAt là nil khi chỉ mới pass public tests")
		}
	})

	// Test B: Sau khi passed_public, học viên sửa code và autosave chạy
	// Trạng thái không bị reset sai (phải giữ passed_public hoặc completed nếu đã complete)
	t.Run("Test B - Autosave After passed_public Preserves Status", func(t *testing.T) {
		prog, err := svc.SaveDraft(studentID, exerciseID, "def test(): return 43 # edited")
		if err != nil {
			t.Fatalf("SaveDraft thất bại: %v", err)
		}

		if prog.Status != "passed_public" {
			t.Errorf("Kỳ vọng status vẫn là 'passed_public' sau khi autosave, nhận '%s'", prog.Status)
		}
	})
}
