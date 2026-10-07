package lesson

import (
	"database/sql"
	"testing"

	"web_python/internal/database"

	_ "modernc.org/sqlite"
)

func setupVisualizationTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Không thể mở sqlite :memory:: %v", err)
	}

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Chạy migrations thất bại: %v", err)
	}

	// Tạo môn học và chương mẫu
	_, _ = db.Exec(`INSERT INTO courses (id, code, name) VALUES (1, 'PY101', 'Python cơ bản')`)
	_, _ = db.Exec(`INSERT INTO chapters (id, course_id, title) VALUES (1, 1, 'Chương 1')`)

	return db
}

func TestLessonVisualizationOwnership(t *testing.T) {
	db := setupVisualizationTestDB(t)
	defer db.Close()

	repo := NewRepository(db)

	// -------------------------------------------------------------
	// Test A: Lesson có visualization độc lập mà không cần exercise nào
	// -------------------------------------------------------------
	t.Run("Test A - Independent Lesson Visualization Without Exercises", func(t *testing.T) {
		l := Lesson{
			ChapterID:         1,
			Title:             "Bài học Binary Search",
			ContentHTML:       "<p>Tìm kiếm nhị phân</p>",
			VisualizationType: "binary_search",
			IsPublished:       true,
		}

		if err := repo.CreateLesson(&l); err != nil {
			t.Fatalf("CreateLesson thất bại: %v", err)
		}

		found, err := repo.FindLessonByID(l.ID)
		if err != nil {
			t.Fatalf("FindLessonByID thất bại: %v", err)
		}

		if found.VisualizationType != "binary_search" {
			t.Errorf("Kỳ vọng VisualizationType của bài học là 'binary_search', nhận: '%s'", found.VisualizationType)
		}
	})

	// -------------------------------------------------------------
	// Test B: Exercise KHÔNG quyết định visualization của Lesson
	// Lesson có VisualizationType = "", bài tập con có visualization_type = "sorting"
	// Lesson phải giữ nguyên VisualizationType = ""
	// -------------------------------------------------------------
	t.Run("Test B - Exercise Does Not Dictate Lesson Visualization", func(t *testing.T) {
		l := Lesson{
			ChapterID:         1,
			Title:             "Bài học không có visualization",
			ContentHTML:       "<p>Lý thuyết</p>",
			VisualizationType: "",
			IsPublished:       true,
		}

		if err := repo.CreateLesson(&l); err != nil {
			t.Fatalf("CreateLesson thất bại: %v", err)
		}

		// Tạo một bài tập con liên kết với bài học này nhưng có visualization_type = 'sorting'
		_, err := db.Exec(`INSERT INTO exercises (course_id, topic_id, lesson_id, title, description, initial_code, visualization_type)
			VALUES (1, 1, ?, 'Bài tập Sorting', 'Mô tả', 'print()', 'sorting')`, l.ID)
		if err != nil {
			t.Fatalf("Insert exercise thất bại: %v", err)
		}

		found, err := repo.FindLessonByID(l.ID)
		if err != nil {
			t.Fatalf("FindLessonByID thất bại: %v", err)
		}

		if found.VisualizationType != "" {
			t.Errorf("Kỳ vọng VisualizationType của bài học vẫn là rỗng '', nhưng bị exercise ghi đè thành: '%s'", found.VisualizationType)
		}
	})

	// -------------------------------------------------------------
	// Test C: UpdateLesson cập nhật và lưu visualization_type chính xác
	// -------------------------------------------------------------
	t.Run("Test C - UpdateLesson Persists Visualization", func(t *testing.T) {
		l := Lesson{
			ChapterID:         1,
			Title:             "Bài học Ngăn xếp",
			ContentHTML:       "<p>Stack</p>",
			VisualizationType: "array",
			IsPublished:       true,
		}

		if err := repo.CreateLesson(&l); err != nil {
			t.Fatalf("CreateLesson thất bại: %v", err)
		}

		// Cập nhật sang 'stack'
		l.VisualizationType = "stack"
		if err := repo.UpdateLesson(&l); err != nil {
			t.Fatalf("UpdateLesson thất bại: %v", err)
		}

		found, err := repo.FindLessonByID(l.ID)
		if err != nil {
			t.Fatalf("FindLessonByID thất bại: %v", err)
		}

		if found.VisualizationType != "stack" {
			t.Errorf("Kỳ vọng VisualizationType sau khi update là 'stack', nhận: '%s'", found.VisualizationType)
		}
	})
}
