package frontend

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"web_python/internal/database"
)

func TestHandleIDE(t *testing.T) {
	// Chuyển working directory về gốc dự án để nạp đúng template web/templates/*.html
	origDir, _ := os.Getwd()
	if strings.HasSuffix(origDir, "frontend") {
		_ = os.Chdir("../../")
		defer os.Chdir(origDir)
	}

	_, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB error: %v", err)
	}
	defer database.CloseDB()

	req := httptest.NewRequest("GET", "/ide", nil)
	rec := httptest.NewRecorder()

	HandleIDE(rec, req)

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
}
