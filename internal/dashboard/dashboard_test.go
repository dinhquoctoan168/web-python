package dashboard

import (
	"database/sql"
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

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Chạy migrations thất bại: %v", err)
	}

	// 1. Tạo khóa học, lớp học và ghi danh sinh viên (id = 3) vào lớp học (id = 1 thuộc course 3)
	if _, err := db.Exec(`INSERT INTO courses (id, code, name) VALUES (3, 'CS101', 'Lập trình Python')`); err != nil {
		t.Fatalf("Lỗi tạo course: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO classes (id, course_id, name, semester, academic_year, teacher_id)
		VALUES (1, 3, 'Lớp Python K15', 'HK1', '2026-2027', 2)`); err != nil {
		t.Fatalf("Lỗi tạo class: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO enrollments (class_id, student_id) VALUES (1, 3)`); err != nil {
		t.Fatalf("Lỗi tạo enrollment: %v", err)
	}

	// 2. Tạo một bài tập assignment cho lớp 1
	if _, err := db.Exec(`INSERT INTO assignments (class_id, title, description, due_at, status, created_by)
		VALUES (1, 'Bài tập tuần 1: Biến & Điều kiện', 'Làm các bài tập', datetime('now', '+3 days'), 'published', 2)`); err != nil {
		t.Fatalf("Lỗi tạo assignment: %v", err)
	}

	// 3. Tạo một ca thi exam cho lớp 1
	if _, err := db.Exec(`INSERT INTO exams (class_id, title, description, duration_minutes, start_at, end_at, status, created_by)
		VALUES (1, 'Thi giữa kỳ: Python Cơ bản', 'Bài thi 45 phút', 45, datetime('now', '+1 day'), datetime('now', '+2 day'), 'published', 2)`); err != nil {
		t.Fatalf("Lỗi tạo exam: %v", err)
	}

	// 4. Ghi nhận 1 bài hoàn thành trong student_exercise_progress
	if _, err := db.Exec(`INSERT INTO student_exercise_progress (student_id, exercise_id, status)
		VALUES (3, 10, 'completed')`); err != nil {
		t.Fatalf("Lỗi tạo progress: %v", err)
	}

	// 5. Ghi nhận 1 submission
	if _, err := db.Exec(`INSERT INTO submissions (student_id, exercise_id, source_code, score, passed_tests, total_tests, status)
		VALUES (3, 10, 'print(1)', 100, 1, 1, 'passed')`); err != nil {
		t.Fatalf("Lỗi tạo submission: %v", err)
	}

	return db
}

func TestGetStudentDashboardData(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	studentUser := &auth.User{
		ID:       3,
		Username: "student",
		FullName: "Sinh Viên Mẫu",
		Role:     "student",
	}

	data, err := svc.GetStudentDashboardData(studentUser)
	if err != nil {
		t.Fatalf("GetStudentDashboardData thất bại: %v", err)
	}

	if data == nil {
		t.Fatalf("Kỳ vọng dữ liệu dashboard không nil")
	}

	// 1. Kiểm tra khóa học đã ghi danh & tiến độ
	if len(data.Courses) == 0 {
		t.Fatalf("Kỳ vọng có ít nhất 1 khóa học đã ghi danh")
	}
	course := data.Courses[0]
	if course.CompletedExercises != 1 {
		t.Errorf("Kỳ vọng số bài tập hoàn thành là 1, nhận %d", course.CompletedExercises)
	}
	if course.TotalExercises <= 0 {
		t.Errorf("Kỳ vọng tổng số bài tập > 0, nhận %d", course.TotalExercises)
	}
	if course.ProgressPercent <= 0 {
		t.Errorf("Kỳ vọng phần trăm tiến độ > 0, nhận %d", course.ProgressPercent)
	}

	// 2. Kiểm tra danh sách sắp tới (Upcoming)
	t.Logf("data.Upcoming: %+v", data.Upcoming)
	if len(data.Upcoming) < 2 {
		t.Errorf("Kỳ vọng ít nhất 2 mục upcoming (assignment + exam), nhận %d", len(data.Upcoming))
	}

	// 3. Kiểm tra bài nộp gần đây (RecentSubmissions)
	if len(data.RecentSubmissions) != 1 {
		t.Errorf("Kỳ vọng 1 submission gần đây, nhận %d", len(data.RecentSubmissions))
	}
	if data.RecentSubmissions[0].Score != 100 {
		t.Errorf("Kỳ vọng điểm 100, nhận %f", data.RecentSubmissions[0].Score)
	}

	// 4. Kiểm tra user không hợp lệ
	_, err = svc.GetStudentDashboardData(nil)
	if err == nil {
		t.Errorf("Kỳ vọng lỗi khi user nil")
	}
}
