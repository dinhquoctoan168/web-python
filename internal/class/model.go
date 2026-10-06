package class

import "time"

// Class đại diện cho một lớp học của một môn học
type Class struct {
	ID           int       `json:"id"`
	CourseID     int       `json:"course_id"`
	CourseCode   string    `json:"course_code,omitempty"`
	CourseName   string    `json:"course_name,omitempty"`
	Name         string    `json:"name"`
	Semester     string    `json:"semester"`
	AcademicYear string    `json:"academic_year"`
	TeacherID    int       `json:"teacher_id"`
	TeacherName  string    `json:"teacher_name,omitempty"`
	Status       string    `json:"status"`
	StudentCount int       `json:"student_count,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// Enrollment đại diện cho một bản ghi ghi danh của sinh viên vào lớp
type Enrollment struct {
	ID         int       `json:"id"`
	ClassID    int       `json:"class_id"`
	StudentID  int       `json:"student_id"`
	EnrolledAt time.Time `json:"enrolled_at"`
}

// EnrolledStudent đại diện cho thông tin sinh viên trong lớp học
type EnrolledStudent struct {
	ID         int       `json:"id"`
	Username   string    `json:"username"`
	FullName   string    `json:"full_name"`
	Email      string    `json:"email"`
	EnrolledAt time.Time `json:"enrolled_at"`
}
