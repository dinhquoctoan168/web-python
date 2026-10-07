package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Không thể mở sqlite :memory:: %v", err)
	}

	queries := []string{
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			full_name TEXT NOT NULL,
			email TEXT,
			role TEXT NOT NULL,
			is_active INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id)
		);`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("Lỗi tạo bảng test: %v", err)
		}
	}
	return db
}

func TestPasswordHashing(t *testing.T) {
	password := "Secret123!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword lỗi: %v", err)
	}

	if !CheckPassword(password, hash) {
		t.Errorf("CheckPassword thất bại với mật khẩu đúng")
	}

	if CheckPassword("WrongPassword", hash) {
		t.Errorf("CheckPassword chấp nhận mật khẩu sai")
	}
}

func TestAuthenticationAndSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	service := NewService(repo)

	hash, _ := HashPassword("student123")
	user := &User{
		Username:     "student",
		PasswordHash: hash,
		FullName:     "Sinh viên A",
		Email:        "sv@test.com",
		Role:         RoleStudent,
		IsActive:     true,
	}
	if err := repo.CreateUser(user); err != nil {
		t.Fatalf("CreateUser lỗi: %v", err)
	}

	// Đăng nhập sai
	_, _, err := service.Authenticate("student", "wrongpass")
	if err == nil {
		t.Errorf("Kỳ vọng lỗi khi sai mật khẩu, nhưng thành công")
	}

	// Đăng nhập đúng
	u, session, err := service.Authenticate("student", "student123")
	if err != nil {
		t.Fatalf("Authenticate thất bại: %v", err)
	}
	if u.Username != "student" {
		t.Errorf("Username không khớp: %s", u.Username)
	}
	if session.ID == "" {
		t.Errorf("Session ID rỗng")
	}

	// Kiểm tra session hợp lệ
	validatedUser, err := service.ValidateSession(session.ID)
	if err != nil {
		t.Fatalf("ValidateSession thất bại: %v", err)
	}
	if validatedUser.ID != u.ID {
		t.Errorf("User ID từ session không khớp: %d vs %d", validatedUser.ID, u.ID)
	}

	// Đăng xuất
	if err := service.Logout(session.ID); err != nil {
		t.Fatalf("Logout thất bại: %v", err)
	}

	// Session sau khi logout phải báo lỗi
	_, err = service.ValidateSession(session.ID)
	if err == nil {
		t.Errorf("Kỳ vọng session không còn hợp lệ sau khi logout")
	}
}

func TestRoleMiddleware(t *testing.T) {
	teacherUser := &User{
		ID:       1,
		Username: "teacher",
		Role:     RoleTeacher,
		IsActive: true,
	}
	studentUser := &User{
		ID:       2,
		Username: "student",
		Role:     RoleStudent,
		IsActive: true,
	}

	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// 1. RequireLogin khi chưa đăng nhập -> chuyển hướng 303 sang /login
	req := httptest.NewRequest("GET", "/protected", nil)
	rec := httptest.NewRecorder()
	RequireLogin(dummyHandler)(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Errorf("Chưa đăng nhập kỳ vọng status 303, nhận được %d", rec.Code)
	}

	// 2. RequireTeacher khi user là student -> 403 Forbidden
	ctx := context.WithValue(context.Background(), userCtxKey, studentUser)
	reqWithStudent := httptest.NewRequest("GET", "/teacher", nil).WithContext(ctx)
	rec = httptest.NewRecorder()
	RequireTeacher(dummyHandler)(rec, reqWithStudent)
	if rec.Code != http.StatusForbidden {
		t.Errorf("Student vào trang teacher kỳ vọng 403, nhận được %d", rec.Code)
	}

	// 3. RequireTeacher khi user là teacher -> 200 OK
	ctxTeacher := context.WithValue(context.Background(), userCtxKey, teacherUser)
	reqWithTeacher := httptest.NewRequest("GET", "/teacher", nil).WithContext(ctxTeacher)
	rec = httptest.NewRecorder()
	RequireTeacher(dummyHandler)(rec, reqWithTeacher)
	if rec.Code != http.StatusOK {
		t.Errorf("Teacher vào trang teacher kỳ vọng 200, nhận được %d", rec.Code)
	}
}

func TestExpiredSession(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	service := NewService(repo)

	hash, _ := HashPassword("password123")
	user := &User{
		Username:     "student_exp",
		PasswordHash: hash,
		FullName:     "Sinh viên hết hạn",
		Role:         RoleStudent,
		IsActive:     true,
	}
	if err := repo.CreateUser(user); err != nil {
		t.Fatalf("CreateUser lỗi: %v", err)
	}

	expiredSession := &Session{
		ID:        "expired-token-123",
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Quá khứ 1 giờ
	}
	if err := repo.CreateSession(expiredSession); err != nil {
		t.Fatalf("CreateSession lỗi: %v", err)
	}

	u, err := service.ValidateSession(expiredSession.ID)
	if !errors.Is(err, ErrSessionExpired) {
		t.Errorf("Kỳ vọng ErrSessionExpired, nhận được: %v (user: %v)", err, u)
	}
}

func TestLoginHandlerWithHTTPTest(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service, "../../web/templates/login.html")

	hash, _ := HashPassword("correct_pass")
	user := &User{
		Username:     "teacher_test",
		PasswordHash: hash,
		FullName:     "Giảng viên Test",
		Role:         RoleTeacher,
		IsActive:     true,
	}
	if err := repo.CreateUser(user); err != nil {
		t.Fatalf("CreateUser lỗi: %v", err)
	}

	// 1. Đăng nhập thành công (login success)
	formSuccess := url.Values{}
	formSuccess.Set("username", "teacher_test")
	formSuccess.Set("password", "correct_pass")

	reqSuccess := httptest.NewRequest("POST", "/login", strings.NewReader(formSuccess.Encode()))
	reqSuccess.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recSuccess := httptest.NewRecorder()

	handler.HandleLogin(recSuccess, reqSuccess)

	if recSuccess.Code != http.StatusSeeOther {
		t.Errorf("Đăng nhập thành công kỳ vọng HTTP 303, nhận %d", recSuccess.Code)
	}
	cookies := recSuccess.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == SessionCookieName {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Fatalf("Không tìm thấy session cookie sau khi đăng nhập thành công")
	}

	// 2. Đăng nhập thất bại (login fail)
	formFail := url.Values{}
	formFail.Set("username", "teacher_test")
	formFail.Set("password", "wrong_password")

	reqFail := httptest.NewRequest("POST", "/login", strings.NewReader(formFail.Encode()))
	reqFail.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recFail := httptest.NewRecorder()

	handler.HandleLogin(recFail, reqFail)

	if recFail.Code != http.StatusOK {
		t.Errorf("Đăng nhập thất bại kỳ vọng HTTP 200 render template, nhận %d", recFail.Code)
	}
	bodyStr := recFail.Body.String()
	if !strings.Contains(strings.ToLower(bodyStr), "tên đăng nhập hoặc mật khẩu không chính xác") {
		t.Errorf("Kỳ vọng body chứa thông báo lỗi đăng nhập, nhận được: %s", bodyStr)
	}

	// 3. Gọi /api/me với cookie hợp lệ (login success)
	reqMe := httptest.NewRequest("GET", "/api/me", nil)
	reqMe.AddCookie(sessionCookie)
	mw := NewMiddleware(service)
	recMe := httptest.NewRecorder()
	mw.AuthenticateMiddleware(http.HandlerFunc(handler.HandleCurrentUser)).ServeHTTP(recMe, reqMe)

	if recMe.Code != http.StatusOK {
		t.Errorf("Gọi /api/me với session hợp lệ kỳ vọng 200, nhận %d", recMe.Code)
	}
	if !strings.Contains(recMe.Body.String(), "teacher_test") {
		t.Errorf("Response /api/me kỳ vọng chứa username, nhận %s", recMe.Body.String())
	}

	// 4. Gọi /api/me khi chưa có cookie -> 401 Unauthorized
	reqMeUnauth := httptest.NewRequest("GET", "/api/me", nil)
	recMeUnauth := httptest.NewRecorder()
	mw.AuthenticateMiddleware(http.HandlerFunc(handler.HandleCurrentUser)).ServeHTTP(recMeUnauth, reqMeUnauth)

	if recMeUnauth.Code != http.StatusUnauthorized {
		t.Errorf("Gọi /api/me chưa xác thực kỳ vọng 401, nhận %d", recMeUnauth.Code)
	}
}
