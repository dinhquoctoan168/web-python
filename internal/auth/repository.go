package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserNotFound    = errors.New("không tìm thấy người dùng")
	ErrSessionNotFound = errors.New("phiên đăng nhập không tồn tại")
)

// Repository chịu trách nhiệm truy vấn database cho auth
type Repository struct {
	db *sql.DB
}

// NewRepository khởi tạo Repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// FindUserByUsername tìm người dùng theo tên đăng nhập
func (r *Repository) FindUserByUsername(username string) (*User, error) {
	row := r.db.QueryRow(`
		SELECT id, username, password_hash, full_name, COALESCE(email, ''), role, is_active, created_at
		FROM users
		WHERE username = ?
	`, username)

	var u User
	var isActiveInt int
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.FullName, &u.Email, &u.Role, &isActiveInt, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("truy vấn user thất bại: %w", err)
	}

	u.IsActive = isActiveInt == 1
	return &u, nil
}

// FindUserByID tìm người dùng theo ID
func (r *Repository) FindUserByID(id int) (*User, error) {
	row := r.db.QueryRow(`
		SELECT id, username, password_hash, full_name, COALESCE(email, ''), role, is_active, created_at
		FROM users
		WHERE id = ?
	`, id)

	var u User
	var isActiveInt int
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.FullName, &u.Email, &u.Role, &isActiveInt, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("truy vấn user thất bại: %w", err)
	}

	u.IsActive = isActiveInt == 1
	return &u, nil
}

// CreateUser thêm một người dùng mới
func (r *Repository) CreateUser(u *User) error {
	isActiveInt := 0
	if u.IsActive {
		isActiveInt = 1
	}

	res, err := r.db.Exec(`
		INSERT INTO users (username, password_hash, full_name, email, role, is_active)
		VALUES (?, ?, ?, ?, ?, ?)
	`, u.Username, u.PasswordHash, u.FullName, u.Email, u.Role, isActiveInt)
	if err != nil {
		return fmt.Errorf("thêm user thất bại: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		u.ID = int(id)
	}
	return nil
}

// CreateSession tạo một session mới trong CSDL
func (r *Repository) CreateSession(s *Session) error {
	_, err := r.db.Exec(`
		INSERT INTO sessions (id, user_id, expires_at)
		VALUES (?, ?, ?)
	`, s.ID, s.UserID, s.ExpiresAt)
	if err != nil {
		return fmt.Errorf("tạo session thất bại: %w", err)
	}
	return nil
}

// FindSession tìm session theo token ID
func (r *Repository) FindSession(sessionID string) (*Session, error) {
	row := r.db.QueryRow(`
		SELECT id, user_id, expires_at, created_at
		FROM sessions
		WHERE id = ?
	`, sessionID)

	var s Session
	err := row.Scan(&s.ID, &s.UserID, &s.ExpiresAt, &s.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("truy vấn session thất bại: %w", err)
	}

	return &s, nil
}

// DeleteSession xoá session khi đăng xuất
func (r *Repository) DeleteSession(sessionID string) error {
	_, err := r.db.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	if err != nil {
		return fmt.Errorf("xoá session thất bại: %w", err)
	}
	return nil
}

// CleanExpiredSessions dọn dẹp các session đã hết hạn
func (r *Repository) CleanExpiredSessions() error {
	_, err := r.db.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now())
	return err
}

// CountUsers đếm số lượng user trong hệ thống
func (r *Repository) CountUsers() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}
