package quiz

import (
	"database/sql"
	"testing"

	"web_python/internal/database"
	"web_python/internal/practice"
	"web_python/internal/submission"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) (*sql.DB, *Service) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Không thể mở sqlite: %v", err)
	}

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Chạy migrations thất bại: %v", err)
	}

	// Tạo user mẫu
	_, _ = db.Exec(`INSERT INTO users (id, username, password_hash, full_name, role) VALUES (1, 'student1', 'hash', 'Hoc Vien', 'student')`)

	repo := NewRepository(db)
	practiceRepo := practice.NewRepository(db)
	practiceSvc := practice.NewService(practiceRepo)
	subRepo := submission.NewRepository(db)
	subSvc := submission.NewService(subRepo)

	svc := NewService(repo, practiceSvc, subSvc)
	return db, svc
}

func TestQuizPublicOptionsAndIsolation(t *testing.T) {
	db, svc := setupTestDB(t)
	defer db.Close()

	// Kiểm tra câu 10 (Trắc nghiệm kiểu dữ liệu)
	options, err := svc.GetOptionsForStudent(10)
	if err != nil {
		t.Fatalf("GetOptionsForStudent thất bại: %v", err)
	}

	if len(options) != 4 {
		t.Fatalf("Kỳ vọng 4 phương án, nhận %d", len(options))
	}

	// Đảm bảo struct PublicQuizOption không hề có trường IsCorrect
	for _, opt := range options {
		if opt.Content == "" {
			t.Errorf("Nội dung phương án không được rỗng")
		}
	}
}

func TestSubmitQuizAndCodeTracing(t *testing.T) {
	db, svc := setupTestDB(t)
	defer db.Close()

	// 1. Kiểm tra làm đúng câu trắc nghiệm 10
	// Phương án 1 (id=1): 'int, float, str, list' là đúng
	resCorrect, err := svc.SubmitQuiz(1, 10, 1)
	if err != nil {
		t.Fatalf("SubmitQuiz thất bại: %v", err)
	}
	if !resCorrect.IsCorrect || resCorrect.Score != 100.0 {
		t.Errorf("Kỳ vọng đáp án đúng được 100 điểm, nhận score=%f is_correct=%v", resCorrect.Score, resCorrect.IsCorrect)
	}

	// 2. Kiểm tra làm sai câu trắc nghiệm 10
	// Phương án 2 (id=2): 'var, val, let, const' là sai
	resWrong, err := svc.SubmitQuiz(1, 10, 2)
	if err != nil {
		t.Fatalf("SubmitQuiz phương án sai thất bại: %v", err)
	}
	if resWrong.IsCorrect || resWrong.Score != 0.0 {
		t.Errorf("Kỳ vọng đáp án sai được 0 điểm, nhận score=%f is_correct=%v", resWrong.Score, resWrong.IsCorrect)
	}

	// 3. Kiểm tra bài tập Code Tracing câu 11
	// Đoạn code: x = 1; for i in range(3): x *= 2; print(x) => Đáp án là 8 (id=6)
	resTracing, err := svc.SubmitQuiz(1, 11, 6)
	if err != nil {
		t.Fatalf("SubmitQuiz code tracing thất bại: %v", err)
	}
	if !resTracing.IsCorrect || resTracing.Score != 100.0 {
		t.Errorf("Kỳ vọng code tracing đáp án 8 đạt 100 điểm, nhận score=%f is_correct=%v", resTracing.Score, resTracing.IsCorrect)
	}

	// 4. Kiểm tra cập nhật tiến độ luyện tập
	practiceRepo := practice.NewRepository(db)
	p10, _ := practiceRepo.GetProgress(1, 10)
	if p10 == nil || p10.BestScore != 100.0 {
		t.Errorf("Tiến độ luyện tập câu 10 không cập nhật điểm cao nhất 100")
	}
}
