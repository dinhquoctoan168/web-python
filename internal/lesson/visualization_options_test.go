package lesson

import (
	"os"
	"strings"
	"testing"
)

func TestVisualizationOptionsInForm(t *testing.T) {
	// Kiểm tra template teacher/lesson_form.html chỉ hiển thị các option đã implement
	contentBytes, err := os.ReadFile("../../web/templates/teacher/lesson_form.html")
	if err != nil {
		// Thử đường dẫn từ thư mục gốc nếu chạy từ root
		contentBytes, err = os.ReadFile("web/templates/teacher/lesson_form.html")
		if err != nil {
			t.Fatalf("Không thể đọc template teacher/lesson_form.html: %v", err)
		}
	}
	formContent := string(contentBytes)

	// Test A1: Bắt buộc có các loại đã implement
	supported := []string{`value="array"`, `value="stack"`, `value="binary_search"`, `value="sorting"`}
	for _, s := range supported {
		if !strings.Contains(formContent, s) {
			t.Errorf("Form bài học thiếu visualization đã implement: %s", s)
		}
	}

	// Test A2: Tuyệt đối KHÔNG có các loại chưa implement (queue, tree, graph)
	unsupported := []string{`value="queue"`, `value="tree"`, `value="graph"`}
	for _, u := range unsupported {
		if strings.Contains(formContent, u) {
			t.Errorf("Form bài học không được chứa visualization chưa cài đặt: %s", u)
		}
	}
}

func TestBackendVisualizationValidation(t *testing.T) {
	db := setupVisualizationTestDB(t)
	defer db.Close()

	repo := NewRepository(db)
	service := NewService(repo)

	// -------------------------------------------------------------
	// Test B: Backend reject unsupported type (graph, tree, queue)
	// -------------------------------------------------------------
	t.Run("Test B - Reject Unsupported Types", func(t *testing.T) {
		unsupportedTypes := []string{"graph", "tree", "queue", "linked_list", "unknown"}
		for _, vt := range unsupportedTypes {
			_, err := service.CreateLesson(1, "Bài test "+vt, "<p>Noi dung</p>", 1, true, vt)
			if err == nil {
				t.Errorf("CreateLesson phải từ chối unsupported type '%s'", vt)
			}
		}

		// Tạo bài học hợp lệ trước, sau đó thử update sang unsupported type
		validLesson, err := service.CreateLesson(1, "Bài hợp lệ", "<p>Noi dung</p>", 1, true, "array")
		if err != nil {
			t.Fatalf("CreateLesson hợp lệ thất bại: %v", err)
		}

		_, err = service.UpdateLesson(validLesson.ID, validLesson.Title, validLesson.ContentHTML, 1, true, "graph")
		if err == nil {
			t.Errorf("UpdateLesson phải từ chối unsupported type 'graph'")
		}
	})

	// -------------------------------------------------------------
	// Test C: Supported types are accepted and persisted
	// -------------------------------------------------------------
	t.Run("Test C - Accept Supported Types", func(t *testing.T) {
		supportedTypes := []string{"array", "stack", "binary_search", "sorting"}
		for i, vt := range supportedTypes {
			l, err := service.CreateLesson(1, "Bài test "+vt, "<p>Noi dung</p>", i+1, true, vt)
			if err != nil {
				t.Fatalf("CreateLesson bị lỗi với supported type '%s': %v", vt, err)
			}
			if l.VisualizationType != vt {
				t.Errorf("Kỳ vọng VisualizationType '%s', nhận '%s'", vt, l.VisualizationType)
			}

			// Kiểm tra DB lưu đúng
			found, err := repo.FindLessonByID(l.ID)
			if err != nil || found.VisualizationType != vt {
				t.Errorf("DB không lưu đúng VisualizationType '%s'", vt)
			}
		}
	})

	// -------------------------------------------------------------
	// Test D: Empty or 'none' type is valid
	// -------------------------------------------------------------
	t.Run("Test D - Empty or None Type Valid", func(t *testing.T) {
		l1, err := service.CreateLesson(1, "Bài không viz 1", "<p>Noi dung</p>", 1, true, "")
		if err != nil {
			t.Errorf("CreateLesson với empty string thất bại: %v", err)
		}
		if l1.VisualizationType != "" {
			t.Errorf("Kỳ vọng empty string, nhận '%s'", l1.VisualizationType)
		}

		l2, err := service.CreateLesson(1, "Bài không viz 2", "<p>Noi dung</p>", 2, true, "none")
		if err != nil {
			t.Errorf("CreateLesson với 'none' thất bại: %v", err)
		}
		if l2.VisualizationType != "" {
			t.Errorf("Kỳ vọng chuẩn hóa về rỗng '', nhận '%s'", l2.VisualizationType)
		}
	})
}
