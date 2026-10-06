package audit

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Không thể mở CSDL in-memory: %v", err)
	}

	query := `CREATE TABLE audit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		action TEXT NOT NULL,
		object_type TEXT NOT NULL,
		object_id INTEGER,
		metadata TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(query); err != nil {
		t.Fatalf("Không thể tạo bảng audit_logs: %v", err)
	}

	return db
}

func TestAuditLogCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	service := NewService(repo)

	userID := 1
	examID := 101

	// 1. Ghi log hành động tạo exam
	err := service.LogAction(&userID, "create_exam", "exam", &examID, map[string]any{
		"title":       "Kiểm tra giữa kỳ",
		"duration":    45,
		"password":    "secret123", // Phải bị redacted
		"token":       "xyz789",    // Phải bị redacted
		"description": "Đề thi chính thức",
	})
	if err != nil {
		t.Fatalf("LogAction thất bại: %v", err)
	}

	// 2. Kiểm tra log được lưu và metadata được redact
	logs, err := service.GetRecentLogs(10)
	if err != nil {
		t.Fatalf("GetRecentLogs thất bại: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("Kỳ vọng 1 bản ghi audit log, nhận được %d", len(logs))
	}

	logEntry := logs[0]
	if logEntry.Action != "create_exam" || logEntry.ObjectType != "exam" {
		t.Errorf("Dữ liệu log không khớp: action=%s, object_type=%s", logEntry.Action, logEntry.ObjectType)
	}
	if *logEntry.UserID != userID || *logEntry.ObjectID != examID {
		t.Errorf("UserID hoặc ObjectID không khớp: user_id=%v, object_id=%v", logEntry.UserID, logEntry.ObjectID)
	}

	// Kiểm tra password và token không được lưu dạng raw
	if strings.Contains(logEntry.Metadata, "secret123") {
		t.Errorf("Mật khẩu bị lộ trong metadata: %s", logEntry.Metadata)
	}
	if strings.Contains(logEntry.Metadata, "xyz789") {
		t.Errorf("Token bị lộ trong metadata: %s", logEntry.Metadata)
	}
	if !strings.Contains(logEntry.Metadata, "[REDACTED]") {
		t.Errorf("Kỳ vọng metadata chứa [REDACTED], nhận được: %s", logEntry.Metadata)
	}

	// 3. Tra cứu theo user
	userLogs, err := service.GetLogsByUser(userID, 10)
	if err != nil || len(userLogs) != 1 {
		t.Fatalf("GetLogsByUser không trả về kết quả đúng")
	}

	// 4. Tra cứu theo object
	objectLogs, err := service.GetLogsByObject("exam", examID, 10)
	if err != nil || len(objectLogs) != 1 {
		t.Fatalf("GetLogsByObject không trả về kết quả đúng")
	}
}

func TestSanitizeFields(t *testing.T) {
	raw := map[string]any{
		"username":      "student01",
		"password":      "plaintext123",
		"session_token": "abc-def-ghi",
		"csrf_token":    "secure-csrf",
		"score":         95.5,
	}

	clean := SanitizeFields(raw)
	if clean["password"] != "[REDACTED]" {
		t.Errorf("Password chưa được redact: %v", clean["password"])
	}
	if clean["session_token"] != "[REDACTED]" {
		t.Errorf("session_token chưa được redact: %v", clean["session_token"])
	}
	if clean["csrf_token"] != "[REDACTED]" {
		t.Errorf("csrf_token chưa được redact: %v", clean["csrf_token"])
	}
	if clean["username"] != "student01" || clean["score"] != 95.5 {
		t.Errorf("Dữ liệu an toàn bị sai lệch: %v", clean)
	}
}

func TestServerLoggingFunctions(t *testing.T) {
	// Kiểm tra gọi các hàm server logging không gây crash/panic
	req := httptest.NewRequest("GET", "/test-error", nil)
	LogRequestError(req, http.StatusInternalServerError, sql.ErrConnDone)
	LogDatabaseError("test_query", sql.ErrTxDone)
	LogLoginFailure("bad_user", "127.0.0.1", "sai mật khẩu")
	LogSubmissionFailure(10, 20, "thời gian chạy vượt ngưỡng")
	LogJudgeError(100, 20, sql.ErrNoRows)

	// Kiểm tra RequestLoggerMiddleware
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	wrapped := RequestLoggerMiddleware(testHandler)
	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Kỳ vọng mã HTTP 400, nhận được %d", rr.Code)
	}
}
