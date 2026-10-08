package assignment

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"web_python/internal/auth"
	"web_python/internal/class"
	"web_python/internal/database"
	"web_python/internal/exercise"

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

func TestAssignmentTemplatesRender(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	svc := NewService(repo)
	classRepo := class.NewRepository(db)
	classSvc := class.NewService(classRepo)
	exRepo := exercise.NewRepository(db)
	exSvc := exercise.NewService(exRepo)
	handler := NewHandler(svc, classSvc, exSvc)

	a, err := svc.CreateAssignment(2, 1, "Bài tập 1", "Mô tả", "", "", []int{101}, []float64{10.0})
	if err != nil {
		t.Fatalf("CreateAssignment failed: %v", err)
	}
	_ = svc.PublishAssignment(a.ID)

	teacherUser := &auth.User{
		ID:       2,
		Username: "teacher1",
		FullName: "Thay Giao",
		Role:     auth.RoleTeacher,
	}

	studentUser := &auth.User{
		ID:       3,
		Username: "student1",
		FullName: "Sinh Vien",
		Role:     auth.RoleStudent,
	}

	// 1. Giảng viên xem danh sách bài tập
	reqTList := httptest.NewRequest("GET", "/teacher/assignments", nil)
	reqTList = reqTList.WithContext(auth.WithUser(reqTList.Context(), teacherUser))
	recTList := httptest.NewRecorder()
	handler.HandleTeacherListAssignments(recTList, reqTList)
	if recTList.Code != http.StatusOK {
		t.Fatalf("HandleTeacherListAssignments kỳ vọng 200 OK, nhận %d: %s", recTList.Code, recTList.Body.String())
	}
	if !strings.Contains(recTList.Body.String(), "breadcrumb") {
		t.Errorf("Kỳ vọng body chứa breadcrumb")
	}

	// 2. Giảng viên mở form tạo bài tập
	reqTForm := httptest.NewRequest("GET", "/teacher/assignment/new", nil)
	reqTForm = reqTForm.WithContext(auth.WithUser(reqTForm.Context(), teacherUser))
	recTForm := httptest.NewRecorder()
	handler.HandleTeacherNewAssignmentForm(recTForm, reqTForm)
	if recTForm.Code != http.StatusOK {
		t.Fatalf("HandleTeacherNewAssignmentForm kỳ vọng 200 OK, nhận %d: %s", recTForm.Code, recTForm.Body.String())
	}
	if !strings.Contains(recTForm.Body.String(), "breadcrumb") {
		t.Errorf("Kỳ vọng body chứa breadcrumb")
	}

	// 3. Giảng viên xem chi tiết bài tập
	reqTDetail := httptest.NewRequest("GET", "/teacher/assignment?id="+strconv.Itoa(a.ID), nil)
	reqTDetail = reqTDetail.WithContext(auth.WithUser(reqTDetail.Context(), teacherUser))
	recTDetail := httptest.NewRecorder()
	handler.HandleTeacherAssignmentDetail(recTDetail, reqTDetail)
	if recTDetail.Code != http.StatusOK {
		t.Fatalf("HandleTeacherAssignmentDetail kỳ vọng 200 OK, nhận %d: %s", recTDetail.Code, recTDetail.Body.String())
	}
	if !strings.Contains(recTDetail.Body.String(), "breadcrumb") {
		t.Errorf("Kỳ vọng body chứa breadcrumb")
	}

	// 4. Sinh viên xem danh sách bài tập
	reqSList := httptest.NewRequest("GET", "/my-assignments", nil)
	reqSList = reqSList.WithContext(auth.WithUser(reqSList.Context(), studentUser))
	recSList := httptest.NewRecorder()
	handler.HandleStudentMyAssignments(recSList, reqSList)
	if recSList.Code != http.StatusOK {
		t.Fatalf("HandleStudentMyAssignments kỳ vọng 200 OK, nhận %d: %s", recSList.Code, recSList.Body.String())
	}
	if !strings.Contains(recSList.Body.String(), "appNavLinks") {
		t.Errorf("Kỳ vọng body chứa appNavLinks")
	}

	// 5. Sinh viên xem chi tiết bài tập
	reqSDetail := httptest.NewRequest("GET", "/assignment?id="+strconv.Itoa(a.ID), nil)
	reqSDetail = reqSDetail.WithContext(auth.WithUser(reqSDetail.Context(), studentUser))
	recSDetail := httptest.NewRecorder()
	handler.HandleStudentAssignmentDetail(recSDetail, reqSDetail)
	if recSDetail.Code != http.StatusOK {
		t.Fatalf("HandleStudentAssignmentDetail kỳ vọng 200 OK, nhận %d: %s", recSDetail.Code, recSDetail.Body.String())
	}
	if !strings.Contains(recSDetail.Body.String(), "breadcrumb") {
		t.Errorf("Kỳ vọng body chứa breadcrumb")
	}
}

