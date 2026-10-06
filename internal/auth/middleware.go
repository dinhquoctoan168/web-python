package auth

import (
	"context"
	"net/http"
)

type contextKey string

const userCtxKey contextKey = "current_user"

// Middleware quản lý việc gán User vào Request Context
type Middleware struct {
	service *Service
}

// NewMiddleware khởi tạo Middleware
func NewMiddleware(service *Service) *Middleware {
	return &Middleware{service: service}
}

// AuthenticateMiddleware đọc cookie session và đưa thông tin user vào context
func (m *Middleware) AuthenticateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(SessionCookieName)
		if err == nil && cookie.Value != "" {
			user, err := m.service.ValidateSession(cookie.Value)
			if err == nil && user != nil {
				ctx := context.WithValue(r.Context(), userCtxKey, user)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// GetUser lấy thông tin User từ Context
func GetUser(ctx context.Context) *User {
	u, ok := ctx.Value(userCtxKey).(*User)
	if !ok {
		return nil
	}
	return u
}

// RequireLogin chuyển hướng sang /login nếu người dùng chưa đăng nhập
func RequireLogin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// RequireTeacher yêu cầu người dùng phải có quyền teacher hoặc admin
func RequireTeacher(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if user.Role != RoleTeacher && user.Role != RoleAdmin {
			http.Error(w, "Truy cập bị từ chối: Yêu cầu quyền Giảng viên", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

// RequireAdmin yêu cầu người dùng phải có quyền admin
func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if user.Role != RoleAdmin {
			http.Error(w, "Truy cập bị từ chối: Yêu cầu quyền Quản trị viên", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}
