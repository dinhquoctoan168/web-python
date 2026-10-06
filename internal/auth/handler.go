package auth

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
)

// Handler xử lý các yêu cầu HTTP liên quan đến Authentication
type Handler struct {
	service     *Service
	tmplPattern string
}

// NewHandler khởi tạo Handler
func NewHandler(service *Service, tmplPattern string) *Handler {
	return &Handler{
		service:     service,
		tmplPattern: tmplPattern,
	}
}

// ShowLoginPage hiển thị trang đăng nhập
func (h *Handler) ShowLoginPage(w http.ResponseWriter, r *http.Request) {
	// Nếu đã đăng nhập, chuyển hướng thẳng vào /ide
	if user := GetUser(r.Context()); user != nil {
		if user.Role == RoleTeacher || user.Role == RoleAdmin {
			http.Redirect(w, r, "/teacher", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		}
		return
	}

	h.renderLoginTemplate(w, "")
}

// HandleLogin xử lý form đăng nhập
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Phương thức không hợp lệ", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		h.renderLoginTemplate(w, "Dữ liệu gửi lên không hợp lệ")
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		h.renderLoginTemplate(w, "Vui lòng nhập đầy đủ tên đăng nhập và mật khẩu")
		return
	}

	user, session, err := h.service.Authenticate(username, password)
	if err != nil {
		h.renderLoginTemplate(w, err.Error())
		return
	}

	// Đặt cookie phiên đăng nhập an toàn
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    session.ID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
	})

	if user.Role == RoleTeacher || user.Role == RoleAdmin {
		http.Redirect(w, r, "/teacher", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

// HandleLogout xử lý đăng xuất
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(SessionCookieName)
	if err == nil && cookie.Value != "" {
		_ = h.service.Logout(cookie.Value)
	}

	// Xoá cookie
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// HandleCurrentUser trả về thông tin user đang đăng nhập dạng JSON
func (h *Handler) HandleCurrentUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	user := GetUser(r.Context())
	if user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok": false,
			"error": map[string]string{
				"code":    "UNAUTHORIZED",
				"message": "Chưa đăng nhập",
			},
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":   true,
		"data": user,
	})
}

func (h *Handler) renderLoginTemplate(w http.ResponseWriter, errorMsg string) {
	tmplPath := h.tmplPattern
	if tmplPath == "" {
		tmplPath = filepath.Join("web", "templates", "login.html")
	}

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Không tìm thấy giao diện đăng nhập: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Error string
	}{
		Error: errorMsg,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}
