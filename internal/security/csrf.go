package security

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
)

const (
	CSRFCookieName = "csrf_token"
	CSRFHeaderName = "X-CSRF-Token"
	CSRFFormField  = "csrf_token"
)

type contextKey string

const csrfCtxKey contextKey = "csrf_token"

// GenerateToken sinh chuỗi CSRF token ngẫu nhiên độ dài 32 bytes (64 ký tự hex)
func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetTokenFromContext lấy CSRF token từ context
func GetTokenFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(csrfCtxKey).(string); ok {
		return val
	}
	return ""
}

// SetCSRFCookie gán cookie CSRF token vào response
func SetCSRFCookie(w http.ResponseWriter, token string, isSecure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: false, // Để client JavaScript có thể đọc và gửi trong header X-CSRF-Token
		SameSite: http.SameSiteLaxMode,
		Secure:   isSecure,
	})
}

// CSRFMiddleware cung cấp middleware bảo vệ chống tấn công CSRF
// Bắt buộc xác thực với các phương thức thay đổi dữ liệu (POST, PUT, DELETE, PATCH)
func CSRFMiddleware(exemptPaths ...string) func(http.Handler) http.Handler {
	exemptMap := make(map[string]bool)
	for _, p := range exemptPaths {
		exemptMap[p] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Lấy hoặc sinh CSRF token từ Cookie
			var token string
			cookie, err := r.Cookie(CSRFCookieName)
			if err == nil && cookie.Value != "" {
				token = cookie.Value
			} else {
				token, _ = GenerateToken()
				SetCSRFCookie(w, token, r.TLS != nil)
			}

			// Gắn token vào context
			ctx := context.WithValue(r.Context(), csrfCtxKey, token)
			r = r.WithContext(ctx)

			// 2. Bỏ qua kiểm tra với các phương thức an toàn
			method := strings.ToUpper(r.Method)
			if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions || method == http.MethodTrace {
				next.ServeHTTP(w, r)
				return
			}

			// 3. Kiểm tra danh sách đường dẫn ngoại lệ (ví dụ: /login trước khi có session)
			if exemptMap[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			// 4. Trích xuất token được gửi lên từ client
			submittedToken := r.Header.Get(CSRFHeaderName)
			if submittedToken == "" {
				// Thử đọc từ Form value
				submittedToken = r.FormValue(CSRFFormField)
			}

			// 5. So sánh token (dùng ConstantTimeCompare để chống timing attack)
			if submittedToken == "" || subtle.ConstantTimeCompare([]byte(submittedToken), []byte(token)) != 1 {
				http.Error(w, "Lỗi bảo mật CSRF: Token không hợp lệ hoặc bị thiếu", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
