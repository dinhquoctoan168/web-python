package ide

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"web_python/internal/assignment"
	"web_python/internal/auth"
	"web_python/internal/course"
	"web_python/internal/database"
	"web_python/internal/exercise"
	"web_python/internal/lesson"
)

func TestHandleIDE(t *testing.T) {
	origDir, _ := os.Getwd()
	if strings.HasSuffix(origDir, "ide") {
		_ = os.Chdir("../../")
		defer os.Chdir(origDir)
	}

	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB error: %v", err)
	}

	cRepo := course.NewRepository(db)
	cService := course.NewService(cRepo)
	lRepo := lesson.NewRepository(db)
	lService := lesson.NewService(lRepo)
	eRepo := exercise.NewRepository(db)
	eService := exercise.NewService(eRepo)
	aRepo := assignment.NewRepository(db)
	aService := assignment.NewService(aRepo)

	handler := NewHandler(cService, lService, eService, aService)

	// 1. Test GET /ide mặc định (tự động chọn môn học đầu tiên, không hardcode DSA301)
	req := httptest.NewRequest("GET", "/ide", nil)
	rec := httptest.NewRecorder()
	handler.HandleIDE(rec, req)

	if rec.Code != 200 {
		t.Fatalf("HandleIDE trả về mã HTTP %d, mong đợi 200. Nội dung: %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, `<meta name="csrf-token"`) {
		t.Errorf("Mong đợi body chứa thẻ meta csrf-token, nhận: %s", body)
	}
	if !strings.Contains(body, "Web Python IDE") {
		t.Errorf("Mong đợi body chứa tiêu đề Web Python IDE")
	}
	if !strings.Contains(body, "window.IDE_CONTEXT") {
		t.Errorf("Mong đợi body chứa window.IDE_CONTEXT")
	}
	if !strings.Contains(body, "nav-compact-bar") {
		t.Errorf("Mong đợi body chứa nav-compact-bar")
	}
	if !strings.Contains(body, "compactMenuDropdown") {
		t.Errorf("Mong đợi body chứa compactMenuDropdown")
	}

	// 2. Test GET /ide?mode=practice&course=PY101
	reqCourse := httptest.NewRequest("GET", "/ide?mode=practice&course=PY101", nil)
	recCourse := httptest.NewRecorder()
	handler.HandleIDE(recCourse, reqCourse)
	if recCourse.Code != 200 {
		t.Fatalf("HandleIDE theo course trả về mã HTTP %d, mong đợi 200", recCourse.Code)
	}

	// 3. Test GET /ide?mode=practice&id=1
	reqEx := httptest.NewRequest("GET", "/ide?mode=practice&id=1", nil)
	recEx := httptest.NewRecorder()
	handler.HandleIDE(recEx, reqEx)
	if recEx.Code != 200 {
		t.Fatalf("HandleIDE theo exercise id trả về mã HTTP %d, mong đợi 200", recEx.Code)
	}

	// 4. Test API Functions
	reqFunc := httptest.NewRequest("GET", "/api/functions?q=len", nil)
	recFunc := httptest.NewRecorder()
	handler.HandleAPIFunctions(recFunc, reqFunc)
	if recFunc.Code != 200 {
		t.Fatalf("HandleAPIFunctions trả về mã HTTP %d, mong đợi 200", recFunc.Code)
	}
}

func TestIDEExcludesLegacyExamMode(t *testing.T) {
	origDir, _ := os.Getwd()
	if strings.HasSuffix(origDir, "ide") {
		_ = os.Chdir("../../")
		defer os.Chdir(origDir)
	}

	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB error: %v", err)
	}

	cRepo := course.NewRepository(db)
	cService := course.NewService(cRepo)
	lRepo := lesson.NewRepository(db)
	lService := lesson.NewService(lRepo)
	eRepo := exercise.NewRepository(db)
	eService := exercise.NewService(eRepo)
	aRepo := assignment.NewRepository(db)
	aService := assignment.NewService(aRepo)

	handler := NewHandler(cService, lService, eService, aService)

	req := httptest.NewRequest("GET", "/ide", nil)
	rec := httptest.NewRecorder()
	handler.HandleIDE(rec, req)

	if rec.Code != 200 {
		t.Fatalf("HandleIDE trả về mã HTTP %d, mong đợi 200", rec.Code)
	}

	body := rec.Body.String()

	// Test A — IDE không có nút Start Exam
	if strings.Contains(body, "btnStartExam") {
		t.Errorf("Giao diện IDE không được chứa 'btnStartExam'")
	}
	if strings.Contains(body, "Bắt đầu thi") {
		t.Errorf("Giao diện IDE không được chứa 'Bắt đầu thi'")
	}

	// Test B — IDE không có fake timer & overlay
	if strings.Contains(body, "examModeContainer") {
		t.Errorf("Giao diện IDE không được chứa 'examModeContainer'")
	}
	if strings.Contains(body, "examSuspendedOverlay") {
		t.Errorf("Giao diện IDE không được chứa 'examSuspendedOverlay'")
	}
	if strings.Contains(body, "45:00") {
		t.Errorf("Giao diện IDE không được chứa fake timer '45:00'")
	}
}

func TestResolveIDEBackContext(t *testing.T) {
	studentUser := &auth.User{ID: 10, Username: "student", Role: auth.RoleStudent}
	teacherUser := &auth.User{ID: 2, Username: "teacher", Role: auth.RoleTeacher}

	tests := []struct {
		name          string
		queryURL      string
		user          *auth.User
		expectedURL   string
		expectedLabel string
	}{
		{
			name:          "Assignment context for student",
			queryURL:      "/ide?mode=assignment&assignment_id=5&id=101",
			user:          studentUser,
			expectedURL:   "/assignment?id=5",
			expectedLabel: "Quay lại Bài tập",
		},
		{
			name:          "Assignment context for teacher",
			queryURL:      "/ide?assignment_id=5",
			user:          teacherUser,
			expectedURL:   "/teacher/assignment?id=5",
			expectedLabel: "Quay lại Bài tập",
		},
		{
			name:          "Lesson context",
			queryURL:      "/ide?lesson_id=12",
			user:          studentUser,
			expectedURL:   "/lesson?id=12",
			expectedLabel: "Quay lại Bài học",
		},
		{
			name:          "Ignore fake exam context for student and fallback to dashboard",
			queryURL:      "/ide?exam_id=3",
			user:          studentUser,
			expectedURL:   "/dashboard",
			expectedLabel: "Quay lại Trang chủ",
		},
		{
			name:          "Ignore fake exam context for teacher and fallback to teacher home",
			queryURL:      "/ide?exam_id=3",
			user:          teacherUser,
			expectedURL:   "/teacher",
			expectedLabel: "Quay lại Bảng điều khiển",
		},
		{
			name:          "Invalid negative assignment_id falls back to student home",
			queryURL:      "/ide?assignment_id=-1",
			user:          studentUser,
			expectedURL:   "/dashboard",
			expectedLabel: "Quay lại Trang chủ",
		},
		{
			name:          "Invalid string lesson_id falls back to student home",
			queryURL:      "/ide?lesson_id=abc",
			user:          studentUser,
			expectedURL:   "/dashboard",
			expectedLabel: "Quay lại Trang chủ",
		},
		{
			name:          "Course ID for student",
			queryURL:      "/ide?course_id=2",
			user:          studentUser,
			expectedURL:   "/course?id=2",
			expectedLabel: "Quay lại Môn học",
		},
		{
			name:          "Course ID for teacher",
			queryURL:      "/ide?course_id=2",
			user:          teacherUser,
			expectedURL:   "/teacher/curriculum?course_id=2",
			expectedLabel: "Quay lại Chương trình",
		},
		{
			name:          "Course code for student",
			queryURL:      "/ide?course=PY101",
			user:          studentUser,
			expectedURL:   "/courses",
			expectedLabel: "Quay lại Môn học",
		},
		{
			name:          "Course code for teacher",
			queryURL:      "/ide?course=PY101",
			user:          teacherUser,
			expectedURL:   "/teacher/courses",
			expectedLabel: "Quay lại Môn học",
		},
		{
			name:          "Default for student",
			queryURL:      "/ide",
			user:          studentUser,
			expectedURL:   "/dashboard",
			expectedLabel: "Quay lại Trang chủ",
		},
		{
			name:          "Default for teacher",
			queryURL:      "/ide",
			user:          teacherUser,
			expectedURL:   "/teacher",
			expectedLabel: "Quay lại Bảng điều khiển",
		},
		{
			name:          "Default for anonymous",
			queryURL:      "/ide",
			user:          nil,
			expectedURL:   "/courses",
			expectedLabel: "Quay lại Môn học",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.queryURL, nil)
			url, label := ResolveIDEBackContext(req, tc.user)
			if url != tc.expectedURL {
				t.Errorf("URL kỳ vọng %s, nhận %s", tc.expectedURL, url)
			}
			if label != tc.expectedLabel {
				t.Errorf("Label kỳ vọng %s, nhận %s", tc.expectedLabel, label)
			}
		})
	}
}

func TestBuildValidatedIDEContext(t *testing.T) {
	studentUser := &auth.User{ID: 10, Username: "student", Role: auth.RoleStudent}
	teacherUser := &auth.User{ID: 2, Username: "teacher", Role: auth.RoleTeacher}

	asgn := &assignment.Assignment{
		ID:    5,
		Title: "Bài tập Tuần 1",
	}
	sampleLesson := &lesson.Lesson{
		ID:       12,
		Title:    "Biến và Kiểu dữ liệu",
		CourseID: 1,
	}
	sampleCourse := &course.Course{
		ID:   1,
		Code: "PY101",
		Name: "Lập trình Python Cơ bản",
	}
	sampleEx := &exercise.ClientExerciseDetail{
		ID:    101,
		Title: "Tính tổng hai số",
	}

	// 1. Assignment mode for student
	ctxAsgnS := BuildValidatedIDEContext(studentUser, "assignment", asgn, nil, sampleCourse, false, sampleEx)
	if ctxAsgnS.BackURL != "/assignment?id=5" {
		t.Errorf("Kỳ vọng BackURL /assignment?id=5, nhận %s", ctxAsgnS.BackURL)
	}
	if len(ctxAsgnS.Breadcrumbs) != 4 || ctxAsgnS.Breadcrumbs[2].Label != "Bài tập Tuần 1" {
		t.Errorf("Breadcrumbs không khớp cho student assignment: %+v", ctxAsgnS.Breadcrumbs)
	}

	// 2. Assignment mode for teacher
	ctxAsgnT := BuildValidatedIDEContext(teacherUser, "assignment", asgn, nil, sampleCourse, false, sampleEx)
	if ctxAsgnT.BackURL != "/teacher/assignment?id=5" {
		t.Errorf("Kỳ vọng BackURL /teacher/assignment?id=5, nhận %s", ctxAsgnT.BackURL)
	}
	if len(ctxAsgnT.Breadcrumbs) != 4 || ctxAsgnT.Breadcrumbs[2].Label != "Bài tập Tuần 1" {
		t.Errorf("Breadcrumbs không khớp cho teacher assignment: %+v", ctxAsgnT.Breadcrumbs)
	}

	// 3. Lesson context
	ctxLesson := BuildValidatedIDEContext(studentUser, "practice", nil, sampleLesson, sampleCourse, false, sampleEx)
	if ctxLesson.BackURL != "/lesson?id=12" {
		t.Errorf("Kỳ vọng BackURL /lesson?id=12, nhận %s", ctxLesson.BackURL)
	}
	if len(ctxLesson.Breadcrumbs) < 3 {
		t.Errorf("Breadcrumbs cho lesson quá ngắn: %+v", ctxLesson.Breadcrumbs)
	}

	// 4. Course context with explicit param
	ctxCourse := BuildValidatedIDEContext(studentUser, "practice", nil, nil, sampleCourse, true, sampleEx)
	if ctxCourse.BackURL != "/course?id=1" {
		t.Errorf("Kỳ vọng BackURL /course?id=1, nhận %s", ctxCourse.BackURL)
	}

	// 5. Direct open
	ctxDirect := BuildValidatedIDEContext(studentUser, "practice", nil, nil, nil, false, nil)
	if ctxDirect.BackURL != "/dashboard" {
		t.Errorf("Kỳ vọng BackURL /dashboard, nhận %s", ctxDirect.BackURL)
	}
	if len(ctxDirect.Breadcrumbs) != 2 {
		t.Errorf("Breadcrumbs direct open kỳ vọng 2 items, nhận %+v", ctxDirect.Breadcrumbs)
	}
}

func TestHandleIDE_NavigationContextRendering(t *testing.T) {
	origDir, _ := os.Getwd()
	if strings.HasSuffix(origDir, "ide") {
		_ = os.Chdir("../../")
		defer os.Chdir(origDir)
	}

	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB error: %v", err)
	}

	cRepo := course.NewRepository(db)
	cService := course.NewService(cRepo)
	lRepo := lesson.NewRepository(db)
	lService := lesson.NewService(lRepo)
	eRepo := exercise.NewRepository(db)
	eService := exercise.NewService(eRepo)
	aRepo := assignment.NewRepository(db)
	aService := assignment.NewService(aRepo)

	handler := NewHandler(cService, lService, eService, aService)

	// 1. Sinh viên vào IDE từ bài học hợp lệ của môn học
	studentUser := &auth.User{ID: 10, Username: "sv_an", FullName: "Nguyen Van An", Role: auth.RoleStudent}
	reqStudent := httptest.NewRequest("GET", "/ide?mode=practice&course=PY101&lesson_id=1", nil)
	reqStudent = reqStudent.WithContext(auth.WithUser(reqStudent.Context(), studentUser))
	recStudent := httptest.NewRecorder()
	handler.HandleIDE(recStudent, reqStudent)

	if recStudent.Code != 200 {
		t.Fatalf("HandleIDE student trả về %d", recStudent.Code)
	}
	bodyS := recStudent.Body.String()
	if !strings.Contains(bodyS, `/lesson?id=1`) {
		t.Errorf("Kỳ vọng body chứa link quay lại bài học /lesson?id=1")
	}
	if !strings.Contains(bodyS, `Quay lại Bài học`) {
		t.Errorf("Kỳ vọng body chứa nhãn 'Quay lại Bài học'")
	}
	if !strings.Contains(bodyS, `Nguyen Van An`) {
		t.Errorf("Kỳ vọng body chứa tên sinh viên 'Nguyen Van An'")
	}

	// 1b. Sinh viên vào IDE với lesson_id xung đột với course khác -> Fallback về course
	reqConflict := httptest.NewRequest("GET", "/ide?mode=practice&course=PY101&lesson_id=99999", nil)
	reqConflict = reqConflict.WithContext(auth.WithUser(reqConflict.Context(), studentUser))
	recConflict := httptest.NewRecorder()
	handler.HandleIDE(recConflict, reqConflict)
	if recConflict.Code != 200 {
		t.Fatalf("HandleIDE conflict trả về %d", recConflict.Code)
	}
	bodyConflict := recConflict.Body.String()
	if strings.Contains(bodyConflict, `lesson_id=99999`) || strings.Contains(bodyConflict, `/lesson?id=99999`) {
		t.Errorf("Ngữ cảnh bài học không tồn tại/xung đột không được xuất hiện trong BackURL")
	}

	// 2. Giảng viên vào IDE
	teacherUser := &auth.User{ID: 2, Username: "gv_binh", FullName: "Tran Van Binh", Role: auth.RoleTeacher}
	reqTeacher := httptest.NewRequest("GET", "/ide", nil)
	reqTeacher = reqTeacher.WithContext(auth.WithUser(reqTeacher.Context(), teacherUser))
	recTeacher := httptest.NewRecorder()
	handler.HandleIDE(recTeacher, reqTeacher)

	if recTeacher.Code != 200 {
		t.Fatalf("HandleIDE teacher trả về %d", recTeacher.Code)
	}
	bodyT := recTeacher.Body.String()
	if !strings.Contains(bodyT, `/teacher`) {
		t.Errorf("Kỳ vọng body chứa link quay lại teacher platform")
	}
	if !strings.Contains(bodyT, `Tran Van Binh`) {
		t.Errorf("Kỳ vọng body chứa tên giảng viên 'Tran Van Binh'")
	}
}
