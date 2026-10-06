package assignment

import (
	"database/sql"
	"testing"
	"time"

	"web_python/internal/database"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Không thể mở database: %v", err)
	}

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Chạy migrations thất bại: %v", err)
	}

	// 1. Tạo users (teacher & student)
	_, _ = db.Exec(`INSERT INTO users (id, username, password_hash, full_name, role) VALUES 
		(2, 'teacher1', 'hash', 'Thay Giao', 'teacher'),
		(3, 'student1', 'hash', 'Sinh Vien', 'student')`)

	// 2. Tạo môn học và lớp học
	_, _ = db.Exec(`INSERT INTO courses (id, code, name, description, status) VALUES 
		(1, 'PY101', 'Python Co Ban', 'Mo ta', 'active')`)
	_, _ = db.Exec(`INSERT INTO classes (id, course_id, name, semester, academic_year, teacher_id) VALUES 
		(1, 1, 'Lop Python 01', 'HK1', '2026-2027', 2)`)

	// 3. Ghi danh sinh viên vào lớp
	_, _ = db.Exec(`INSERT INTO enrollments (student_id, class_id) VALUES (3, 1)`)

	// 4. Tạo câu hỏi trong ngân hàng bài tập
	_, _ = db.Exec(`INSERT INTO exercises (id, course_id, topic_id, title, difficulty, description, initial_code) VALUES 
		(101, 1, 1, 'Bài 1: In Xin Chào', 'Dễ', 'Mô tả bài 1', 'print("Hello")'),
		(102, 1, 1, 'Bài 2: Tính Tổng', 'Trung bình', 'Mô tả bài 2', 'def add(a, b): return a + b')`)

	return db
}

func TestAssignmentFlow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)

	// 1. Giảng viên tạo bài tập (kèm 2 câu hỏi, điểm 2.0 và 3.0)
	// Đặt hạn nộp vào ngày mai để kiểm tra tính năng IsUrgent
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02T15:04")
	a, err := svc.CreateAssignment(2, 1, "Bài tập kiểm tra 15p", "Làm trong 15 phút", "", tomorrow, []int{101, 102}, []float64{2.0, 3.0})
	if err != nil {
		t.Fatalf("Tạo assignment thất bại: %v", err)
	}
	if a.ID == 0 {
		t.Fatalf("Assignment ID phải > 0")
	}
	if a.Status != StatusDraft {
		t.Errorf("Kỳ vọng status ban đầu là draft, nhận %s", a.Status)
	}

	// 2. Lấy chi tiết bài tập
	detail, err := svc.GetAssignment(a.ID)
	if err != nil {
		t.Fatalf("GetAssignment thất bại: %v", err)
	}
	if detail.ClassName != "Lop Python 01" {
		t.Errorf("Kỳ vọng ClassName 'Lop Python 01', nhận %s", detail.ClassName)
	}
	if len(detail.Exercises) != 2 {
		t.Fatalf("Kỳ vọng 2 bài tập con, nhận %d", len(detail.Exercises))
	}
	if detail.TotalPoints != 5.0 {
		t.Errorf("Kỳ vọng tổng điểm là 5.0, nhận %f", detail.TotalPoints)
	}

	// 3. Giảng viên xem danh sách bài tập của mình
	tAssignments, err := svc.ListTeacherAssignments(2)
	if err != nil {
		t.Fatalf("ListTeacherAssignments thất bại: %v", err)
	}
	if len(tAssignments) != 1 {
		t.Errorf("Kỳ vọng giảng viên có 1 bài tập, nhận %d", len(tAssignments))
	}

	// 4. Sinh viên kiểm tra danh sách bài tập trước khi phát hành (phải rỗng)
	sAssignmentsDraft, err := svc.ListStudentAssignments(3)
	if err != nil {
		t.Fatalf("ListStudentAssignments thất bại: %v", err)
	}
	if len(sAssignmentsDraft) != 0 {
		t.Errorf("Bài tập chưa publish không được hiển thị cho sinh viên, nhận %d", len(sAssignmentsDraft))
	}

	// 5. Phát hành bài tập
	if err := svc.PublishAssignment(a.ID); err != nil {
		t.Fatalf("PublishAssignment thất bại: %v", err)
	}

	// 6. Sinh viên kiểm tra danh sách bài tập sau khi phát hành
	sAssignmentsPub, err := svc.ListStudentAssignments(3)
	if err != nil {
		t.Fatalf("ListStudentAssignments sau publish thất bại: %v", err)
	}
	if len(sAssignmentsPub) != 1 {
		t.Fatalf("Kỳ vọng sinh viên thấy 1 bài tập đã phát hành, nhận %d", len(sAssignmentsPub))
	}

	sa := sAssignmentsPub[0]
	if sa.ExerciseCount != 2 {
		t.Errorf("Kỳ vọng ExerciseCount = 2, nhận %d", sa.ExerciseCount)
	}
	if sa.TotalPoints != 5.0 {
		t.Errorf("Kỳ vọng TotalPoints = 5.0, nhận %f", sa.TotalPoints)
	}
	if !sa.IsUrgent {
		t.Errorf("Hạn nộp vào ngày mai nên IsUrgent phải là true")
	}
}
