package lesson

import (
	"strings"
	"testing"
)

func TestLegacyLessonSanitizeOnRender(t *testing.T) {
	db := setupVisualizationTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	service := NewService(repo)

	// -------------------------------------------------------------
	// Test A: Legacy Script Tag in DB -> Render SafeHTML must be clean
	// -------------------------------------------------------------
	t.Run("Test A - Legacy Script Tag Stripped on Render", func(t *testing.T) {
		res, err := db.Exec(`INSERT INTO lessons (chapter_id, title, content_html, order_num, is_published) 
			VALUES (1, 'Legacy XSS 1', '<p>Hello</p><script>alert(1)</script>', 1, 1)`)
		if err != nil {
			t.Fatalf("Chèn dữ liệu cũ thất bại: %v", err)
		}
		lessonID, _ := res.LastInsertId()

		les, _, err := service.GetLessonDetail(int(lessonID), false)
		if err != nil {
			t.Fatalf("GetLessonDetail thất bại: %v", err)
		}

		safeStr := string(les.SafeHTML)
		if strings.Contains(safeStr, "<script") {
			t.Errorf("SafeHTML không được chứa '<script', nhưng nhận: %s", safeStr)
		}
		if strings.Contains(safeStr, "alert(") {
			t.Errorf("SafeHTML không được chứa 'alert(', nhưng nhận: %s", safeStr)
		}
	})

	// -------------------------------------------------------------
	// Test B: Legacy Inline Event Handler (onerror, ...)
	// -------------------------------------------------------------
	t.Run("Test B - Legacy Inline Event Handler Stripped on Render", func(t *testing.T) {
		res, err := db.Exec(`INSERT INTO lessons (chapter_id, title, content_html, order_num, is_published) 
			VALUES (1, 'Legacy XSS 2', '<img src="x" onerror="alert(1)">', 2, 1)`)
		if err != nil {
			t.Fatalf("Chèn dữ liệu cũ thất bại: %v", err)
		}
		lessonID, _ := res.LastInsertId()

		les, _, err := service.GetLessonDetail(int(lessonID), false)
		if err != nil {
			t.Fatalf("GetLessonDetail thất bại: %v", err)
		}

		safeStr := string(les.SafeHTML)
		if strings.Contains(safeStr, "onerror") {
			t.Errorf("SafeHTML không được chứa 'onerror', nhưng nhận: %s", safeStr)
		}
	})

	// -------------------------------------------------------------
	// Test C: Legacy Javascript URL
	// -------------------------------------------------------------
	t.Run("Test C - Legacy Javascript URL Stripped on Render", func(t *testing.T) {
		res, err := db.Exec(`INSERT INTO lessons (chapter_id, title, content_html, order_num, is_published) 
			VALUES (1, 'Legacy XSS 3', '<a href="javascript:alert(1)">click</a>', 3, 1)`)
		if err != nil {
			t.Fatalf("Chèn dữ liệu cũ thất bại: %v", err)
		}
		lessonID, _ := res.LastInsertId()

		les, _, err := service.GetLessonDetail(int(lessonID), false)
		if err != nil {
			t.Fatalf("GetLessonDetail thất bại: %v", err)
		}

		safeStr := string(les.SafeHTML)
		if strings.Contains(safeStr, "javascript:") {
			t.Errorf("SafeHTML không được chứa 'javascript:', nhưng nhận: %s", safeStr)
		}
	})

	// -------------------------------------------------------------
	// Test D: Valid HTML Structure Preserved on Render
	// -------------------------------------------------------------
	t.Run("Test D - Valid HTML Structure Preserved on Render", func(t *testing.T) {
		validHTML := `<h2>Binary Search</h2><p>Hello</p><pre><code class="language-python">print("x")</code></pre>`
		res, err := db.Exec(`INSERT INTO lessons (chapter_id, title, content_html, order_num, is_published) 
			VALUES (1, 'Valid Lesson', ?, 4, 1)`, validHTML)
		if err != nil {
			t.Fatalf("Chèn dữ liệu cũ thất bại: %v", err)
		}
		lessonID, _ := res.LastInsertId()

		les, _, err := service.GetLessonDetail(int(lessonID), false)
		if err != nil {
			t.Fatalf("GetLessonDetail thất bại: %v", err)
		}

		safeStr := string(les.SafeHTML)
		if !strings.Contains(safeStr, "<h2>") || !strings.Contains(safeStr, "Binary Search") {
			t.Errorf("SafeHTML làm mất thẻ h2 hoặc nội dung: %s", safeStr)
		}
		if !strings.Contains(safeStr, "<p>") || !strings.Contains(safeStr, "Hello") {
			t.Errorf("SafeHTML làm mất thẻ p hoặc nội dung: %s", safeStr)
		}
		if !strings.Contains(safeStr, "<pre>") || !strings.Contains(safeStr, "<code") {
			t.Errorf("SafeHTML làm mất thẻ pre/code: %s", safeStr)
		}
	})
}
