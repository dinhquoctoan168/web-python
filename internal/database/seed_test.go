package database

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestSeedComprehensiveCourses(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Không thể mở in-memory sqlite: %v", err)
	}
	defer db.Close()

	if err := RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations thất bại: %v", err)
	}

	// Lần chạy 1: Nạp dữ liệu học tập chi tiết
	if err := seedComprehensiveCourses(db); err != nil {
		t.Fatalf("seedComprehensiveCourses thất bại lần 1: %v", err)
	}

	// Kiểm tra số lượng chapters
	var chapterCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM chapters").Scan(&chapterCount); err != nil {
		t.Fatalf("Lỗi đếm chapters: %v", err)
	}
	if chapterCount < 10 {
		t.Errorf("Kỳ vọng ít nhất 10 chapters cho 3 môn học, nhận được %d", chapterCount)
	}

	// Kiểm tra số lượng lessons
	var lessonCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM lessons").Scan(&lessonCount); err != nil {
		t.Fatalf("Lỗi đếm lessons: %v", err)
	}
	if lessonCount < 20 {
		t.Errorf("Kỳ vọng ít nhất 20 lessons cho 3 môn học, nhận được %d", lessonCount)
	}

	// Kiểm tra số lượng bài tập và test case
	var exCount, tcCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM exercises").Scan(&exCount)
	_ = db.QueryRow("SELECT COUNT(*) FROM exercise_test_cases").Scan(&tcCount)

	if exCount < 10 {
		t.Errorf("Kỳ vọng ít nhất 10 exercises mẫu, nhận được %d", exCount)
	}
	if tcCount < 20 {
		t.Errorf("Kỳ vọng ít nhất 20 test cases mẫu, nhận được %d", tcCount)
	}

	// Lần chạy 2: Kiểm tra tính idempotent (không bị nhân đôi dữ liệu)
	if err := seedComprehensiveCourses(db); err != nil {
		t.Fatalf("seedComprehensiveCourses thất bại lần 2 (idempotency): %v", err)
	}

	var chapterCount2, lessonCount2, exCount2 int
	_ = db.QueryRow("SELECT COUNT(*) FROM chapters").Scan(&chapterCount2)
	_ = db.QueryRow("SELECT COUNT(*) FROM lessons").Scan(&lessonCount2)
	_ = db.QueryRow("SELECT COUNT(*) FROM exercises").Scan(&exCount2)

	if chapterCount2 != chapterCount || lessonCount2 != lessonCount || exCount2 != exCount {
		t.Errorf("Lỗi idempotency: Dữ liệu bị nhân đôi (chap: %d vs %d, les: %d vs %d, ex: %d vs %d)",
			chapterCount, chapterCount2, lessonCount, lessonCount2, exCount, exCount2)
	}
}
