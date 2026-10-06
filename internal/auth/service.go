package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	SessionCookieName = "session_id"
	SessionDuration   = 24 * time.Hour
	hashIterations    = 10000
)

var (
	ErrInvalidCredentials = errors.New("tên đăng nhập hoặc mật khẩu không chính xác")
	ErrAccountInactive    = errors.New("tài khoản hiện đang bị vô hiệu hoá")
	ErrSessionExpired     = errors.New("phiên đăng nhập đã hết hạn")
)

// Service cung cấp logic xác thực và quản lý phiên
type Service struct {
	repo *Repository
}

// NewService khởi tạo Service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// HashPassword băm mật khẩu với salt ngẫu nhiên sử dụng crypto/sha256 nhiều vòng
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("không thể tạo salt: %w", err)
	}

	data := append(salt, []byte(password)...)
	for i := 0; i < hashIterations; i++ {
		h := sha256.Sum256(data)
		data = h[:]
	}

	return fmt.Sprintf("%s$%s", hex.EncodeToString(salt), hex.EncodeToString(data)), nil
}

// CheckPassword kiểm tra tính trùng khớp của mật khẩu một cách an toàn (timing-safe)
func CheckPassword(password, encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 2 {
		return false
	}

	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false
	}

	expectedHash, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}

	data := append(salt, []byte(password)...)
	for i := 0; i < hashIterations; i++ {
		h := sha256.Sum256(data)
		data = h[:]
	}

	return subtle.ConstantTimeCompare(data, expectedHash) == 1
}

// GenerateSessionID tạo chuỗi token ngẫu nhiên an toàn cho session
func GenerateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Authenticate kiểm tra thông tin đăng nhập và tạo session nếu hợp lệ
func (s *Service) Authenticate(username, password string) (*User, *Session, error) {
	u, err := s.repo.FindUserByUsername(username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	if !u.IsActive {
		return nil, nil, ErrAccountInactive
	}

	if !CheckPassword(password, u.PasswordHash) {
		return nil, nil, ErrInvalidCredentials
	}

	sessID, err := GenerateSessionID()
	if err != nil {
		return nil, nil, fmt.Errorf("không thể tạo session ID: %w", err)
	}

	session := &Session{
		ID:        sessID,
		UserID:    u.ID,
		ExpiresAt: time.Now().Add(SessionDuration),
	}

	if err := s.repo.CreateSession(session); err != nil {
		return nil, nil, err
	}

	return u, session, nil
}

// ValidateSession xác thực token session và trả về User tương ứng
func (s *Service) ValidateSession(sessionID string) (*User, error) {
	if sessionID == "" {
		return nil, ErrSessionNotFound
	}

	sess, err := s.repo.FindSession(sessionID)
	if err != nil {
		return nil, err
	}

	if time.Now().After(sess.ExpiresAt) {
		_ = s.repo.DeleteSession(sessionID)
		return nil, ErrSessionExpired
	}

	u, err := s.repo.FindUserByID(sess.UserID)
	if err != nil {
		return nil, err
	}

	if !u.IsActive {
		return nil, ErrAccountInactive
	}

	return u, nil
}

// Logout huỷ phiên đăng nhập
func (s *Service) Logout(sessionID string) error {
	if sessionID == "" {
		return nil
	}
	return s.repo.DeleteSession(sessionID)
}

// GetRepository trả về repo liên kết
func (s *Service) GetRepository() *Repository {
	return s.repo
}
