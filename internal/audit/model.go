package audit

import "time"

// AuditLog đại diện cho một bản ghi kiểm toán nghiệp vụ hệ thống
type AuditLog struct {
	ID         int       `json:"id"`
	UserID     *int      `json:"user_id,omitempty"`
	Action     string    `json:"action"`
	ObjectType string    `json:"object_type"`
	ObjectID   *int      `json:"object_id,omitempty"`
	Metadata   string    `json:"metadata"`
	CreatedAt  time.Time `json:"created_at"`
}
