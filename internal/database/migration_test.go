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

func TestMigration015_MigrateTopicsToChapters(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Không thể mở in-memory sqlite: %v", err)
	}
	defer db.Close()

	if err := RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations thất bại: %v", err)
	}

	// 1. Thêm một topic cũ giả lập
	resT, err := db.Exec(`INSERT INTO topics (name, description, order_num) VALUES ('Bảng băm Legacy', 'Cấu trúc băm', 10)`)
	if err != nil {
		t.Fatalf("Lỗi tạo topic test: %v", err)
	}
	topicID, _ := resT.LastInsertId()

	// 2. Thêm bài tập cũ gắn với topicID và lesson_id = NULL
	resE, err := db.Exec(`INSERT INTO exercises (title, description, topic_id, difficulty) VALUES ('Bài tập băm', 'Mô tả bài tập', ?, 'Dễ')`, topicID)
	if err != nil {
		t.Fatalf("Lỗi tạo exercise test: %v", err)
	}
	exID, _ := resE.LastInsertId()

	// 3. Thực thi migration 15 Up trong một transaction
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Lỗi bắt đầu tx: %v", err)
	}
	var mig15 Migration
	for _, m := range migrations {
		if m.Version == 15 {
			mig15 = m
			break
		}
	}
	if mig15.Version != 15 {
		t.Fatalf("Không tìm thấy migration 15")
	}

	if err := mig15.Up(tx); err != nil {
		_ = tx.Rollback()
		t.Fatalf("Chạy Migration 15 Up thất bại: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("Lỗi commit tx: %v", err)
	}

	// 4. Kiểm chứng:
	// - Chapter 'Bảng băm Legacy' đã được tạo
	var chapID int
	err = db.QueryRow(`SELECT id FROM chapters WHERE title = 'Bảng băm Legacy'`).Scan(&chapID)
	if err != nil {
		t.Fatalf("Không tìm thấy chapter chuyển đổi từ topic: %v", err)
	}

	// - Lesson trong chapter này đã được tạo
	var lesID int
	err = db.QueryRow(`SELECT id FROM lessons WHERE chapter_id = ?`, chapID).Scan(&lesID)
	if err != nil {
		t.Fatalf("Không tìm thấy lesson tương ứng trong chapter: %v", err)
	}

	// - Bài tập cũ được cập nhật lesson_id và course_id
	var updatedLesID, updatedCourseID int
	err = db.QueryRow(`SELECT lesson_id, course_id FROM exercises WHERE id = ?`, exID).Scan(&updatedLesID, &updatedCourseID)
	if err != nil {
		t.Fatalf("Không tìm thấy bài tập sau migration: %v", err)
	}

	if updatedLesID != lesID {
		t.Errorf("Kỳ vọng lesson_id = %d, nhận được %d", lesID, updatedLesID)
	}
	if updatedCourseID <= 0 {
		t.Errorf("Kỳ vọng course_id > 0, nhận được %d", updatedCourseID)
	}

	// - Dữ liệu cũ trong bảng topics vẫn còn nguyên (chưa xóa)
	var oldTopicCount int
	_ = db.QueryRow(`SELECT COUNT(*) FROM topics WHERE id = ?`, topicID).Scan(&oldTopicCount)
	if oldTopicCount != 1 {
		t.Errorf("Kỳ vọng topic cũ vẫn còn trong DB, nhận %d", oldTopicCount)
	}
}
