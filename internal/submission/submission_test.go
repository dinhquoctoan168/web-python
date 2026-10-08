package submission

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"web_python/internal/auth"
	"web_python/internal/database"
	"web_python/internal/exercise"
	"web_python/internal/judge"

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

func TestJudgeAndSubmitFlow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Thêm test case cho bài tập 1
	_, _ = db.Exec(`INSERT INTO exercise_test_cases (exercise_id, call_expression, expected_output, is_hidden, weight, order_num) 
		VALUES (1, 'test()', '42', 0, 1.0, 1)`)

	repo := NewRepository(db)
	svc := NewService(repo)

	exRepo := exercise.NewRepository(db)
	exSvc := exercise.NewService(exRepo)
	judgeSvc := judge.NewService(exSvc)
	svc.SetJudgeService(judgeSvc)

	sub, judgeRes, err := svc.JudgeAndSubmit(1, 1, "def test(): return 42")
	if err != nil {
		t.Fatalf("JudgeAndSubmit thất bại: %v", err)
	}

	if sub.Score != 100.0 || sub.Status != "pass" {
		t.Errorf("Kỳ vọng submission pass với 100 điểm, nhận score=%f status=%s", sub.Score, sub.Status)
	}
	if judgeRes.PassedTests != 1 || judgeRes.TotalTests != 1 {
		t.Errorf("Kỳ vọng 1/1 test pass, nhận %d/%d", judgeRes.PassedTests, judgeRes.TotalTests)
	}
}

func TestEditScoreWithAudit(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	sub, err := svc.SubmitCode(1, 1, "def test(): return 10", 50.0, 1, 2)
	if err != nil {
		t.Fatalf("SubmitCode thất bại: %v", err)
	}

	teacherID := 99
	newScore := 85.0
	if err := svc.EditScore(teacherID, sub.ID, newScore); err != nil {
		t.Fatalf("EditScore thất bại: %v", err)
	}

	updated, err := svc.GetSubmissionDetail(sub.ID)
	if err != nil {
		t.Fatalf("GetSubmissionDetail thất bại: %v", err)
	}
	if updated.Score != 85.0 {
		t.Errorf("Kỳ vọng điểm được sửa thành 85.0, nhận %f", updated.Score)
	}
}

func TestSubmissionTemplatesRender(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)
	handler := NewHandler(svc)

	sub, err := svc.SubmitCode(1, 1, "def test(): return 10", 50.0, 1, 2)
	if err != nil {
		t.Fatalf("SubmitCode thất bại: %v", err)
	}

	teacherUser := &auth.User{
		ID:       2,
		Username: "teacher1",
		FullName: "Thay Giao",
		Role:     auth.RoleTeacher,
	}

	// 1. Giảng viên xem danh sách bài nộp
	reqList := httptest.NewRequest("GET", "/teacher/submissions", nil)
	reqList = reqList.WithContext(auth.WithUser(reqList.Context(), teacherUser))
	recList := httptest.NewRecorder()
	handler.HandleTeacherSubmissions(recList, reqList)
	if recList.Code != http.StatusOK {
		t.Fatalf("HandleTeacherSubmissions kỳ vọng 200 OK, nhận %d: %s", recList.Code, recList.Body.String())
	}
	if !strings.Contains(recList.Body.String(), "breadcrumb") {
		t.Errorf("Kỳ vọng body chứa breadcrumb")
	}

	// 2. Giảng viên xem chi tiết một bài nộp
	reqDetail := httptest.NewRequest("GET", "/teacher/submission/view?id="+strconv.Itoa(sub.ID), nil)
	reqDetail = reqDetail.WithContext(auth.WithUser(reqDetail.Context(), teacherUser))
	recDetail := httptest.NewRecorder()
	handler.HandleTeacherSubmissionView(recDetail, reqDetail)
	if recDetail.Code != http.StatusOK {
		t.Fatalf("HandleTeacherSubmissionView kỳ vọng 200 OK, nhận %d: %s", recDetail.Code, recDetail.Body.String())
	}
	if !strings.Contains(recDetail.Body.String(), "breadcrumb") {
		t.Errorf("Kỳ vọng body chứa breadcrumb")
	}
}


