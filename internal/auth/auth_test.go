package auth

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

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
