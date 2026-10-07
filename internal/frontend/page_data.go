package frontend

import (
	"net/http"
	"os"
	"path/filepath"

	"web_python/internal/auth"
	"web_python/internal/security"
)

// BasePageData chứa các thông tin dùng chung cho mọi trang giao diện HTML
type BasePageData struct {
	Title     string     `json:"title"`
	User      *auth.User `json:"user"`
	CSRFToken string     `json:"csrf_token"`
	ActiveNav string     `json:"active_nav"`
}

// NewBasePageData tạo BasePageData chuẩn từ http.Request
func NewBasePageData(r *http.Request, title string) BasePageData {
	return BasePageData{
		Title:     title,
		User:      auth.GetUser(r.Context()),
		CSRFToken: security.GetTokenFromContext(r.Context()),
	}
}

// InjectCSRFToMap tự động thêm CSRFToken và User vào map dữ liệu nếu chưa có
func InjectCSRFToMap(r *http.Request, data map[string]any) map[string]any {
	if data == nil {
		data = make(map[string]any)
	}
	if _, ok := data["CSRFToken"]; !ok {
		data["CSRFToken"] = security.GetTokenFromContext(r.Context())
	}
	if _, ok := data["User"]; !ok {
		if u := auth.GetUser(r.Context()); u != nil {
			data["User"] = u
		}
	}
	return data
}

// ResolveTemplatePath tìm đường dẫn template hợp lệ (hỗ trợ cả khi chạy từ test package con)
func ResolveTemplatePath(tmplPath string) string {
	if _, err := os.Stat(tmplPath); err == nil {
		return tmplPath
	}
	alt := filepath.Join("..", "..", tmplPath)
	if _, err := os.Stat(alt); err == nil {
		return alt
	}
	return tmplPath
}

