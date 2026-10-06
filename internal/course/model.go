package course

import "time"

// Trạng thái môn học
const (
	StatusActive   = "active"
	StatusArchived = "archived"
)

// Course đại diện cho một môn học / học phần trong hệ thống
type Course struct {
	ID          int       `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedBy   int       `json:"created_by"`
	TeacherName string    `json:"teacher_name,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
