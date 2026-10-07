package lesson

import (
	"strings"
	"testing"
)

func TestLessonSanitize(t *testing.T) {
	db := setupVisualizationTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	service := NewService(repo)

	// -------------------------------------------------------------
	// Test A: Loại bỏ thẻ <script> và nội dung thực thi độc hại
	// -------------------------------------------------------------
	t.Run("Test A - Script Tag Removal", func(t *testing.T) {
		input := `<p>Hello</p><script>alert(1)</script>`
		les, err := service.CreateLesson(1, "Bài học XSS 1", input, 1, true)
		if err != nil {
			t.Fatalf("CreateLesson thất bại: %v", err)
		}

		if strings.Contains(les.ContentHTML, "<script") {
			t.Errorf("Kỳ vọng không chứa '<script', nhưng nhận: %s", les.ContentHTML)
		}
		if strings.Contains(les.ContentHTML, "alert(") {
			t.Errorf("Kỳ vọng không chứa 'alert(', nhưng nhận: %s", les.ContentHTML)
		}

		// Kiểm tra đọc lại từ DB
		found, err := repo.FindLessonByID(les.ID)
		if err != nil {
			t.Fatalf("FindLessonByID thất bại: %v", err)
		}
		if strings.Contains(found.ContentHTML, "<script") || strings.Contains(found.ContentHTML, "alert(") {
			t.Errorf("Dữ liệu lưu trong DB chưa được sanitize: %s", found.ContentHTML)
		}
	})

	// -------------------------------------------------------------
	// Test B: Loại bỏ inline event handler (onerror, onclick, ...)
	// -------------------------------------------------------------
	t.Run("Test B - Inline Event Handler Removal", func(t *testing.T) {
		input := `<img src="x" onerror="alert(1)">`
		les, err := service.CreateLesson(1, "Bài học XSS 2", input, 2, true)
		if err != nil {
			t.Fatalf("CreateLesson thất bại: %v", err)
		}

		if strings.Contains(les.ContentHTML, "onerror") {
			t.Errorf("Kỳ vọng không chứa 'onerror', nhưng nhận: %s", les.ContentHTML)
		}
	})

	// -------------------------------------------------------------
	// Test C: Loại bỏ javascript: scheme trong đường dẫn href
	// -------------------------------------------------------------
	t.Run("Test C - Javascript URL Removal", func(t *testing.T) {
		input := `<a href="javascript:alert(1)">click</a>`
		les, err := service.CreateLesson(1, "Bài học XSS 3", input, 3, true)
		if err != nil {
			t.Fatalf("CreateLesson thất bại: %v", err)
		}

		if strings.Contains(les.ContentHTML, "javascript:") {
			t.Errorf("Kỳ vọng không chứa 'javascript:', nhưng nhận: %s", les.ContentHTML)
		}
	})

	// -------------------------------------------------------------
	// Test D: Giữ lại các thẻ HTML hợp lệ và an toàn
	// -------------------------------------------------------------
	t.Run("Test D - Valid HTML Preservation", func(t *testing.T) {
		input := `<h2>Binary Search</h2><p>Nội dung</p><pre><code>print("hello")</code></pre>`
		les, err := service.CreateLesson(1, "Bài học hợp lệ", input, 4, true)
		if err != nil {
			t.Fatalf("CreateLesson thất bại: %v", err)
		}

		if !strings.Contains(les.ContentHTML, "<h2>") || !strings.Contains(les.ContentHTML, "Binary Search") {
			t.Errorf("Mất thẻ h2 hoặc nội dung tiêu đề: %s", les.ContentHTML)
		}
		if !strings.Contains(les.ContentHTML, "<p>") || !strings.Contains(les.ContentHTML, "Nội dung") {
			t.Errorf("Mất thẻ p: %s", les.ContentHTML)
		}
		if !strings.Contains(les.ContentHTML, "<pre>") || !strings.Contains(les.ContentHTML, "<code>") {
			t.Errorf("Mất thẻ code block (pre/code): %s", les.ContentHTML)
		}
	})
}
