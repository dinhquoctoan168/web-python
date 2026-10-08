package frontend_test

import (
	"testing"

	"web_python/internal/auth"
	"web_python/internal/frontend"
)

func TestBuildNavigationDataStudent(t *testing.T) {
	student := &auth.User{
		ID:       10,
		Username: "student1",
		FullName: "Nguyễn Văn Sinh Viên",
		Role:     auth.RoleStudent,
	}

	nav := frontend.BuildNavigationData(student, "dashboard", nil, "", "", "csrf-123")

	if nav.HomeURL != "/dashboard" {
		t.Errorf("Expected HomeURL /dashboard, got %s", nav.HomeURL)
	}
	if nav.Role != auth.RoleStudent {
		t.Errorf("Expected Role student, got %s", nav.Role)
	}
	if nav.UserName != "Nguyễn Văn Sinh Viên" {
		t.Errorf("Expected UserName Nguyễn Văn Sinh Viên, got %s", nav.UserName)
	}
	if nav.CSRFToken != "csrf-123" {
		t.Errorf("Expected CSRFToken csrf-123, got %s", nav.CSRFToken)
	}

	expectedKeys := []string{"dashboard", "courses", "classes", "assignments", "exams", "ide"}
	if len(nav.Items) != len(expectedKeys) {
		t.Fatalf("Expected %d items, got %d", len(expectedKeys), len(nav.Items))
	}

	var foundActive bool
	for i, k := range expectedKeys {
		if nav.Items[i].Key != k {
			t.Errorf("Item %d key expected %s, got %s", i, k, nav.Items[i].Key)
		}
		if nav.Items[i].Key == "dashboard" && nav.Items[i].Active {
			foundActive = true
		}
	}
	if !foundActive {
		t.Errorf("Expected dashboard to be active")
	}
}

func TestBuildNavigationDataTeacher(t *testing.T) {
	teacher := &auth.User{
		ID:       2,
		Username: "teacher1",
		FullName: "Trần Thầy Giáo",
		Role:     auth.RoleTeacher,
	}

	nav := frontend.BuildNavigationData(teacher, "courses", nil, "", "", "csrf-456")

	if nav.HomeURL != "/teacher" {
		t.Errorf("Expected HomeURL /teacher, got %s", nav.HomeURL)
	}
	if nav.Role != auth.RoleTeacher {
		t.Errorf("Expected Role teacher, got %s", nav.Role)
	}

	expectedKeys := []string{"dashboard", "courses", "classes", "assignments", "exams", "submissions", "ide"}
	if len(nav.Items) != len(expectedKeys) {
		t.Fatalf("Expected %d items, got %d", len(expectedKeys), len(nav.Items))
	}

	var coursesActive bool
	for _, item := range nav.Items {
		if item.Key == "courses" && item.Active {
			coursesActive = true
		}
	}
	if !coursesActive {
		t.Errorf("Expected courses item to be active")
	}
}

func TestBuildNavigationDataAnonymous(t *testing.T) {
	nav := frontend.BuildNavigationData(nil, "", nil, "", "", "")
	if nav.HomeURL != "/login" {
		t.Errorf("Expected HomeURL /login, got %s", nav.HomeURL)
	}
	if len(nav.Items) != 1 || nav.Items[0].Key != "login" {
		t.Errorf("Expected single login item for anonymous, got %+v", nav.Items)
	}
}
