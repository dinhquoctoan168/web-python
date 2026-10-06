package audit

import (
	"encoding/json"
	"strings"
)

// Service cung cấp các tác vụ ghi và tra cứu nhật ký kiểm toán
type Service struct {
	repo *Repository
}

// NewService khởi tạo audit service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// SanitizeFields loại bỏ hoặc che các thông tin nhạy cảm (password, session token, cookie...)
func SanitizeFields(fields map[string]any) map[string]any {
	if fields == nil {
		return map[string]any{}
	}
	clean := make(map[string]any, len(fields))
	sensitiveKeys := []string{"password", "pass", "pwd", "token", "session", "csrf", "secret", "authorization", "cookie"}

	for k, v := range fields {
		kLower := strings.ToLower(k)
		isSensitive := false
		for _, s := range sensitiveKeys {
			if strings.Contains(kLower, s) {
				isSensitive = true
				break
			}
		}
		if isSensitive {
			clean[k] = "[REDACTED]"
		} else {
			clean[k] = v
		}
	}
	return clean
}

// LogAction ghi nhận một hành động nghiệp vụ với siêu dữ liệu đã được lọc an toàn
func (s *Service) LogAction(userID *int, action, objectType string, objectID *int, metadata map[string]any) error {
	sanitized := SanitizeFields(metadata)
	metaBytes, err := json.Marshal(sanitized)
	metaStr := "{}"
	if err == nil {
		metaStr = string(metaBytes)
	}

	entry := &AuditLog{
		UserID:     userID,
		Action:     action,
		ObjectType: objectType,
		ObjectID:   objectID,
		Metadata:   metaStr,
	}
	return s.repo.Record(entry)
}

// GetRecentLogs lấy danh sách audit logs mới nhất
func (s *Service) GetRecentLogs(limit int) ([]*AuditLog, error) {
	return s.repo.ListRecent(limit)
}

// GetLogsByUser lấy audit logs của một người dùng
func (s *Service) GetLogsByUser(userID, limit int) ([]*AuditLog, error) {
	return s.repo.ListByUser(userID, limit)
}

// GetLogsByObject lấy audit logs theo đối tượng thao tác
func (s *Service) GetLogsByObject(objectType string, objectID, limit int) ([]*AuditLog, error) {
	return s.repo.ListByObject(objectType, objectID, limit)
}
