package database

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestRunMigrations(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Không thể mở in-memory sqlite: %v", err)
	}
	defer db.Close()

	// Lần chạy 1: áp dụng migration
	if err := RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations thất bại lần 1: %v", err)
	}

	// Kiểm tra bảng schema_migrations
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = 1").Scan(&count); err != nil {
		t.Fatalf("Truy vấn schema_migrations thất bại: %v", err)
	}
	if count != 1 {
		t.Fatalf("Kỳ vọng version 1 đã được áp dụng, nhận được count = %d", count)
	}

	// Kiểm tra bảng topics tồn tại
	if err := db.QueryRow("SELECT COUNT(*) FROM topics").Scan(&count); err != nil {
		t.Fatalf("Bảng topics chưa được tạo: %v", err)
	}

	// Lần chạy 2: idempotent, không lỗi khi chạy lại
	if err := RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations thất bại lần 2 (idempotency): %v", err)
	}
}
