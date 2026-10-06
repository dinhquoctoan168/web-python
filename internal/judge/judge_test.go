package judge

import (
	"database/sql"
	"testing"
	"time"

	"web_python/internal/database"
	"web_python/internal/exercise"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) (*sql.DB, *exercise.Service) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Không thể mở sqlite: %v", err)
	}

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Chạy migrations thất bại: %v", err)
	}

	exRepo := exercise.NewRepository(db)
	exService := exercise.NewService(exRepo)

	return db, exService
}

func TestJudge_FullPassAndHiddenMasking(t *testing.T) {
	db, exService := setupTestDB(t)
	defer db.Close()

	// Tạo bài tập 101 với 1 public test và 1 hidden test
	_, err := db.Exec(`INSERT INTO exercises (id, course_id, title, description, initial_code) 
		VALUES (101, 1, 'Tính bình phương', 'Viết hàm square(n)', 'def square(n): pass')`)
	if err != nil {
		t.Fatalf("Tạo bài tập thất bại: %v", err)
	}

	_, _ = db.Exec(`INSERT INTO exercise_test_cases (exercise_id, input_data, call_expression, expected_output, is_hidden, weight, order_num) 
		VALUES (101, '4', 'square(4)', '16', 0, 1.0, 1)`)
	_, _ = db.Exec(`INSERT INTO exercise_test_cases (exercise_id, input_data, call_expression, expected_output, is_hidden, weight, order_num) 
		VALUES (101, '10', 'square(10)', '100', 1, 1.0, 2)`)

	judgeSvc := NewService(exService)

	correctCode := `def square(n):
    return n * n
`
	res, err := judgeSvc.Evaluate(JudgeRequest{
		ExerciseID: 101,
		SourceCode: correctCode,
	})
	if err != nil {
		t.Fatalf("Evaluate thất bại: %v", err)
	}

	if res.Score != 100.0 {
		t.Errorf("Kỳ vọng 100 điểm, nhận %f", res.Score)
	}
	if res.PassedTests != 2 || res.TotalTests != 2 {
		t.Errorf("Kỳ vọng 2/2 tests pass, nhận %d/%d", res.PassedTests, res.TotalTests)
	}
	if res.Status != "pass" {
		t.Errorf("Kỳ vọng status 'pass', nhận %s", res.Status)
	}

	// Kiểm tra bảo mật: Test case ẩn phải được che giấu
	for _, tc := range res.Tests {
		if tc.IsHidden {
			if tc.Expected != "[Ẩn]" || tc.Actual != "[Ẩn]" || tc.Input != "[Test case ẩn]" {
				t.Errorf("Test case ẩn bị lộ thông tin: %+v", tc)
			}
		} else {
			if tc.Expected != "16" || tc.Actual != "16" {
				t.Errorf("Test case public không hiển thị đúng kết quả: %+v", tc)
			}
		}
	}
}

func TestJudge_PartialPass(t *testing.T) {
	db, exService := setupTestDB(t)
	defer db.Close()

	_, _ = db.Exec(`INSERT INTO exercises (id, course_id, title, description, initial_code) 
		VALUES (102, 1, 'Bài test', 'Mô tả', 'pass')`)
	_, _ = db.Exec(`INSERT INTO exercise_test_cases (exercise_id, input_data, call_expression, expected_output, is_hidden, weight, order_num) 
		VALUES (102, '1', 'check(1)', 'True', 0, 1.0, 1)`)
	_, _ = db.Exec(`INSERT INTO exercise_test_cases (exercise_id, input_data, call_expression, expected_output, is_hidden, weight, order_num) 
		VALUES (102, '2', 'check(2)', 'False', 1, 1.0, 2)`)

	judgeSvc := NewService(exService)

	// Code luôn trả về True (đạt test 1, rớt test 2)
	partialCode := `def check(n):
    return True
`
	res, err := judgeSvc.Evaluate(JudgeRequest{
		ExerciseID: 102,
		SourceCode: partialCode,
	})
	if err != nil {
		t.Fatalf("Evaluate thất bại: %v", err)
	}

	if res.Score != 50.0 {
		t.Errorf("Kỳ vọng 50 điểm, nhận %f", res.Score)
	}
	if res.PassedTests != 1 || res.TotalTests != 2 {
		t.Errorf("Kỳ vọng 1/2 tests pass, nhận %d/%d", res.PassedTests, res.TotalTests)
	}
	if res.Status != "partial" {
		t.Errorf("Kỳ vọng status 'partial', nhận %s", res.Status)
	}
}

func TestJudge_SecurityViolation(t *testing.T) {
	db, exService := setupTestDB(t)
	defer db.Close()

	_, _ = db.Exec(`INSERT INTO exercises (id, course_id, title, description) VALUES (103, 1, 'Bài test', 'Mô tả')`)
	_, _ = db.Exec(`INSERT INTO exercise_test_cases (exercise_id, expected_output) VALUES (103, '1')`)

	judgeSvc := NewService(exService)

	maliciousCode := `import os
print(os.listdir('.'))
`
	res, err := judgeSvc.Evaluate(JudgeRequest{
		ExerciseID: 103,
		SourceCode: maliciousCode,
	})
	if err != nil {
		t.Fatalf("Evaluate không được trả lỗi hệ thống mà phải trả kết quả vi phạm: %v", err)
	}

	if res.Score != 0 || res.Status != "fail" {
		t.Errorf("Code vi phạm bảo mật phải bị fail 0 điểm, nhận score=%f, status=%s", res.Score, res.Status)
	}
	if res.ExecutionError == "" {
		t.Errorf("Cần có thông báo lỗi bảo mật trong ExecutionError")
	}
}

func TestJudge_TimeoutExecution(t *testing.T) {
	db, exService := setupTestDB(t)
	defer db.Close()

	_, _ = db.Exec(`INSERT INTO exercises (id, course_id, title, description, time_limit_ms) VALUES (104, 1, 'Vòng lặp', 'Mô tả', 500)`)
	_, _ = db.Exec(`INSERT INTO exercise_test_cases (exercise_id, call_expression, expected_output) VALUES (104, 'loop_fn()', '1')`)

	judgeSvc := NewService(exService)
	judgeSvc.SetDefaultTimeout(800 * time.Millisecond)

	infiniteLoopCode := `def loop_fn():
    while True:
        pass
    return 1
`
	res, err := judgeSvc.Evaluate(JudgeRequest{
		ExerciseID: 104,
		SourceCode: infiniteLoopCode,
	})
	if err != nil {
		t.Fatalf("Evaluate thất bại: %v", err)
	}

	if res.Score != 0 || res.Status != "fail" {
		t.Errorf("Kỳ vọng fail do timeout, nhận score=%f, status=%s", res.Score, res.Status)
	}
	if len(res.Tests) > 0 && res.Tests[0].Error == "" {
		t.Errorf("Kỳ vọng ghi nhận lỗi timeout trong Tests[0]")
	}
}
