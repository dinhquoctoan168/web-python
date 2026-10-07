package ide

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"web_python/internal/assignment"
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
	defer database.CloseDB()

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
