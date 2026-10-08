package teacher

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"web_python/internal/auth"
	"web_python/internal/database"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Không thể mở in-memory sqlite: %v", err)
	}
	db.SetMaxOpenConns(1)

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Chạy migrations thất bại: %v", err)
	}

	// 1. Giảng viên và sinh viên
	_, _ = db.Exec(`INSERT INTO users (id, username, password_hash, full_name, role) VALUES 
		(2, 'teacher', 'hash', 'Thầy Giáo Mẫu', 'teacher'),
		(3, 'student1', 'hash', 'Nguyễn Văn An', 'student')`)

	// 2. Khóa học và chương học
	_, _ = db.Exec(`INSERT INTO courses (id, code, name, created_by) VALUES (3, 'CS101', 'Lập trình Python', 2)`)
	_, _ = db.Exec(`INSERT INTO chapters (id, course_id, title, order_num) VALUES 
		(1, 3, 'Chương 1: Biến & Điều kiện', 1),
		(2, 3, 'Chương 2: Vòng lặp for & while', 2)`)
	_, _ = db.Exec(`INSERT INTO lessons (id, chapter_id, title, order_num) VALUES (1, 1, 'Bài 1: Giới thiệu', 1)`)

	// 3. Bài tập
	_, _ = db.Exec(`UPDATE exercises SET lesson_id = 1 WHERE id IN (10, 11)`)

	// 4. Lớp học và ghi danh
	_, _ = db.Exec(`INSERT INTO classes (id, course_id, name, semester, academic_year, teacher_id) 
		VALUES (1, 3, '23CNTT1', 'HK1', '2026-2027', 2)`)
	_, _ = db.Exec(`INSERT INTO enrollments (class_id, student_id) VALUES (1, 3)`)

	// 5. Tiến độ & Lịch sử hành vi học viên (tạo tình huống sinh viên gặp khó khăn: hint >= 2 mà chưa giải xong)
	if _, err := db.Exec(`INSERT INTO student_exercise_progress (student_id, exercise_id, status, attempts, best_score)
		VALUES (3, 10, 'in_progress', 3, 50)`); err != nil {
		t.Fatalf("Lỗi insert student_exercise_progress: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO exercise_attempts (student_id, exercise_id, run_count, test_count, hint_count)
		VALUES (3, 10, 8, 4, 3)`); err != nil {
		t.Fatalf("Lỗi insert exercise_attempts: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO submissions (student_id, exercise_id, source_code, score, passed_tests, total_tests, status)
		VALUES (3, 10, 'print(x)', 50, 1, 2, 'failed')`); err != nil {
		t.Fatalf("Lỗi insert submissions: %v", err)
	}

	return db
}

func TestTeacherDashboardAndClassAnalytics(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo, db)

	teacherUser := &auth.User{
		ID:       2,
		Username: "teacher",
		FullName: "Thầy Giáo Mẫu",
		Role:     "teacher",
	}

	var countEx, countAtt, countProg int
	_ = db.QueryRow(`SELECT COUNT(*) FROM exercises WHERE course_id = 3`).Scan(&countEx)
	_ = db.QueryRow(`SELECT COUNT(*) FROM exercise_attempts WHERE student_id = 3`).Scan(&countAtt)
	_ = db.QueryRow(`SELECT COUNT(*) FROM student_exercise_progress WHERE student_id = 3`).Scan(&countProg)
	t.Logf("countEx: %d, countAtt: %d, countProg: %d", countEx, countAtt, countProg)

	// 1. Kiểm tra Dashboard Page Data
	dashData, err := svc.GetDashboardPageData(teacherUser)
	if err != nil {
		t.Fatalf("GetDashboardPageData thất bại: %v", err)
	}
	if dashData.Summary.TotalClasses != 1 {
		t.Errorf("Kỳ vọng 1 lớp học phụ trách, nhận %d", dashData.Summary.TotalClasses)
	}
	if dashData.Summary.TotalStudents != 1 {
		t.Errorf("Kỳ vọng 1 sinh viên phụ trách, nhận %d", dashData.Summary.TotalStudents)
	}
	if len(dashData.Classes) != 1 {
		t.Fatalf("Kỳ vọng danh sách 1 lớp học, nhận %d", len(dashData.Classes))
	}
	if dashData.Classes[0].Name != "23CNTT1" {
		t.Errorf("Tên lớp không đúng: %s", dashData.Classes[0].Name)
	}

	// 2. Kiểm tra Class Analytics (Phân tích lớp học)
	classAnalytics, err := svc.GetClassAnalytics(1)
	if err != nil {
		t.Fatalf("GetClassAnalytics thất bại: %v", err)
	}
	if classAnalytics.ClassName != "23CNTT1" {
		t.Errorf("Tên lớp phân tích không đúng: %s", classAnalytics.ClassName)
	}
	if classAnalytics.StudentCount != 1 {
		t.Errorf("Sĩ số lớp kỳ vọng 1, nhận %d", classAnalytics.StudentCount)
	}
	if len(classAnalytics.Students) != 1 {
		t.Fatalf("Kỳ vọng 1 sinh viên trong bảng phân tích, nhận %d", len(classAnalytics.Students))
	}

	stRow := classAnalytics.Students[0]
	if stRow.FullName != "Nguyễn Văn An" {
		t.Errorf("Họ tên sinh viên không đúng: %s", stRow.FullName)
	}
	if !stRow.HasStruggling {
		t.Errorf("Kỳ vọng cờ HasStruggling = true do sinh viên đã thử nhiều lần chưa hoàn thành")
	}

	// 3. Kiểm tra Student Diagnostics (Definition of Done: Xác định bài tập gặp khó)
	diagData, err := svc.GetStudentDiagnostics(3, 1, teacherUser)
	if err != nil {
		t.Fatalf("GetStudentDiagnostics thất bại: %v", err)
	}
	if diagData.Student.FullName != "Nguyễn Văn An" {
		t.Errorf("Tên sinh viên trong chẩn đoán không khớp: %s", diagData.Student.FullName)
	}

	var foundStruggle bool
	for _, d := range diagData.Diagnostics {
		if d.ExerciseID == 10 {
			if !d.IsStruggling {
				t.Errorf("Kỳ vọng bài tập 10 được đánh dấu IsStruggling = true")
			}
			if d.HintCount != 3 {
				t.Errorf("Kỳ vọng hint_count = 3, nhận %d", d.HintCount)
			}
			foundStruggle = true
		}
	}
	if !foundStruggle {
		t.Errorf("Kỳ vọng tìm thấy chẩn đoán cho bài tập 10")
	}
}

func TestTeacherDashboardTemplateRender(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo, db)
	handler := NewHandler(svc)

	teacherUser := &auth.User{
		ID:       2,
		Username: "teacher",
		FullName: "Thầy Giáo Mẫu",
		Role:     "teacher",
	}

	req := httptest.NewRequest("GET", "/teacher", nil)
	req = req.WithContext(auth.WithUser(req.Context(), teacherUser))
	rec := httptest.NewRecorder()

	handler.HandleTeacherDashboard(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Teacher") {
		t.Errorf("Expected body to contain 'Teacher'")
	}
	if !strings.Contains(body, "Thầy Giáo Mẫu") {
		t.Errorf("Expected body to contain user name 'Thầy Giáo Mẫu'")
	}
	if !strings.Contains(body, "appNavLinks") {
		t.Errorf("Expected body to contain appNavLinks from shared nav template")
	}
}

func TestTeacherAnalyticsAndStudentDetailTemplatesRender(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo, db)
	handler := NewHandler(svc)

	teacherUser := &auth.User{
		ID:       2,
		Username: "teacher",
		FullName: "Thầy Giáo Mẫu",
		Role:     "teacher",
	}

	// 1. Phân tích lớp học
	reqAnalytics := httptest.NewRequest("GET", "/teacher/class/analytics?class_id=1", nil)
	reqAnalytics = reqAnalytics.WithContext(auth.WithUser(reqAnalytics.Context(), teacherUser))
	recAnalytics := httptest.NewRecorder()
	handler.HandleClassAnalytics(recAnalytics, reqAnalytics)
	if recAnalytics.Code != http.StatusOK {
		t.Fatalf("HandleClassAnalytics kỳ vọng 200 OK, nhận %d: %s", recAnalytics.Code, recAnalytics.Body.String())
	}
	if !strings.Contains(recAnalytics.Body.String(), "breadcrumb") {
		t.Errorf("Kỳ vọng body chứa breadcrumb")
	}

	// 2. Chi tiết học viên
	reqStudent := httptest.NewRequest("GET", "/teacher/student/detail?student_id=3&class_id=1", nil)
	reqStudent = reqStudent.WithContext(auth.WithUser(reqStudent.Context(), teacherUser))
	recStudent := httptest.NewRecorder()
	handler.HandleStudentDetail(recStudent, reqStudent)
	if recStudent.Code != http.StatusOK {
		t.Fatalf("HandleStudentDetail kỳ vọng 200 OK, nhận %d: %s", recStudent.Code, recStudent.Body.String())
	}
	if !strings.Contains(recStudent.Body.String(), "breadcrumb") {
		t.Errorf("Kỳ vọng body chứa breadcrumb")
	}
}


