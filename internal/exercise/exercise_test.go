package exercise

import (
	"database/sql"
	"strings"
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

	// Thêm bài tập mẫu có cả public test và hidden test và solution_code
	res, err := db.Exec(`INSERT INTO exercises 
		(course_id, topic_id, title, difficulty, description, initial_code, solution_code, solution_hint, allowed_functions) 
		VALUES (3, 1, 'Test Min Max', 'Dễ', 'Tìm min max', 'def find_min_max(arr): pass', 'def find_min_max(arr): return min(arr), max(arr)', 'Gợi ý', '["len"]')`)
	if err != nil {
		t.Fatalf("Thêm bài tập thất bại: %v", err)
	}
	exID, _ := res.LastInsertId()

	// Thêm 2 public test cases
	_, _ = db.Exec(`INSERT INTO exercise_test_cases 
		(exercise_id, input_data, call_expression, expected_output, is_hidden, weight, order_num) 
		VALUES (?, '[1, 2, 3]', 'find_min_max([1, 2, 3])', '(1, 3)', 0, 1.0, 1)`, exID)
	_, _ = db.Exec(`INSERT INTO exercise_test_cases 
		(exercise_id, input_data, call_expression, expected_output, is_hidden, weight, order_num) 
		VALUES (?, '[-5, -1]', 'find_min_max([-5, -1])', '(-5, -1)', 0, 1.0, 2)`, exID)

	// Thêm 1 hidden test case
	_, _ = db.Exec(`INSERT INTO exercise_test_cases 
		(exercise_id, input_data, call_expression, expected_output, is_hidden, weight, order_num) 
		VALUES (?, '[100, 200, 50]', 'find_min_max([100, 200, 50])', '(50, 200)', 1, 2.0, 99)`, exID)

	return db
}

func TestGetExerciseForClient_SecurityIsolation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	clientEx, err := svc.GetExerciseForClient(1)
	if err != nil {
		t.Fatalf("GetExerciseForClient thất bại: %v", err)
	}
	if clientEx == nil {
		t.Fatalf("Kỳ vọng tìm thấy bài tập với ID = 1")
	}

	// 1. Kiểm tra không chứa hidden test cases
	if len(clientEx.TestCases) != 2 {
		t.Fatalf("Kỳ vọng 2 public test cases, thực tế nhận %d", len(clientEx.TestCases))
	}
	for _, tc := range clientEx.TestCases {
		if strings.Contains(tc.InputData, "100") || strings.Contains(tc.ExpectedOutput, "50") {
			t.Fatalf("BẢO MẬT: Bị lộ hidden test case trong ClientExerciseDetail: %+v", tc)
		}
	}

	// 2. Kiểm tra TestCasesJSON không chứa dữ liệu hidden test
	if strings.Contains(clientEx.TestCasesJSON, "100, 200") {
		t.Fatalf("BẢO MẬT: Bị lộ hidden test case trong chuỗi TestCasesJSON")
	}

	// 3. Kiểm tra ClientExerciseDetail không có trường SolutionCode
	// (Được đảm bảo bởi thiết kế struct ClientExerciseDetail không chứa SolutionCode)
	if clientEx.Title != "Test Min Max" {
		t.Errorf("Tiêu đề bài tập không đúng: %s", clientEx.Title)
	}
}

func TestGetExerciseForJudge_IncludesAll(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	judgeEx, err := svc.GetExerciseForJudge(1)
	if err != nil {
		t.Fatalf("GetExerciseForJudge thất bại: %v", err)
	}
	if judgeEx == nil {
		t.Fatalf("Kỳ vọng tìm thấy bài tập cho Judge")
	}

	// Server Judge phải nhận toàn bộ 3 test cases (2 public + 1 hidden)
	if len(judgeEx.TestCases) != 3 {
		t.Fatalf("Kỳ vọng 3 test cases cho judge, thực tế nhận %d", len(judgeEx.TestCases))
	}

	var hasHidden bool
	for _, tc := range judgeEx.TestCases {
		if tc.IsHidden {
			hasHidden = true
		}
	}
	if !hasHidden {
		t.Fatalf("Judge phải nhận được ít nhất 1 hidden test case")
	}

	if judgeEx.SolutionCode == "" {
		t.Fatalf("Judge phải có solution_code để đối chiếu")
	}
}
