package exam

import (
	"database/sql"
	"errors"
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
	"web_python/internal/judge"

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

	// 4. Tạo câu hỏi trong ngân hàng bài tập kèm test case
	_, _ = db.Exec(`INSERT INTO exercises (id, course_id, topic_id, title, difficulty, description, initial_code) VALUES 
		(101, 1, 1, 'Câu 1: Hàm nhân đôi', 'Dễ', 'Viết hàm double_val(n)', 'def double_val(n): pass'),
		(102, 1, 1, 'Câu 2: Hàm lập phương', 'Dễ', 'Viết hàm cube_val(n)', 'def cube_val(n): pass')`)

	_, _ = db.Exec(`INSERT INTO exercise_test_cases (exercise_id, call_expression, expected_output, is_hidden, weight, order_num) VALUES 
		(101, 'double_val(5)', '10', 0, 1.0, 1),
		(102, 'cube_val(3)', '27', 0, 1.0, 1)`)

	exRepo := exercise.NewRepository(db)
	exSvc := exercise.NewService(exRepo)
	judgeSvc := judge.NewService(exSvc)

	examRepo := NewRepository(db)
	examSvc := NewService(examRepo, judgeSvc)

	return db, examSvc
}

func TestExamFullLifecycle(t *testing.T) {
	db, svc := setupTestDB(t)
	defer db.Close()

	// 1. Giảng viên tạo đề thi 60 phút gồm 2 câu hỏi (câu 1: 4 điểm, câu 2: 6 điểm)
	nowStr := time.Now().Add(-1 * time.Hour).Format("2006-01-02T15:04")
	laterStr := time.Now().Add(2 * time.Hour).Format("2006-01-02T15:04")

	exam, err := svc.CreateExam(2, 1, "Kiểm tra giữa kỳ", "Quy chế thi nghiêm túc", 60, nowStr, laterStr, []int{101, 102}, []float64{4.0, 6.0})
	if err != nil {
		t.Fatalf("CreateExam thất bại: %v", err)
	}
	if exam.ID == 0 {
		t.Fatalf("Exam ID phải > 0")
	}

	// 2. Phát hành đề thi
	if err := svc.PublishExam(exam.ID); err != nil {
		t.Fatalf("PublishExam thất bại: %v", err)
	}

	// 3. Sinh viên xem danh sách bài thi
	studentExams, err := svc.ListStudentExams(3)
	if err != nil {
		t.Fatalf("ListStudentExams thất bại: %v", err)
	}
	if len(studentExams) != 1 {
		t.Fatalf("Kỳ vọng sinh viên thấy 1 bài thi, nhận %d", len(studentExams))
	}
	if studentExams[0].TotalPoints != 10.0 {
		t.Errorf("Kỳ vọng tổng điểm là 10.0, nhận %f", studentExams[0].TotalPoints)
	}

	// 4. Sinh viên vào phòng thi (Bắt đầu phiên thi)
	session, exDetail, answers, err := svc.StartOrResumeSession(exam.ID, 3)
	if err != nil {
		t.Fatalf("StartOrResumeSession thất bại: %v", err)
	}
	if session.Status != SessionInProgress {
		t.Errorf("Kỳ vọng trạng thái session là in_progress, nhận %s", session.Status)
	}
	if session.RemainingSeconds <= 0 {
		t.Errorf("RemainingSeconds phải > 0, nhận %d", session.RemainingSeconds)
	}
	if len(exDetail.Questions) != 2 {
		t.Errorf("Kỳ vọng 2 câu hỏi trong đề thi, nhận %d", len(exDetail.Questions))
	}
	if len(answers) != 0 {
		t.Errorf("Kỳ vọng answers ban đầu rỗng, nhận %d", len(answers))
	}

	// 5. Sinh viên lưu nháp bài làm (Câu 1 làm đúng, câu 2 làm sai)
	codeQ1 := "def double_val(n): return n * 2"
	codeQ2 := "def cube_val(n): return 0"

	// Kiểm tra quyền sở hữu: student khác không thể lưu nháp vào session này
	if err := svc.SaveAnswerDraft(session.ID, 999, 101, codeQ1); err == nil {
		t.Fatalf("Kỳ vọng lỗi khi student 999 cố lưu nháp vào session của student %d", session.StudentID)
	}

	if err := svc.SaveAnswerDraft(session.ID, session.StudentID, 101, codeQ1); err != nil {
		t.Fatalf("SaveAnswerDraft câu 1 thất bại: %v", err)
	}
	if err := svc.SaveAnswerDraft(session.ID, session.StudentID, 102, codeQ2); err != nil {
		t.Fatalf("SaveAnswerDraft câu 2 thất bại: %v", err)
	}

	// Kiểm tra quyền sở hữu: student khác không thể nộp bài session này
	if _, err := svc.SubmitExam(session.ID, 999); err == nil {
		t.Fatalf("Kỳ vọng lỗi khi student 999 cố nộp bài session của student %d", session.StudentID)
	}

	// 6. Nộp bài thi và chấm điểm tổng kết bằng Server-Side Judge
	finalScore, err := svc.SubmitExam(session.ID, session.StudentID)
	if err != nil {
		t.Fatalf("SubmitExam thất bại: %v", err)
	}

	// Câu 1 đúng được trọn 4.0 điểm, câu 2 sai được 0 điểm => Tổng điểm phải là 4.0
	if finalScore != 4.0 {
		t.Errorf("Kỳ vọng điểm tổng kết là 4.0, nhận %f", finalScore)
	}

	// Kiểm tra lại trạng thái sinh viên sau khi nộp
	studentExamsAfter, _ := svc.ListStudentExams(3)
	if studentExamsAfter[0].SessionStatus != SessionSubmitted {
		t.Errorf("Kỳ vọng SessionStatus là submitted, nhận %s", studentExamsAfter[0].SessionStatus)
	}
	if studentExamsAfter[0].FinalScore == nil || *studentExamsAfter[0].FinalScore != 4.0 {
		t.Errorf("Kỳ vọng FinalScore là 4.0")
	}
}

func TestExam_StudentCannotStartEarly(t *testing.T) {
	db, svc := setupTestDB(t)
	defer db.Close()

	// Tạo ca thi có StartAt cách thời điểm hiện tại 2 giờ trong tương lai
	startStr := time.Now().Add(2 * time.Hour).Format("2006-01-02T15:04")
	endStr := time.Now().Add(4 * time.Hour).Format("2006-01-02T15:04")

	exam, err := svc.CreateExam(2, 1, "Thi cuối kỳ (chưa mở)", "Mô tả", 60, startStr, endStr, []int{101}, []float64{10.0})
	if err != nil {
		t.Fatalf("CreateExam thất bại: %v", err)
	}
	if err := svc.PublishExam(exam.ID); err != nil {
		t.Fatalf("PublishExam thất bại: %v", err)
	}

	// Sinh viên 3 cố gắng vào thi sớm -> Hệ thống từ chối
	_, _, _, err = svc.StartOrResumeSession(exam.ID, 3)
	if err == nil {
		t.Fatalf("Kỳ vọng lỗi khi sinh viên vào thi sớm, nhưng hàm trả về thành công")
	}
	if !strings.Contains(err.Error(), "chưa đến giờ mở ca thi") {
		t.Errorf("Kỳ vọng thông báo 'chưa đến giờ mở ca thi', nhận được: %v", err)
	}
}

func TestExam_StudentCannotSubmitAfterDeadline(t *testing.T) {
	db, svc := setupTestDB(t)
	defer db.Close()

	startStr := time.Now().Add(-3 * time.Hour).Format("2006-01-02T15:04")
	endStr := time.Now().Add(1 * time.Hour).Format("2006-01-02T15:04")

	// Đề thi 30 phút, nhưng đã bắt đầu cách đây 2 giờ
	exam, err := svc.CreateExam(2, 1, "Kiểm tra 30 phút", "Mô tả", 30, startStr, endStr, []int{101}, []float64{10.0})
	if err != nil {
		t.Fatalf("CreateExam thất bại: %v", err)
	}
	if err := svc.PublishExam(exam.ID); err != nil {
		t.Fatalf("PublishExam thất bại: %v", err)
	}

	// Sinh viên bắt đầu làm bài
	session, _, _, err := svc.StartOrResumeSession(exam.ID, 3)
	if err != nil {
		t.Fatalf("StartOrResumeSession thất bại: %v", err)
	}

	// Giả lập thời gian phiên thi đã bắt đầu từ 2 tiếng trước (quá thời lượng 30 phút)
	twoHoursAgo := time.Now().Add(-2 * time.Hour)
	_, err = db.Exec("UPDATE exam_sessions SET started_at = ? WHERE id = ?", twoHoursAgo, session.ID)
	if err != nil {
		t.Fatalf("Lỗi update started_at: %v", err)
	}

	// 1. Lưu nháp code sau hạn chót -> Bị từ chối ErrSessionClosed
	err = svc.SaveAnswerDraft(session.ID, session.StudentID, 101, "def double_val(n): return n * 2")
	if !errors.Is(err, ErrSessionClosed) {
		t.Errorf("Lưu nháp sau deadline kỳ vọng ErrSessionClosed, nhận: %v", err)
	}

	// 2. Nộp bài sau hạn chót -> Bị từ chối ErrSessionClosed
	_, err = svc.SubmitExam(session.ID, session.StudentID)
	if !errors.Is(err, ErrSessionClosed) {
		t.Errorf("Nộp bài sau deadline kỳ vọng ErrSessionClosed, nhận: %v", err)
	}
}

func TestExam_TimerCalculatedByServer(t *testing.T) {
	db, svc := setupTestDB(t)
	defer db.Close()

	startStr := time.Now().Add(-1 * time.Hour).Format("2006-01-02T15:04")
	endStr := time.Now().Add(2 * time.Hour).Format("2006-01-02T15:04")

	// Đề thi 60 phút
	exam, err := svc.CreateExam(2, 1, "Thi tính giờ Server", "Mô tả", 60, startStr, endStr, []int{101}, []float64{10.0})
	if err != nil {
		t.Fatalf("CreateExam thất bại: %v", err)
	}
	if err := svc.PublishExam(exam.ID); err != nil {
		t.Fatalf("PublishExam thất bại: %v", err)
	}

	session, _, _, err := svc.StartOrResumeSession(exam.ID, 3)
	if err != nil {
		t.Fatalf("StartOrResumeSession thất bại: %v", err)
	}

	// Giả lập phiên thi đã bắt đầu từ 20 phút trước
	twentyMinAgo := time.Now().Add(-20 * time.Minute)
	_, _ = db.Exec("UPDATE exam_sessions SET started_at = ? WHERE id = ?", twentyMinAgo, session.ID)

	// Sinh viên resume phiên thi -> Server phải tự tính remaining_seconds dựa trên đồng hồ server
	resumedSession, _, _, err := svc.StartOrResumeSession(exam.ID, 3)
	if err != nil {
		t.Fatalf("Resume session thất bại: %v", err)
	}

	// Thời lượng 60 phút - 20 phút đã trôi qua = 40 phút = 2400 giây
	// Kỳ vọng remaining trong khoảng 2380 - 2420 giây
	if resumedSession.RemainingSeconds < 2380 || resumedSession.RemainingSeconds > 2420 {
		t.Errorf("Thời gian RemainingSeconds tính toán sai: %d (kỳ vọng xấp xỉ 2400 giây)", resumedSession.RemainingSeconds)
	}
}

func TestExamTemplatesRender(t *testing.T) {
	db, svc := setupTestDB(t)
	defer db.Close()

	exRepo := exercise.NewRepository(db)
	exSvc := exercise.NewService(exRepo)
	classRepo := class.NewRepository(db)
	classSvc := class.NewService(classRepo)

	handler := NewHandler(svc, classSvc, exSvc)

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

	// Tạo bài thi mẫu
	nowStr := time.Now().Add(-1 * time.Hour).Format("2006-01-02T15:04")
	laterStr := time.Now().Add(2 * time.Hour).Format("2006-01-02T15:04")
	exam, err := svc.CreateExam(2, 1, "Kiểm tra giữa kỳ", "Quy chế thi nghiêm túc", 60, nowStr, laterStr, []int{101, 102}, []float64{4.0, 6.0})
	if err != nil {
		t.Fatalf("CreateExam thất bại: %v", err)
	}

	// 1. Giảng viên xem danh sách bài thi
	reqTList := httptest.NewRequest("GET", "/teacher/exams", nil)
	reqTList = reqTList.WithContext(auth.WithUser(reqTList.Context(), teacherUser))
	recTList := httptest.NewRecorder()
	handler.HandleTeacherListExams(recTList, reqTList)
	if recTList.Code != http.StatusOK {
		t.Fatalf("HandleTeacherListExams kỳ vọng 200 OK, nhận %d: %s", recTList.Code, recTList.Body.String())
	}
	if !strings.Contains(recTList.Body.String(), "appNavLinks") {
		t.Errorf("Kỳ vọng body chứa appNavLinks")
	}
	if !strings.Contains(recTList.Body.String(), "breadcrumb") {
		t.Errorf("Kỳ vọng body chứa breadcrumb")
	}

	// 2. Giảng viên mở form tạo bài thi mới
	reqTForm := httptest.NewRequest("GET", "/teacher/exam/new", nil)
	reqTForm = reqTForm.WithContext(auth.WithUser(reqTForm.Context(), teacherUser))
	recTForm := httptest.NewRecorder()
	handler.HandleTeacherNewExamForm(recTForm, reqTForm)
	if recTForm.Code != http.StatusOK {
		t.Fatalf("HandleTeacherNewExamForm kỳ vọng 200 OK, nhận %d: %s", recTForm.Code, recTForm.Body.String())
	}
	if !strings.Contains(recTForm.Body.String(), "breadcrumb") {
		t.Errorf("Kỳ vọng body chứa breadcrumb")
	}

	// 3. Giảng viên xem trang giám sát bài thi
	reqTMon := httptest.NewRequest("GET", "/teacher/exam/monitoring?id="+strconv.Itoa(exam.ID), nil)
	reqTMon = reqTMon.WithContext(auth.WithUser(reqTMon.Context(), teacherUser))
	recTMon := httptest.NewRecorder()
	handler.HandleTeacherExamMonitoring(recTMon, reqTMon)
	if recTMon.Code != http.StatusOK {
		t.Fatalf("HandleTeacherExamMonitoring kỳ vọng 200 OK, nhận %d: %s", recTMon.Code, recTMon.Body.String())
	}
	if !strings.Contains(recTMon.Body.String(), "breadcrumb") {
		t.Errorf("Kỳ vọng body chứa breadcrumb")
	}

	// 4. Sinh viên xem danh sách bài thi
	reqSList := httptest.NewRequest("GET", "/my-exams", nil)
	reqSList = reqSList.WithContext(auth.WithUser(reqSList.Context(), studentUser))
	recSList := httptest.NewRecorder()
	handler.HandleStudentMyExams(recSList, reqSList)
	if recSList.Code != http.StatusOK {
		t.Fatalf("HandleStudentMyExams kỳ vọng 200 OK, nhận %d: %s", recSList.Code, recSList.Body.String())
	}
	if !strings.Contains(recSList.Body.String(), "appNavLinks") {
		t.Errorf("Kỳ vọng body chứa appNavLinks")
	}
}
