package progress_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"web_python/internal/auth"
	"web_python/internal/database"
	"web_python/internal/progress"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) (*sql.DB, *progress.Service) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Không thể mở CSDL in-memory: %v", err)
	}
	db.SetMaxOpenConns(1)

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Lỗi migrate: %v", err)
	}

	// 1. Tạo users
	_, err = db.Exec(`
		INSERT INTO users (id, username, password_hash, full_name, role) 
		VALUES (1, 'teacher1', 'hash', 'Giảng viên A', 'teacher'),
		       (2, 'student1', 'hash', 'Sinh viên B', 'student');
	`)
	if err != nil {
		t.Fatalf("Lỗi tạo user: %v", err)
	}

	// 2. Tạo môn học
	_, err = db.Exec(`
		INSERT INTO courses (id, code, name, description, created_by)
		VALUES (1, 'CS101', 'Lập trình Python', 'Môn học nền tảng', 1);
	`)
	if err != nil {
		t.Fatalf("Lỗi tạo course: %v", err)
	}

	// 3. Tạo chương mục (2 chương)
	_, err = db.Exec(`
		INSERT INTO chapters (id, course_id, title, order_num)
		VALUES (1, 1, 'Chương 1: Cú pháp cơ bản', 1),
		       (2, 1, 'Chương 2: Cấu trúc điều khiển', 2);
	`)
	if err != nil {
		t.Fatalf("Lỗi tạo chapters: %v", err)
	}

	// 4. Tạo bài học (Chương 1: 2 bài, Chương 2: 1 bài)
	_, err = db.Exec(`
		INSERT INTO lessons (id, chapter_id, title, order_num, is_published)
		VALUES (1, 1, 'Bài 1: Biến và kiểu dữ liệu', 1, 1),
		       (2, 1, 'Bài 2: Toán tử và biểu thức', 2, 1),
		       (3, 2, 'Bài 3: Câu lệnh If Else', 1, 1);
	`)
	if err != nil {
		t.Fatalf("Lỗi tạo lessons: %v", err)
	}

	// 5. Tạo bài tập (Bài 1: 2 bài, Bài 2: 1 bài, Bài 3: 1 bài => Tổng 4 bài)
	_, err = db.Exec(`
		INSERT INTO exercises (id, course_id, lesson_id, title, exercise_type, status, description)
		VALUES (1, 1, 1, 'Bài tập 1.1', 'coding', 'active', 'Mô tả 1.1'),
		       (2, 1, 1, 'Bài tập 1.2', 'multiple_choice', 'active', 'Mô tả 1.2'),
		       (3, 1, 2, 'Bài tập 2.1', 'coding', 'active', 'Mô tả 2.1'),
		       (4, 1, 3, 'Bài tập 3.1', 'coding', 'active', 'Mô tả 3.1');
	`)
	if err != nil {
		t.Fatalf("Lỗi tạo exercises: %v", err)
	}

	repo := progress.NewRepository(db)
	return db, progress.NewService(repo)
}

func TestProgressCalculation(t *testing.T) {
	_, svc := setupTestDB(t)

	// Giai đoạn ban đầu: Chưa làm bài nào
	lp1, err := svc.GetLessonProgress(2, 1)
	if err != nil {
		t.Fatalf("Lỗi lấy tiến độ bài 1: %v", err)
	}
	if lp1.TotalExercises != 2 || lp1.CompletedExercises != 0 || lp1.ProgressPercent != 0 || lp1.IsCompleted {
		t.Errorf("Kỳ vọng Lesson 1 ban đầu 0%% hoàn thành, nhận được: %+v", lp1)
	}

	cp1, err := svc.GetChapterProgress(2, 1)
	if err != nil {
		t.Fatalf("Lỗi lấy tiến độ chương 1: %v", err)
	}
	if cp1.TotalLessons != 2 || cp1.CompletedLessons != 0 || cp1.ProgressPercent != 0 || cp1.IsCompleted {
		t.Errorf("Kỳ vọng Chapter 1 ban đầu 0%% hoàn thành, nhận được: %+v", cp1)
	}

	courseProg, err := svc.GetCourseProgress(2, 1)
	if err != nil {
		t.Fatalf("Lỗi lấy tiến độ môn học: %v", err)
	}
	if courseProg.TotalActivities != 4 || courseProg.CompletedActivities != 0 || courseProg.ProgressPercent != 0 || courseProg.IsCompleted {
		t.Errorf("Kỳ vọng Course ban đầu 0%% hoàn thành, nhận được: %+v", courseProg)
	}
}

func TestProgressEngineEndToEnd(t *testing.T) {
	db, svc := setupTestDB(t)
	handler := progress.NewHandler(svc)

	// Step 1: Hoàn thành bài tập 1.1 (1/2 bài trong Bài 1)
	_, err := db.Exec(`INSERT INTO student_exercise_progress (student_id, exercise_id, status, best_score) VALUES (2, 1, 'completed', 100)`)
	if err != nil {
		t.Fatalf("Lỗi insert bài 1.1: %v", err)
	}

	lp1, _ := svc.GetLessonProgress(2, 1)
	if lp1.CompletedExercises != 1 || lp1.ProgressPercent != 50 || lp1.IsCompleted {
		t.Errorf("Bài 1 phải đạt 50%% và IsCompleted = false, nhận được: %+v", lp1)
	}

	cp1, _ := svc.GetChapterProgress(2, 1)
	if cp1.CompletedLessons != 0 || cp1.ProgressPercent != 0 {
		t.Errorf("Chương 1 chưa có bài học nào hoàn thành trọn vẹn, nhận được: %+v", cp1)
	}

	courseP, _ := svc.GetCourseProgress(2, 1)
	if courseP.CompletedActivities != 1 || courseP.ProgressPercent != 25 {
		t.Errorf("Khóa học phải đạt 25%% tiến độ, nhận được: %+v", courseP)
	}

	// Step 2: Hoàn thành bài tập 1.2 => Lesson 1 hoàn thành 100%!
	_, err = db.Exec(`INSERT INTO student_exercise_progress (student_id, exercise_id, status, best_score) VALUES (2, 2, 'completed', 100)`)
	if err != nil {
		t.Fatalf("Lỗi insert bài 1.2: %v", err)
	}

	lp1, _ = svc.GetLessonProgress(2, 1)
	if !lp1.IsCompleted || lp1.ProgressPercent != 100 {
		t.Errorf("Bài 1 phải hoàn thành 100%%, nhận được: %+v", lp1)
	}

	cp1, _ = svc.GetChapterProgress(2, 1)
	if cp1.CompletedLessons != 1 || cp1.ProgressPercent != 50 {
		t.Errorf("Chương 1 có 1/2 bài hoàn thành => 50%%, nhận được: %+v", cp1)
	}

	// Step 3: Hoàn thành bài tập 2.1 => Lesson 2 hoàn thành => Chapter 1 hoàn thành 100%!
	_, err = db.Exec(`INSERT INTO student_exercise_progress (student_id, exercise_id, status, best_score) VALUES (2, 3, 'completed', 100)`)
	if err != nil {
		t.Fatalf("Lỗi insert bài 2.1: %v", err)
	}

	cp1, _ = svc.GetChapterProgress(2, 1)
	if !cp1.IsCompleted || cp1.ProgressPercent != 100 || cp1.CompletedLessons != 2 {
		t.Errorf("Chương 1 phải hoàn thành 100%%, nhận được: %+v", cp1)
	}

	// Step 4: Hoàn thành bài tập 3.1 => Chapter 2 và Course hoàn thành 100%!
	_, err = db.Exec(`INSERT INTO student_exercise_progress (student_id, exercise_id, status, best_score) VALUES (2, 4, 'completed', 100)`)
	if err != nil {
		t.Fatalf("Lỗi insert bài 3.1: %v", err)
	}

	courseP, _ = svc.GetCourseProgress(2, 1)
	if !courseP.IsCompleted || courseP.ProgressPercent != 100 || courseP.CompletedActivities != 4 {
		t.Errorf("Khóa học phải hoàn thành 100%%, nhận được: %+v", courseP)
	}

	// Step 5: Test API endpoints với HTTP Request
	studentUser := &auth.User{ID: 2, Username: "student1", Role: auth.RoleStudent}

	// Test GET /api/progress/course?id=1
	req := httptest.NewRequest(http.MethodGet, "/api/progress/course?id=1", nil)
	req = req.WithContext(auth.WithUser(req.Context(), studentUser))
	w := httptest.NewRecorder()

	handler.HandleGetCourseProgress(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Kỳ vọng StatusOK 200, nhận %d", w.Code)
	}

	var resCourse progress.CourseProgress
	if err := json.Unmarshal(w.Body.Bytes(), &resCourse); err != nil {
		t.Fatalf("Lỗi unmarshal JSON: %v", err)
	}
	if !resCourse.IsCompleted || resCourse.ProgressPercent != 100 {
		t.Errorf("API Course progress sai lệch: %+v", resCourse)
	}

	// Test GET /api/progress/chapter?id=1
	req = httptest.NewRequest(http.MethodGet, "/api/progress/chapter?id=1", nil)
	req = req.WithContext(auth.WithUser(req.Context(), studentUser))
	w = httptest.NewRecorder()

	handler.HandleGetChapterProgress(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Kỳ vọng StatusOK 200, nhận %d", w.Code)
	}

	var resChapter progress.ChapterProgress
	if err := json.Unmarshal(w.Body.Bytes(), &resChapter); err != nil {
		t.Fatalf("Lỗi unmarshal JSON: %v", err)
	}
	if !resChapter.IsCompleted || resChapter.ProgressPercent != 100 {
		t.Errorf("API Chapter progress sai lệch: %+v", resChapter)
	}

	// Test GET /api/progress/lesson?id=1
	req = httptest.NewRequest(http.MethodGet, "/api/progress/lesson?id=1", nil)
	req = req.WithContext(auth.WithUser(req.Context(), studentUser))
	w = httptest.NewRecorder()

	handler.HandleGetLessonProgress(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Kỳ vọng StatusOK 200, nhận %d", w.Code)
	}

	var resLesson progress.LessonProgress
	if err := json.Unmarshal(w.Body.Bytes(), &resLesson); err != nil {
		t.Fatalf("Lỗi unmarshal JSON: %v", err)
	}
	if !resLesson.IsCompleted || resLesson.ProgressPercent != 100 {
		t.Errorf("API Lesson progress sai lệch: %+v", resLesson)
	}
}
