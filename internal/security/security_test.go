package security_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
	"web_python/internal/class"
	"web_python/internal/database"
	"web_python/internal/exam"
	"web_python/internal/security"

	_ "modernc.org/sqlite"
)

// 1. Kiểm thử sinh CSRF token
func TestCSRFTokenGeneration(t *testing.T) {
	t1, err := security.GenerateToken()
	if err != nil {
		t.Fatalf("Lỗi sinh CSRF token: %v", err)
	}
	if len(t1) != 64 {
		t.Errorf("Kỳ vọng token độ dài 64 hex chars, nhận %d", len(t1))
	}

	t2, _ := security.GenerateToken()
	if t1 == t2 {
		t.Errorf("Token ngẫu nhiên bị trùng lặp: %s == %s", t1, t2)
	}
}

// 2. Kiểm thử CSRF Middleware
func TestCSRFMiddleware(t *testing.T) {
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mw := security.CSRFMiddleware("/login", "/logout")
	handler := mw(mockHandler)

	// Test 1: GET request được cho qua và nhận được cookie csrf_token
	reqGET := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	recGET := httptest.NewRecorder()
	handler.ServeHTTP(recGET, reqGET)

	if recGET.Code != http.StatusOK {
		t.Errorf("Kỳ vọng GET thành công 200, nhận: %d", recGET.Code)
	}

	var csrfCookie *http.Cookie
	for _, c := range recGET.Result().Cookies() {
		if c.Name == security.CSRFCookieName {
			csrfCookie = c
			break
		}
	}
	if csrfCookie == nil || csrfCookie.Value == "" {
		t.Fatalf("Kỳ vọng nhận được cookie %s", security.CSRFCookieName)
	}

	validToken := csrfCookie.Value

	// Test 2: POST request không gửi token -> Bị chặn 403 Forbidden
	reqPOSTNoToken := httptest.NewRequest(http.MethodPost, "/api/submissions", nil)
	reqPOSTNoToken.AddCookie(csrfCookie)
	recPOSTNoToken := httptest.NewRecorder()
	handler.ServeHTTP(recPOSTNoToken, reqPOSTNoToken)

	if recPOSTNoToken.Code != http.StatusForbidden {
		t.Errorf("Kỳ vọng POST thiếu token bị chặn 403, nhận: %d", recPOSTNoToken.Code)
	}

	// Test 3: POST request gửi sai token -> Bị chặn 403 Forbidden
	reqPOSTWrongToken := httptest.NewRequest(http.MethodPost, "/api/submissions", nil)
	reqPOSTWrongToken.AddCookie(csrfCookie)
	reqPOSTWrongToken.Header.Set(security.CSRFHeaderName, "token_sai_hoan_toan_12345")
	recPOSTWrongToken := httptest.NewRecorder()
	handler.ServeHTTP(recPOSTWrongToken, reqPOSTWrongToken)

	if recPOSTWrongToken.Code != http.StatusForbidden {
		t.Errorf("Kỳ vọng POST sai token bị chặn 403, nhận: %d", recPOSTWrongToken.Code)
	}

	// Test 4: POST request gửi đúng token qua Header X-CSRF-Token -> Cho qua 200 OK
	reqPOSTHeader := httptest.NewRequest(http.MethodPost, "/api/submissions", nil)
	reqPOSTHeader.AddCookie(csrfCookie)
	reqPOSTHeader.Header.Set(security.CSRFHeaderName, validToken)
	recPOSTHeader := httptest.NewRecorder()
	handler.ServeHTTP(recPOSTHeader, reqPOSTHeader)

	if recPOSTHeader.Code != http.StatusOK {
		t.Errorf("Kỳ vọng POST gửi đúng header thành công 200, nhận: %d", recPOSTHeader.Code)
	}

	// Test 5: POST request gửi đúng token qua form field csrf_token -> Cho qua 200 OK
	form := url.Values{}
	form.Set(security.CSRFFormField, validToken)
	reqPOSTForm := httptest.NewRequest(http.MethodPost, "/teacher/course/create", strings.NewReader(form.Encode()))
	reqPOSTForm.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqPOSTForm.AddCookie(csrfCookie)
	recPOSTForm := httptest.NewRecorder()
	handler.ServeHTTP(recPOSTForm, reqPOSTForm)

	if recPOSTForm.Code != http.StatusOK {
		t.Errorf("Kỳ vọng POST gửi đúng form token thành công 200, nhận: %d", recPOSTForm.Code)
	}

	// Test 6: POST request tới đường dẫn ngoại lệ (/login) -> Cho qua 200 OK mà không cần token
	reqExempt := httptest.NewRequest(http.MethodPost, "/login", nil)
	recExempt := httptest.NewRecorder()
	handler.ServeHTTP(recExempt, reqExempt)

	if recExempt.Code != http.StatusOK {
		t.Errorf("Kỳ vọng POST /login ngoại lệ thành công 200, nhận: %d", recExempt.Code)
	}
}

// 3. Kiểm thử phân quyền giáo viên sở hữu lớp học (Teacher Ownership Authorization)
func TestTeacherClassOwnershipAuthorization(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Lỗi mở sqlite: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Lỗi migration: %v", err)
	}

	_, _ = db.Exec(`
		INSERT INTO users (id, username, password_hash, full_name, role) VALUES 
		(1, 'teacher_a', 'hash', 'Thầy A', 'teacher'),
		(2, 'teacher_b', 'hash', 'Thầy B', 'teacher');
		INSERT INTO courses (id, code, name, created_by) VALUES (1, 'CS101', 'Python', 1);
		INSERT INTO classes (id, course_id, name, teacher_id) VALUES (10, 1, 'Lớp của Thầy A', 1);
	`)

	classRepo := class.NewRepository(db)
	classService := class.NewService(classRepo)

	// Thầy A kiểm tra quyền sở hữu lớp 10 -> Hợp lệ
	if err := classService.VerifyTeacherOwnership(10, 1, false); err != nil {
		t.Errorf("Thầy A phải sở hữu lớp 10, nhận lỗi: %v", err)
	}

	// Thầy B kiểm tra quyền sở hữu lớp 10 -> Bị từ chối
	err = classService.VerifyTeacherOwnership(10, 2, false)
	if err != class.ErrUnauthorizedTeacher {
		t.Errorf("Kỳ vọng ErrUnauthorizedTeacher khi Thầy B can thiệp lớp của Thầy A, nhận: %v", err)
	}

	// Admin kiểm tra quyền lớp 10 -> Cho phép
	if err := classService.VerifyTeacherOwnership(10, 2, true); err != nil {
		t.Errorf("Admin phải được phép can thiệp lớp, nhận lỗi: %v", err)
	}
}

// 4. Kiểm thử phân quyền học viên vào thi (Student Exam Enrollment Authorization)
func TestStudentExamEnrollmentAuthorization(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Lỗi mở sqlite: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Lỗi migration: %v", err)
	}

	now := time.Now()
	startAt := now.Add(-10 * time.Minute)
	endAt := now.Add(60 * time.Minute)

	_, _ = db.Exec(`
		INSERT INTO users (id, username, password_hash, full_name, role) VALUES 
		(1, 'teacher', 'hash', 'Thầy Giáo', 'teacher'),
		(2, 'student_in_class', 'hash', 'Sinh viên Lớp', 'student'),
		(3, 'student_outsider', 'hash', 'Sinh viên Ngoài', 'student');
		INSERT INTO courses (id, code, name, created_by) VALUES (1, 'CS101', 'Python', 1);
		INSERT INTO classes (id, course_id, name, teacher_id) VALUES (1, 1, '23CNTT1', 1);
		INSERT INTO enrollments (class_id, student_id) VALUES (1, 2);
	`)

	// Tạo bài thi gán cho class_id = 1, status = 'published'
	_, _ = db.Exec(`
		INSERT INTO exams (id, class_id, title, description, duration_minutes, start_at, end_at, status, created_by)
		VALUES (100, 1, 'Kiểm tra giữa kỳ', 'Mô tả bài thi', 45, ?, ?, 'published', 1);
	`, startAt, endAt)

	examRepo := exam.NewRepository(db)
	examService := exam.NewService(examRepo, nil)

	// Sinh viên 2 (đã ghi danh vào lớp 1) -> Được phép bắt đầu làm bài thi
	session, _, _, err := examService.StartOrResumeSession(100, 2)
	if err != nil {
		t.Fatalf("Sinh viên trong lớp phải được vào thi, nhận lỗi: %v", err)
	}
	if session == nil {
		t.Fatalf("Kỳ vọng phiên thi được khởi tạo thành công")
	}

	// Sinh viên 3 (chưa ghi danh vào lớp 1) -> Server từ chối phân quyền
	_, _, _, err = examService.StartOrResumeSession(100, 3)
	if err == nil {
		t.Errorf("Kỳ vọng bị chặn khi sinh viên ngoài lớp cố gắng vào thi, nhưng lại thành công")
	} else if !strings.Contains(err.Error(), "không thuộc danh sách lớp") {
		t.Errorf("Lỗi trả về không đúng kỳ vọng: %v", err)
	}
}
