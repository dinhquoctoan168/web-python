package lesson

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"web_python/internal/auth"

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
			visualization_type TEXT DEFAULT '',
			visualization_config TEXT DEFAULT '',
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

	// 8. Kiểm tra VisualizationType trực thuộc bài học (Phase 3)
	_, err = service.UpdateLesson(l1.ID, l1.Title, l1.ContentHTML, l1.OrderNum, l1.IsPublished, "binary_search")
	if err != nil {
		t.Fatalf("UpdateLesson lỗi: %v", err)
	}

	vDetail, _, err := service.GetLessonDetail(l1.ID, false)
	if err != nil {
		t.Fatalf("GetLessonDetail lỗi: %v", err)
	}
	if vDetail.VisualizationType != "binary_search" {
		t.Errorf("Mong đợi VisualizationType 'binary_search', nhận: '%s'", vDetail.VisualizationType)
	}
}

func TestLessonTemplatesRender(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, _ = db.Exec(`INSERT INTO courses (id, code, name) VALUES (1, 'PY101', 'Python cơ bản')`)

	repo := NewRepository(db)
	service := NewService(repo)
	handler := NewHandler(service)

	ch, err := service.CreateChapter(1, "Chương 1", "Mô tả", 1)
	if err != nil {
		t.Fatalf("Lỗi tạo chapter: %v", err)
	}
	l, err := service.CreateLesson(ch.ID, "Bài 1", "<p>Nội dung</p>", 1, true, "")
	if err != nil {
		t.Fatalf("Lỗi tạo lesson: %v", err)
	}

	teacherUser := &auth.User{
		ID:       1,
		Username: "teacher1",
		FullName: "Thầy Giáo",
		Role:     auth.RoleTeacher,
	}

	// 1. Sinh viên / Người dùng xem chi tiết bài học
	reqLesson := httptest.NewRequest("GET", "/lesson?id=1", nil)
	reqLesson = reqLesson.WithContext(auth.WithUser(reqLesson.Context(), teacherUser))
	recLesson := httptest.NewRecorder()
	handler.HandleStudentLesson(recLesson, reqLesson)
	if recLesson.Code != http.StatusOK {
		t.Fatalf("HandleStudentLesson kỳ vọng 200 OK, nhận %d: %s", recLesson.Code, recLesson.Body.String())
	}
	if !strings.Contains(recLesson.Body.String(), "breadcrumb") {
		t.Errorf("Kỳ vọng trang bài học chứa breadcrumb")
	}

	// 2. Giảng viên xem quản lý chương trình đào tạo
	reqCurr := httptest.NewRequest("GET", "/teacher/curriculum?course_id=1", nil)
	reqCurr = reqCurr.WithContext(auth.WithUser(reqCurr.Context(), teacherUser))
	recCurr := httptest.NewRecorder()
	handler.HandleTeacherCurriculum(recCurr, reqCurr)
	if recCurr.Code != http.StatusOK {
		t.Fatalf("HandleTeacherCurriculum kỳ vọng 200 OK, nhận %d: %s", recCurr.Code, recCurr.Body.String())
	}
	if !strings.Contains(recCurr.Body.String(), "breadcrumb") {
		t.Errorf("Kỳ vọng trang curriculum chứa breadcrumb")
	}

	// 3. Giảng viên mở form tạo bài học mới
	reqNew := httptest.NewRequest("GET", "/teacher/lesson/new?chapter_id=1", nil)
	reqNew = reqNew.WithContext(auth.WithUser(reqNew.Context(), teacherUser))
	recNew := httptest.NewRecorder()
	handler.HandleTeacherNewLessonForm(recNew, reqNew)
	if recNew.Code != http.StatusOK {
		t.Fatalf("HandleTeacherNewLessonForm kỳ vọng 200 OK, nhận %d: %s", recNew.Code, recNew.Body.String())
	}
	if !strings.Contains(recNew.Body.String(), "breadcrumb") {
		t.Errorf("Kỳ vọng form bài học chứa breadcrumb")
	}

	// 4. Giảng viên mở form chỉnh sửa bài học
	reqEdit := httptest.NewRequest("GET", "/teacher/lesson/edit?id=1", nil)
	reqEdit = reqEdit.WithContext(auth.WithUser(reqEdit.Context(), teacherUser))
	recEdit := httptest.NewRecorder()
	handler.HandleTeacherEditLessonForm(recEdit, reqEdit)
	if recEdit.Code != http.StatusOK {
		t.Fatalf("HandleTeacherEditLessonForm kỳ vọng 200 OK, nhận %d: %s", recEdit.Code, recEdit.Body.String())
	}
	if !strings.Contains(recEdit.Body.String(), "breadcrumb") {
		t.Errorf("Kỳ vọng form chỉnh sửa bài học chứa breadcrumb")
	}
	_ = l
}
