package frontend

import (
	"web_python/internal/auth"
)

// NavItem đại diện cho một mục điều hướng trên thanh menu
type NavItem struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	URL    string `json:"url"`
	Active bool   `json:"active"`
}

// Breadcrumb đại diện cho một nút trong chuỗi đường dẫn phân cấp
type Breadcrumb struct {
	Label string `json:"label"`
	URL   string `json:"url"` // Rỗng đối với trang hiện tại
}

// NavigationData chứa toàn bộ dữ liệu phục vụ render navigation thống nhất
type NavigationData struct {
	HomeURL     string       `json:"home_url"`
	HomeLabel   string       `json:"home_label"`
	UserName    string       `json:"user_name"`
	Role        string       `json:"role"`
	Items       []NavItem    `json:"items"`
	Breadcrumbs []Breadcrumb `json:"breadcrumbs"`
	BackURL     string       `json:"back_url"`
	BackLabel   string       `json:"back_label"`
	CSRFToken   string       `json:"csrf_token"`
}

// BuildNavigationData xây dựng NavigationData chuẩn hóa theo vai trò người dùng
func BuildNavigationData(user *auth.User, activeKey string, breadcrumbs []Breadcrumb, backURL, backLabel, csrfToken string) NavigationData {
	data := NavigationData{
		HomeURL:     "/",
		HomeLabel:   "Web Python Platform",
		Breadcrumbs: breadcrumbs,
		BackURL:     backURL,
		BackLabel:   backLabel,
		CSRFToken:   csrfToken,
	}

	if user == nil {
		data.HomeURL = "/login"
		data.Items = []NavItem{
			{Key: "login", Label: "Đăng nhập", URL: "/login", Active: activeKey == "login"},
		}
		return data
	}

	data.UserName = user.FullName
	data.Role = user.Role

	switch user.Role {
	case auth.RoleTeacher:
		data.HomeURL = "/teacher"
		data.HomeLabel = "Teacher Platform"
		data.Items = []NavItem{
			{Key: "dashboard", Label: "Tổng quan", URL: "/teacher", Active: activeKey == "dashboard"},
			{Key: "courses", Label: "Môn học", URL: "/teacher/courses", Active: activeKey == "courses"},
			{Key: "classes", Label: "Lớp học", URL: "/teacher/classes", Active: activeKey == "classes"},
			{Key: "assignments", Label: "Bài tập", URL: "/teacher/assignments", Active: activeKey == "assignments"},
			{Key: "exams", Label: "Kiểm tra", URL: "/teacher/exams", Active: activeKey == "exams"},
			{Key: "submissions", Label: "Bài nộp", URL: "/teacher/submissions", Active: activeKey == "submissions"},
			{Key: "ide", Label: "IDE", URL: "/ide", Active: activeKey == "ide"},
		}

	case auth.RoleAdmin:
		data.HomeURL = "/teacher"
		data.HomeLabel = "Admin Platform"
		data.Items = []NavItem{
			{Key: "dashboard", Label: "Tổng quan", URL: "/teacher", Active: activeKey == "dashboard"},
			{Key: "courses", Label: "Môn học", URL: "/teacher/courses", Active: activeKey == "courses"},
			{Key: "classes", Label: "Lớp học", URL: "/teacher/classes", Active: activeKey == "classes"},
			{Key: "assignments", Label: "Bài tập", URL: "/teacher/assignments", Active: activeKey == "assignments"},
			{Key: "exams", Label: "Kiểm tra", URL: "/teacher/exams", Active: activeKey == "exams"},
			{Key: "submissions", Label: "Bài nộp", URL: "/teacher/submissions", Active: activeKey == "submissions"},
			{Key: "ide", Label: "IDE", URL: "/ide", Active: activeKey == "ide"},
		}

	default: // Student
		data.HomeURL = "/dashboard"
		data.HomeLabel = "Web Python Platform"
		data.Items = []NavItem{
			{Key: "dashboard", Label: "Tổng quan", URL: "/dashboard", Active: activeKey == "dashboard"},
			{Key: "courses", Label: "Môn học", URL: "/courses", Active: activeKey == "courses"},
			{Key: "classes", Label: "Lớp học", URL: "/my-classes", Active: activeKey == "classes"},
			{Key: "assignments", Label: "Bài tập", URL: "/my-assignments", Active: activeKey == "assignments"},
			{Key: "exams", Label: "Kiểm tra", URL: "/my-exams", Active: activeKey == "exams"},
			{Key: "ide", Label: "IDE", URL: "/ide", Active: activeKey == "ide"},
		}
	}

	return data
}
