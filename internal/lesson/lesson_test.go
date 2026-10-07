package lesson

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Không thể mở sqlite :memory:: %v", err)
	}

	queries := []string{
		`CREATE TABLE courses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			description TEXT,
			status TEXT DEFAULT 'active'
		);`,
		`CREATE TABLE chapters (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			course_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			order_num INTEGER DEFAULT 0,
			FOREIGN KEY(course_id) REFERENCES courses(id)
		);`,
		`CREATE TABLE lessons (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			chapter_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content_html TEXT,
			order_num INTEGER DEFAULT 0,
			is_published INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(chapter_id) REFERENCES chapters(id)
		);`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("Lỗi tạo bảng test: %v", err)
		}
	}
	return db
}

func TestLessonAndCurriculum(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, _ = db.Exec(`INSERT INTO courses (id, code, name) VALUES (1, 'PY101', 'Python cơ bản')`)

	repo := NewRepository(db)
	service := NewService(repo)

	// 1. Tạo chương mới
	chap, err := service.CreateChapter(1, "Chương 1", "Mô tả chương 1", 1)
	if err != nil {
		t.Fatalf("CreateChapter lỗi: %v", err)
	}
	if chap.ID <= 0 {
		t.Errorf("ID chương không hợp lệ: %d", chap.ID)
	}

	// 2. Tạo bài học xuất bản (published) và bài học bản nháp (draft)
	l1, err := service.CreateLesson(chap.ID, "Bài 1", "<p>Nội dung 1</p>", 1, true)
	if err != nil {
		t.Fatalf("CreateLesson lỗi: %v", err)
	}
	l2, err := service.CreateLesson(chap.ID, "Bài 2 Nháp", "<p>Bản nháp</p>", 2, false)
	if err != nil {
		t.Fatalf("CreateLesson draft lỗi: %v", err)
	}

	// 3. Sinh viên chỉ thấy bài đã xuất bản
	studentCurr, err := service.GetCurriculum(1, false)
	if err != nil {
		t.Fatalf("GetCurriculum cho sinh viên lỗi: %v", err)
	}
	if len(studentCurr.Chapters) != 1 {
		t.Fatalf("Kỳ vọng 1 chương, nhận %d", len(studentCurr.Chapters))
	}
	if len(studentCurr.Chapters[0].Lessons) != 1 {
		t.Errorf("Sinh viên kỳ vọng thấy 1 bài xuất bản, nhận %d", len(studentCurr.Chapters[0].Lessons))
	}

	// 4. Giảng viên thấy cả bài xuất bản và bài nháp
	teacherCurr, err := service.GetCurriculum(1, true)
	if err != nil {
		t.Fatalf("GetCurriculum cho giảng viên lỗi: %v", err)
	}
	if len(teacherCurr.Chapters[0].Lessons) != 2 {
		t.Errorf("Giảng viên kỳ vọng thấy 2 bài, nhận %d", len(teacherCurr.Chapters[0].Lessons))
	}

	// 5. Chi tiết bài học
	lessonDetail, sidebarCurr, err := service.GetLessonDetail(l1.ID, false)
	if err != nil {
		t.Fatalf("GetLessonDetail lỗi: %v", err)
	}
	if lessonDetail.Title != "Bài 1" {
		t.Errorf("Tiêu đề bài học không khớp: %s", lessonDetail.Title)
	}
	if len(sidebarCurr.Chapters) == 0 {
		t.Errorf("Sidebar curriculum không có chương")
	}

	// 6. Sinh viên không thể xem bài nháp
	_, _, err = service.GetLessonDetail(l2.ID, false)
	if err == nil {
		t.Errorf("Kỳ vọng lỗi khi sinh viên truy cập bài nháp")
	}

	// 7. Cập nhật bài học
	updated, err := service.UpdateLesson(l2.ID, "Bài 2 Đã Sửa", "<p>Nội dung mới</p>", 2, true)
	if err != nil {
		t.Fatalf("UpdateLesson lỗi: %v", err)
	}
	if !updated.IsPublished {
		t.Errorf("Bài học chưa được cập nhật sang xuất bản")
	}

	// 8. Kiểm tra VisualizationType nếu bài học có bài tập liên kết
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS exercises (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		lesson_id INTEGER,
		title TEXT,
		visualization_type TEXT
	);`)
	_, _ = db.Exec(`INSERT INTO exercises (lesson_id, title, visualization_type) VALUES (?, 'Bài tập tìm kiếm', 'binary_search')`, l1.ID)

	vDetail, _, err := service.GetLessonDetail(l1.ID, false)
	if err != nil {
		t.Fatalf("GetLessonDetail lỗi: %v", err)
	}
	if vDetail.VisualizationType != "binary_search" {
		t.Errorf("Mong đợi VisualizationType 'binary_search', nhận: '%s'", vDetail.VisualizationType)
	}
}
