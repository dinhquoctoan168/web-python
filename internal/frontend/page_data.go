package frontend

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"web_python/internal/auth"
	"web_python/internal/security"
)

// BasePageData chứa các thông tin dùng chung cho mọi trang giao diện HTML
type BasePageData struct {
	Title     string         `json:"title"`
	User      *auth.User     `json:"user"`
	CSRFToken string         `json:"csrf_token"`
	ActiveNav string         `json:"active_nav"`
	Nav       NavigationData `json:"nav"`
}

// NewBasePageData tạo BasePageData chuẩn từ http.Request
func NewBasePageData(r *http.Request, title string, activeNav ...string) BasePageData {
	user := auth.GetUser(r.Context())
	csrf := security.GetTokenFromContext(r.Context())
	activeKey := ""
	if len(activeNav) > 0 {
		activeKey = activeNav[0]
	}
	return BasePageData{
		Title:     title,
		User:      user,
		CSRFToken: csrf,
		ActiveNav: activeKey,
		Nav:       BuildNavigationData(user, activeKey, nil, "", "", csrf),
	}
}

// InjectCSRFToMap tự động thêm CSRFToken, User và Nav vào map dữ liệu nếu chưa có
func InjectCSRFToMap(r *http.Request, data map[string]any) map[string]any {
	if data == nil {
		data = make(map[string]any)
	}
	csrf := security.GetTokenFromContext(r.Context())
	if _, ok := data["CSRFToken"]; !ok {
		data["CSRFToken"] = csrf
	}
	var u *auth.User
	if existingUser, ok := data["User"].(*auth.User); ok {
		u = existingUser
	} else if userObj := auth.GetUser(r.Context()); userObj != nil {
		data["User"] = userObj
		u = userObj
	}
	if _, ok := data["Nav"]; !ok {
		activeKey := ""
		if ak, ok := data["ActiveNav"].(string); ok {
			activeKey = ak
		}
		data["Nav"] = BuildNavigationData(u, activeKey, nil, "", "", csrf)
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

// GetSharedTemplatePaths trả về danh sách đường dẫn tới các shared partial templates
func GetSharedTemplatePaths() []string {
	paths := []string{
		filepath.Join("web", "templates", "shared", "app_nav.html"),
		filepath.Join("web", "templates", "shared", "breadcrumb.html"),
	}
	var resolved []string
	for _, p := range paths {
		resolved = append(resolved, ResolveTemplatePath(p))
	}
	return resolved
}

// ParseFilesWithShared phân giải template bao gồm cả các shared templates (main template luôn ở vị trí đầu)
func ParseFilesWithShared(filenames ...string) (*template.Template, error) {
	if len(filenames) == 0 {
		return nil, os.ErrInvalid
	}
	mainFile := ResolveTemplatePath(filenames[0])
	allFiles := []string{mainFile}
	allFiles = append(allFiles, GetSharedTemplatePaths()...)
	for _, f := range filenames[1:] {
		allFiles = append(allFiles, ResolveTemplatePath(f))
	}
	return template.New(filepath.Base(mainFile)).ParseFiles(allFiles...)
}
